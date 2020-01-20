package main

/*
WriteJSON方法直接返回json对象格式，不会按照render一样返回数据。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewCore()
	httptest.NewClient(app).Stop(0)
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteJSON(map[string]interface{}{
			"name":    "eudore",
			"message": "hello eudore",
		})
	})
	app.Listen(":8088")
	app.Run()
}
