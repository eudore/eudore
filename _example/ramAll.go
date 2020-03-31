package main

/*
example使用Acl、Pbac混合鉴权，参考RamAcl.go、RamPbac.go
*/

import (
	"fmt"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
	"github.com/eudore/eudore/middleware/ram"
)

func main() {
	acl := ram.NewAcl()
	acl.AddPermission(1, "1")
	acl.AddPermission(2, "2")
	acl.AddPermission(3, "3")
	acl.AddPermission(4, "4")
	acl.AddPermission(5, "5")
	acl.AddPermission(6, "6")
	acl.AddPermission(10, "hello")
	acl.BindAllowPermission(0, 1)
	acl.BindAllowPermission(0, 2)
	acl.BindAllowPermission(1, 3)
	acl.BindAllowPermission(1, 4)
	acl.BindDenyPermission(2, 4)
	acl.BindAllowPermission(2, 5)
	acl.BindAllowPermission(2, 6)

	rbac := ram.NewRbac()
	rbac.AddPermission(1, "1")
	rbac.AddPermission(2, "2")
	rbac.AddPermission(3, "3")
	rbac.AddPermission(4, "4")
	rbac.AddPermission(5, "5")
	rbac.AddPermission(6, "6")
	rbac.BindPermissions(1, 1, 2, 3)
	rbac.BindPermissions(2, 4, 5, 6)
	rbac.BindRole(2, 1)

	pbac := ram.NewPbac()
	pbac.AddPolicyStringJson(1, `{"version":"1","description":"AdministratorAccess","statement":[{"effect":true,"action":["*"],"resource":["*"]}]}`)
	pbac.AddPolicyStringJson(2, `{"version":"1","description":"Get method allow","statement":[{"effect":true,"action":["*"],"resource":["*"],"conditions":{"method":["GET"]}}]}`)
	pbac.AddPolicyStringJson(3, `{"version":"1","description":"Get method allow","statement":[{"effect":true,"action":["3"],"resource":["*"]}]}`)
	pbac.BindPolicy(1, 2)
	pbac.BindPolicy(1, 3)

	app := eudore.NewCore()
	app.AddMiddleware(middleware.NewLoggerFunc(app.App, "action", "ram", "route", "resource", "browser"))
	// 测试给予参数 UID=2  即用户id为2，实际应由jwt、seession、token、cookie等方法计算得到UID。
	app.AddMiddleware(func(ctx eudore.Context) {
		ctx.SetParam(eudore.ParamUID, "2")
	})
	app.AddMiddleware(ram.NewMiddleware(acl, rbac, pbac))
	for _, i := range []int{1, 2, 3, 4, 5, 6} {
		app.AnyFunc(fmt.Sprintf("/%d action=%d", i, i), eudore.HandlerEmpty)
	}
	app.AnyFunc("/* action=hello", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("PUT", "/hello").Do().CheckStatus(200)
	client.NewRequest("PUT", "/1").Do().CheckStatus(200)
	client.NewRequest("PUT", "/2").Do().CheckStatus(200)
	client.NewRequest("PUT", "/3").Do().CheckStatus(200)
	client.NewRequest("PUT", "/4").Do().CheckStatus(200)
	client.NewRequest("PUT", "/5").Do().CheckStatus(200)
	client.NewRequest("PUT", "/6").Do().CheckStatus(200)
	for client.Next() {
		app.Error(client.Error())
	}
	client.Stop(0)

	app.Run()
}
