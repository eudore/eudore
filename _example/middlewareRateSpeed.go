package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewRateSpeedFunc(16*1024, 64*1024, app.Context))
	app.PostFunc("/post", func(ctx eudore.Context) {
		ctx.Debug(string(ctx.Body()))
	})
	app.AnyFunc("/srv", func(ctx eudore.Context) {
		ctx.WriteString("rate speed 16kB")
	})
	app.AnyFunc("/*", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("POST", "/post").WithBody("return body").Do().CheckStatus(200)
	client.NewRequest("PUT", "/srv").Do().CheckStatus(200)

	app.Listen(":8088")
	app.Run()

}
