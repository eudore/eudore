package main

/*
通过context.Context接口在不同的处理函数直接有状态传递数据。
*/

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(func(ctx eudore.Context) {
		ctx.SetValue("val", "this is val")
	})
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Debugf("val: %s", ctx.Value("val"))
	})

	app.NewRequest(app, "PUT", "/fl", eudore.NewClientCheckStatus(201))
	app.Listen(":8088")
	app.Run()
}
