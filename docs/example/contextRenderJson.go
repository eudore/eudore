package main

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewCore()
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteJSON(map[string]interface{}{
			"name":    "eudore",
			"message": "hello eudore",
		})
	})
	app.Listen(":8088")
	app.Run()
}
