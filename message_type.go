//消息类型

package wechat

const (
	MsgTypeText       = "text"       //文本消息
	MsgTypeImg        = "image"      //图片消息
	MsgTypeVoice      = "voice"      //声音消息
	MsgTypeVideo      = "video"      //视频消息
	MsgTypeVideoShort = "shortvideo" //短视频消息
	MsgTypeLocation   = "location"   //地理位置消息
	MsgTypeLink       = "link"       //链接消息
	MsgTypeMusic      = "music"      //音乐消息
	MsgTypeNews       = "news"       //图文消息
	MsgTypeEncrypt    = "encrypt"    //加密消息
	//inner type
	MsgTypeAuth = "auth"
)

type (
	MsgEncrypt struct {
		Encrypt string `xml:"Encrypt"`
	}
	MsgType struct {
		MsgType string `xml:"MsgType"`
	}
	MsgCommon struct {
		ToUserName   string `xml:"ToUserName"`
		FromUserName string `xml:"FromUserName"`
		CreateTime   int    `xml:"CreateTime"`
		MsgType      string `xml:"MsgType"`
		MsgId        int64  `xml:"MsgId"`
		Encrypt      string `xml:"Encrypt"`
	}
	MsgText struct {
		MsgCommon
		Content string `xml:"Content"`
	}
	MsgImg struct {
		MsgCommon
		PicUrl  string `xml:"PicUrl"`
		MediaId string `xml:"MediaId"`
	}
	MsgVoice struct {
		MsgCommon
		Format string `xml:"Format"`
	}
	MsgVideo struct {
		MsgCommon
		MediaId      string `xml:"MediaId"`
		ThumbMediaId string `xml:"ThumbMediaId"`
		MsgId        string `xml:"MsgId"`
	}
	MsgShortVideo struct {
		MsgCommon
		MediaId      string `xml:"MediaId"`
		ThumbMediaId string `xml:"ThumbMediaId"`
		MsgId        string `xml:"MsgId"`
	}
	MsgLocation struct {
		MsgCommon
		LocationX float64 `xml:"LocationX"`
		LocationY float64 `xml:"LocationY"`
		Scale     int     `xml:"Scale"`
		Label     string  `xml:"Label"`
	}
	MsgLink struct {
		MsgCommon
		Title       string `xml:"Title"`
		Description string `xml:"Description"`
		Url         string `xml:"Url"`
	}
	MsgEventCommon struct {
		ToUserName   string `xml:"ToUserName"`
		FromUserName string `xml:"FromUserName"`
		CreateTime   int    `xml:"CreateTime"`
		MsgType      string `xml:"MsgType"`
		Event        string `xml:"Event"`
	}
	MsgEventSubscribe struct {
		MsgEventCommon
	}
	MsgEventUnSubscribe struct {
		MsgEventCommon
	}
	MsgEventQrcode struct {
		MsgEventCommon
		EventKey string `xml:"EventKey"`
		Ticket   string `xml:"Ticket"`
	}
	MsgEventScan struct {
		MsgEventCommon
		EventKey string `xml:"EventKey"`
		Ticket   string `xml:"Ticket"`
	}
	MsgEventClick struct {
		MsgEventCommon
		EventKey string `xml:"EventKey"`
	}
	MsgEventLocation struct {
		MsgEventCommon
		Latitude  string `xml:"Latitude"`
		Longitude string `xml:"Longitude"`
		Precision string `xml:"Precision"`
	}
	MsgEventView struct {
		MsgEventCommon
		EventKey string `xml:"EventKey"`
	}
	MsgTypeSt struct {
		MsgType string `xml:"MsgType"`
	}

	//Response struct
	//密文返回值结构
	MsgResponseEncrypt struct {
		XMLName      struct{} `xml:"xml"`
		Encrypt      string   `xml:"Encrypt"`
		MsgSignature string   `xml:"MsgSignature"`
		TimeStamp    string   `xml:"TimeStamp"`
		Nonce        string   `xml:"Nonce"`
	}
	MsgResponseCommon struct {
		XMLName      struct{} `xml:"xml"`
		ToUserName   string   `xml:"ToUserName"`
		FromUserName string   `xml:"FromUserName"`
		CreateTime   int64    `xml:"CreateTime"`
		MsgType      string   `xml:"MsgType"`
	}
	MsgResponseText struct {
		MsgResponseCommon
		Content string `xml:"Content"`
	}
	MsgResponseImg struct {
		MsgResponseCommon
		Image struct {
			MediaId string
		} `xml:"Image"`
	}
	MsgResponseVoice struct {
		MsgResponseCommon
		Voice struct {
			MediaId string
		} `xml:"Voice"`
	}
	MsgResponseVideo struct {
		MsgResponseCommon
		Video struct {
			MediaId     string
			Title       string
			Description string
		} `xml:"Video"`
	}
	MsgResponseMusic struct {
		MsgResponseCommon
		Music struct {
			Title        string
			Description  string
			MusicUrl     string
			HQMusicUrl   string
			ThumbMediaId string
		} `xml:"Music"`
	}
	MsgResponseMusicNoMedia struct {
		MsgResponseCommon
		Music struct {
			Title       string
			Description string
			MusicUrl    string
			HQMusicUrl  string
		} `xml:"Music"`
	}
	MsgResponseShortVideo struct {
		MsgResponseCommon
		Video struct {
			MediaId     string
			Title       string
			Description string
		} `xml:"Video"`
	}
	MsgResponseArticle struct {
		Title       string
		Description string
		PicUrl      string
		Url         string
	}
	MsgResponseNews struct {
		MsgResponseCommon
		ArticleCount int                  `xml:"ArticleCount"`
		Articles     []MsgResponseArticle `xml:"Articles>item"`
	}
	MsgResponseLocation struct {
		MsgResponseCommon
	}
	MsgResponseLink struct {
		MsgResponseCommon
	}
	MsgResponseEventSubscribe struct {
		MsgResponseCommon
	}
	MsgResponseEventUnSubscribe struct {
		MsgResponseCommon
	}
	MsgResponseEventQrcode struct {
		MsgResponseCommon
	}
	MsgResponseEventScan struct {
		MsgResponseCommon
	}
	MsgResponseEventClick struct {
		MsgResponseCommon
	}
	MsgResponseEventLocation struct {
		MsgResponseCommon
	}
	MsgResponseEventView struct {
		MsgResponseCommon
	}
)
