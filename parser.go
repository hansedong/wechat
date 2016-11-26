package wechat

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"strconv"
	"strings"
	"time"
)

type (
	//微信消息
	WechatContext struct {
		requestString     string //微信发来的有效xml内容（如果是加密的，则解密）
		msgTypeActu       string //微信消息实际的类型（如果是加密的，则解密拿到实际类型）
		msgTypeFromWechat string //微信原始消息类型（如果消息不是加密的，则和MsgTypeActu一样，否则就是MsgTypeEncrypt）
	}
)

//实例化微信消息解析器
func NewXmlParser(requestString string) *WechatContext {
	return &WechatContext{requestString: requestString}
}

//获取微信的原始消息类型，其值，是所有 MsgType* 的可能，定义在 message.go 中
func (wc *WechatContext) GetOriginXMLMsgType() string {
	return wc.msgTypeFromWechat
}

//获取解析微信消息XML后的实际消息类型，其值，是所有 MsgType* 的可能，定义在 message.go 中
func (wc *WechatContext) GetActuXMLMsgType() string {
	return wc.msgTypeActu
}

//获取有效的（如果是加密的，则使用解密后的XML）的请求XML数据
func (wc *WechatContext) GetMsgXMLContent() string {
	return wc.requestString
}

//检查微信发来的是否是密文消息
func (wc *WechatContext) CheckIfEncryptMsgType() bool {
	return wc.msgTypeFromWechat == MsgTypeEncrypt
}

//将微信消息解析为文本结构类型
func (wc *WechatContext) DecodeTextMsg() (MsgText, error) {
	msg := wc.requestString
	v := MsgText{}
	err := xml.Unmarshal([]byte(msg), &v)
	return v, err
}

//将微信消息解析为图片结构类型
func (wc *WechatContext) DecodeImgMsg() (MsgImg, error) {
	msg := wc.requestString
	v := MsgImg{}
	err := xml.Unmarshal([]byte(msg), &v)
	return v, err
}

//将微信消息解析为声音结构类型
func (wc *WechatContext) DecodeVoiceMsg() (MsgVoice, error) {
	msg := wc.requestString
	v := MsgVoice{}
	err := xml.Unmarshal([]byte(msg), &v)
	return v, err
}

//将微信消息解析为视频结构类型
func (wc *WechatContext) DecodeVideoMsg() (MsgVideo, error) {
	msg := wc.requestString
	v := MsgVideo{}
	err := xml.Unmarshal([]byte(msg), &v)
	return v, err
}

//将微信消息解析短视频结构类型
func (wc *WechatContext) DecodeShortVideoMsg() (MsgShortVideo, error) {
	msg := wc.requestString
	v := MsgShortVideo{}
	err := xml.Unmarshal([]byte(msg), &v)
	return v, err
}

//将微信消息解析为地理位置结构类型
func (wc *WechatContext) DecodeLocationMsg() (MsgLocation, error) {
	msg := wc.requestString
	v := MsgLocation{}
	err := xml.Unmarshal([]byte(msg), &v)
	return v, err
}

//将微信消息解析为链接结构类型
func (wc *WechatContext) DecodeLinkMsg() (MsgLink, error) {
	msg := wc.requestString
	v := MsgLink{}
	err := xml.Unmarshal([]byte(msg), &v)
	return v, err
}

//将微信消息解析为图文结构类型
func (wc *WechatContext) DecodeNewsMsg() (MsgLink, error) {
	msg := wc.requestString
	v := MsgLink{}
	err := xml.Unmarshal([]byte(msg), &v)
	return v, err
}

//获取微信发来的XML内容
func (wc *WechatContext) GetXmlRequestContent() string {
	return wc.requestString
}

//获取消息类型，如果是加密的消息，则解出实际类型
func (wc *WechatContext) DecodeMsg(app *WechatApp) error {
	var tObj MsgCommon = MsgCommon{}
	var err error = nil
	//获取应用参数
	appId := app.Configure.AppId
	encodingKey := app.Configure.EncodingAESKey
	//判断是否为加密类型，如果是，做解密操作，得到实际类型
	if strings.Contains(wc.requestString, "<Encrypt>") {
		wc.msgTypeFromWechat = MsgTypeEncrypt
		msgEncrypt := MsgEncrypt{}
		err = xml.Unmarshal([]byte(wc.requestString), &msgEncrypt)
		if err == nil {
			wc.requestString, err = DecryptMsg(msgEncrypt.Encrypt, encodingKey, appId)
			err = xml.Unmarshal([]byte(wc.requestString), &tObj)
		}
	} else {
		err = xml.Unmarshal([]byte(wc.requestString), &tObj)
		if err == nil {
			wc.msgTypeFromWechat = tObj.MsgType
		}
	}
	if err != nil {
		return err
	}
	wc.msgTypeActu = tObj.MsgType
	return err
}

func (wc *WechatContext) GetDefaultTextMsg() (MsgText, error) {
	v := MsgText{
		MsgCommon: MsgCommon{
			FromUserName: "from user",
			ToUserName:   "to user name",
			MsgType:      "text",
			CreateTime:   123456789,
		},
		Content: "this is a example show",
	}
	return v, nil
}

//获取一个文本类型的返回数据
func (wc *WechatContext) GetMsgTextResponse() *MsgResponseText {
	msg, err := wc.DecodeTextMsg()
	if err != nil {
		return nil
	}
	res := &MsgResponseText{
		MsgResponseCommon: MsgResponseCommon{
			FromUserName: msg.ToUserName,
			ToUserName:   msg.FromUserName,
			MsgType:      MsgTypeText,
			CreateTime:   getMsgResponseTimestamp(),
		},
		Content: "SUCCESS",
	}
	return res
}

//获取一个图片素材类型的返回数据
func (wc *WechatContext) GetMsgImgResponse() *MsgResponseImg {
	msg, err := wc.DecodeImgMsg()
	if err != nil {
		return nil
	}
	res := &MsgResponseImg{
		MsgResponseCommon: MsgResponseCommon{
			FromUserName: msg.ToUserName,
			ToUserName:   msg.FromUserName,
			MsgType:      MsgTypeImg,
			CreateTime:   getMsgResponseTimestamp(),
		},
	}
	return res
}

//获取一个语音型的返回数据
func (wc *WechatContext) GetMsgVoiceResponse() *MsgResponseVoice {
	msg, err := wc.DecodeVoiceMsg()
	if err != nil {
		return nil
	}
	res := &MsgResponseVoice{
		MsgResponseCommon: MsgResponseCommon{
			FromUserName: msg.ToUserName,
			ToUserName:   msg.FromUserName,
			MsgType:      MsgTypeImg,
			CreateTime:   getMsgResponseTimestamp(),
		},
	}
	return res
}

//获取一个视频类型的返回数据
func (wc *WechatContext) GetMsgVideoResponse() *MsgResponseVideo {
	msg, err := wc.DecodeVideoMsg()
	if err != nil {
		return nil
	}
	res := &MsgResponseVideo{
		MsgResponseCommon: MsgResponseCommon{
			FromUserName: msg.ToUserName,
			ToUserName:   msg.FromUserName,
			MsgType:      MsgTypeVideo,
			CreateTime:   getMsgResponseTimestamp(),
		},
	}
	return res
}

//获取一个短视频类型的返回数据
func (wc *WechatContext) GetMsgShortVideoResponse() *MsgResponseShortVideo {
	msg, err := wc.DecodeShortVideoMsg()
	if err != nil {
		return nil
	}
	res := &MsgResponseShortVideo{
		MsgResponseCommon: MsgResponseCommon{
			FromUserName: msg.ToUserName,
			ToUserName:   msg.FromUserName,
			MsgType:      MsgTypeVideo,
			CreateTime:   getMsgResponseTimestamp(),
		},
	}
	return res
}

//获取一个地理位置类型的返回数据
func (wc *WechatContext) GetMsgLocationResponse() *MsgResponseLocation {
	msg, err := wc.DecodeLocationMsg()
	if err != nil {
		return nil
	}
	res := &MsgResponseLocation{
		MsgResponseCommon: MsgResponseCommon{
			FromUserName: msg.ToUserName,
			ToUserName:   msg.FromUserName,
			MsgType:      MsgTypeLocation,
			CreateTime:   getMsgResponseTimestamp(),
		},
	}
	return res
}

//获取一个超链接类型的返回数据
func (wc *WechatContext) GetMsgLinkResponse() *MsgResponseLink {
	msg, err := wc.DecodeLinkMsg()
	if err != nil {
		return nil
	}
	res := &MsgResponseLink{
		MsgResponseCommon: MsgResponseCommon{
			FromUserName: msg.ToUserName,
			ToUserName:   msg.FromUserName,
			MsgType:      MsgTypeLink,
			CreateTime:   getMsgResponseTimestamp(),
		},
	}
	return res
}

//获取一个图文类型的返回数据
func (wc *WechatContext) GetMsgNewsResponse() *MsgResponseNews {
	msg, err := wc.DecodeLinkMsg()
	if err != nil {
		return nil
	}
	res := &MsgResponseNews{
		MsgResponseCommon: MsgResponseCommon{
			FromUserName: msg.ToUserName,
			ToUserName:   msg.FromUserName,
			MsgType:      MsgTypeNews,
			CreateTime:   getMsgResponseTimestamp(),
		},
	}
	return res
}

//结合 GetMsgNewsResponse 一起使用，添加一个图文文章（可添加多个）
func (res *MsgResponseNews) AddArticle(item MsgResponseArticle) *MsgResponseNews {
	res.ArticleCount += 1
	res.Articles = append(res.Articles, item)
	return res
}

//获取一个音乐类型的返回数据
func (wc *WechatContext) GetMsgMusicResponse() *MsgResponseMusic {
	msg, err := wc.DecodeLinkMsg()
	if err != nil {
		return nil
	}
	res := &MsgResponseMusic{
		MsgResponseCommon: MsgResponseCommon{
			FromUserName: msg.ToUserName,
			ToUserName:   msg.FromUserName,
			MsgType:      MsgTypeMusic,
			CreateTime:   getMsgResponseTimestamp(),
		},
	}
	return res
}

//获取一个图文类型的返回数据（和 GetMsgMusicResponse 相比，无需素材id，如果你的公众号没有认证，可以用这个）
func (wc *WechatContext) GetMsgMusicNoMediaResponse() *MsgResponseMusicNoMedia {
	msg, err := wc.DecodeLinkMsg()
	if err != nil {
		return nil
	}
	res := &MsgResponseMusicNoMedia{
		MsgResponseCommon: MsgResponseCommon{
			FromUserName: msg.ToUserName,
			ToUserName:   msg.FromUserName,
			MsgType:      MsgTypeMusic,
			CreateTime:   getMsgResponseTimestamp(),
		},
	}
	return res
}

//获取时间戳
func getMsgResponseTimestamp() int64 {
	return time.Now().Unix()
}

func EncodeResponse(msg interface{}) ([]byte, error) {
	//	var data []byte
	xmlData, err := xml.MarshalIndent(msg, "", "\r\n")
	return xmlData, err
	//	xmlDataHeader := []byte(xml.Header)
	//	data = append(xmlDataHeader, xmlData...)
}

//将返回给用户的数据，做加密处理
func EncodeResponseEncrypt(msg interface{}, app *WechatApp) ([]byte, error) {
	var xmlData []byte
	var err error
	//先把消息转为XML
	xmlData, err = xml.MarshalIndent(msg, "", "\r\n")
	if err != nil {
		return xmlData, err
	}
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	//获取加密串
	var encryptStr string
	encryptStr, err = MakeEncryptXmlData(xmlData, []byte(app.Configure.AppId), encodingAESKey2AESKey(app.Configure.EncodingAESKey))
	if err != nil {
		return xmlData, err
	}
	//实例化结构体
	responseEncrypt := MsgResponseEncrypt{}
	responseEncrypt.Encrypt = encryptStr
	responseEncrypt.TimeStamp = timestamp
	responseEncrypt.Nonce = "sdsjjj"
	responseEncrypt.MsgSignature = MakeMsgSignature(app.Configure.Token, timestamp, encryptStr, responseEncrypt.Nonce)
	//最终结构体加密
	xmlData, err = xml.MarshalIndent(responseEncrypt, "", "\r\n")
	return xmlData, err
}

//解码accesstoken信息，本质上，就是获取字符串类型的 accesstoken、过期时间
func DecodeAccessToken(body []byte) (string, int64, error) {
	var data struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	err := json.Unmarshal(body, &data)
	if err != nil {
		return "", 0, err
	}
	if data.ErrCode != 0 {
		return "", 0, errors.New("get access token error with wexin api error message : " + string(body))
	}
	return data.AccessToken, data.ExpiresIn, nil
}
