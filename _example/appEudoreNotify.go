package main

/*
实现参考notify.Init
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/component/notify"
)

func main() {
	app := eudore.NewEudore()
	httptest.NewClient(app).Stop(0)
	// 设置编译命令、启动命令、监听目录
	app.Config.Set("component.notify.buildcmd", "go build -o server appEudoreNotify.go")
	app.Config.Set("component.notify.startcmd", "./server")
	app.Config.Set("component.notify.watchdir", ".")

	app.RegisterInit("init-notify", 0x015, notify.Init)
	app.RegisterInit("init-listen", 0x016, func(app *eudore.Eudore) error {
		app.GetFunc("/*", func(ctx eudore.Context) {
			ctx.WriteString("server notify")
		})
		return app.Listen(":8088")
	})
	app.Run()
}
