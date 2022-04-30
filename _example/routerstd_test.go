package eudore_test

import (
	"testing"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func TestRouterStdAny(t *testing.T) {
	// 扩展RouterStd允许的方法
	eudore.RouterAllMethod = append(eudore.RouterAllMethod, "LOCK", "UNLOCK")
	eudore.RouterAnyMethod = append(eudore.RouterAnyMethod, "LOCK", "UNLOCK")
	defer func() {
		eudore.RouterAllMethod = eudore.RouterAllMethod[:len(eudore.RouterAllMethod)-2]
		eudore.RouterAnyMethod = eudore.RouterAnyMethod[:len(eudore.RouterAnyMethod)-2]
	}()

	app := eudore.NewApp()
	client := eudore.NewClientWarp()
	app.SetValue(eudore.ContextKeyClient, client)
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
	client.NewRequest("GET", "/get/1").Do()
	client.NewRequest("POST", "/get/2").Do()
	client.NewRequest("PUT", "/get/3").Do()
	client.NewRequest("LOCK", "/get/4").Do()
	client.NewRequest("COPY", "/get/5").Do()
	client.NewRequest("GET", "/get").Do()
	client.NewRequest("POST", "/get").Do()
	client.NewRequest("PUT", "/get").Do()
	client.NewRequest("PUT", "/3").Do()
	client.NewRequest("put", "/3").Do()
	client.NewRequest("POST", "/index").Do()

	app.CancelFunc()
	app.Run()
}

func TestRouterStdCheck(t *testing.T) {
	client := eudore.NewClientWarp()
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyClient, client)
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
	client.NewRequest("GET", "/1/1").Do()
	client.NewRequest("POST", "/1/222").Do()
	client.NewRequest("PUT", "/2/3").Do()
	client.NewRequest("PUT", "/3/11/3").Do()
	client.NewRequest("PUT", "/3/11/22").Do()
	client.NewRequest("PUT", "/4/22").Do()
	client.NewRequest("PUT", "/5/22").Do()
	client.NewRequest("PUT", "/:{num}").Do()

	app.CancelFunc()
	app.Run()
}

func TestRouterStdDelete(t *testing.T) {
	eudore.RouterAllMethod = append(eudore.RouterAllMethod, "LOCK", "UNLOCK")
	eudore.RouterAnyMethod = append(eudore.RouterAnyMethod, "LOCK", "UNLOCK")
	defer func() {
		eudore.RouterAllMethod = eudore.RouterAllMethod[:len(eudore.RouterAllMethod)-2]
		eudore.RouterAnyMethod = eudore.RouterAnyMethod[:len(eudore.RouterAnyMethod)-2]
	}()

	echoStringHandler := func(str string) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			ctx.WriteString(str)
		}
	}

	app := eudore.NewApp()
	client := eudore.NewClientWarp()
	app.SetValue(eudore.ContextKeyClient, client)
	app.SetValue(eudore.ContextKeyRouter, eudore.NewRouterStd(eudore.NewRouterCoreLock(nil)))

	register := app.Group(" register=off")
	app.AnyFunc("/version", echoStringHandler("any version"))
	app.AnyFunc("/version")

	app.AnyFunc("/version", echoStringHandler("any version"))
	app.AnyFunc("/version1", echoStringHandler("any version"))
	client.NewRequest("GET", "/version").Do().Callback(eudore.NewResponseReaderCheckStatus(200), eudore.NewResponseReaderCheckBody("any version"))
	app.GetFunc("/version", echoStringHandler("get version"))
	client.NewRequest("GET", "/version").Do().Callback(eudore.NewResponseReaderCheckStatus(200), eudore.NewResponseReaderCheckBody("get version"))
	register.AddHandler("GET,POST", "/version", echoStringHandler("get version"))
	client.NewRequest("GET", "/version").Do().Callback(eudore.NewResponseReaderCheckStatus(200), eudore.NewResponseReaderCheckBody("any version"))
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
	client.NewRequest("GET", "/version").Do().Callback(eudore.NewResponseReaderCheckStatus(200), eudore.NewResponseReaderCheckBody("any version"))
	app.GetFunc("/version", echoStringHandler("get version"))
	client.NewRequest("GET", "/version").Do().Callback(eudore.NewResponseReaderCheckStatus(200), eudore.NewResponseReaderCheckBody("get version"))
	register.AddHandler("GET,POST", "/version", echoStringHandler("get version"))
	client.NewRequest("GET", "/version").Do().Callback(eudore.NewResponseReaderCheckStatus(200), eudore.NewResponseReaderCheckBody("any version"))
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
