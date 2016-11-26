//七牛对象存储对象，你可以在此之上做一个Manager，管理多个bucket，以节省每次实例化造成的内存开销
package wechat

import (
	"log"
	"os"

	"github.com/qiniu/api.v7/kodo"
	"github.com/qiniu/api.v7/kodocli"
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
	//初始化AK，SK
	conf := &kodo.Config{AccessKey: qi.AccessKey, SecretKey: qi.SecretKey}
	//创建一个Client
	c := kodo.New(0, conf)
	//设置上传的策略
	policy := &kodo.PutPolicy{
		Scope: qi.Bucket,
		//设置Token过期时间
		Expires: 3600,
		SaveKey: uploadFileName,
	}
	//生成一个上传token
	token := c.MakeUptoken(policy)
	//构建一个uploader
	zone := 0
	uploader := kodocli.NewUploader(zone, nil)
	var ret PutRet
	//调用PutFileWithoutKey方式上传，没有设置saveasKey以文件的hash命名
	err := uploader.PutFileWithoutKey(nil, &ret, token, fileSrc, nil)
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
