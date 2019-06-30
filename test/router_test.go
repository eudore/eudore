package test

import (
	"fmt"
	"github.com/eudore/eudore"
	"testing"
)

func TestSubRouter(t *testing.T) {
	e := eudore.NewCore()
	e.RegisterMiddleware("", "", eudore.HandlerFuncs{echoHandle})
	e.RegisterMiddleware("GET", "", eudore.HandlerFuncs{echoHandle})
	e.AnyFunc("/*", echoHandle)

	api := e.Group("/api/* group:api2")
	api.AddMiddleware(argHandle("3", "true"))
	api.AnyFunc("/*", echoHandle)

	v := api.Group("/:name")
	v.AnyFunc("/:id", echoHandle)

	eudore.TestHttpHandler(e, "HEAD", "/api/22")
	eudore.TestHttpHandler(e, "GET", "/api")
	eudore.TestHttpHandler(e, "POST", "/api/name/22")
	eudore.TestHttpHandler(e, "GET", "/api/name/22/info")

}
func argHandle(key, val string) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		ctx.AddParam(key, val)
	}
}

func echoHandle(ctx eudore.Context) {
	fmt.Println(ctx.GetParam("Route"), ctx.Path())
}

func TestRouterEmpty(t *testing.T) {
	app := eudore.NewCore()
	app.RegisterComponent(eudore.ComponentRouterEmptyName, eudore.HandlerFunc(func(ctx eudore.Context) {
		ctx.WriteString(app.Router.Version())
		t.Log(app.Router.Version())
	}))
	eudore.TestHttpHandler(app, "GET", "/api/name/22/info")
}
