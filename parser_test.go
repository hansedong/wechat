package wechat

import (
	"testing"
)

func TestDecodeTextInput(t *testing.T) {
	ctx := &WechatContext{}
	ctx.RequestString = `"<xml>
<ToUserName><![CDATA[toUser]]></ToUserName>
<FromUserName><![CDATA[fromUser]]></FromUserName> 
<CreateTime>1348831860</CreateTime>
<MsgType><![CDATA[text]]></MsgType>
<Content><![CDATA[this is a test]]></Content>
<MsgId>1234567890123456</MsgId>
</xml>"`
	textInput, err := ctx.DecodeTextMsg()
	if err != nil {
		t.Fatalf("parse xml error:" + err.Error())
	}
	stData := MsgText{
		MsgCommon: MsgCommon{
			ToUserName:   "toUser",
			FromUserName: "fromUser",
			CreateTime:   int(1348831860),
			MsgType:      "text",
			MsgId:        int64(1234567890123456),
		},
		Content: "this is a test",
	}
	if stData != textInput {
		t.Fatalf("parse xml input to struct data  error")
	}
}

func TestDecodeImgInput(t *testing.T) {
	ctx := &WechatContext{}
	ctx.RequestString = `"<xml>
<ToUserName><![CDATA[toUser]]></ToUserName>
<FromUserName><![CDATA[fromUser]]></FromUserName>
<CreateTime>1348831860</CreateTime>
<MsgType><![CDATA[image]]></MsgType>
<PicUrl><![CDATA[this is a url]]></PicUrl>
<MediaId><![CDATA[media_id]]></MediaId>
<MsgId>1234567890123456</MsgId>
</xml>"`
	imgInput, err := ctx.DecodeImgMsg()
	if err != nil {
		t.Fatalf("parse xml error:" + err.Error())
	}
	stData := MsgImg{
		MsgCommon: MsgCommon{
			ToUserName:   "toUser",
			FromUserName: "fromUser",
			CreateTime:   int(1348831860),
			MsgType:      "image",
			MsgId:        int64(1234567890123456),
		},
		PicUrl:  "this is a url",
		MediaId: "media_id",
	}
	if stData != imgInput {
		t.Fatalf("parse xml input to struct data  error")
	}
}

func TestDecodeVoiceInput(t *testing.T) {
	ctx := &WechatContext{}
	ctx.RequestString = `"<xml>
<ToUserName><![CDATA[toUser]]></ToUserName>
<FromUserName><![CDATA[fromUser]]></FromUserName>
<CreateTime>1357290913</CreateTime>
<MsgType><![CDATA[voice]]></MsgType>
<MediaId><![CDATA[media_id]]></MediaId>
<Format><![CDATA[Format]]></Format>
<MsgId>1234567890123456</MsgId>
</xml>"`
	voiceInput, err := ctx.DecodeVoiceMsg()
	if err != nil {
		t.Fatalf("parse xml error:" + err.Error())
	}
	stData := MsgVoice{
		MsgCommon: MsgCommon{
			ToUserName:   "toUser",
			FromUserName: "fromUser",
			CreateTime:   int(1357290913),
			MsgType:      "voice",
			MsgId:        int64(1234567890123456),
		},
		Format: "Format",
	}
	if stData != voiceInput {
		t.Fatalf("parse xml input to struct data  error")
	}
}

func TestDecodeAccessToken(t *testing.T) {
	body := []byte(`{"access_token":"bJxAt8DOO3XJv4xc6gWs_HDDZmldEE5c3m_rejF3zc4nH7w2MyQtg3yAa1wdlsMbyKOb5-wvCQC-HgFjascTjMS8gFpim1mxsXTfNdrRg0COFBZnySwZOTeOinnjZ63gMECfAEACUS","expires_in":7200}`)
	token, expire, err := DecodeAccessToken(body)
	if err != nil {
		t.Fatalf("decode wechat access token response error:" + err.Error())
	}
	if token == "" || expire == 0 {
		t.Fatalf("decode wechat access token response error")
	}
}
