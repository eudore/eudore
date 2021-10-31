package main

/*
eudore.NewControllerError会传递创建controller的error，由RouterAddController处理。
*/

import (
	"errors"
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewApp()
	app.AddController(NewErrController(-10))
	app.AddController(NewErrController(10))

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

type errController struct {
	eudore.ControllerAutoRoute
}

func NewErrController(i int) eudore.Controller {
	ctl := new(errController)
	if i < 0 {
		// controller创建错误处理。
		return eudore.NewControllerError(ctl, errors.New("int must grate 0"))
	}
	return ctl
}

func (ctl *errController) Any(ctx eudore.Context) {
	ctx.Info("errController Any")
}

// Get 方法注册不存在的扩展函数，触发注册error。
func (*errController) Get(eudore.ControllerAutoRoute) interface{} {
	return "get errController"
}
