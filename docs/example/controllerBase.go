package main

import (
	"github.com/eudore/eudore"
)

type MyBaseController struct {
	eudore.ControllerBase
}

func (ctl *MyBaseController) Any() {
	ctl.Info("MyBaseController Any")
}
func (*MyBaseController) Get() interface{} {
	return "get MyBaseController"
}
func (ctl *MyBaseController) GetInfoById() interface{} {
	return ctl.GetParam("id")
}

func main() {
	app := eudore.NewCore()
	app.AddController(new(MyBaseController))

	app.Listen(":8084")
	app.Run()
}
