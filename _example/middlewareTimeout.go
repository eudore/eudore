package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
	"time"
)

func main() {
	app := eudore.NewCore()
	app.AddMiddleware(middleware.NewTimeoutFunc(3 * time.Second / 10))
	app.AnyFunc("/*", func(ctx eudore.ContextData) {
		time.Sleep(time.Duration(ctx.GetParamInt64("*")) * time.Second / 10)
		ctx.WriteString("hellp")
	})

	client := httptest.NewClient(app)
	client.NewRequest("PUT", "/1").Do().CheckStatus(200)
	client.NewRequest("PUT", "/2").Do().CheckStatus(200)
	client.NewRequest("PUT", "/3").Do().CheckStatus(200)
	client.NewRequest("PUT", "/4").Do().CheckStatus(200)
	for client.Next() {
		app.Error(client.Error())
	}
	client.Stop(0)

	app.Listen(":8088")
	app.Run()
}
