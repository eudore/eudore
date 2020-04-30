package main

/*
Core是对eudore.App对象的简单封装，实现Listen和Run。
*/

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewApp()
	app.Set("workdir", ".")
	app.Options(app.Parse())
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore core")
	})
	app.AnyFunc("/data", func(ctx eudore.Context) interface{} {
		return map[string]interface{}{
			"aa": 11,
			"bb": 22,
		}
	})
	app.Listen(":8088")
	app.CancelFunc()
	app.Run()
}
