package main

/*
Ram处理接口
type Handler interface {
	Match(int, string, eudore.Context) (bool, bool)
	// return1 验证结果 return2 是否验证
}
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/component/ram"
	"github.com/eudore/eudore/middleware"
)

func main() {
	acl := ram.NewACL()
	acl.AddPermission(1, "GetUserInfo")
	acl.BindAllowPermission(2, 1)

	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	// 测试给予id等于请求参数id，实际应由jwt、seession、token、cookie等方法计算得到UID。
	app.AddMiddleware(func(ctx eudore.Context) {
		ctx.SetParam(eudore.ParamUID, ctx.GetQuery("id"))
		ctx.SetParam(eudore.ParamUNAME, ctx.GetQuery("id"))
	})
	// acl处理/version 默认allow
	app.AnyFunc("/version action=GetInfo", ram.NewMiddleware(acl, ram.Allow{}), eudore.HandlerEmpty)
	// acl处理请求，默认deny
	app.AddMiddleware(ram.NewMiddleware(acl, ram.Deny{}))
	app.GetFunc("/api/id/:userid/info action=GetUserInfo", eudore.HandlerEmpty)
	app.GetFunc("/api/name/:username/info action=GetUserInfo", eudore.HandlerEmpty)
	app.AnyFunc("/info action=GetInfo", eudore.HandlerEmpty)
	app.AnyFunc("/*", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/?id=1").Do().CheckStatus(200)
	client.NewRequest("GET", "/info?id=1").Do().CheckStatus(200)
	client.NewRequest("GET", "/version?id=1").Do().CheckStatus(200)
	client.NewRequest("GET", "/api/id/1/info?id=1").Do().CheckStatus(200)
	client.NewRequest("GET", "/api/id/1/info?id=2").Do().CheckStatus(200)
	client.NewRequest("GET", "/api/name/1/info?id=1").Do().CheckStatus(200)
	client.NewRequest("GET", "/api/name/1/info?id=2").Do().CheckStatus(200)

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
