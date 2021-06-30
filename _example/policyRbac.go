package main

import (
	"fmt"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
	"github.com/eudore/eudore/policy"
)

func main() {
	app := eudore.NewApp(eudore.Renderer(eudore.RenderJSON))
	policys := policy.NewPolicys()

	{
		// 默认用户0 绑定角色1
		policys.AddRole(0, 1)
		// 角色1绑定权限
		policys.AddPermission(1, "hello", "1", "2")
	}
	app.GetFunc("/policys/runtime", policys.HandleRuntime)

	app.AddMiddleware(middleware.NewLoggerFunc(app, "route", "action", "Policy", "Resource", "Userid"))
	app.AddMiddleware(policys.HandleHTTP)
	for _, i := range []int{1, 2, 3, 4, 5, 6} {
		app.AnyFunc(fmt.Sprintf("/%d action=%d", i, i), eudore.HandlerEmpty)
	}
	app.AnyFunc("/* action=hello resource-prefix=/", eudore.HandlerEmpty)

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
