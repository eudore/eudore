package main

/*
WriteJSON方法直接返回json对象格式，不会按照render一样返回数据。
如果设置默认Render为eudore.Renderer(eudore.RenderJSON)，那么所有Render都返回json。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	// 设置默认Render为JSON，所有Render数据均使用json返回
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyRender, eudore.RenderJSON)
	app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))

	app.AddMiddleware(middleware.NewRequestIDFunc(nil))
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.Render(map[string]interface{}{
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
	app.Run()
}
