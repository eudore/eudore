package main

/*
RouterRadix是eudore默认路由，使用基数树算法实现。

具有路由匹配优先级： 常量匹配 > 变量匹配 > 通配符匹配
方法优先级： 具体方法 > Any方法

在路径中使用'{}'包裹的一段字符串为块模式，切分时将整块紧跟上一个字符串，这样允许在校验规则内使用任何字符,
字符空格、冒号、星号、前花括号、后花括号、斜杠均为特殊符号（' '、':'、'*'、'{'、'}'、'/'），一定需要使用块模式包裹字符串。

可以使用app.AddHandler("TEST","/api/v:v/user/*name", eudore.HandlerEmpty)查看debug信息。
例如路径切割的切片，首字符为':'是变量匹配，首字符为'*'是通配符匹配，其他都是常量字符串匹配。
变量匹配从当前到下一个斜杠('/')处或结尾，通配符匹配当前位置到结尾，常量匹配对应的字符串。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(func(ctx eudore.Context) {
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
	app.GetFunc("", func(ctx eudore.Context) {
		ctx.WriteString("root request: path is /")
	})
	app.AddHandler("404", "", eudore.HandlerRouter404)
	app.AddHandler("405", "", eudore.HandlerRouter405)

	// 请求测试
	client := httptest.NewClient(app)
	client.NewRequest("GET", "/get").Do().CheckStatus(200).CheckBodyContainString("get", "/*path")
	client.NewRequest("GET", "/get/ha").Do().CheckStatus(200).CheckBodyContainString("/get/:name")
	client.NewRequest("GET", "/get/eudore").Do().CheckStatus(200).CheckBodyContainString("/get/eudore")
	client.NewRequest("PUT", "/get/eudore").Do().CheckStatus(200).CheckBodyContainString("any", "/*path")

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
