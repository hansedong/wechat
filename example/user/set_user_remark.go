//【根据openId，设置用户备注名】
//1、你需要在配置好公众号后台
//2、执行 go run set_user_remark.go

package main

import (
	"fmt"

	"github.com/donghongshuai/gocircum/wechat"
)

func main() {
	setUserRemark()
}

//设置用户备注名
func setUserRemark() {

	//实例化一个微信公众号应用
	app := wechat.NewApp()
	app.Configure.Token = "Your Token"         //token信息
	app.Configure.AppId = "Your AppId"         //应用id
	app.Configure.AppSecret = "Your AppSecret" //应用密钥
	app.Configure.AppNo = "应用编号"               //应用编号，每个应用必须不同（建议和AppId用同一个值）

	//根据openId，获取用户列表
	err := app.SetUserRemark("oIiWbv21bPbKJRowPeoJhAILevL0", "大么么")
	if err != nil {
		fmt.Println(err)
	}
}
