package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewEudore()
	httptest.NewClient(app).Stop(0)
	app.RegisterInit("init-router", 0x015, func(app *eudore.Eudore) error {
		// 添加静态文件处理
		app.AddStatic("/js/*", "static")

		// WriteFile 调用http.ServeFile实现，可以额外添加etag计算等逻辑，文件路径拼接需要注意清理。
		app.GetFunc("/css/*", func(ctx eudore.Context) {
			ctx.WriteFile("static" + ctx.Path())
		})

		app.GetFunc("/*", func(ctx eudore.Context) {
			ctx.WriteString("hello eudore")
		})
		return nil
	})
	app.RegisterInit("init-listen", 0x016, func(app *eudore.Eudore) error {
		return app.Listen(":8088")
	})
	app.Run()
}
