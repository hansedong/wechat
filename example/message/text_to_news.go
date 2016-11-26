//【处理文本消息的例子，本例子为，返回给用户一个图文消息】
//1、你需要在配置好公众号后台
//2、执行 go run text_to_news.go 开启server

package main

import (
	"github.com/donghongshuai/gocircum/wechat"
)

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
	app.Configure.EnableMsgCrypt = true                                          //启用消息安全模式（前提是你的公众号消息加解密方式那里，选择的是安全模式），注意：此模式会有一定的影响性能
	app.Configure.EncodingAESKey = "Zd8P9Ba51FhoWH8NXAJQV2Ghhdssa9zVitQdCqRf7H6" //消息加解密需要的编码密钥（EnableMsgCrypt设置为true时有效，这个可以在公众号后台做设置）
	//设置文本消息处理器
	app.AddTextHandler(MyTextHandler)

	//3、实例化一个公众号应用管理器
	appManager := wechat.NewWechatAppManager()
	appManager.AddApp(app)

	//4、绑定web服务器和应用管理器
	server.HandleWechat(appManager)
	//5、开启web服务并处理请求
	server.Listen(":9898")
}

//文本消息处理器
//本例为，处理用户消息，返回一个文章
func MyTextHandler(ctx *wechat.WechatContext, app *wechat.WechatApp) interface{} {
	res := ctx.GetMsgNewsResponse()
	article1 := wechat.MsgResponseArticle{Title: "测试文章", Description: "测试描述信息。。。。。", PicUrl: "http://mmbiz.qpic.cn/mmbiz_jpg/BviaV2pUMyNYdlDxIlDX1MicTBBRUTtE32k67u7MzACfdfrK03plOxOVujCaLYO6FkPAFJCjJhF2voBjPslLJ0KA/0", Url: "http://www.baidu.com"}
	res.AddArticle(article1)
	return res
}
