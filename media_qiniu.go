//七牛对象存储对象，你可以在此之上做一个Manager，管理多个bucket，以节省每次实例化造成的内存开销
package wechat

import (
	"log"
	"os"

	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	"golang.org/x/net/context"
)

//构造返回值字段
type PutRet struct {
	Hash string `json:"hash"`
	Key  string `json:"key"`
}

type (
	QiniuH struct {
		AccessKey    string
		SecretKey    string
		Bucket       string
		DownloadPath string
	}
)

var Logger *log.Logger

func init() {
	Logger = log.New(os.Stdout, "", log.LstdFlags)
}

//实例化一个七牛云存储对象
func NewQiniuCos(accessKey, secretKey, bucket string) *QiniuH {
	return &QiniuH{AccessKey: accessKey, SecretKey: secretKey, Bucket: bucket}
}

//将HTTP相应数据保存
func (qi *QiniuH) Save(fileSrc string) bool {
	if fileSrc == "" {
		return false
	}
	//异步处理文件下载和上传，防止微信消息阻塞超时
	err := qi.Upload(fileSrc)
	if err != nil {
		return false
	}
	return true
}

func (qi *QiniuH) Upload(fileSrc string) error {
	uploadFileName := getUploadFileName(fileSrc)

	localFile := fileSrc
	bucket := qi.Bucket
	key := uploadFileName
	putPolicy := storage.PutPolicy{
		Scope: bucket,
	}
	mac := qbox.NewMac(accessKey, secretKey)
	upToken := putPolicy.UploadToken(mac)
	cfg := storage.Config{}
	// 空间对应的机房
	cfg.Zone = &storage.ZoneHuadong
	// 是否使用https域名
	cfg.UseHTTPS = false
	// 上传是否使用CDN上传加速
	cfg.UseCdnDomains = false
	// 构建表单上传的对象
	formUploader := storage.NewFormUploader(&cfg)
	ret := storage.PutRet{}
	err := formUploader.PutFile(context.Background(), &ret, upToken, key, localFile, nil)
	//重新发起上传尝试
	if err != nil {
		Logger.Printf("七牛云存储：%v: %v", uploadFileName, "上传失败（发起重试），失败信息："+err.Error())
		err = uploader.PutFileWithoutKey(nil, &ret, token, fileSrc, nil)
		return err
	}
	if err == nil {
		Logger.Printf("七牛云存储：%v: %v", uploadFileName, "上传成功")
	} else {
		Logger.Printf("七牛云存储：%v: %v", uploadFileName, "上传失败，失败信息："+err.Error())
	}
	return nil
}
