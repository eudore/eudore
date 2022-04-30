package eudore_test

import (
	"fmt"
	"testing"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

type Test014Controller struct {
	eudore.ControllerAutoRoute
}

func (ctl *Test014Controller) Get(ctx eudore.Context) interface{} {
	return "Test014Controller"
}

func TestRouterStdAdd(t *testing.T) {
	type String struct {
		Data string
	}

	app := eudore.NewApp()
	app.AddHandlerExtend(func(str String) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			ctx.WriteString(str.Data)
		}
	})
	app.AddMiddleware(middleware.NewRecoverFunc())
	app.AddController(&Test014Controller{})

	api := app.Group("/method")
	api.AddHandler("TEST", "/*", String{"test"})
	api.AddHandler("LOCK", "/*", String{"lock"})
	api.AddHandler("UNLOCK", "/*", String{"unlock"})
	api.AddHandler("MOVE", "/*", String{"lock"})
	api.AnyFunc("/*", String{"any /*"})
	api.GetFunc("/", String{"get"})
	api.PostFunc("/", String{"post"})
	api.PutFunc("/", String{"put"})
	api.DeleteFunc("/", String{"delete"})
	api.HeadFunc("/", String{"head"})
	api.PatchFunc("/", String{"patch"})

	app.CancelFunc()
	app.Run()
}

func TestRouterError(t *testing.T) {
	app := eudore.NewApp()
	app.AddHandlerExtend("/api", TestRouterError)
	app.AddController(&Test015Controller{})
	app.AddController(eudore.NewControllerError(&Test015Controller{}, fmt.Errorf("test controller error")))
	app.GetFunc("{/*}", eudore.HandlerEmpty)
	app.GetFunc("/*path|check", eudore.HandlerEmpty)
	app.GetFunc("/err", func(*testing.T) {})

	app.CancelFunc()
	app.Run()
}

type Test015Controller struct {
	eudore.ControllerAutoRoute
}

func (Test015Controller) String() string {
	return "015"
}

func TestRouterStd404_405(t *testing.T) {
	app := eudore.NewApp()
	client := eudore.NewClientWarp()
	app.SetValue(eudore.ContextKeyClient, client)
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.AddHandler("404", "", eudore.HandlerRouter404)
	app.AddHandler("405", "", eudore.HandlerRouter405)
	app.GetFunc("/index", eudore.HandlerEmpty)

	client.NewRequest("GET", "/").Do().Callback(eudore.NewResponseReaderCheckStatus(404))
	client.NewRequest("PUT", "/").Do().Callback(eudore.NewResponseReaderCheckStatus(404))
	client.NewRequest("GET", "/index").Do().Callback(eudore.NewResponseReaderCheckStatus(200))
	client.NewRequest("PUT", "/index").Do().Callback(eudore.NewResponseReaderCheckStatus(405))

	app.CancelFunc()
	app.Run()
}

func TestRouterMiddleware2(t *testing.T) {
	app := eudore.NewApp()
	app.AddMiddleware()
	app.AddMiddleware("/api", func(int) {})
	app.AddMiddleware(func(int) {})
	app.AddHandlerExtend()

	app.AddMiddleware(middleware.NewRecoverFunc(), middleware.NewLoggerFunc(app))
	app.AddMiddleware("/api/v2", eudore.HandlerEmpty)
	app.AddMiddleware("/api/v1", eudore.HandlerEmpty)
	app.AddMiddleware("/api", eudore.HandlerEmpty)
	app.AnyFunc("/api/v1", eudore.HandlerEmpty)

	apiv1 := app.Group("/api/v1")
	apiv1.AnyFunc("/users", eudore.HandlerEmpty)

	app.CancelFunc()
	app.Run()
}

func TestRouterCoreLock(t *testing.T) {
	app := eudore.NewApp()
	client := eudore.NewClientWarp()
	app.SetValue(eudore.ContextKeyClient, client)
	app.SetValue(eudore.ContextKeyRouter, eudore.NewRouterStd(eudore.NewRouterCoreLock(nil)))
	app.Info(app.Router.(interface{ Metadata() interface{} }).Metadata())
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.GetFunc("/", eudore.HandlerEmpty)

	client.NewRequest("GET", "/").Do().Callback(eudore.NewResponseReaderCheckStatus(1200))
	app.CancelFunc()
	app.Run()
}

func TestRouterCoreDebug(t *testing.T) {
	app := eudore.NewApp()
	client := eudore.NewClientWarp()
	app.SetValue(eudore.ContextKeyClient, client)
	app.SetValue(eudore.ContextKeyRouter, eudore.NewRouterStd(eudore.NewRouterCoreDebug(nil)))
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.GetFunc("/", eudore.HandlerEmpty)
	app.GetFunc("/index", eudore.HandlerEmpty)
	app.GetFunc("/health", func(ctx eudore.Context) interface{} {
		return app.Router.(interface{ Metadata() interface{} }).Metadata()
	})
	app.GetFunc("/delete", eudore.HandlerEmpty)
	app.GetFunc("/delete")

	client.NewRequest("GET", "/eudore/debug/router/data").AddHeader(eudore.HeaderAccept, eudore.MimeApplicationJSON).Do().Callback(eudore.NewResponseReaderCheckStatus(1200))
	client.NewRequest("GET", "/health").Do()
	app.CancelFunc()
	app.Run()
}

func TestRouterCoreHost(t *testing.T) {
	echoHandleHost := func(ctx eudore.Context) {
		ctx.WriteString(ctx.GetParam("host"))
	}

	app := eudore.NewApp()
	client := eudore.NewClientWarp()
	app.SetValue(eudore.ContextKeyClient, client)
	app.SetValue(eudore.ContextKeyRouter, eudore.NewRouterStd(eudore.NewRouterCoreHost(nil)))
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.AnyFunc("/* host=eudore.com", echoHandleHost)
	app.AnyFunc("/* host=eudore.com:8088", echoHandleHost)
	app.AnyFunc("/* host=eudore.cn", echoHandleHost)
	app.AnyFunc("/* host=eudore.*", echoHandleHost)
	app.AnyFunc("/* host=example.com", echoHandleHost)
	app.AnyFunc("/* host=www.*.cn", echoHandleHost)
	app.AnyFunc("/api/* host=*", echoHandleHost)
	app.AnyFunc("/api/* host=eudore.com,eudore.cn", echoHandleHost)
	app.AnyFunc("/*", echoHandleHost)

	client.NewRequest("GET", "/").Do().Callback(eudore.NewResponseReaderCheckStatus(200), eudore.NewResponseReaderCheckBody(""))
	client.NewRequest("GET", "/").AddHeader("Host", "eudore.cn").Do().Callback(eudore.NewResponseReaderCheckStatus(200), eudore.NewResponseReaderCheckBody("eudore.cn"))
	client.NewRequest("GET", "/").AddHeader("Host", "eudore.com").Do().Callback(eudore.NewResponseReaderCheckStatus(200), eudore.NewResponseReaderCheckBody("eudore.com"))
	client.NewRequest("GET", "/").AddHeader("Host", "eudore.com:8088").Do().Callback(eudore.NewResponseReaderCheckStatus(200), eudore.NewResponseReaderCheckBody("eudore.com"))
	client.NewRequest("GET", "/").AddHeader("Host", "eudore.com:8089").Do().Callback(eudore.NewResponseReaderCheckStatus(200), eudore.NewResponseReaderCheckBody("eudore.com"))
	client.NewRequest("GET", "/").AddHeader("Host", "eudore.net").Do().Callback(eudore.NewResponseReaderCheckStatus(200), eudore.NewResponseReaderCheckBody("eudore.*"))
	client.NewRequest("GET", "/").AddHeader("Host", "www.eudore.cn").Do().Callback(eudore.NewResponseReaderCheckStatus(200), eudore.NewResponseReaderCheckBody("www.*.cn"))
	client.NewRequest("GET", "/").AddHeader("Host", "example.com").Do().Callback(eudore.NewResponseReaderCheckStatus(200), eudore.NewResponseReaderCheckBody("example.com"))
	client.NewRequest("GET", "/").AddHeader("Host", "www.example").Do().Callback(eudore.NewResponseReaderCheckStatus(200), eudore.NewResponseReaderCheckBody(""))
	client.NewRequest("GET", "/api/v1").AddHeader("Host", "example.com").Do().Callback(eudore.NewResponseReaderCheckStatus(200), eudore.NewResponseReaderCheckBody("*"))
	client.NewRequest("GET", "/api/v1").AddHeader("Host", "eudore.com").Do().Callback(eudore.NewResponseReaderCheckStatus(200), eudore.NewResponseReaderCheckBody("eudore.com,eudore.cn"))

	app.CancelFunc()
	app.Run()
}
