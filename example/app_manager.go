//应用管理器说明
//1、你需要在配置好公众号后台
//2、执行 go run app_manager.go 开启server
//3、设置你的公众号的接口URL为：域名/callback/你程序设置的AppNo

package main

import (
	"github.com/hansedong/wechat"
)

func main() {
	NetHttpServe()
}

//公众号应用的详细设置
func NetHttpServe() {

	//1.1、开启一个 net/http server
	server := wechat.NewNetHttpServer()

	//2.1、再实例化一个微信公众号应用
	app := wechat.NewApp()
	app.Configure.Token = "你的Token"                                              //token信息
	app.Configure.AppId = "你的AppId"                                              //应用id
	app.Configure.AppSecret = "你的密钥"                                             //应用密钥
	app.Configure.AppNo = "可自由设置的应用编号"                                           //应用编号，每个应用必须不同（建议和AppId用同一个值）
	app.Configure.EnableMsgCrypt = true                                          //启用消息安全模式（前提是你的公众号消息加解密方式那里，选择的是安全模式），注意：此模式会有一定的影响性能
	app.Configure.EncodingAESKey = "Zd8P9Ba51FhoWH8NXAJQV2Ghhdssa9zVitQdCqRf7H6" //消息加解密需要的编码密钥（EnableMsgCrypt设置为true时有效，这个可以在公众号后台做设置）
	//设置文本消息处理器
	app.AddTextHandler(MyTextHandler)

	//3.1、实例化一个公众号应用管理器，这种情况下，你的公众号回调地址，必须是：http://xxx.com/callback/你的应用编号
	appManager := wechat.NewWechatAppManager()

	//设置回调路径：假如你觉得回调地址为 http://xxx.com/callback/你的应用编号 不好，想把 "callback" 设置为 "try"这个字符，可以这样设置：
	//appManager.CallBackUri = "/try/"
	//或者一开始的时候，这样做实例化：appManager := wechat.NewWechatAppManager("/try/")

	//3.1、想公众号管理器中，添加一个公用号应用（你可以添加任意多个）
	appManager.AddApp(app)

	//4、绑定web服务器和应用管理器
	server.HandleWechat(appManager)
	//5、开启web服务并处理请求
	server.Listen(":9898")
}

//文本消息处理器
func MyTextHandler(ctx *wechat.WechatContext, app *wechat.WechatApp) interface{} {
	res := ctx.GetMsgTextResponse()
	res.Content = "Hello World"
	return res

}
