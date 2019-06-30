package main

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewEudore()
	app.GetFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("server daemon")
	})

	// app.Listen(":8088")
	app.Run()
}
