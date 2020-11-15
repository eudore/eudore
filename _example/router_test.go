package eudore_test

import (
	"testing"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func TestRouterStdAny2(t *testing.T) {
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
	app.AddHandler("404,444", "", eudore.HandlerRouter404)
	app.AddHandler("405", "", eudore.HandlerRouter405)

	// 请求测试
	client := httptest.NewClient(app)
	client.NewRequest("GET", "/get/1").Do().CheckStatus(200).CheckBodyContainString("get").OutBody()
	client.NewRequest("POST", "/get/2").Do().CheckStatus(200).CheckBodyContainString("post").OutBody()
	client.NewRequest("PUT", "/get/3").Do().CheckStatus(200).CheckBodyContainString("any").OutBody()
	client.NewRequest("GET", "/get").Do().CheckStatus(404)
	client.NewRequest("POST", "/get").Do().CheckStatus(404)
	client.NewRequest("PUT", "/get").Do().CheckStatus(404)
	client.NewRequest("PUT", "/3").Do().CheckStatus(404)
	client.NewRequest("put", "/3").Do().CheckStatus(405)

	app.CancelFunc()
	app.Run()
}

func TestRouterStdCheck2(t *testing.T) {
	app := eudore.NewApp()

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
	app.AddHandler(eudore.MethodOptions, "/", eudore.HandlerEmpty)
	app.AddHandler(eudore.MethodConnect, "/", eudore.HandlerEmpty)
	app.AddHandler(eudore.MethodTrace, "/", eudore.HandlerEmpty)

	// 请求测试
	client := httptest.NewClient(app)
	client.NewRequest("GET", "/1/1").Do().CheckStatus(200)
	client.NewRequest("POST", "/1/222").Do().CheckStatus(200)
	client.NewRequest("PUT", "/2/3").Do().CheckStatus(200)
	client.NewRequest("PUT", "/3/11/3").Do().CheckStatus(200)
	client.NewRequest("PUT", "/3/11/22").Do().CheckStatus(200)
	client.NewRequest("PUT", "/4/22").Do().CheckStatus(200)
	client.NewRequest("PUT", "/5/22").Do().CheckStatus(404)

	app.CancelFunc()
	app.Run()
}

func TestRouterMiddleware2(t *testing.T) {
	app := eudore.NewApp()
	app.AddMiddleware()
	app.AddMiddleware(func(int) {})
	app.AddHandlerExtend()
	app.CancelFunc()
	app.Run()
}

func TestRouter2(t *testing.T) {
	app := eudore.NewApp()
	app.AddMiddleware("/api/user", eudore.HandlerEmpty)
	app.AddMiddleware("/api/", eudore.HandlerEmpty)
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))

	api1 := app.Group("/api/v1")
	api1.AnyFunc("/any", eudore.HandlerEmpty)
	api1.DeleteFunc("/delete", eudore.HandlerEmpty)
	api1.HeadFunc("/head", eudore.HandlerEmpty)
	api1.PatchFunc("route=/patch", eudore.HandlerEmpty)
	api1.AddHandler(eudore.MethodOptions, "route=/options", eudore.HandlerEmpty)
	api1.AddHandler(eudore.MethodConnect, "/", eudore.HandlerEmpty)
	api1.AddHandler(eudore.MethodTrace, "/", eudore.HandlerEmpty)
	app.CancelFunc()
	app.Run()
}
