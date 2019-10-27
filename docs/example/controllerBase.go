package main

import (
	"github.com/eudore/eudore"
)

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

func main() {
	app := eudore.NewCore()
	app.AddController(new(myBaseController))

	app.Listen(":8084")
	app.Run()
}
