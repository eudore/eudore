package main

/*
实现参考eudore.ConfigParseRead和eudore.ConfigParseConfig内容
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"time"
)

func main() {
	go func() {
		app := eudore.NewCore()
		app.AnyFunc("/*", map[string]interface{}{
			"route": "/*",
			"name":  "eudore",
		})
		app.Listen(":8089")
		app.Run()
	}()
	time.Sleep(100 * time.Millisecond)
	app := eudore.NewCore()
	httptest.NewClient(app).Stop(0)
	app.Set("keys.config", "http://127.0.0.1:8089/xxx")
	app.Set("keys.help", true)
	app.Listen(":8088")
	app.Run()
}
