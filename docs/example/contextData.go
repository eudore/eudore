package main

/*
ContextData额外增加了数据类型转换方法。

访问 /?id=333

*/

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewCore()
	app.AnyFunc("/*", func(ctx eudore.ContextData) {
		var id int = ctx.GetQueryInt("id")
		ctx.WriteString("hello eudore core")
		ctx.Infof("id is %d", id)
	})
	app.Listen(":8088")
	app.Run()
}
