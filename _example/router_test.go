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
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.AddHandler("404", "", eudore.HandlerRouter404)
	app.AddHandler("405", "", eudore.HandlerRouter405)
	app.GetFunc("/index", eudore.HandlerEmpty)

	app.NewRequest(nil, "GET", "/", eudore.NewClientCheckStatus(404))
	app.NewRequest(nil, "PUT", "/", eudore.NewClientCheckStatus(404))
	app.NewRequest(nil, "GET", "/index", eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "PUT", "/index", eudore.NewClientCheckStatus(405))

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
	apiv1.AddMiddleware(func(int) {})
	apiv1.AnyFunc("/users", eudore.HandlerEmpty)

	app.CancelFunc()
	app.Run()
}

func TestRouterCoreLock(t *testing.T) {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyRouter, eudore.NewRouterStd(eudore.NewRouterCoreLock(nil)))
	app.Info(app.Router.(interface{ Metadata() interface{} }).Metadata())
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.GetFunc("/", eudore.HandlerEmpty)

	app.NewRequest(nil, "GET", "/", eudore.NewClientCheckStatus(200))
	app.CancelFunc()
	app.Run()
}

func TestRouterCoreDebug(t *testing.T) {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyRouter, eudore.NewRouterStd(eudore.NewRouterCoreDebug(nil)))
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.GetFunc("/", eudore.HandlerEmpty)
	app.GetFunc("/index", eudore.HandlerEmpty)
	app.GetFunc("/health", func(ctx eudore.Context) interface{} {
		return app.Router.(interface{ Metadata() interface{} }).Metadata()
	})
	app.GetFunc("/delete", eudore.HandlerEmpty)
	app.GetFunc("/delete")

	app.NewRequest(nil, "GET", "/eudore/debug/router/data",
		eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeApplicationJSON),
		eudore.NewClientCheckStatus(200),
	)
	app.NewRequest(nil, "GET", "/health")
	app.CancelFunc()
	app.Run()
}

func TestRouterCoreHost(t *testing.T) {
	echoHandleHost := func(ctx eudore.Context) {
		ctx.WriteString(ctx.GetParam("host"))
	}

	app := eudore.NewApp()
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

	app.NewRequest(nil, "GET", "/", eudore.NewClientCheckStatus(200), eudore.NewClientCheckBody(""))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHost("eudore.cn"), eudore.NewClientCheckStatus(200), eudore.NewClientCheckBody("eudore.cn"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHost("eudore.com"), eudore.NewClientCheckStatus(200), eudore.NewClientCheckBody("eudore.com"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHost("eudore.com:8088"), eudore.NewClientCheckStatus(200), eudore.NewClientCheckBody("eudore.com"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHost("eudore.com:8089"), eudore.NewClientCheckStatus(200), eudore.NewClientCheckBody("eudore.com"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHost("eudore.net"), eudore.NewClientCheckStatus(200), eudore.NewClientCheckBody("eudore.*"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHost("www.eudore.cn"), eudore.NewClientCheckStatus(200), eudore.NewClientCheckBody("www.*.cn"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHost("example.com"), eudore.NewClientCheckStatus(200), eudore.NewClientCheckBody("example.com"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHost("www.example"), eudore.NewClientCheckStatus(200), eudore.NewClientCheckBody(""))
	app.NewRequest(nil, "GET", "/api/v1", eudore.NewClientHost("example.com"), eudore.NewClientCheckStatus(200), eudore.NewClientCheckBody("*"))
	app.NewRequest(nil, "GET", "/api/v1", eudore.NewClientHost("eudore.com"), eudore.NewClientCheckStatus(200), eudore.NewClientCheckBody("eudore.com,eudore.cn"))

	app.CancelFunc()
	app.Run()
}
