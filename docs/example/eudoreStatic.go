package main

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewEudore()
	app.RegisterInit("init-router", 0x015, func(app *eudore.Eudore) error {
		// 添加静态文件处理
		app.AddStatic("/js/*", "static")

		// WriteFile 与http.ServeFile效果相同，可以额外添加etag计算等逻辑。
		app.GetFunc("/css/*", func(ctx eudore.Context) {
			ctx.WriteFile("static" + ctx.Path())
		})

		app.GetFunc("/*", func(ctx eudore.Context) {
			ctx.WriteString("hello eudore")
		})
		return nil
	})
	app.RegisterInit("init-listen", 0x016, func(app *eudore.Eudore) error {
		app.Listen(":8088")
		return nil
	})
	app.Run()
}
