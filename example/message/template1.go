//【主动推送模板消息的例子，执行此go文件，可以直接向某用户发送模板】
//1、你需要在配置好公众号后台
//2、执行 go run template1.go

package main

import (
	"fmt"

	"github.com/donghongshuai/gocircum/wechat"
)

func main() {
	sendTplMsg()
}

//发送模板消息示例
func sendTplMsg() {

	//1、初始化应用
	app := wechat.NewApp()
	app.Configure.Token = "Your Token"         //token信息
	app.Configure.AppId = "Your AppId"         //应用id
	app.Configure.AppSecret = "Your AppSecret" //应用密钥
	app.Configure.AppNo = "应用编号"               //应用编号，每个应用必须不同（建议和AppId用同一个值）

	//2、初始化一个模板消息内容
	tplData := wechat.NewTplMsg()
	tplData.SetUserId("用户openId").SetTplId("模板消息Id")
	tplData.AddDataMeta("first", "尊敬的客户：\r\n近1小时前，服务器已处理并上传了你发来的上传任务", "")
	tplData.AddDataMeta("keynote1", "已处理任务数：5", "")
	tplData.AddDataMeta("keynote2", "2014年5月20日  12：24", "")
	tplData.AddDataMeta("remark", "感谢您选择我们的服务", "")

	//3、初始化模板消息处理器
	tpl, err := wechat.NewTPL(app)
	if err != nil {
		return
	}
	//4、发送模板消息
	err = tpl.SendTPLMsg(tplData)
	if err != nil {
		fmt.Println(err.Error())
	}
}
