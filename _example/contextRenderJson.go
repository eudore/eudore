package main

/*
WriteJSON方法直接返回json对象格式，不会按照render一样返回数据。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewApp()
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteJSON(map[string]interface{}{
			"name":    "eudore",
			"message": "hello eudore",
		})
	})
	app.GetFunc("/json2", func(ctx eudore.Context) interface{} {
		return map[string]interface{}{
			"name":    "eudore",
			"message": "hello eudore",
		}
	})
	app.GetFunc("/json1", func(ctx eudore.Context) interface{} {
		return "hello eudore"
	})

	client := httptest.NewClient(app).AddHeaderValue(eudore.HeaderAccept, eudore.MimeApplicationJSON)
	client.NewRequest("GET", "/").Do().OutBody()
	client.NewRequest("GET", "/json1").Do().OutBody()
	client.NewRequest("GET", "/json2").Do().OutBody()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
