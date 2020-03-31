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
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteJSON(map[string]interface{}{
			"name":    "eudore",
			"message": "hello eudore",
		})
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/").Do().Out()

	app.Run()
}
