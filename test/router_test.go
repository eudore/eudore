package test

import (
	"fmt"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"testing"
)

func TestSubRouter(t *testing.T) {
	app := eudore.NewCore()
	app.AddMiddleware(echoHandle)
	app.AnyFunc("/*", echoHandle)

	api := app.Group("/api/* group:api2")
	api.AddMiddleware(argHandle("3", "true"))
	api.AnyFunc("/*", echoHandle)

	v := api.Group("/:name")
	v.AnyFunc("/:id", echoHandle)

	client := httptest.NewClient(app)
	client.NewRequest("HEAD", "/api/22").Do()
	client.NewRequest("GET", "/api").Do()
	client.NewRequest("POST", "/api/name/22").Do()
	client.NewRequest("GET", "/api/name/22/info").Do()
}

func argHandle(key, val string) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		ctx.AddParam(key, val)
	}
}

func echoHandle(ctx eudore.Context) {
	fmt.Println(ctx.GetParam("Route"), ctx.Path())
}
