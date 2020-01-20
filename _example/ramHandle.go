package main

/*
Ram处理接口
type Handler interface {
	Match(int, string, eudore.Context) (bool, bool)
	// return1 验证结果 return2 是否验证
}
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

	app := eudore.NewCore()
	app.AddMiddleware(middleware.NewLoggerFunc(app.App, "route"))
	// 测试给予参数 UID=2  即用户id为2，实际应由jwt、seession、token、cookie等方法计算得到UID。
	app.AddMiddleware(func(ctx eudore.Context) {
		ctx.SetParam(eudore.ParamUID, "2")
	})
	app.AddMiddleware(func(ctx eudore.Context) {
		// 获取用户id
		uid := eudore.GetStringInt(ctx.GetParam(eudore.ParamUID))
		// 获取权限
		perm := ctx.Path()[1:]
		// 匹配 返回(是否分配，匹配结果)
		_, ok := acl.Match(uid, perm, ctx)
		// 如果匹配不成功就返回403
		if !ok {
			ctx.WriteHeader(403)
			ctx.End()
		}
	})
	for _, i := range []int{1, 2, 3, 4, 5, 6} {
		app.AnyFunc(fmt.Sprintf("/%d", i), fmt.Sprintf("hello %d", i))
	}
	app.AnyFunc("/*", "hello")

	client := httptest.NewClient(app)
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

	app.Listen(":8088")
	app.Run()
}
