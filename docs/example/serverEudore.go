package main

/*
具有Response Header Server: eudore
不推荐使用该Server，不成熟可能有未知bug。
*/

import (
	"github.com/eudore/eudore"
	eserver "github.com/eudore/eudore/component/server/eudore"
)

func main() {
	app := eudore.NewCore()
	app.Server = eserver.NewServer()
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("start fasthttp server, this default page.\n")
		ctx.WriteString("your path is " + ctx.Path())
	})
	app.Listen(":8088")
	app.Run()
}
