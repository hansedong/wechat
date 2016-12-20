package wechat

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type (
	//NetHttp Web服务器实体
	NetHttpServer struct {
		Configure    *HttpConfiguration
		Router       *httprouter.Router
		AppManager   *WechatAppManager
		ErrorHandler ErrorHandler //你可以注册自己的错误处理器，来接管wechatgo的错误信息
	}
	//NetHttp配置
	HttpConfiguration struct {
		Gzip bool
	}
)

//实例化一个基于Go语言内置net/http的Web Server
//serverConf本质是一个切片，用户如果没有传入此参数，则系统默认实例化一个FastHttp的Web服务的配置，如果用户传入了配置实例，则用用户的配置实例
func NewNetHttpServer(serverConf ...HttpConfiguration) *NetHttpServer {
	if len(serverConf) == 0 {
		serverConf = append(serverConf, HttpConfiguration{})
	}
	conf := serverConf[0]
	server := &NetHttpServer{Configure: &conf}
	server.ErrorHandler = func(err ErrorObject) { fmt.Println(err.Error.Error()) }
	return server
}

//分析并设置HTTP请求的处理方式
func (httpServer *NetHttpServer) ServeHTTP() {
	m := httpServer.AppManager
	router := httpServer.Router
	callbackUri := m.CallBackUri
	router.GET(callbackUri+":appNo", httpServer.authFunc)
	router.POST(callbackUri+":appNo", httpServer.callbackFunc)
}

//获取微信公众号的认证echo信息
//用户在设置微信公众号服务器配置，并开启后，微信会发送一次认证请求，此函数即做此验证用
func (httpServer *NetHttpServer) getAuthEchostr(token string, r *http.Request) []byte {
	signature := r.FormValue("signature")
	timestamp := r.FormValue("timestamp")
	nonce := r.FormValue("nonce")
	echostr := r.FormValue("echostr")
	return AuthWechatServer(signature, echostr, token, timestamp, nonce)
}

//校验微信来源是否有效
func (httpServer *NetHttpServer) checkSourceSignature(app *WechatApp, r *http.Request) bool {
	token := app.Configure.Token
	signature := r.FormValue("signature")
	timestamp := r.FormValue("timestamp")
	nonce := r.FormValue("nonce")
	return CheckWechatAuthSignature(signature, token, timestamp, nonce)
}

//校验微信消息授权签名是否有效（在开启了消息安全模式的时候调用）
func (httpServer *NetHttpServer) checkMsgSignature(app *WechatApp, r *http.Request, encryptMsgStr string) bool {
	token := app.Configure.Token
	msg_signature := r.FormValue("msg_signature")
	timestamp := r.FormValue("timestamp")
	nonce := r.FormValue("nonce")
	return CheckWechatAuthSignature(msg_signature, token, timestamp, nonce, encryptMsgStr)
}

//校验微信公众号回调请求，成功返回：校验结果、消息类型、此次消息对应的微信公众号应用、xml解析后的上下文
//msgCheck：是否进行消息类型检测
func (httpServer *NetHttpServer) checkCallbackFuncParams(w http.ResponseWriter, r *http.Request, params httprouter.Params, msgCheck bool) (bool, string, *WechatApp, *WechatContext) {
	var app *WechatApp
	var xmlParser *WechatContext
	var msgType string
	var err error
	var valiad, appexists bool
	appNo := params.ByName("appNo")
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
	if !httpServer.checkSourceSignature(app, r) {
		err = errors.New("请求非法，签名验证未通过")
		goto CHECK_END
	}
	//获取消息类型
	if msgCheck {
		//获取xml解析实例
		xmlParser = NewXmlParser(httpServer.getHttpRequestXML(r))
		err = xmlParser.DecodeMsg(app)
		if err != nil {
			err = errors.New("微信消息解析失败：" + err.Error())
			goto CHECK_END
		}
		msgType = xmlParser.GetActuXMLMsgType()
		if msgType == "" {
			err = errors.New("未获取到微信消息类型，应用编号：" + appNo + "，XML信息：" + xmlParser.GetXmlRequestContent())
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
		httpServer.ErrorHandler(ErrorObject{Error: err, AppId: app.Configure.AppId})
	}
	return valiad, msgType, app, xmlParser
}

func (httpServer *NetHttpServer) callbackFunc(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	//微信消息参数校验
	valiad, msgType, app, xmlParser := httpServer.checkCallbackFuncParams(w, r, params, true)
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
	w.Header().Add("Server", "wechat server")
	w.Header().Add("Content-Type", "text/xml")
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
		w.Write(outPut)
	}
	return
}

func (httpServer *NetHttpServer) authFunc(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	valiad, _, app, _ := httpServer.checkCallbackFuncParams(w, r, params, false)
	if valiad {
		ret := httpServer.getAuthEchostr(app.Configure.Token, r)
		//header
		w.Header().Add("Server", "wechat server")
		w.Header().Add("Content-Type", "text/plain")
		w.Write(ret)
	}
	return
}

//获取此次请求的XML字符串
func (httpServer *NetHttpServer) getHttpRequestXML(r *http.Request) string {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return ""
	}
	str := string(data)
	return str
}

//处理微信公众号服务
func (httpServer *NetHttpServer) HandleWechat(m *WechatAppManager) {
	router := httprouter.New()
	httpServer.Router = router
	httpServer.AppManager = m
	httpServer.ServeHTTP()
}

//开始监听HTTP服务
func (httpServer *NetHttpServer) Listen(address string) {
	router := httpServer.Router
	http.ListenAndServe(address, router)
}

//开始监听HTTPS服务
func (httpServer *NetHttpServer) ListenTLS(address, cerFile, keyFile string) {
	router := httpServer.Router
	httpServer.ListenAndServeTLS(address, cer, key, router)
}
