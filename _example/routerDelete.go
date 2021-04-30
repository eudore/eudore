package main

/*
运行时动态增删路由规则需要路由器核心带锁(RouterCoreLock包装)，防止数据修改(非原子操作)中路由匹配数据混乱。
路由注册存在参数'register=off'或处理函数为nil时，会移除方法和路由路径完全相同的路由节点。
移除的方法和Route Path必须和注册时完全一致
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp(
		eudore.NewRouterStd(eudore.NewRouterCoreLock(nil)),
	)

	client := httptest.NewClient(app)

	register := app.Group(" register=off")
	app.AnyFunc("/verison", echoStringHandler("any verison"))
	app.AnyFunc("/verison")

	app.AnyFunc("/verison", echoStringHandler("any verison"))
	app.AnyFunc("/verison1", echoStringHandler("any verison"))
	client.NewRequest("GET", "/verison").Do().CheckStatus(200).CheckBodyString("any verison")
	app.GetFunc("/verison", echoStringHandler("get verison"))
	client.NewRequest("GET", "/verison").Do().CheckStatus(200).CheckBodyString("get verison")
	register.AddHandler("GET,POST", "/verison", echoStringHandler("get verison"))
	client.NewRequest("GET", "/verison").Do().CheckStatus(200).CheckBodyString("any verison")
	register.AnyFunc("/verison*", echoStringHandler("any verison"))
	register.AnyFunc("/verison0", echoStringHandler("any verison"))
	register.AnyFunc("/verison2", echoStringHandler("any verison"))
	register.AnyFunc("/verison1", echoStringHandler("any verison"))
	register.AnyFunc("/verison", echoStringHandler("any verison"))

	app.AnyFunc("/api/v:v1/*", eudore.HandlerEmpty)
	app.AnyFunc("/api/v:v2/*", eudore.HandlerEmpty)
	register.AnyFunc("/api/v:ve/*", eudore.HandlerEmpty)
	register.AnyFunc("/api/v:v1/*", eudore.HandlerEmpty)
	register.AnyFunc("/api/v:v2/*", eudore.HandlerEmpty)

	// ---------------- 测试 ----------------

	app.Options(eudore.NewRouterStd(eudore.NewRouterCoreLock(nil)))
	register = app.Group(" register=off")
	app.AnyFunc("/eudore/debug/look/*", middleware.NewLookFunc(app))
	app.AnyFunc("/verison", echoStringHandler("any verison"))
	app.AnyFunc("/verison1", echoStringHandler("any verison"))
	app.AnyFunc("/verison2", echoStringHandler("any verison"))
	client.NewRequest("GET", "/verison").Do().CheckStatus(200).CheckBodyString("any verison")
	app.GetFunc("/verison", echoStringHandler("get verison"))
	client.NewRequest("GET", "/verison").Do().CheckStatus(200).CheckBodyString("get verison")
	register.AddHandler("GET,POST", "/verison", echoStringHandler("get verison"))
	client.NewRequest("GET", "/verison").Do().CheckStatus(200).CheckBodyString("any verison")
	register.GetFunc("/verison", echoStringHandler("get verison"))
	register.AnyFunc("/verison*", echoStringHandler("any verison"))
	register.AnyFunc("/verison0", echoStringHandler("any verison"))
	register.AnyFunc("/verison1", echoStringHandler("any verison"))
	register.AnyFunc("/verison3", echoStringHandler("any verison"))
	register.AnyFunc("/verison2", echoStringHandler("any verison"))
	register.AnyFunc("/verison", echoStringHandler("any verison"))

	app.AnyFunc("/api/v:v1/*", eudore.HandlerEmpty)
	app.AnyFunc("/api/v:v2/*", eudore.HandlerEmpty)
	app.AddHandler("TEST", "/api/v:v2/*", eudore.HandlerEmpty)
	register.AnyFunc("/api/v:ve/*", eudore.HandlerEmpty)
	register.AnyFunc("/api/v:v1/*", eudore.HandlerEmpty)
	register.AnyFunc("/api/v:v2/*", eudore.HandlerEmpty)

	app.AnyFunc("/api/v1/user/id/:id", eudore.HandlerEmpty)
	app.AnyFunc("/api/v1/user/name/*name", eudore.HandlerEmpty)
	app.AnyFunc("/api/v1/user/:id|isnum", eudore.HandlerEmpty)
	app.AnyFunc("/api/v1/user/*name|nozero", eudore.HandlerEmpty)
	register.AnyFunc("/api/v1/user/:id|isnum/", eudore.HandlerEmpty)
	register.AnyFunc("/api/v1/user/:id|isnum", eudore.HandlerEmpty)
	register.AnyFunc("/api/v1/user/*name|nozero", eudore.HandlerEmpty)
	register.AnyFunc("/api/v1/user/id/:id", eudore.HandlerEmpty)
	register.AnyFunc("/api/v1/user/name/*name", eudore.HandlerEmpty)

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

func echoStringHandler(str string) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		ctx.WriteString(str)
	}
}
