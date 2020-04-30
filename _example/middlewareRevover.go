package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.AddMiddleware(middleware.NewRecoverFunc())
	app.AnyFunc("/*", func(eudore.Context) {
		panic("test error")
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/").Do()

	app.CancelFunc()
	app.Run()
}
