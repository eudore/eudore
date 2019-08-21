package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/command"
)

func main() {
	app := eudore.NewEudore()
	app.GetFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("server daemon")
	})

	app.RegisterInit("eudore-command", 0x00a, command.InitCommand)
	app.RegisterInit("init-listen", 0x016, func(app *eudore.Eudore) error {
		app.Listen(":8088")
		return nil
	})
	app.Run()
}

// go build -o server
// ./server --command=daemon
