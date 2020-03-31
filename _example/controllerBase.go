package main

/*
eudore.Controller是一个接口，可自行实现，eudore.ControllerBase只是其中一种默认实现。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewCore()
	app.AddController(new(myBaseController))

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/index").Do()
	client.NewRequest("GET", "/info/22").Do()
	client.NewRequest("POST", "/").Do()
	for client.Next() {
		app.Error(client.Error())
	}

	app.Run()
}

type myBaseController struct {
	eudore.ControllerBase
}

func (ctl *myBaseController) Any() {
	ctl.Info("myBaseController Any")
}
func (*myBaseController) Get() interface{} {
	return "get myBaseController"
}
func (ctl *myBaseController) GetInfoById() interface{} {
	return ctl.GetParam("id")
}

func (ctl *myBaseController) Help(ctx eudore.Context) {}

func (*myBaseController) ControllerRoute() map[string]string {
	return map[string]string{
		// 修改Path方法的路由注册
		"Help": "",
	}
}
