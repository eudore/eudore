package main

/*
查看日志中显示的访问参数：

/3 status=403 ram=deny 由于acl没有给UID=2绑定权限id3，所以acl没有匹配这个请求，由默认处理deny处理，deny的处理结构统一为拒绝，所以返回403。

/4 status=403 ram=acl acl给UID2绑定的拒绝权限id4，所以acl匹配了这个请求ram处理者就是acl，但是绑定的是拒绝所以返回403。

/5 status=200 ram=acl acl给UID2绑定的允许权限id5，所以ram=acl(acl为时间ram处理者)，绑定的允许权限，所以访问通过返回200。
*/

import (
	"fmt"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/component/ram"
	"github.com/eudore/eudore/middleware"
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
	acl.BindDenyPermission(0, 2)
	acl.BindAllowPermission(1, 3)
	acl.BindAllowPermission(1, 4)
	acl.BindPermission(2, 4, false)
	acl.BindPermission(2, 5, true)
	acl.BindPermission(2, 6, true)

	acl.UnbindPermission(0, 6)
	acl.UnbindPermission(0, 1)
	acl.UnbindPermission(0, 2)
	acl.DeletePermission("6")

	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "action", "ram", "route", "resource", "browser"))
	// 测试给予参数 UID=2  即用户id为2，实际应由jwt、seession、token、cookie等方法计算得到UID。
	app.AddMiddleware(func(ctx eudore.Context) {
		ctx.SetParam(eudore.ParamUID, "2")
	})
	app.AddMiddleware(ram.NewMiddleware(acl))
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

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
