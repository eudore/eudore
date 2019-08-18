package test

import (
	"fmt"
	"github.com/eudore/eudore"
	"testing"
)

func TestSubRouter(t *testing.T) {
	app := eudore.NewCore()
	app.RegisterMiddleware("", "", eudore.HandlerFuncs{echoHandle})
	app.RegisterMiddleware("GET", "", eudore.HandlerFuncs{echoHandle})
	app.AnyFunc("/*", echoHandle)

	api := app.Group("/api/* group:api2")
	api.AddMiddleware("ANY", "", argHandle("3", "true"))
	api.AnyFunc("/*", echoHandle)

	v := api.Group("/:name")
	v.AnyFunc("/:id", echoHandle)

	eudore.TestAppRequest(app, "HEAD", "/api/22", nil)
	eudore.TestAppRequest(app, "GET", "/api", nil)
	eudore.TestAppRequest(app, "POST", "/api/name/22", nil)
	eudore.TestAppRequest(app, "GET", "/api/name/22/info", nil)

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
	/*	app.RegisterComponent(eudore.ComponentRouterEmptyName, eudore.HandlerFunc(func(ctx eudore.Context) {
		ctx.WriteString(app.Router.Version())
		t.Log(app.Router.Version())
	}))*/
	eudore.TestAppRequest(app, "GET", "/api/name/22/info", nil)
}
