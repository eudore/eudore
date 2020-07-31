package main

/*
注册的扩展转换函数实现Context处理对象转换，通过转换后对象实现扩展。
*/

import (
	"fmt"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewApp()
	app.AddHandlerExtend("", func(i interface{}) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			ctx.WriteString("我是全局扩展 " + fmt.Sprint(i))
		}
	})
	app.GetFunc("/*", 1)

	// Group创建新的Router拥有独立的处理扩展，有效使用。 链式逆向匹配类型
	api := app.Group("/api")
	api.AddHandlerExtend(func(i interface{}) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			ctx.WriteString("我是api扩展 " + fmt.Sprint(i))
		}
	})
	api.GetFunc("/*", 2)

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/file/").Do().Out()
	client.NewRequest("GET", "/api/file").Do().Out()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
