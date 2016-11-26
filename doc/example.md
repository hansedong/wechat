###包的简单实用示例
本例代码作用是，开启web服务，用户给公众号发送任意文字消息之后，公众号将公众号的access token返回给用户。
注意：此示例不安全，仅测试演示使用，access token实际上不可以暴漏给外界。

```
package main

import (
	"github.com/donghongshuai/gocircum/wechat"
)

func main() {
	FastHttpServe()
}

// fasthttp server
func FastHttpServe() {
	server := wechat.NewFastHttpServer()
	wt := wechat.NewWechat()
	wt.Configure.Token = "U9Js0jStxxxxxxxxfxijAbqsp199"
	wt.Configure.AppId = "wxeeewewewaaa1ce1"
	wt.Configure.AppSecret = "08766555b58bb27c34e5a8afcd65841b"
	wt.AddTextHandler(MyTextHandler)
	wt.AddImgHandler(MyImageHandler)
	wt.AddShortVideoHandler(MyShortVideoHandler)
	wt.AddVideoHandler(MyVideoHandler)
	server.HandleWechat(wt)
	server.Listen(":9898")
}

func MyTextHandler(ctx *wechat.WechatContext) interface{} {
	var res interface{}
	accessToken, _, _ := wechat.WechatApp.GetAccessToken()
	res := ctx.GetMsgTextResponse()
	res.Content = "FastHttp：这是文字！！ accessToken：" + accessToken
	return res
}
```

代码说明：

```
// fasthttp server
func FastHttpServe() {

	//实例化一个web服务
	server := wechat.NewFastHttpServer()
	
	//实例化微信处理对象，然后设置公众号相关信息，如：Token、AppId、AppSecret等信息
	wt := wechat.NewWechat()
	wt.Configure.Token = "U9Js0jStxxxxxxxxfxijAbqsp199"
	wt.Configure.AppId = "wxeeewewewaaa1ce1"
	wt.Configure.AppSecret = "08766555b58bb27c34e5a8afcd65841b"
	
	//设置文本消息处理器，用户发送的文本消息，将由开发者自己定义的函数来处理
	wt.AddTextHandler(MyTextHandler)
	
	//设置图片消息处理器，用户发送的图片消息，由开发者自己定义的函数来处理
	wt.AddImgHandler(MyImageHandler)
	
	//设置短视频消息处理器
	wt.AddShortVideoHandler(MyShortVideoHandler)

	//绑定微信处理对象和web server
	server.HandleWechat(wt)
	
	//开始监听HTTP端口，对外服务
	server.Listen(":9898")
}
```

```
//文本消息处理器
//用户自定义的文本消息处理器，必须实现接口 wechat.MsgTextHandler，同理 图片处理器、
//短视频处理器等都要实现预定义的接口，以文本消息处理器为例，函数参数必须是微信的上下文
//对象指针：*wechat.WechatContext。函数的返回值必须是 interface类型。
func MyTextHandler(ctx *wechat.WechatContext) interface{} {
	//设置返回变量
	var res interface{}
	//获取微信公众号的access token信息
	accessToken, _, _ := wechat.WechatApp.GetAccessToken()
	//我们可以返回用户一个文本消息（你也可以返回News信息或图片，这取决于你自己，
	//但本例中，是以返回文本内容为例的，如果你要返回一个图片信息给用户，那可以这样写：
	//res := ctx.GetMsgImgResponse()）
	res := ctx.GetMsgTextResponse()
	//设置返回给用户的具体文本信息
	res.Content = "FastHttp：这是文字！！ accessToken：" + accessToken
	//返回数据（必须）
	return res
}
```

如上，几行代码，即可开启一个web服务，并处理用户消息返回特定内容。开发者不需要关心消息的加密解密、消息有效性验证、access token如何获取等内容。