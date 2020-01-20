package main

/*
通过重写控制器的GetRouteParam方法使用pkg、name、method创建action参数再附加到路由参数中。
*/

import (
	"fmt"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"strings"
)

type (
	ramActionController struct {
		eudore.ControllerBase
	}

	ramBaseController struct {
		ramActionController
	}
)

func main() {
	app := eudore.NewCore()
	httptest.NewClient(app).Stop(0)
	app.AddController(new(ramBaseController))

	app.Listen(":8088")
	app.Run()
}

// GetRouteParam 方法添加路由参数信息。
func (ctl *ramActionController) GetRouteParam(pkg, name, method string) string {
	pos := strings.LastIndexByte(pkg, '/') + 1
	if pos != 0 {
		pkg = pkg[pos:]
	}
	if strings.HasSuffix(name, "Controller") {
		name = name[:len(name)-len("Controller")]
	}
	if pkg == "task" {
		return ""
	}
	return fmt.Sprintf("action=%s:%s:%s", pkg, name, method)
}

func (ctl *ramBaseController) Any() {
	ctl.Info("ramBaseController Any")
}
func (*ramBaseController) Get() interface{} {
	return "get ramBaseController"
}
func (ctl *ramBaseController) GetInfoById() interface{} {
	return ctl.GetParam("id")
}
