package wechat

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type (
	MsgTextHandler             func(*WechatContext, *WechatApp) interface{}
	MsgImgHandler              func(*WechatContext, *WechatApp) interface{}
	MsgVoiceHandler            func(*WechatContext, *WechatApp) interface{}
	MsgVideoHandler            func(*WechatContext, *WechatApp) interface{}
	MsgShortVideoHandler       func(*WechatContext, *WechatApp) interface{}
	MsgLocationHandler         func(*WechatContext, *WechatApp) interface{}
	MsgLinkHandler             func(*WechatContext, *WechatApp) interface{}
	MsgTplHandler              func(*WechatApp, chan Notice) //模板消息处理器
	MsgEventSubscribeHandler   func(*WechatContext, *WechatApp) interface{}
	MsgEventUnSubscribeHandler func(*WechatContext, *WechatApp) interface{}
	MsgEventQrcodeHandler      func(*WechatContext, *WechatApp) interface{}
	MsgEventScanHandler        func(*WechatContext, *WechatApp) interface{}
	MsgEventViewHandler        func(*WechatContext, *WechatApp) interface{}
	MsgEventLocationHandler    func(*WechatContext, *WechatApp) interface{}
	MsgEventClickHandler       func(*WechatContext, *WechatApp) interface{}
	AccessTokenOtherHandler    func(*WechatApp) (AccessToken, error) //access token处理器

	ErrorObject struct {
		AppId string
		XML   string
		Error error
	}
	ErrorHandler func(errobj ErrorObject)
)

type (
	WechatConfigure struct {
		AppNo          string //应用编号，每个应用必须不同，建议和AppId采用同一个值，http://www.xx.com/callback/"Your CallBackCode"
		AppId          string //公众号AppId
		AppSecret      string //公众号密钥
		Token          string //公众号Token
		EncodingAESKey string //公众号设置消息为安全模式时，此内容必填
		AccessTokenUrl string //获取AccessToken的URL（程序内使用）
		MediaUrl       string //媒体url
		AppEnable      bool   //公众号状态，启用还是关闭（暂时无效）
		EnableMsgCrypt bool   //true：消息密文模式（公众号要设置安全模式），false：消息明文模式（公众号默认设置）。开启后可以提高安全性，但会降低性能
	}

	WechatApp struct {
		//应用配置
		Configure WechatConfigure
		//消息处理器
		TextHandler       MsgTextHandler
		ImagHandler       MsgImgHandler
		VoiceHandler      MsgVoiceHandler
		VideoHandler      MsgVideoHandler
		ShortVideoHandler MsgShortVideoHandler
		LocaltionHandler  MsgLocationHandler
		LinkHandler       MsgLinkHandler
		//模板消息处理器
		TplHandler MsgTplHandler
		//AccessToken处理器
		AccessTokenOtherHandler AccessTokenOtherHandler
		//事件处理器
		EventSubscribeHandler   MsgEventSubscribeHandler
		EventUnSubscribeHandler MsgEventUnSubscribeHandler
		EventQrcodeHandler      MsgEventQrcodeHandler
		EventScanHandler        MsgEventScanHandler
		EventViewHandler        MsgEventViewHandler
		EventLocationHandler    MsgEventLocationHandler
		EventClickHandler       MsgEventClickHandler
	}
)

//实例化一个应用
func NewApp(configure ...WechatConfigure) *WechatApp {
	if len(configure) == 0 {
		configure = append(configure, WechatConfigure{
			AccessTokenUrl: "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
			MediaUrl:       "https://api.weixin.qq.com/cgi-bin/media/upload?access_token=%s&type=%s",
		})
	}
	conf := configure[0]
	wt := &WechatApp{Configure: conf}
	wt.init()

	return wt
}

//应用初始化
func (wt *WechatApp) init() {
	//初始化默认的第三方的AccessTokenHandler，此Handler将accessToken保存于进程内部，时效为2小时
	if wt.AccessTokenOtherHandler == nil {
		wt.AccessTokenOtherHandler = GetAccessTokenFromOther
	}
}

func (wt *WechatApp) AddTextHandler(handler MsgTextHandler) {
	wt.TextHandler = handler
}

func (wt *WechatApp) AddImgHandler(handler MsgImgHandler) {
	wt.ImagHandler = handler
}

func (wt *WechatApp) AddVoiceHandler(handler MsgVoiceHandler) {
	wt.VoiceHandler = handler
}

func (wt *WechatApp) AddVideoHandler(handler MsgVideoHandler) {
	wt.VideoHandler = handler
}

func (wt *WechatApp) AddShortVideoHandler(handler MsgShortVideoHandler) {
	wt.ShortVideoHandler = handler
}

func (wt *WechatApp) AddLocationHandler(handler MsgLocationHandler) {
	wt.LocaltionHandler = handler
}

func (wt *WechatApp) AddLinkHandler(handler MsgLinkHandler) {
	wt.LinkHandler = handler
}

//增加模板消息处理器
func (wt *WechatApp) AddTplHandler(handler MsgTplHandler) {
	wt.TplHandler = handler
}

//Event handler
func (wt *WechatApp) AddEventSubscribeHandler(handler MsgEventSubscribeHandler) {
	wt.EventSubscribeHandler = handler
}

func (wt *WechatApp) AddEventUnSubscribeHandler(handler MsgEventUnSubscribeHandler) {
	wt.EventUnSubscribeHandler = handler
}

func (wt *WechatApp) AddEventQrcodeHandler(handler MsgEventQrcodeHandler) {
	wt.EventQrcodeHandler = handler
}

func (wt *WechatApp) AddEventScanHandler(handler MsgEventScanHandler) {
	wt.EventScanHandler = handler
}

func (wt *WechatApp) AddEventViewHandler(handler MsgEventViewHandler) {
	wt.EventViewHandler = handler
}

func (wt *WechatApp) AddEventLocationHandler(handler MsgEventLocationHandler) {
	wt.EventLocationHandler = handler
}

func (wt *WechatApp) AddEventClickHandler(handler MsgEventClickHandler) {
	wt.EventClickHandler = handler
}

func (wt *WechatApp) AddAuthWechatServerHandler(handler MsgLinkHandler) {
	wt.LinkHandler = handler
}

func (wt *WechatApp) AddAccessTokenOtherHandler(handler AccessTokenOtherHandler) {
	wt.AccessTokenOtherHandler = handler
}

//check wether ther msg is from wechat web server
func AuthWechatServer(signature string, echostr string, sortParams ...string) []byte {
	var ret []byte
	if CheckWechatAuthSignature(signature, sortParams...) {
		ret = []byte(echostr)
	}
	return ret
}

//验证微信消息签名是否正确（支持验证微信请求中的signature和msg_signature）
func CheckWechatAuthSignature(signature string, sortParams ...string) bool {
	var ret bool
	mySignature := MakeMsgSignature(sortParams...)
	if mySignature == signature {
		ret = true
	}
	return ret
}

//获取AccesToken，函数返回 accesstoken、accesstoken过期时间，错误信息
func (wt *WechatApp) GetAccessToken() (AccessToken, error) {
	var accessToken AccessToken
	var err error
	//优先从用户定义的处理器中，获取AccessToken
	if wt.AccessTokenOtherHandler != nil {
		accessToken, err = wt.AccessTokenOtherHandler(wt)
	}
	if err != nil || accessToken.Token == "" {
		appId := wt.Configure.AppId
		appSecret := wt.Configure.AppSecret
		accessTokenUrl := wt.Configure.AccessTokenUrl
		var token string
		var expire int64
		token, expire, err = getAccessTokenByParams(appId, appSecret, accessTokenUrl)
		accessToken.Expire = expire
		accessToken.Token = token
	}
	return accessToken, err
}

//通过微信API，实时获取AccesToken，函数返回 accesstoken、accesstoken过期时间，错误信息
//注意，此方法，尽量不要从外部私自调用，因为调用后，以前的accesstoken将会时效，你需要知道自己要做什么以及如何做
func (wt *WechatApp) GetAccessTokenRealTime() (AccessToken, error) {
	appId := wt.Configure.AppId
	appSecret := wt.Configure.AppSecret
	accessTokenUrl := wt.Configure.AccessTokenUrl
	token, expire, err := getAccessTokenByParams(appId, appSecret, accessTokenUrl)
	return AccessToken{Token: token, Expire: expire}, err
}

func getAccessTokenByParams(appId string, appSecret string, accessTokenUrl string) (string, int64, error) {
	apiUrl := fmt.Sprintf(accessTokenUrl, appId, appSecret)
	resp, err := http.Get(apiUrl)
	if err != nil {
		fmt.Println(err.Error())
		return "", 0, err
	}
	defer resp.Body.Close()
	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}
	//decode json
	return DecodeAccessToken(body)
}
