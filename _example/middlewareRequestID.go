package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
	// "github.com/google/uuid"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.AddMiddleware(middleware.NewRequestIDFunc(nil))
/*	app.AddMiddleware(middleware.NewRequestIDFunc(func() string {
		return uuid.New().String()
	}))	*/
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Debug("hello 世界")
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/").Do()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
