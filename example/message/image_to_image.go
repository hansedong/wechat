//素材处理——图片
//【处理图片消息的例子，本例子为，拿到用户发送的图片，服务端将用户发送的图片，保存到本地，并上传到七牛、腾讯云的对象存储】
//1、你需要在配置好公众号后台
//2、执行 go run image_to_image.go 开启server

package main

import (
	"github.com/hansedong/wechat"
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

	//2.2、设置媒体资源处理器
	mediaHandler = initMediaSaver(app)

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

func initMediaSaver(app *wechat.WechatApp) *wechat.Media {
	//2、实例化七牛对象存储处理器
	qiniuOcs := wechat.NewQiniuCos("你的AccessKey", "你的SecretKey", "你的bucket")
	//3、实例化腾讯云对象存储处理器
	qCos := wechat.NewQCos("你的AppId", "你的SecretId", "你的SecretKey", "你的bucket")
	//4、将七牛存储处理器和腾讯云存储处理器注册到媒体对象中，并保存素材资源到云上
	media, _ := wechat.NewMedia(app, "", 1, 256, 1, 256)
	media.AddMediaSaveHandler(qiniuOcs)
	media.AddMediaSaveHandler(qCos)
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
func validateOpenId(string) (bl bool) {
	return true
}
