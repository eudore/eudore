package main

/*
RouterRadix是eudore默认路由，使用基数数算法实现。

具有路由匹配优先级： 常量匹配 > 变量匹配 > 通配符匹配
方法优先级： 具体方法 > Any方法

测试命令：
curl -XGET localhost:8088/get
curl -XGET localhost:8088/get/ha
curl -XGET localhost:8088/get/eudore
curl -XPUT localhost:8088/get/eudore

*/
import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewCore()
	app.AddMiddleware("ANY", "", func(ctx eudore.Context) {
		ctx.WriteString("route: " + ctx.GetParam("route") + "\n")
	})
	app.GetFunc("/get/:name", func(ctx eudore.Context) {
		ctx.WriteString("get name: " + ctx.GetParam("name") + "\n")
	})
	// /get/eudore是常量匹配优先于/get/:name
	app.GetFunc("/get/eudore", func(ctx eudore.Context) {
		ctx.WriteString("get eudore\n")
	})
	app.GetFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString("get path: /" + ctx.GetParam("path") + "\n")
	})
	// Any方法不会覆盖Get方法。
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString("any path: /" + ctx.GetParam("path") + "\n")
	})
	// 启动server
	app.Listen(":8088")
	app.Run()
}
