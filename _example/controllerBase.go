package main

/*
eudore.Controller是一个接口，可自行实现，eudore.ControllerBase只是其中一种默认实现。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewApp()
	app.AddController(new(myBaseController))

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/index").Do()
	client.NewRequest("GET", "/info/22").Do()
	client.NewRequest("POST", "/").Do()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

type myBaseController struct {
	eudore.ControllerBase
}

// Any 方法注册 Any /*路径。
func (ctl *myBaseController) Any() {
	ctl.Info("myBaseController Any")
}

// Get 方法注册 Get /*路径。
func (ctl *myBaseController) Get() interface{} {
	ctl.Debug("ctl get")
	return "get myBaseController"
}

// GetInfoById 方法注册GET /info/:id 路由路径。
func (ctl *myBaseController) GetInfoById() interface{} {
	return ctl.GetParam("id")
}

// String 方法返回控制器名称，响应Router.AddController输出的名称。
func (ctl *myBaseController) String() string {
	return "hello.myBaseController"
}

// Help 方法定义一个控制器本身的方法。
func (ctl *myBaseController) Help(ctx eudore.Context) {}

// ControllerRoute 方法返回控制器路由推导修改信息。
func (*myBaseController) ControllerRoute() map[string]string {
	return map[string]string{
		// 修改Path方法的路由注册,路径为空就是忽略改方法
		"Help":   "",
		"String": "",
		// 给Any方法自动生成的路径添加参数iaany
		"Any": " isany=1",
	}
}
