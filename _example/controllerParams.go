package main

/*
通过重写控制器的ControllerParam方法,开源使用pkg、name、method生成额外的路由默认参数，默认附加额外参数。
*/

import (
	"fmt"
	"github.com/eudore/eudore"
)

type (
	myParamsController struct {
		eudore.ControllerAutoRoute
	}
)

func main() {
	app := eudore.NewApp()
	app.AddController(new(myParamsController))

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

// ControllerParam 方法添加路由参数信息。
func (ctl *myParamsController) ControllerParam(pkg, name, method string) string {
	return fmt.Sprintf("source=ControllerParam cpkg=%s cname=%s cmethod=%s", pkg, name, method)
}

func (ctl *myParamsController) ControllerGroup(string) string {
	return "myctl"
}

func (ctl *myParamsController) Any(ctx eudore.Context) {
	ctx.Info("myParamsController Any")
}
func (*myParamsController) Get() interface{} {
	return "get myParamsController"
}
func (ctl *myParamsController) GetInfoById(ctx eudore.Context) interface{} {
	return ctx.GetParam("id")
}
