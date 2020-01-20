package main

/*
Core是对eudore.App对象的简单封装，实现Listen和Run。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewCore()
	httptest.NewClient(app).Stop(0)
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore core")
	})
	app.Listen(":8088")
	app.Run()
}
