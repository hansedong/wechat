//详细展开公众号的设置
//1、你需要在配置好公众号后台
//2、执行 go run app.go 开启server

package main

import (
	"errors"
	"time"

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
	app2 := wechat.NewApp()
	app2.Configure.Token = "你的Token"
	app2.Configure.AppId = "你的AppId"
	app2.Configure.AppSecret = "你的密钥"
	app2.Configure.AppNo = "可自由设置的应用编号"

	//启用消息安全模式（前提是你的公众号消息加解密方式那里，选择的是安全模式），注意：此模式会有一定的影响性能
	app.Configure.EnableMsgCrypt = true

	//消息加解密需要的编码密钥（EnableMsgCrypt设置为true时有效，这个可以在公众号后台做设置）
	app.Configure.EncodingAESKey = "消息加密解密秘钥"

	//设置处理器（具体例子，可参考example/message/*）
	app.AddTextHandler(MyTextHandler)       //设置文本消息处理器
	app.AddImgHandler(MyTextHandler)        //设置图片消息处理器
	app.AddVideoHandler(MyTextHandler)      //设置视频消息处理器（注意：如果你要想获取视频内容，会用到素材处理部分，需要你有素材接口的API权限，需要公众号认证）
	app.AddShortVideoHandler(MyTextHandler) //设置短视频消息处理器
	//....

	//3、实例化一个公众号应用管理器
	appManager := wechat.NewWechatAppManager()
	appManager.AddApp(app)

	//4、设置AccessToken处理器
	//①：由于【access token】的获取，微信是限制了每天的请求次数的，所以，如果是分布式机器，
	//或者是为了避免频繁请求微信API获取【access token】，你可能需要将其保存在缓存（如 Redis 中）作
	//为中间存储供各个业务集中使用。
	//②：默认情况下，我们提供了获取access token的接口（在消息处理器中，使用
	// accessToken, _ := app.GetAccessToken() 获取）,它会将accessToken缓存到进程中，2小时后再
	//次调用会此接口，会重新获取accessToken信息，有效期内则使用缓存。但做单机服务比较合适，
	//做分布式是显然不合适的，所以，如果要做分布式服务，你就需要自己实现实现一个 accessToken的Handler了。
	//③：下面是一个例子（本SDK使用的默认就是这个），你可以举一反三。
	var accessToken wechat.AccessToken
	tokenhandler := func(wt *wechat.WechatApp) (wechat.AccessToken, error) {
		var err error
		timeStamp := time.Now().Unix()
		//判断进程内缓存的accesstoken是否已过期
		if timeStamp > accessToken.Expire {
			accessToken, err = wt.GetAccessTokenRealTime()
			if err == nil {
				accessToken.Expire += timeStamp //时效2小时
			}
		}
		if err != nil || accessToken.Token == "" {
			return accessToken, errors.New("无效的AccessToken")
		}
		return accessToken, err
	}
	app.AccessTokenOtherHandler = tokenhandler

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
