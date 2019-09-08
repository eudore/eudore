package main

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewCore()
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.Redirect(302, "/hello")
	})
	app.GetFunc("/hello", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore")
	})
	app.Listen(":8088")
	app.Run()
}
