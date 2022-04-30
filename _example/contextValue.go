package main

/*
通过context.Context接口在不同的处理函数直接有状态传递数据。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(func(ctx eudore.Context) {
		ctx.SetValue("val", "this is val")
	})
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Debugf("val: %s", ctx.Value("val"))
	})

	client := httptest.NewClient(app)
	client.NewRequest("PUT", "/fl").Do().CheckStatus(200)

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
