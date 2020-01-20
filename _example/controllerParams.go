package main

/*
通过重写控制器的GetRouteParam方法,开源使用pkg、name、method生成额外的路由默认参数，默认附加额外参数。
*/

import (
	"fmt"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

type (
	myParamsController struct {
		eudore.ControllerBase
	}
)

func main() {
	app := eudore.NewCore()
	httptest.NewClient(app).Stop(0)
	app.AddController(new(myParamsController))

	app.Listen(":8088")
	app.Run()
}

// GetRouteParam 方法添加路由参数信息。
func (ctl *myParamsController) GetRouteParam(pkg, name, method string) string {
	return fmt.Sprintf("source=GetRouteParam cpkg=%s cname=%s cmethod=%s", pkg, name, method)
}

func (ctl *myParamsController) Any() {
	ctl.Info("myParamsController Any")
}
func (*myParamsController) Get() interface{} {
	return "get myParamsController"
}
func (ctl *myParamsController) GetInfoById() interface{} {
	return ctl.GetParam("id")
}
