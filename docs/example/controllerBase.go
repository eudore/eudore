package main

import (
	"github.com/eudore/eudore"
)

type MyBaseController struct {
	eudore.ControllerBase
}

func (*MyBaseController) Any()             {}
func (*MyBaseController) Get()             {}
func (*MyBaseController) GetInfoByIdName() {}
func (*MyBaseController) GetIndex()        {}
func (*MyBaseController) GetContent()      {}

// ControllerRoute 方法实现ControllerRoute接口，允许修改注册的路由路径。
func (*MyBaseController) ControllerRoute() map[string]string {
	m := map[string]string{
		// 禁用该方法的路由注册
		"GetIndex": "",
	}
	return m
}

func main() {
	app := eudore.NewCore()
	app.AddController(new(MyBaseController))

	app.Listen(":8088")
	app.Run()
}
