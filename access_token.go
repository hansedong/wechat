//accessToken处理器
//1、对于分布式服务，或者请求量较大的服务来说。由于微信严格显示每个应用每天请求accesstoken接口的次数，所
//以，每次都实时（在消息处理器中，通过：accessToken, _, _ := app.GetAccessToken()）获取
//accessToken并不是一个好的处理方式。
//2、此SDK此考虑到此种方式，并默认提供了一种在进程内缓存accessToken的方式。但此种方式只适合单机服务，
//开发者可以自行根据自己需要拓展自己的方式，比如将accessToken缓存在Redis中。
//3、Tips：我们在NewApp的时候，其实内部会有一次初始化，该初始化会设置一个 AccessTokenHandler 。我们在消息
//处理器中，使用 app.GetAccessToken() 的时候，其实默认会先从AccessTokenHandler获取，获取不到才会
//走实时获取的逻辑。

package wechat

import (
	"errors"
	"time"
)

type (
	AccessToken struct {
		Token  string
		Expire int64
	}
)

var accessToken AccessToken = AccessToken{}

//自定义的第三方获取AccessToken服务，实现了 wechat.AccessTokenOtherHandler 接口
//你可以在实例化WechatApp之后，使用自己实现的Handler替换more的WechatApp.AccessTokenOtherHandler
func GetAccessTokenFromOther(wt *WechatApp) (AccessToken, error) {
	var err error
	timeStamp := time.Now().Unix()
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
