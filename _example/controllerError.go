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
	app.AddController(new(myErrController))

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

type myErrController struct {
	eudore.ControllerBase
}

func (*myErrController) Inject(ctl eudore.Controller, router eudore.Router) error {
	err := eudore.ControllerInjectStateful(ctl, router)
	if err != nil {
		return err
	}
	return errors.New("inject test error")
}

func (ctl *myErrController) Any() {
	ctl.Info("myErrController Any")
}
func (*myErrController) Get(eudore.ControllerBase) interface{} {
	return "get myErrController"
}
func (ctl *myErrController) GetInfoById() interface{} {
	return ctl.GetParam("id")
}
