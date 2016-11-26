###2、wechatgo设计思想

为了尽可能的让开发者方便使用、便于理解，wechatgo其实分为几个主要部分

* HTTP SERVER 部分
* 微信处理对象部分
* 素材管理部分

> HTTP SERVER 部分：相对简单，主要是描述如何提供Web供微信开放平台消息调用。
> 微信处理对象部分：内容较多，主要包含消息处理、事件处理、Token处理等等。
> 素材管理部分：主要为处理用户素材（图片、视频等等）。

基于这几部分，代码层面大致划分如下：


```
// web server 包，用Go语言内置包开启处理微信消息的web服务
server_http.go

// web server 包，用fasthttp包开启处理微信消息的web服务（fasthttp比go内置的net/http包性能要高）
//fasthttp：https://github.com/valyala/fasthttp
server_fasthttp.go

// 微信消息解析处理对象
parser.go

// 微信消息处理对象包
wechat.go
```

在之前的章节，我们已经用一个例子做演示，说明了如何开启一个web服务，以及利用这个web服务处理微信消息。继续以之前的例子来说明：

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


如果要做的是响应用户发给微信公众号消息的话，你仅仅需要做三个事情：

* 设置微信处理对象 wt
* 设置HTTP 服务对象 server ，并绑定微信处理对象
* 为微信处理对象，添加好你的消息处理器，比如上面例子中设置的文本消息处理器 MyTextHandler，专门用于处理用户发给微信的文本消息。你可以为一个微信处理对象，绑定多个不同的消息处理器。

对于只是想获取access token这种情况，更简单，实例化一个微信处理对象即可。后面我们会详细展示，具体说明。


