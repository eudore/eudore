package main

/*
通过重写控制器的ControllerParam方法使用pkg、name、method创建action参数再附加到路由参数中。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/policy"
)

type policyActionController struct {
	eudore.ControllerAutoRoute
	policy.ControllerAction
}

func main() {
	app := eudore.NewApp()
	app.AddController(new(policyActionController))

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

func (ctl *policyActionController) Any(ctx eudore.Context) {
	ctx.Info("ramBaseController Any")
}
