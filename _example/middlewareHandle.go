package main

/*
eudore middleware是一个eudore.HandlerFunc类型的函数

Context中定义middleware方法
type Context interface {
	Next()
	End()
	...
}

eudore.Router.Match方法返回的数据类型是[]eudore.HandlerFunc即多个请求处理函数，前面的为中间件处理函数，最后一个或多个是注册的请求处理函数。
Next()方法调用下一个请求处理函数，再Next()方法定义的内容就是后置处理函数。
End()方法终止后续请求处理函数。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewCore()
	app.AddMiddleware(middleware.NewLoggerFunc(app.App, "route"))
	app.AddMiddleware(func(ctx eudore.Context) {
		ctx.WriteString("pre\n")
		ctx.Next()
		ctx.WriteString("\npost")
	})
	app.AnyFunc("/*", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("PUT", "/1").Do().Out()
	for client.Next() {
		app.Error(client.Error())
	}

	app.Run()
}
