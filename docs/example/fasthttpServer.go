package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/server/fasthttp"
)

func main() {
	app := eudore.NewCore()
	app.Server = fasthttp.NewServer(nil)
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("start fasthttp server, this default page.\n")
		ctx.WriteString("your path is " + ctx.Path())
	})
	app.Listen(":8084")
	app.Run()
}
