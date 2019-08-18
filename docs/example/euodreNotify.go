package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/notify"
)

func main() {
	app := eudore.NewEudore()
	app.GetFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("server notify")
	})

	app.Config.Set("component.notify.buildcmd", "go build -o server coreNotify.go")
	app.Config.Set("component.notify.startcmd", "./server")
	app.Config.Set("component.notify.watchdir", ".")

	app.RegisterInit("init-notify", 0x015, notify.Init)
	app.RegisterInit("init-listen", 0x016, func(app *eudore.Eudore) error {
		app.Listen(":8088")
		return nil
	})
	app.Run()
}
