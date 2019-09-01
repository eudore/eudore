package main

/*
command包可以通过Daemon()函数后台启动程序，也可以通过命令解析启动程序。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/command"
)

func main() {
	command.Daemon()

	app := eudore.NewEudore()
	app.GetFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("server daemon")
	})

	app.RegisterInit("init-listen", 0x016, func(app *eudore.Eudore) error {
		app.Listen(":8088")
		return nil
	})
	app.Run()
}
