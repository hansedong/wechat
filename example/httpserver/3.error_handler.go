//例子：设置错误处理器，在控制台输出wechatgo抛出的错误信息
//当服务发生异常情况的时候，wechatgo允许你定义一个错误处理器，接受wechatgo推送的错误信息，你可以自己依据错误内容做后续处理
//为了方便看效果，假如你配置的回调地址为：http://xxx.com/callback/wxxxjjs329323，那么直接在浏览器访问就可以在控制台看到错误信息了（因为无参数，签名验证失败）
//1、你需要在配置好公众号后台
//2、执行 go run 3.error_handler.go 开启server
package main

import (
	"fmt"

	"github.com/hansedong/wechat"
)

func main() {
	FastHttpServe()
}

//以fasthttp的方式启动web server对外提供服务
func FastHttpServe() {

	//1.1、开启一个 fasthttp server
	server := wechat.NewFastHttpServer()

	//1.2、设置错误处理器，接收wechatgo的错误
	var errorHandler wechat.ErrorHandler
	errorHandler = func(err wechat.ErrorObject) {
		fmt.Println("这是我的自定义错误输出：：：：：" + err.Error.Error())
	}
	server.ErrorHandler = errorHandler

	//2.1、再实例化一个微信公众号应用
	app := wechat.NewApp()
	app.Configure.Token = "你的Token"
	app.Configure.AppId = "你的AppId"
	app.Configure.AppSecret = "你的密钥"
	app.Configure.AppNo = "可自由设置的应用编号" //应用编号，每个应用必须不同（建议和AppId用同一个值）
	//设置消息处理器
	app.AddTextHandler(MyTextHandler)

	//3、实例化一个公众号应用管理器
	appManager := wechat.NewWechatAppManager()
	appManager.AddApp(app)

	//4、绑定web服务器和应用管理器
	server.HandleWechat(appManager)
	//5、开启web服务并处理请求
	server.Listen(":9898")

	//6、公众号回调地址都需要设置为：http://xxx.com/callback/你的应用编号。
}

//文本消息处理器
func MyTextHandler(ctx *wechat.WechatContext, app *wechat.WechatApp) interface{} {
	res := ctx.GetMsgTextResponse()
	res.Content = "FastHttp：这是文字！！ 你的应用Id：" + app.Configure.AppId
	return res

}
