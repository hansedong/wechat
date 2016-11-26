//腾讯云对象存储对象，你可以在此之上做一个Manager，管理多个bucket，以节省每次实例化造成的内存开销
package wechat

import (
	Cos "github.com/laurence6/cos-go"
)

type (
	QCos struct {
		AppId     string
		SecretID  string
		SecretKey string
		Bucket    string
	}
)

//实例化一个七牛云存储对象
func NewQCos(appId, secretId, secretKey, bucket string) *QCos {
	return &QCos{AppId: appId, SecretID: secretId, SecretKey: secretKey, Bucket: bucket}
}

//将HTTP相应数据保存
func (q *QCos) Save(fileSrc string) bool {
	if fileSrc == "" {
		return false
	}
	//异步处理文件下载和上传，防止微信消息阻塞超时
	err := q.Upload(fileSrc)
	//上次失败，且不是因为上传了重复文件，则认为是真的上传失败，重新再传一次
	if (err != nil) && (err.Error() != "ErrSameFileUpload") {
		err = q.Upload(fileSrc)
		if err != nil {
			return false
		}
	}
	return true
}

func (q *QCos) Upload(fileSrc string) error {
	//初始化AK，SK
	cos := Cos.New(q.AppId, q.SecretID, q.SecretKey)
	//上传
	fileName := getUploadFileName(fileSrc)
	_, err := cos.UploadFile(fileSrc, q.Bucket, fileName)
	if err != nil {
		return err
	}
	return nil
}
