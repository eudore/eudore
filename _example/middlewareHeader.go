package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
	"net/http"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.AddMiddleware(middleware.NewHeaderWithSecureFunc(http.Header{
		"Cache-Control": []string{"no-cache"},
	}))
	app.AddMiddleware(middleware.NewHeaderFunc(nil))
	app.AnyFunc("/*", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/1").Do().Out()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
