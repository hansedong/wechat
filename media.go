package wechat

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

const (
	mediaApi = "https://api.weixin.qq.com/cgi-bin/media/get?access_token=%s&media_id=%s"
)

var mediaGlobal *Media

type (
	MediaFileInfo struct { //资源内容信息
	}
	UpChanSt struct {
		FromUser string
		FileSrc  string
	}
	Media struct {
		SaveHandlers []MediaSaver //素材/媒体资源存储处理器，可定义多个，比如同时存储在七牛和腾讯云
		App          *WechatApp
		downloadDir  string
		downChan     chan struct{} //此channel，主要用于阻塞文件下载，防止下载文件太多，造成网络阻塞，对应服务器的上行带宽
		upChan       chan UpChanSt //文件上传channel
		downloadRate int           //下载速度（单位为MB/s）
		noticeChan   chan Notice
	}
	//通知数据
	Notice struct {
		FromUser string //发送上传任务的用户
		UpRes    bool   //处理结果
	}
	MediaSaver interface {
		Save(string) bool //素材转存接口，必须实现了Save方法，该方法要求一个文件在磁盘的路径
	}
)

//实例化媒体资源处理
//app（微信对象）、downloadDir（素材下载目录）、downChanNum（下载队列大小）、downRate（下载速度，单位KB）、upChanNum（上传队列大小）、upRate（上传速度）
func NewMedia(app *WechatApp, downloadDir string, downChanNum int, downRate int, upChanNum int, upRate int) (*Media, error) {
	//如果对象已存在
	if mediaGlobal != nil {
		return mediaGlobal, nil
	}
	//下载路径的目录
	if downloadDir == "" {
		downloadDir = "./"
	}
	//上传和下载队列不能完全阻塞
	if upChanNum < 1 {
		upChanNum = 1
	}
	if downChanNum < 1 {
		downChanNum = 1
	}
	media := &Media{App: app, downloadDir: downloadDir, upChan: make(chan UpChanSt, upChanNum), downChan: make(chan struct{}, downChanNum), downloadRate: downRate}
	mediaGlobal = media
	media.initMediaConsumer() //初始化上传消费者
	return media, nil
}

func (media *Media) initMediaConsumer() {
	//文件转存
	go func(media *Media) {
		for upSt := range media.upChan {
			wg := &sync.WaitGroup{}
			wg.Add(len(media.SaveHandlers))
			res := true
			for _, saveHandler := range media.SaveHandlers {
				//go func(saveHandler MediaSaver) {	//并行上传
				func(saveHandler MediaSaver) { //非并行上传
					pres := saveHandler.Save(upSt.FileSrc)
					//有一个处理失败，就任务此次上传，是失败的
					if pres == false {
						res = pres
					}
					wg.Done()
				}(saveHandler)
			}
			wg.Wait()
			//每个对象存储处理器都处理完成之后，删除文件素材文件
			media.delFile(upSt.FileSrc)
			//发送通知
			media.noticeChan <- Notice{FromUser: upSt.FromUser, UpRes: res}
		}
	}(media)
}

//添加媒体资源存储处理器
func (media *Media) SetNoticeChan(ch chan Notice) bool {
	var ok bool
	if ch != nil {
		media.noticeChan = ch
		ok = true
	}
	return ok
}

//添加媒体资源存储处理器
func (media *Media) AddMediaSaveHandler(s MediaSaver) bool {
	var ok bool
	if s != nil {
		media.SaveHandlers = append(media.SaveHandlers, s)
		ok = true
	}
	return ok
}

//转存下载的素材文件。返回文件大小和错误信息
func (media *Media) Save(mediaId string, fromUser string) (bool, error) {
	var ok bool
	var err error
	//检测是否已设置下载文件的存储器
	if len(media.SaveHandlers) == 0 {
		err = errors.New("尚未定义文件存储处理器")
		return ok, err
	}
	var accessToken AccessToken
	accessToken, err = media.App.GetAccessToken()
	if err == nil {
		var resp *http.Response
		resp, err = media.getMediaFromWechatApi(mediaId, accessToken.Token)
		if err != nil {
			return ok, err
		}
		if resp == nil {
			return ok, errors.New("未从微信获取到有效的素材下载数据")
		}
		//异步处理文件下载和上传，防止微信消息阻塞超时
		//下载文件（由于服务器上行带宽和下行带宽不一致，所以文件的下载和上传是分开的协程）
		go func(resp *http.Response, media *Media) {
			//释放链接
			defer resp.Body.Close()
			fileSrc, err := media.download(resp)
			if err != nil {
				//下载重试
				var retryRes = "成功"
				var err1 error
				fileSrc, err1 = media.download(resp)
				if err1 != nil {
					retryRes = "失败"
					//下载失败，直接
				}
				fmt.Println("下载素材失败，失败原因：" + err.Error() + "，发起重试并下载：" + retryRes)
				if err1 != nil {
					return
				}
			}
			//放入channel，供消费者下载
			media.upChan <- UpChanSt{FileSrc: fileSrc, FromUser: fromUser}
		}(resp, media)
	}
	ok = true
	return ok, err
}

//下载文件，返回下载后的存储路径
func (media *Media) download(resp *http.Response) (downloadSrc string, err error) {
	media.downChan <- struct{}{}
	defer func() {
		<-media.downChan
	}()
	//获取文件名
	fileName := media.parseFileName(resp.Header)
	if fileName != "" {
		if media.downloadDir == "" {
			media.downloadDir = "./"
		}
		downloadSrc = media.downloadDir + fileName
		//判断文件是否存在，如果存在，不再下载，直接错误信息
		_, err = os.Stat(downloadSrc)
		if err == nil || os.IsExist(err) {
			return "", errors.New("要下载的文件已存在，可能是微信二次通知所致")
		}
		var f *os.File
		//创建文件
		f, err = os.OpenFile(downloadSrc, os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			return downloadSrc, err
		}
		_, err = media.CopyLimitRate(f, resp.Body)
	}
	return downloadSrc, nil
}

//限速下载的实现
func (media *Media) CopyLimitRate(dst io.Writer, src io.Reader) (written int64, err error) {
	buf := make([]byte, 32*1024)               //32k
	dTimes := float64(media.downloadRate / 32) //1S内需要的下载次数
	loop := time.Duration(math.Floor(1000 / dTimes))
	timeTicker := time.NewTicker(loop * time.Millisecond)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
		<-timeTicker.C
	}
	timeTicker.Stop()
	return written, err
}

//删除文件
func (media *Media) delFile(fileSrc string) error {
	err := os.Remove(fileSrc)
	return err
}

func (media *Media) parseFileName(headers http.Header) string {
	var fileName string
	if headers == nil {
		return fileName
	}
	contentDis, ok := headers["Content-Disposition"]
	if ok {
		fileName = strings.Replace(contentDis[0], "attachment; filename=\"", "", 1)
		fileName = strings.Replace(fileName, "\"", "", 1)
	}
	return fileName
}

//获取微信的素材文件的http response信息
func (media *Media) getMediaFromWechatApi(mediaId string, accessToken string) (*http.Response, error) {
	var resp *http.Response
	var err error
	//下载链接
	apiUrl := fmt.Sprintf(mediaApi, accessToken, mediaId)
	//发起请求
	var req http.Request
	req.Method = "GET"
	req.Close = true
	req.URL, _ = url.Parse(apiUrl)
	header := http.Header{}
	header.Set("Range", "bytes=0-")
	req.Header = header
	resp, err = http.DefaultClient.Do(&req)
	return resp, err
}

//临时方法，获取文件的hash值
func getUploadFileName(fileSrc string) string {
	file, err := os.Open(fileSrc)
	if err != nil {
		return ""
	}
	defer file.Close()
	md5h := md5.New()
	io.Copy(md5h, file)
	return fmt.Sprintf("%x", md5h.Sum([]byte(""))) + path.Ext(fileSrc)
}
