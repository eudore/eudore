package main

/*
ContextData额外增加了数据类型转换方法。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewCore()
	app.AnyFunc("/*", func(ctx eudore.ContextData) {
		var id int = ctx.GetQueryInt("id")
		ctx.WriteString("hello eudore core")
		ctx.Infof("id is %d", id)
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/?id=333").Do().Out()
	for client.Next() {
		app.Error(client.Error())
	}
	client.Stop(0)

	app.Listen(":8088")
	app.Run()
}
