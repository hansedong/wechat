//【被推送模板消息的例子，这个例子的主要作用是：用户可发送图片或视频给公众号，公众号会转存到腾讯云和七牛的对象存储服务商，
//服务端，会在用户不再做发送操作的1小时候，给用户推送模板消息，告知用户上传了多少文件，成功多少，失败多少等】
//1、你需要在配置好公众号后台
//2、执行 go run template2.go 开启server

package main

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/donghongshuai/gocircum/wechat"
)

var mediaHandler *wechat.Media

func main() {
	FastHttpServe()
}

//公众号应用的详细设置
func FastHttpServe() {

	//1.1、开启一个 fasthttp server
	server := wechat.NewFastHttpServer()

	//2.1、再实例化一个微信公众号应用
	app := wechat.NewApp()
	app.Configure.Token = "Your Token"                                           //token信息
	app.Configure.AppId = "Your AppId"                                           //应用id
	app.Configure.AppSecret = "Your AppSecret"                                   //应用密钥
	app.Configure.AppNo = "应用编号"                                                 //应用编号，每个应用必须不同（建议和AppId用同一个值）
	app.Configure.EnableMsgCrypt = false                                         //启用消息安全模式（前提是你的公众号消息加解密方式那里，选择的是安全模式），注意：此模式会有一定的影响性能
	app.Configure.EncodingAESKey = "RH4kqCbwHle8QbXwHkn4jJshWFtjQujko6HKPBjRSEF" //消息加解密需要的编码密钥（EnableMsgCrypt设置为true时有效，这个可以在公众号后台做设置）
	//设置文本消息处理器
	app.AddImgHandler(MyImageHandler)
	app.AddVideoHandler(MyVideoHandler)
	app.AddShortVideoHandler(MyShortVideoHandler)

	//2.3置模板消息处理器，并执行
	tplChan := make(chan wechat.Notice, 10000) //初始化一个队列，用于在模板消息处理器与媒体资源处理器之间通信
	app.AddTplHandler(TplMsgHandler)
	app.TplHandler(app, tplChan)

	//2.2、设置媒体资源处理器，并设置通道参数，2.2和2.3结合，上传成功后会给模板消息处理器通知的，用于后续做模板消息发送
	mediaHandler = initMediaSaver(app, tplChan)

	//3、实例化一个公众号应用管理器
	appManager := wechat.NewWechatAppManager()
	appManager.AddApp(app)

	//4、绑定web服务器和应用管理器
	server.HandleWechat(appManager)
	//5、开启web服务并处理请求
	server.ListenTLS(":443", "a.crt", "a.key")
}

//图片消息处理器
//本例为，处理用户消息，解析出用户发送的图片，然后保存到七牛云存储
func MyImageHandler(ctx *wechat.WechatContext, app *wechat.WechatApp) interface{} {
	handleMedia(ctx, app)
	res := ctx.GetMsgTextResponse()
	res.Content = "图片上传成功"
	return res
}

//视频消息处理器
//本例为，处理用户消息，解析出用户发送的视频，然后保存到七牛云存储和腾讯云对象存储
func MyVideoHandler(ctx *wechat.WechatContext, app *wechat.WechatApp) interface{} {
	handleMedia(ctx, app)
	res := ctx.GetMsgTextResponse()
	res.Content = "视频上传成功"
	return res
}

//短视频消息处理器
//本例为，处理用户消息，解析出用户发送的短视频，然后保存到七牛云存储和腾讯云对象存储
func MyShortVideoHandler(ctx *wechat.WechatContext, app *wechat.WechatApp) interface{} {
	handleMedia(ctx, app)
	res := ctx.GetMsgTextResponse()
	res.Content = "短视频上传成功"
	return res
}

func initMediaSaver(app *wechat.WechatApp, ch chan wechat.Notice) *wechat.Media {
	//2、实例化七牛对象存储处理器
	qiniuOcs := wechat.NewQiniuCos("你的AccessKey", "你的SecretKey", "你的bucket")
	//3、实例化腾讯云对象存储处理器
	qCos := wechat.NewQCos("你的AppId", "你的SecretId", "你的SecretKey", "你的bucket")
	//4、将七牛存储处理器和腾讯云存储处理器注册到媒体对象中，并保存素材资源到云上
	media, _ := wechat.NewMedia(app, "", 1, 256, 1, 256)
	media.AddMediaSaveHandler(qiniuOcs)
	media.AddMediaSaveHandler(qCos)
	media.SetNoticeChan(ch)
	return media
}

//媒体资源保存上传
func handleMedia(ctx *wechat.WechatContext, app *wechat.WechatApp) {
	//1、获取mediaId
	videoInfo, _ := ctx.DecodeVideoMsg()
	mediaId := videoInfo.MediaId
	mediaHandler.Save(mediaId, videoInfo.FromUserName)
}

//验证OpenId，只有指定的openId用户，发送的素材资源才会处理上传，不是每个关注公众号的人都会处理
func validateOpenId(string) bool {
	return true
}

//模板消息处理器
func TplMsgHandler(app *wechat.WechatApp, ch chan wechat.Notice) {
	if app == nil || ch == nil {
		panic("无效的应用和通信队列")
	}
	//初始化模板消息处理器
	tpl, err := wechat.NewTPL(app)
	if err != nil {
		return
	}
	type Item struct {
		SussessCount int
		FailureCount int
		Time         int64
	}
	userMap := make(map[string]Item)
	var lock sync.RWMutex
	//5s遍历一次
	ticker := time.NewTicker(10 * time.Second)
	//接受处理结果
	go func() {
		for notice := range ch {
			//如果用户存在，更新计数器和时间戳
			lock.Lock()
			v, ok := userMap[notice.FromUser]
			if ok {
				v.Time = time.Now().Unix()
			} else {
				v = Item{Time: time.Now().Unix()}
			}
			if notice.UpRes == true {
				v.SussessCount += 1
			} else {
				v.FailureCount += 1
			}
			userMap[notice.FromUser] = v
			lock.Unlock()
		}
	}()
	//判断并发送模板消息给用户（从用户主动发素材类消息给公众号开始，到1小时未给公众号发消息后，公众号会提送模板消息给用户）
	go func() {
		var timeRange int64 = 600 //多久之后开始发送模板消息通知
		for range ticker.C {
			for openId, v := range userMap {
				now := time.Now().Unix()
				if now-v.Time >= timeRange {
					//发送模板消息
					tplData := wechat.NewTplMsg()
					tplData.SetUserId(openId)
					tplData.SetTplId("4aGR_y_c9z7dK-BoSlgEFHOQJa91JoeuuXZtmjc8Qjo")
					userInfo, err := app.GetUserBaseInfo(openId)
					if err != nil {
						continue
					}
					tplData.AddDataMeta("first", "Hi，"+userInfo.Nickname+"：\r\n近10分钟前，服务器已处理并上传了你发来的上传任务", "")
					msg := "任务总数：" + strconv.Itoa(v.SussessCount+v.FailureCount) + "个，成功：" + strconv.Itoa(v.SussessCount) + "个，失败：" + strconv.Itoa(v.FailureCount) + "个"
					tplData.AddDataMeta("keynote1", msg, "")
					tplData.AddDataMeta("keynote2", time.Unix(now-timeRange, 0).Format("2006-01-02 15:04:05"), "")
					tplData.AddDataMeta("remark", "感谢使用", "")
					//发送模板消息
					err = tpl.SendTPLMsg(tplData)
					if err != nil {
						fmt.Println(err.Error())
					}
					lock.Lock()
					delete(userMap, openId)
					lock.Unlock()
				}
			}

		}

	}()
}
