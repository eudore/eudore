package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/notify"
)

func main() {
	app := eudore.NewCore()

	app.Config.Set("component.notify.buildcmd", "go build -o server coreNotify.go")
	app.Config.Set("component.notify.startcmd", "./server")
	app.Config.Set("component.notify.watchdir", ".")
	notify.NewNotify(app.App).Run()

	app.Listen(":8088")
	app.Run()
}
