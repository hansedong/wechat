//用户相关操作
package wechat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/panthesingh/goson"
)

const (
	API_USER_BASE_INFO  = "https://api.weixin.qq.com/cgi-bin/user/info?access_token=%s&openid=%s&lang=zh_CN"
	API_USER_LIST       = "https://api.weixin.qq.com/cgi-bin/user/get?access_token=%s&next_openid=%s"
	API_SET_USER_REMARK = "https://api.weixin.qq.com/cgi-bin/user/info/updateremark?access_token=%s"
)

type (
	UserBaseInfo struct {
		Subscribe     int    `json:"subscribe"`      //用户是否订阅该公众号标识，值为0时，代表此用户没有关注该公众号，拉取不到其余信息。
		OpenId        string `json:"openid"`         //用户的标识，对当前公众号唯一
		Nickname      string `json:"nickname"`       //用户的昵称
		Sex           int    `json:"sex"`            //用户的性别，值为1时是男性，值为2时是女性，值为0时是未知
		Language      string `json:"language"`       //用户所在城市
		City          string `json:"city"`           //用户所在国家
		Province      string `json:"province"`       //用户所在省份
		Country       string `json:"country"`        //用户的语言，简体中文为zh_CN
		Headimgurl    string `json:"headimgurl"`     //用户头像，最后一个数值代表正方形头像大小（有0、46、64、96、132数值可选，0代表640*640正方形头像），用户没有头像时该项为空。若用户更换头像，原有头像URL将失效。
		SubscribeTime int64  `json:"subscribe_time"` //用户关注时间，为时间戳。如果用户曾多次关注，则取最后关注时间
		Unionid       string `json:"unionid"`        //只有在用户将公众号绑定到微信开放平台帐号后，才会出现该字段
		Remark        string `json:"remark"`         //公众号运营者对粉丝的备注，公众号运营者可在微信公众平台用户管理界面对粉丝添加备注
		Groupid       int    `json:"groupid"`        //用户所在的分组ID
		TagList       []int  `json:"tagid_list"`     //用户被打上的标签ID列表
	}
	UserListDataItem struct {
		OpenId     []string `json:"openid"`      //列表数据，OPENID的列表
		NextOpenId string   `json:"next_openid"` //拉取列表的最后一个用户的OPENID
	}
	UserList struct {
		Total int              `json:"total"` //关注该公众账号的总用户数
		Count int              `json:"count"` //拉取的OPENID个数，最大值为10000
		Data  UserListDataItem `json:"data"`
	}
	UserRemark struct {
		OpenId string `json:"openid"` //用户openId
		Remark string `json:"remark"` //用户备注名（必须小于30个字符）
	}
	Ret struct {
		Errcode int    `json:"errcode"` //错误码
		Errmsg  string `json:"errmsg"`  //错误信息
	}
)

//获取用户的基本信息
//此方法，配合 app.go 中的 NewApp 结合使用
func (app *WechatApp) GetUserInfo(openId string) (userInfo UserBaseInfo, err error) {
	//检测openId
	if openId == "" {
		err = errors.New("无效的openId")
		return
	}
	//获取accessToken
	var accessTokenSt AccessToken
	accessTokenSt, err = app.GetAccessToken()
	if err != nil {
		return
	}
	apiUrl := fmt.Sprintf(API_USER_BASE_INFO, accessTokenSt.Token, openId)
	//创建request信息，并发送请求
	var req *http.Request
	var resp *http.Response
	client := &http.Client{}
	req, err = http.NewRequest("POST", apiUrl, nil)
	if err != nil {
		return
	}
	resp, err = client.Do(req)
	if err != nil {
		return
	}
	//defer
	defer resp.Body.Close()
	//初始化返回数据结构
	var data []byte
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var g *goson.Goson
	g, err = goson.Parse(data)
	if err != nil {
		return
	}
	_, ok := g.Get("errcode").Value().(string) //看是否有errorcode，如果有，说明接口调用出错
	if ok {
		err = errors.New("获取用户信息出错，错误信息：" + string(data))
		return
	}
	err = json.Unmarshal(data, &userInfo)
	return
}

//获取用户列表。如果当前页不足1000，则返回结果中的 NextOpenId 为空
//此方法，配合 app.go 中的 NewApp 结合使用
//当公众号关注者数量超过10000时，可通过填写next_openid的值，从而多次拉取列表的方式来满足需求。具体而言，
//就是在调用接口时，将上一次调用得到的返回中的next_openid值，作为下一次调用中的next_openid值。
func (app *WechatApp) GetUserList(nextOpenId string) (userList UserList, err error) {
	//获取accessToken
	var accessTokenSt AccessToken
	accessTokenSt, err = app.GetAccessToken()
	if err != nil {
		return
	}
	apiUrl := fmt.Sprintf(API_USER_LIST, accessTokenSt.Token, nextOpenId)
	//创建request信息，并发送请求
	var req *http.Request
	var resp *http.Response
	client := &http.Client{}
	req, err = http.NewRequest("POST", apiUrl, nil)
	if err != nil {
		return
	}
	resp, err = client.Do(req)
	if err != nil {
		return
	}
	//defer
	defer resp.Body.Close()
	//初始化返回数据结构
	var data []byte
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	var g *goson.Goson
	g, err = goson.Parse(data)
	if err != nil {
		return
	}
	_, ok := g.Get("errcode").Value().(string) //看是否有errorcode，如果有，说明接口调用出错
	if ok {
		err = errors.New("获取用户信息出错，错误信息：" + string(data))
		return
	}
	err = json.Unmarshal(data, &userList)
	return
}

//设置用户标签
//此方法，配合 app.go 中的 NewApp 结合使用
func (app *WechatApp) SetUserRemark(openId, remark string) (err error) {
	//获取accessToken
	var accessTokenSt AccessToken
	accessTokenSt, err = app.GetAccessToken()
	if err != nil {
		return
	}
	apiUrl := fmt.Sprintf(API_SET_USER_REMARK, accessTokenSt.Token)
	//创建request信息，并发送请求
	var req *http.Request
	var resp *http.Response
	client := &http.Client{}
	var data []byte
	data, err = json.Marshal(UserRemark{OpenId: openId, Remark: remark})
	if err != nil {
		return
	}
	dataReader := bytes.NewReader(data)
	req, err = http.NewRequest("POST", apiUrl, dataReader)
	if err != nil {
		return
	}
	resp, err = client.Do(req)
	if err != nil {
		return
	}
	//defer
	defer resp.Body.Close()
	//初始化返回数据结构
	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	ret := &Ret{}
	err = json.Unmarshal(data, &ret)
	if err != nil {
		err = errors.New("设置用户备注名出错，错误信息：" + string(data))
		return
	}
	return nil
}
