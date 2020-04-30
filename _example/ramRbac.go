package main

/*
rbac 是给予角色访问控制，用户拥有多个角色，角色拥有多个权限，那么就是用户拥有多项权限，如果用户拥有某项权限访问控制即为通过。
*/

import (
	"fmt"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/component/ram"
	"github.com/eudore/eudore/middleware"
)

func main() {
	rbac := ram.NewRbac()
	// 创建权限id=1 name=1
	rbac.AddPermission(1, "1")
	rbac.AddPermission(2, "2")
	rbac.AddPermission(3, "3")
	rbac.AddPermission(4, "4")
	rbac.AddPermission(5, "5")
	rbac.AddPermission(6, "6")
	// 角色id1 绑定 权限id1 2 3
	rbac.BindPermissions(1, 1, 2, 3)
	// 角色id2 绑定 权限id4 5 6
	rbac.BindPermissions(2, 4, 5, 6)
	// 用户id 绑定 角色id1
	rbac.BindRole(2, 1)

	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "action", "ram", "route", "resource", "browser"))
	// 测试给予参数 UID=2  即用户id为2，实际应由jwt、seession、token、cookie等方法计算得到UID。
	app.AddMiddleware(func(ctx eudore.Context) {
		ctx.SetParam(eudore.ParamUID, "2")
	})
	app.AddMiddleware(ram.NewMiddleware(rbac))
	for _, i := range []int{1, 2, 3, 4, 5, 6} {
		app.AnyFunc(fmt.Sprintf("/%d action=%d", i, i), eudore.HandlerEmpty)
	}
	app.AnyFunc("/* action=hello", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("PUT", "/hello").Do().CheckStatus(200)
	client.NewRequest("GET", "/1").Do().CheckStatus(200)
	client.NewRequest("GET", "/2").Do().CheckStatus(200)
	client.NewRequest("PUT", "/3").Do().CheckStatus(200)
	client.NewRequest("PUT", "/4").Do().CheckStatus(200)
	client.NewRequest("PUT", "/5").Do().CheckStatus(200)
	client.NewRequest("PUT", "/6").Do().CheckStatus(200)
	for client.Next() {
		app.Error(client.Error())
	}

	app.CancelFunc()
	app.Run()
}
