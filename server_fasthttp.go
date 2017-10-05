package wechat

import (
	"errors"
	"fmt"

	"github.com/buaazp/fasthttprouter"
	"github.com/valyala/fasthttp"
)

type (
	//FastHttp Web服务器实体
	FastHttpServer struct {
		Configure    *FastHttpConfiguration
		Router       *fasthttprouter.Router
		AppManager   *WechatAppManager
		ErrorHandler ErrorHandler //你可以注册自己的错误处理器，来接管wechatgo的错误信息
	}
	//FastHttp配置
	FastHttpConfiguration struct {
		Gzip bool
	}
)

//实例化一个基于FastHttp的Web Server
//serverConf本质是一个切片，用户如果没有传入此参数，则系统默认实例化一个FastHttp的Web服务的配置，如果用户传入了配置实例，则用用户的配置实例
func NewFastHttpServer(serverConf ...FastHttpConfiguration) *FastHttpServer {
	if len(serverConf) == 0 {
		serverConf = append(serverConf, FastHttpConfiguration{})
	}
	conf := serverConf[0]
	server := &FastHttpServer{Configure: &conf}
	server.ErrorHandler = func(err ErrorObject) { fmt.Println(err.Error.Error()) }
	return server
}

//分析并设置HTTP请求的处理方式
func (httpServer *FastHttpServer) ServeHTTP() {
	m := httpServer.AppManager
	router := httpServer.Router
	callbackUri := m.CallBackUri
	router.GET(callbackUri+":appNo", httpServer.authFunc)
	router.POST(callbackUri+":appNo", httpServer.callbackFunc)
}

//获取微信公众号的认证echo信息
//用户在设置微信公众号服务器配置，并开启后，微信会发送一次认证请求，此函数即做此验证用
func (httpServer *FastHttpServer) getAuthEchostr(token string, ctx *fasthttp.RequestCtx) []byte {
	urlParams := ctx.QueryArgs()
	signature := string(urlParams.Peek("signature"))
	timestamp := string(urlParams.Peek("timestamp"))
	nonce := string(urlParams.Peek("nonce"))
	echostr := string(urlParams.Peek("echostr"))
	return AuthWechatServer(signature, echostr, token, timestamp, nonce)
}

//校验微信来源是否有效
func (httpServer *FastHttpServer) checkSourceSignature(app *WechatApp, ctx *fasthttp.RequestCtx) bool {
	token := app.Configure.Token
	urlParams := ctx.QueryArgs()
	signature := string(urlParams.Peek("signature"))
	timestamp := string(urlParams.Peek("timestamp"))
	nonce := string(urlParams.Peek("nonce"))
	return CheckWechatAuthSignature(signature, token, timestamp, nonce)
}

//校验微信消息授权签名是否有效（在开启了消息安全模式的时候调用）
func (httpServer *FastHttpServer) checkMsgSignature(app *WechatApp, ctx *fasthttp.RequestCtx, encryptMsgStr string) bool {
	token := app.Configure.Token
	urlParams := ctx.QueryArgs()
	msg_signature := string(urlParams.Peek("msg_signature"))
	timestamp := string(urlParams.Peek("timestamp"))
	nonce := string(urlParams.Peek("nonce"))
	return CheckWechatAuthSignature(msg_signature, token, timestamp, nonce, encryptMsgStr)
}

//校验微信公众号回调请求，成功返回：校验结果、消息类型、此次消息对应的微信公众号应用、xml解析后的上下文
//msgCheck：是否进行消息类型检测
func (httpServer *FastHttpServer) checkCallbackFuncParams(ctx *fasthttp.RequestCtx, msgCheck bool) (bool, string, *WechatApp, *WechatContext) {
	var app *WechatApp
	var xmlParser *WechatContext
	var msgType, xmlStr string
	var err error
	var valiad, appexists bool
	appNo := ctx.UserValue("appNo").(string)
	if appNo == "" {
		err = errors.New("未找到能处理此次请求的公众号应用，应用编号：" + appNo)
		goto CHECK_END
	}
	//获取应该相应此次请求的微信公众号应用（如果出错，调用用户注册的错误处理函数，然后直接返回return）
	app, appexists = httpServer.AppManager.GetApp(appNo)
	if !appexists {
		err = errors.New("未找到能处理此次请求的公众号应用，应用编号：" + appNo)
		goto CHECK_END
	}
	//校验消息来源的有效性（是否来自于微信）
	if !httpServer.checkSourceSignature(app, ctx) {
		err = errors.New("请求非法，签名验证未通过")
		goto CHECK_END
	}
	//获取消息类型
	if msgCheck {
		//获取xml解析实例
		xmlStr = httpServer.getHttpRequestXML(ctx)
		xmlParser = NewXmlParser(xmlStr)
		err = xmlParser.DecodeMsg(app)
		if err != nil {
			err = errors.New("微信消息解析失败：" + err.Error())
			goto CHECK_END
		}
		msgType = xmlParser.GetActuXMLMsgType()
		if msgType == "" {
			err = errors.New("未获取到微信消息类型，应用编号：" + appNo + "，XML信息：" + xmlStr)
			goto CHECK_END
		}
		if xmlParser.GetXmlRequestContent() == "" {
			err = errors.New("未获取到微信消息的XML数据")
			goto CHECK_END
		}
	}
	//当开启消息签名验证时，验证消息的有效性
	if msgCheck && app.Configure.EnableMsgCrypt {
		if app.Configure.EncodingAESKey == "" {
			err = errors.New("已开启消息安全模式，但应用的EncodingAESKey还未设置")
			goto CHECK_END
		}
		msgEncrypt := xmlParser.GetOriginXMLMsgType()
		if msgEncrypt != MsgTypeEncrypt {
			err = errors.New("此应用已设置消息安全模式，但未从微信POST的XML内容中，获取Encrypt信息，请检查微信公众号后台是否已确认开启")
			goto CHECK_END
		}
	}
	valiad = true

CHECK_END:
	if !valiad && err != nil && httpServer.ErrorHandler != nil {
		httpServer.ErrorHandler(ErrorObject{Error: err})
	}
	return valiad, msgType, app, xmlParser
}

//处理微信公众号的回调消息
func (httpServer *FastHttpServer) callbackFunc(ctx *fasthttp.RequestCtx) {
	//微信消息参数校验
	valiad, msgType, app, xmlParser := httpServer.checkCallbackFuncParams(ctx, true)
	if !valiad {
		return
	}
	var ret interface{}
	err := errors.New("此应用未定义消息对应的处理器，应用编号：" + app.Configure.AppNo + "，消息类型：" + msgType)
	//处理特定消息类型的请求
	switch msgType {
	case MsgTypeImg:
		if app.ImagHandler != nil {
			ret = app.ImagHandler(xmlParser, app)
		} else {
			httpServer.ErrorHandler(ErrorObject{Error: err})
		}
	case MsgTypeVoice:
		if app.VoiceHandler != nil {
			ret = app.VoiceHandler(xmlParser, app)
		} else {
			httpServer.ErrorHandler(ErrorObject{Error: err})
		}
	case MsgTypeVideo:
		if app.VideoHandler != nil {
			ret = app.VideoHandler(xmlParser, app)
		} else {
			httpServer.ErrorHandler(ErrorObject{Error: err})
		}
	case MsgTypeVideoShort:
		if app.ShortVideoHandler != nil {
			ret = app.ShortVideoHandler(xmlParser, app)
		} else {
			httpServer.ErrorHandler(ErrorObject{Error: err})
		}
	case MsgTypeLocation:
		if app.LocaltionHandler != nil {
			ret = app.LocaltionHandler(xmlParser, app)
		} else {
			httpServer.ErrorHandler(ErrorObject{Error: err})
		}
	case MsgTypeLink:
		if app.LinkHandler != nil {
			ret = app.LinkHandler(xmlParser, app)
		} else {
			httpServer.ErrorHandler(ErrorObject{Error: err})
		}
	case MsgTypeText:
		if app.TextHandler != nil {
			ret = app.TextHandler(xmlParser, app)
		} else {
			httpServer.ErrorHandler(ErrorObject{Error: err})
		}
	default:
		if app.TextHandler != nil {
			ret = app.TextHandler(xmlParser, app)
		} else {
			httpServer.ErrorHandler(ErrorObject{Error: err})
		}
	}
	//设置输出信息
	var outPut []byte
	ctx.SetContentType("text/xml")
	//判断当前应用是否已设置安全模式
	if app.Configure.EnableMsgCrypt {
		outPut, err = EncodeResponseEncrypt(ret, app)
	} else {
		outPut, err = EncodeResponse(ret)
	}
	//错误检查
	if err != nil && httpServer.ErrorHandler != nil {
		httpServer.ErrorHandler(ErrorObject{Error: errors.New("生成微信消息返回值出错：" + err.Error())})
	} else {
		ctx.Write(outPut)
	}
	return
}

//处理微信公众号的认证消息（此函数一般只调用一次，当用户设置微信公众号服务器配置，并启用的时候，会发生此次调用）
func (httpServer *FastHttpServer) authFunc(ctx *fasthttp.RequestCtx) {
	valiad, _, app, _ := httpServer.checkCallbackFuncParams(ctx, false)
	if valiad {
		ret := httpServer.getAuthEchostr(app.Configure.Token, ctx)
		//header
		ctx.SetContentType("text/plain")
		ctx.Write(ret)
	}
	return
}

//获取此次请求的XML字符串
func (httpServer *FastHttpServer) getHttpRequestXML(ctx *fasthttp.RequestCtx) string {
	data := ctx.PostBody()
	str := string(data)
	return str
}

//处理微信公众号服务
func (httpServer *FastHttpServer) HandleWechat(m *WechatAppManager) {
	router := fasthttprouter.New()
	httpServer.Router = router
	httpServer.AppManager = m
	httpServer.ServeHTTP()
}

//开始监听HTTP服务
func (httpServer *FastHttpServer) Listen(address string) {
	router := httpServer.Router
	fasthttp.ListenAndServe(address, router.Handler)
}

//开始监听HTTP服务
func (httpServer *FastHttpServer) ListenTLS(address, cer, key string) {
	router := httpServer.Router
	fasthttp.ListenAndServeTLS(address, cer, key, router.Handler)
}
