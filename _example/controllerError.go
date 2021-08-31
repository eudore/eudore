package main

/*
eudore.Controller是一个接口，可自行实现，eudore.ControllerBase只是其中一种默认实现。
*/

import (
	"errors"
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewApp()
	app.AddController(NewMyErrController(-10))
	app.AddController(NewMyErrController(10))

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

type myErrController struct {
	eudore.ControllerAutoRoute
}

func NewMyErrController(i int) eudore.Controller {
	ctl := new(myErrController)
	if i < 0 {
		return eudore.NewControllerError(ctl, errors.New("int must grate 0"))
	}
	return ctl
}

func (ctl *myErrController) Any(ctx eudore.Context) {
	ctx.Info("myErrController Any")
}
func (*myErrController) Get(eudore.ControllerAutoRoute) interface{} {
	return "get myErrController"
}
func (ctl *myErrController) GetInfoById(ctx eudore.Context) interface{} {
	return ctx.GetParam("id")
}
