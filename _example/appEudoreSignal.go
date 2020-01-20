package main

/*
NewEudore函数定义了信号初始函数，默认处理终端端口(Ctrl+C)、重启(kill -12)、关闭(kill -15)三种信号。
app.RegisterInit("eudore-signal", 0x00c, InitSignal)

可以使用app.RegisterInit("eudore-signal", 0x000, nil)删除或者重新定义。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewEudore()
	httptest.NewClient(app).Stop(0)
	// 删除信号处理。
	app.RegisterInit("eudore-signal", 0x000, nil)
	app.RegisterInit("init-router", 0x015, func(app *eudore.Eudore) error {
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
