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
	httptest.NewClient(app).Stop(0)
	app.AddController(new(myBaseController))
	app.Listen(":8088")
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
