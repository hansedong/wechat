//模板消息
package wechat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	API_INDUSTRY_SET = "https://api.weixin.qq.com/cgi-bin/template/api_set_industry?access_token=%s"
	API_SEND_TMP_MSG = "https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=%s"
)

type (
	TPL struct {
		app *WechatApp
	}
	IndustrySendData struct {
		IndustryId1 string `json:industry_id1`
		IndustryId2 string `json:industry_id2`
	}
	//模板消息结构体
	MsgData struct {
		ToUser   string                 `json:"touser"`
		TPLId    string                 `json:"template_id"`
		Url      string                 `json:"url"`
		TopColor string                 `json:"topcolor"`
		Data     map[string]MsgDataMeta `json:"data"`
	}
	//模板消息元素数据

	MsgDataMeta struct {
		Value string `json:"value"`
		Color string `json:"color"`
	}
)

//实例化一个模板消息内容，并初始化。具体的消息数据，可赋值
func NewTplMsg() *MsgData {
	msg := &MsgData{Data: make(map[string]MsgDataMeta)}
	msg.TopColor = "#FF0000"
	return msg
}

//设置模板消息的id
func (msg *MsgData) SetTplId(tplId string) *MsgData {
	if msg == nil || tplId == "" {
		return msg
	}
	msg.TPLId = tplId
	return msg
}

//设置模板消息接收用户
func (msg *MsgData) SetUserId(openId string) *MsgData {
	if msg == nil || openId == "" {
		return msg
	}
	msg.ToUser = openId
	return msg
}

//为模板消息添加具体的数据元素 metaKey为元数据的节点key，metaValue为数据，color为颜色，如果传空，会使用默认值"#173177"
func (msg *MsgData) AddDataMeta(metaKey string, metaValue string, color string) *MsgData {
	if msg == nil || metaKey == "" || metaValue == "" {
		return msg
	}
	if color == "" {
		color = "#173177"
	}
	msg.Data[metaKey] = MsgDataMeta{Color: color, Value: metaValue}
	return msg
}

//清空模板消息的数据元素
func (msg *MsgData) ClearDataMeta() *MsgData {
	if msg == nil {
		return msg
	}
	msg.Data = make(map[string]MsgDataMeta)
	return msg
}

//实例化一个模板消息处理对象
func NewTPL(app *WechatApp) (*TPL, error) {
	if app == nil {
		return nil, errors.New("无效的微信应用对象")
	}
	tpl := &TPL{app: app}
	return tpl, nil
}

//获取所属主营行业和副营行业，行业编码参考：http://mp.weixin.qq.com/wiki?t=resource/res_main&id=mp1433751277&token=&lang=zh_CN
func (tpl *TPL) SetIndustry(mainIndustry string, subIndustry string) error {
	accessTokenSt, err := tpl.app.GetAccessToken()
	if err != nil {
		return err
	}
	apiUrl := fmt.Sprintf(API_INDUSTRY_SET, accessTokenSt.Token)
	//设置请求数据
	data := IndustrySendData{IndustryId1: mainIndustry, IndustryId2: subIndustry}
	//发送请求
	client := &http.Client{}
	dataBytes, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", apiUrl, bytes.NewReader(dataBytes))
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	//defer
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	return err
}

//发送模板消息，tId为模板消息id，data为消息内容对象
func (tpl *TPL) SendTPLMsg(data *MsgData) error {
	accessTokenSt, err := tpl.app.GetAccessToken()
	if err != nil {
		return err
	}
	apiUrl := fmt.Sprintf(API_SEND_TMP_MSG, accessTokenSt.Token)
	//发送请求
	client := &http.Client{}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", apiUrl, bytes.NewReader(dataBytes))
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	//defer
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	return err
}
