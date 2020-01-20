package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewCore()
	httptest.NewClient(app).Stop(0)
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.Redirect(302, "/hello")
	})
	app.GetFunc("/hello", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore")
	})
	app.Listen(":8088")
	app.Run()
}
