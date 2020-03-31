package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewCore()
	app.AddMiddleware(middleware.NewLoggerFunc(app.App, "route"))
	app.AddMiddleware(middleware.NewRecoverFunc())
	app.AnyFunc("/*", func(eudore.Context) {
		panic("test error")
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/").Do()

	app.Run()
}
