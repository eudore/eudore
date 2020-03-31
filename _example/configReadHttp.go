package main

/*
实现参考eudore.ConfigParseRead和eudore.ConfigParseConfig内容
*/

import (
	"github.com/eudore/eudore"
	"time"
)

func main() {
	go func() {
		app := eudore.NewCore()
		app.AnyFunc("/*", func(ctx eudore.Context) {
			ctx.WriteJSON(map[string]interface{}{
				"route": "/*",
				"name":  "eudore",
			})
		})
		app.Listen(":8089")
		app.Run()
	}()
	time.Sleep(100 * time.Millisecond)

	app := eudore.NewCore()
	app.Set("keys.config", []string{"http://127.0.0.1:8087/xxx", "http://127.0.0.1:8089/xxx"})
	app.Set("keys.help", true)
	err := app.Parse()
	if err != nil {
		panic(err)
	}
	app.Run()
}
