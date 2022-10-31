package eudore_test

import (
	"testing"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func TestRouterStdAny(t *testing.T) {
	// 扩展RouterStd允许的方法
	eudore.DefaultRouterAllMethod = append(eudore.DefaultRouterAllMethod, "LOCK", "UNLOCK")
	eudore.DefaultRouterAnyMethod = append(eudore.DefaultRouterAnyMethod, "LOCK", "UNLOCK")
	defer func() {
		eudore.DefaultRouterAllMethod = eudore.DefaultRouterAllMethod[:len(eudore.DefaultRouterAllMethod)-2]
		eudore.DefaultRouterAnyMethod = eudore.DefaultRouterAnyMethod[:len(eudore.DefaultRouterAnyMethod)-2]
	}()

	app := eudore.NewApp()
	// Any方法覆盖
	app.GetFunc("/get/:val", func(ctx eudore.Context) {
		ctx.WriteString("method is get\n")
	})
	app.AnyFunc("/get/:val", func(ctx eudore.Context) {
		ctx.WriteString("method is any\n")
	})
	app.PostFunc("/get/:val", func(ctx eudore.Context) {
		ctx.WriteString("method is post\n")
	})
	app.AddHandler("LOCK", "/get/:val", func(ctx eudore.Context) {
		ctx.WriteString("method is lock\n")
	})
	app.GetFunc("/index", eudore.HandlerEmpty)
	app.AddHandler("404,444", "", eudore.HandlerRouter404)
	app.AddHandler("405", "", eudore.HandlerRouter405)

	// 请求测试
	app.NewRequest(nil, "GET", "/get/1")
	app.NewRequest(nil, "POST", "/get/2")
	app.NewRequest(nil, "PUT", "/get/3")
	app.NewRequest(nil, "LOCK", "/get/4")
	app.NewRequest(nil, "COPY", "/get/5")
	app.NewRequest(nil, "GET", "/get")
	app.NewRequest(nil, "POST", "/get")
	app.NewRequest(nil, "PUT", "/get")
	app.NewRequest(nil, "PUT", "/3")
	app.NewRequest(nil, "put", "/3")
	app.NewRequest(nil, "POST", "/index")

	app.CancelFunc()
	app.Run()
}

func TestRouterStdCheck(t *testing.T) {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyFuncCreator, eudore.NewFuncCreator())
	app.SetValue(eudore.ContextKeyRouter, eudore.NewRouterStd(nil))

	app.AnyFunc("/1/:num|isnum version=1", eudore.HandlerEmpty)
	app.AnyFunc("/1/222", eudore.HandlerEmpty)
	app.AnyFunc("/2/:num|num", eudore.HandlerEmpty)
	app.AnyFunc("/2/:num|", eudore.HandlerEmpty)
	app.AnyFunc("/2/:", eudore.HandlerEmpty)
	app.AnyFunc("/3/:num|isnum/22", eudore.HandlerEmpty)
	app.AnyFunc("/3/:num|isnum/*", eudore.HandlerEmpty)
	app.AnyFunc("/4/*num|isnum", eudore.HandlerEmpty)
	app.AnyFunc("/4/*num|isnum", eudore.HandlerEmpty)
	app.AnyFunc("/4/*", eudore.HandlerEmpty)
	app.AnyFunc("/5/*num|num", eudore.HandlerEmpty)
	app.AnyFunc("/api/v1/2", eudore.HandlerEmpty)
	app.AnyFunc("/api/v1/1", eudore.HandlerEmpty)
	app.AnyFunc("/*num|^\\d+$", eudore.HandlerEmpty)
	app.AnyFunc("/api/v1/*|{^0/api\\S+$}", eudore.HandlerEmpty)
	app.AnyFunc("/api/v1/*|{\\s+{}}", eudore.HandlerEmpty)
	app.AnyFunc("{/api/v1/*}", eudore.HandlerEmpty)
	app.AnyFunc("/api/v1/{{*}}", eudore.HandlerEmpty)
	app.AnyFunc("/api/*", eudore.HandlerEmpty)
	app.AnyFunc("/api/*", eudore.HandlerEmpty)
	app.AddHandler(eudore.MethodOptions, "/", eudore.HandlerEmpty)
	app.AddHandler(eudore.MethodConnect, "/", eudore.HandlerEmpty)
	app.AddHandler(eudore.MethodTrace, "/", eudore.HandlerEmpty)

	// 请求测试
	app.NewRequest(nil, "GET", "/1/1")
	app.NewRequest(nil, "POST", "/1/222")
	app.NewRequest(nil, "PUT", "/2/3")
	app.NewRequest(nil, "PUT", "/3/11/3")
	app.NewRequest(nil, "PUT", "/3/11/22")
	app.NewRequest(nil, "PUT", "/4/22")
	app.NewRequest(nil, "PUT", "/5/22")
	app.NewRequest(nil, "PUT", "/:{num}")

	app.CancelFunc()
	app.Run()
}

func TestRouterStdDelete(t *testing.T) {
	eudore.DefaultRouterAllMethod = append(eudore.DefaultRouterAllMethod, "LOCK", "UNLOCK")
	eudore.DefaultRouterAnyMethod = append(eudore.DefaultRouterAnyMethod, "LOCK", "UNLOCK")
	defer func() {
		eudore.DefaultRouterAllMethod = eudore.DefaultRouterAllMethod[:len(eudore.DefaultRouterAllMethod)-2]
		eudore.DefaultRouterAnyMethod = eudore.DefaultRouterAnyMethod[:len(eudore.DefaultRouterAnyMethod)-2]
	}()

	echoStringHandler := func(str string) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			ctx.WriteString(str)
		}
	}

	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyRouter, eudore.NewRouterStd(eudore.NewRouterCoreLock(nil)))

	register := app.Group(" register=off")
	app.AnyFunc("/version", echoStringHandler("any version"))
	app.AnyFunc("/version")

	app.AnyFunc("/version", echoStringHandler("any version"))
	app.AnyFunc("/version1", echoStringHandler("any version"))
	app.NewRequest(nil, "GET", "/version", eudore.NewClientCheckStatus(200), eudore.NewClientCheckBody("any version"))
	app.GetFunc("/version", echoStringHandler("get version"))
	app.NewRequest(nil, "GET", "/version", eudore.NewClientCheckStatus(200), eudore.NewClientCheckBody("get version"))
	register.AddHandler("GET,POST", "/version", echoStringHandler("get version"))
	app.NewRequest(nil, "GET", "/version", eudore.NewClientCheckStatus(200), eudore.NewClientCheckBody("any version"))
	register.AnyFunc("/version*", echoStringHandler("any version"))
	register.AnyFunc("/version0", echoStringHandler("any version"))
	register.AnyFunc("/version2", echoStringHandler("any version"))
	register.AnyFunc("/version1", echoStringHandler("any version"))
	register.AnyFunc("/version", echoStringHandler("any version"))

	app.AnyFunc("/api/v:v1/*", eudore.HandlerEmpty)
	app.AnyFunc("/api/v:v2/*", eudore.HandlerEmpty)
	app.AddHandler("LOCK", "/api/v:v2/*", eudore.HandlerEmpty)
	app.AddHandler("LOCK", "/api/v:v3/*", eudore.HandlerEmpty)
	register.AnyFunc("/api/v:ve/*", eudore.HandlerEmpty)
	register.AnyFunc("/api/v:v1/*", eudore.HandlerEmpty)
	register.AddHandler("LOCK", "/api/v:v2/*", eudore.HandlerEmpty)
	register.AnyFunc("/api/v:v2/*", eudore.HandlerEmpty)
	register.AddHandler("LOCK", "/api/v:v2/*", eudore.HandlerEmpty)
	register.AddHandler("LOCK", "/api/v:v3/*", eudore.HandlerEmpty)

	// ---------------- 测试 ----------------
	app.SetValue(eudore.ContextKeyRouter, eudore.NewRouterStd(eudore.NewRouterCoreLock(nil)))
	register = app.Group(" register=off")
	app.AnyFunc("/eudore/debug/look/*", middleware.NewLookFunc(app))
	app.AnyFunc("/version", echoStringHandler("any version"))
	app.AnyFunc("/version1", echoStringHandler("any version"))
	app.AnyFunc("/version2", echoStringHandler("any version"))
	app.NewRequest(nil, "GET", "/version", eudore.NewClientCheckStatus(200), eudore.NewClientCheckBody("any version"))
	app.GetFunc("/version", echoStringHandler("get version"))
	app.NewRequest(nil, "GET", "/version", eudore.NewClientCheckStatus(200), eudore.NewClientCheckBody("get version"))
	register.AddHandler("GET,POST", "/version", echoStringHandler("get version"))
	app.NewRequest(nil, "GET", "/version", eudore.NewClientCheckStatus(200), eudore.NewClientCheckBody("any version"))
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
	app.AnyFunc("/api/v1/user/:id|isnum", eudore.HandlerEmpty)
	app.AnyFunc("/api/v1/user/*name|nozero", eudore.HandlerEmpty)
	register.AnyFunc("/api/v1/user/:id|isnum/", eudore.HandlerEmpty)
	register.AnyFunc("/api/v1/user/:id|isnum", eudore.HandlerEmpty)
	register.AnyFunc("/api/v1/user/*name|nozero", eudore.HandlerEmpty)
	register.AnyFunc("/api/v1/user/id/:id", eudore.HandlerEmpty)
	register.AnyFunc("/api/v1/user/name/*name", eudore.HandlerEmpty)

	app.CancelFunc()
	app.Run()
}
