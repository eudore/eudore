package main

/*
运行时动态增删路由规则需要路由器核心带锁(RouterCoreLock包装)，防止数据修改(非原子操作)中路由匹配数据混乱。
路由注册存在参数'register=off'或处理函数为nil时，会移除方法和路由路径完全相同的路由节点。
移除的方法和Route Path必须和注册时完全一致
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyRouter, eudore.NewRouter(eudore.NewRouterCoreLock(nil)))
	client := app.WithClient(eudore.NewClientCheckStatus(200))

	register := app.Group(" register=off")
	app.AnyFunc("/version", echoStringHandler("any version"))
	app.AnyFunc("/version")

	app.AnyFunc("/version", echoStringHandler("any version"))
	app.AnyFunc("/version1", echoStringHandler("any version"))
	client.NewRequest(nil, "GET", "/version", eudore.NewClientCheckBody("any version"))
	app.GetFunc("/version", echoStringHandler("get version"))
	client.NewRequest(nil, "GET", "/version", eudore.NewClientCheckBody("get version"))
	register.AddHandler("GET,POST", "/version", echoStringHandler("get version"))
	client.NewRequest(nil, "GET", "/version", eudore.NewClientCheckBody("any version"))
	register.AnyFunc("/version*", echoStringHandler("any version"))
	register.AnyFunc("/version0", echoStringHandler("any version"))
	register.AnyFunc("/version2", echoStringHandler("any version"))
	register.AnyFunc("/version1", echoStringHandler("any version"))
	register.AnyFunc("/version", echoStringHandler("any version"))

	app.AnyFunc("/api/v:v1/*", eudore.HandlerEmpty)
	app.AnyFunc("/api/v:v2/*", eudore.HandlerEmpty)
	register.AnyFunc("/api/v:ve/*", eudore.HandlerEmpty)
	register.AnyFunc("/api/v:v1/*", eudore.HandlerEmpty)
	register.AnyFunc("/api/v:v2/*", eudore.HandlerEmpty)

	// ---------------- 测试 ----------------

	app.SetValue(eudore.ContextKeyRouter, eudore.NewRouter(eudore.NewRouterCoreLock(nil)))
	register = app.Group(" register=off")
	app.AnyFunc("/eudore/debug/look/*", middleware.NewLookFunc(app))
	app.AnyFunc("/version", echoStringHandler("any version"))
	app.AnyFunc("/version1", echoStringHandler("any version"))
	app.AnyFunc("/version2", echoStringHandler("any version"))
	client.NewRequest(nil, "GET", "/version", eudore.NewClientCheckBody("any version"))
	app.GetFunc("/version", echoStringHandler("get version"))
	client.NewRequest(nil, "GET", "/version", eudore.NewClientCheckBody("get version"))
	register.AddHandler("GET,POST", "/version", echoStringHandler("get version"))
	client.NewRequest(nil, "GET", "/version", eudore.NewClientCheckBody("any version"))
	register.GetFunc("/version", echoStringHandler("get version"))
	register.AnyFunc("/version*", echoStringHandler("any version"))
	register.AnyFunc("/version0", echoStringHandler("any version"))
	register.AnyFunc("/version1", echoStringHandler("any version"))
	register.AnyFunc("/version3", echoStringHandler("any version"))
	register.AnyFunc("/version2", echoStringHandler("any version"))
	register.AnyFunc("/version", echoStringHandler("any version"))

	app.AnyFunc("/api/v:v1/*", eudore.HandlerEmpty)
	app.AnyFunc("/api/v:v2/*", eudore.HandlerEmpty)
	app.AddHandler("TEST", "/api/v:v2/*", eudore.HandlerEmpty)
	register.AnyFunc("/api/v:ve/*", eudore.HandlerEmpty)
	register.AnyFunc("/api/v:v1/*", eudore.HandlerEmpty)
	register.AnyFunc("/api/v:v2/*", eudore.HandlerEmpty)

	app.AnyFunc("/api/v1/user/id/:id", eudore.HandlerEmpty)
	app.AnyFunc("/api/v1/user/name/*name", eudore.HandlerEmpty)
	app.AnyFunc("/api/v1/user/:id|num", eudore.HandlerEmpty)
	app.AnyFunc("/api/v1/user/*name|nozero", eudore.HandlerEmpty)
	register.AnyFunc("/api/v1/user/:id|num/", eudore.HandlerEmpty)
	register.AnyFunc("/api/v1/user/:id|num", eudore.HandlerEmpty)
	register.AnyFunc("/api/v1/user/*name|nozero", eudore.HandlerEmpty)
	register.AnyFunc("/api/v1/user/id/:id", eudore.HandlerEmpty)
	register.AnyFunc("/api/v1/user/name/*name", eudore.HandlerEmpty)

	app.Listen(":8088")
	app.Run()
}

func echoStringHandler(str string) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		ctx.WriteString(str)
	}
}
