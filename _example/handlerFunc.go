package main

/*
注册的扩展转换函数实现Context处理对象转换，通过转换后对象实现扩展。

func(MyContext, string) -> eudore.HandlerFunc
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewApp()
	app.AddHandlerExtend(func(fn func(eudore.Context, string)) eudore.HandlerFunc {
		name := "eudore"
		return func(ctx eudore.Context) {
			fn(ctx, name)
		}
	})
	app.GetFunc("/*", func(ctx eudore.Context, name string) {
		ctx.WriteString("name is " + name)
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/file/").Do().Out()
	for client.Next() {
		app.Error(client.Error())
	}

	app.CancelFunc()
	app.Run()
}
