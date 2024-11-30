package eudore_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	. "github.com/eudore/eudore"
)

type Test014Controller struct {
	ControllerAutoRoute
}

func (ctl *Test014Controller) Get(ctx Context) interface{} {
	return "Test014Controller"
}

func TestRouterStdAdd(t *testing.T) {
	type String struct {
		Data string
	}
	ctx := context.WithValue(context.Background(),
		ContextKeyHandlerExtender, DefaultHandlerExtender,
	)
	ctx = context.WithValue(ctx,
		ContextKeyLogger, DefaultLoggerNull,
	)

	r := NewRouter(nil)
	r.(interface{ Mount(context.Context) }).Mount(ctx)
	r.AddHandlerExtend(func(str String) HandlerFunc {
		return func(ctx Context) {
			ctx.WriteString(str.Data)
		}
	})
	r.Group(" loggerkind=all")
	r.Group(" loggerkind=~all")
	r.Group(" loggerkind=middleware").AddController(&Test014Controller{})

	api := r.Group("/api")
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
	api.AnyFunc("/allow", String{"405"})

	r.Match(MethodOptions, "/api/allow", &Params{"route", ""})
}

func TestRouterError(t *testing.T) {
	var h HandlerFunc
	r := NewRouter(nil)
	r.Group(" loggerkind=~handler|metadata").GetFunc("/", HandlerEmpty)
	r.AddHandlerExtend("/api", TestRouterError)
	r.AddController(&Test015Controller{})
	r.AddController(NewControllerError(&Test015Controller{}, fmt.Errorf("test controller error")))
	r.Group("/api").AddController(NewControllerError(&Test015Controller{}, fmt.Errorf("test controller error")))
	r.GetFunc("{/*}", HandlerEmpty)
	r.GetFunc("/*path|check", HandlerEmpty)
	r.GetFunc("/err", func(*testing.T) {})
	r.GetFunc("/nil", h)
}

type Test015Controller struct {
	ControllerAutoRoute
}

func (Test015Controller) String() string {
	return "015"
}

func TestRouterMiddleware(t *testing.T) {
	r := NewRouter(nil)
	r.AddMiddleware()
	r.AddMiddleware("/api", func(int) {})
	r.AddMiddleware(func(int) {})
	r.AddHandlerExtend()
	r.AddMiddleware("/api/v2", HandlerEmpty)
	r.AddMiddleware("/api/v1", HandlerEmpty)
	r.AddMiddleware("/api", HandlerEmpty)
	r.AnyFunc("/api/v1", HandlerEmpty)

	apiv1 := r.Group("/api/v1")
	apiv1.AddMiddleware(func(int) {})
	apiv1.AnyFunc("/users", HandlerEmpty)
}

func TestRouterCoreHost(t *testing.T) {
	echoHandleHost := func(ctx Context) {
		ctx.WriteString(ctx.GetParam("route-host"))
	}
	s := NewServer(nil)
	c := NewClient()
	r := NewRouter(NewRouterCoreHost(nil))
	get := NewContextBaseFunc(context.Background())

	s.SetHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := get()
		ctx.Reset(w, req)
		ctx.SetHandlers(-1, r.Match(ctx.Method(), ctx.Path(), ctx.Params()))
		ctx.Next()
	}))

	r.AnyFunc("/* route-host=*.eudore.cn", echoHandleHost)
	r.AnyFunc("/* route-host=www.example.*", echoHandleHost)
	r.AnyFunc("/* route-host=eudore.cn", echoHandleHost)
	r.AnyFunc("/* route-host=eudore.*", echoHandleHost)
	r.AnyFunc("/* route-host=eudore.*:8080", echoHandleHost)
	r.AddHandler("404", "", HandlerRouter404)
	r.(interface{ Metadata() interface{} }).Metadata()

	routes := []struct {
		path   string
		host   string
		status int
		body   string
	}{
		{"/", "www.eudore.net", 404, "404"},
		{"/", "eudore.cn", 200, "eudore.cn"},
		{"/", "eudore.cn2", 200, "eudore.*"},
		{"/", "eudore.com", 200, "eudore.*"},
		{"/", "eudore.com:80", 200, "eudore.*"},
		{"/", "godoc.eudore.cn", 200, "*.eudore.cn"},
		{"/", "www.example.cn", 200, "www.example.*"},
	}
	ctx := context.WithValue(context.Background(),
		ContextKeyServer, s,
	)
	for _, route := range routes {
		err := c.NewRequest("GET", route.path, ctx,
			NewClientOptionHost(route.host),
			NewClientCheckStatus(route.status),
			NewClientCheckBody(route.body),
		)
		if err != nil {
			t.Error(err)
		}
	}
	r.(interface{ Mount(context.Context) }).Mount(context.Background())
	r.(interface{ Unmount(context.Context) }).Unmount(context.Background())
}
