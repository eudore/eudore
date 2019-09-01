package main

/*
Core是对eudore.App对象的简单封装，实现不到百行。
*/

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewCore()
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore core")
	})
	app.Listen(":8088")
	app.Run()
}
