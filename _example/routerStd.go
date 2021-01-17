package main

/*
RouterStd是eudore的默认路由器，使用基数树算法独立实现，性能与httprouter相识。

具有路由匹配优先级： 常量匹配 > 变量校验匹配 >变量匹配 > 通配符校验匹配 > 通配符匹配
方法优先级： 具体方法 > Any方法

用法：在正常变量和通配符后，使用'|'符号分割，后为校验规则，isnum是校验函数；{min:100}为动态检验函数，min是动态校验函数名称，':'后为参数；如果为'^'开头为正则校验,并且要使用'$'作为结尾。

在路径中使用'{}'包裹的一段字符串为块模式，切分时将整块紧跟上一个字符串，这样允许在校验规则内使用任何字符,
字符空格、冒号、星号、前花括号、后花括号、斜杠均为特殊符号（' '、':'、'*'、'{'、'}'、'/'），一定需要使用块模式包裹字符串。

可以使用app.AddHandler("TEST","/api/v:v/user/*name", eudore.HandlerEmpty)查看debug信息。
例如路径切割的切片，首字符为':'是变量匹配，首字符为'*'是通配符匹配，其他都是常量字符串匹配。
变量匹配从当前到下一个斜杠('/')处或结尾，通配符匹配当前位置到结尾，常量匹配对应的字符串。
```
:num|isnum
:num|{min:100}
:num|{^0.*$}
*num|isnum
*num|{min:100}
*num|{^0.*$}
```
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	// 默认路由器就是 NewRouterStd(nil)
	app := eudore.NewApp()
	app.AnyFunc("/eudore/debug/look/*", middleware.NewLookFunc(app.Router))

	app.AddMiddleware(func(ctx eudore.Context) {
		ctx.WriteString("route: " + ctx.GetParam("route") + "\n")
	})
	app.GetFunc("/get/:name", func(ctx eudore.Context) {
		ctx.WriteString("get name: " + ctx.GetParam("name") + "\n")
	})
	app.GetFunc("/get/eudore", func(ctx eudore.Context) {
		ctx.WriteString("get eudore\n")
	})
	app.GetFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString("get path: /" + ctx.GetParam("path") + "\n")
	})
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString("any path: /" + ctx.GetParam("path") + "\n")
	})
	app.AddHandler("404", "", eudore.HandlerRouter404)
	app.AddHandler("405", "", eudore.HandlerRouter405)

	// ---------- 分割线 -----上面是基础路由 下面注册校验路由 ----------

	// 正则校验，相当于 regexp:{^0.*$}，是一个动态校验函数。
	app.GetFunc("/get/:num|{^0.*$}", func(ctx eudore.Context) {
		ctx.WriteString("first char is '0', num is: " + ctx.GetParam("num") + "\n")
	})
	// 动态校验函数，min闭包生成校验函数。
	app.GetFunc("/get/:num|{min:100}", func(ctx eudore.Context) {
		ctx.WriteString("num great 100, num is: " + ctx.GetParam("num") + "\n")
	})
	// 校验函数，使用校验函数isnum。
	app.GetFunc("/get/:num|isnum", func(ctx eudore.Context) {
		ctx.WriteString("isnum num is: " + ctx.GetParam("num") + "\n")
	})

	// 通配符研究不写了，和变量校验相同。
	app.GetFunc("/*path|{^0.*$}", func(ctx eudore.Context) {
		ctx.WriteString("get path first char is '0', path is: " + ctx.GetParam("path") + "\n")
	})

	app.GetFunc("/:path|haha", eudore.HandlerRouter404)
	// ---------- 分割线 运行测试请求 ----------

	// 测试
	client := httptest.NewClient(app)
	client.NewRequest("GET", "/get").Do().CheckStatus(200).CheckBodyContainString("get")
	client.NewRequest("GET", "/get/ha").Do().CheckStatus(200).CheckBodyContainString("/get/:name")
	client.NewRequest("GET", "/get/eudore").Do().CheckStatus(200).CheckBodyContainString("/get/eudore")
	client.NewRequest("PUT", "/get/eudore").Do().CheckStatus(405)

	client.NewRequest("GET", "/get/2").Do().CheckStatus(200).CheckBodyContainString("isnum")
	client.NewRequest("GET", "/get/22").Do().CheckStatus(200).CheckBodyContainString("isnum")
	client.NewRequest("GET", "/get/222").Do().CheckStatus(200).CheckBodyContainString("num great 100", "222")
	client.NewRequest("GET", "/get/0xx").Do().CheckStatus(200).CheckBodyContainString("first char is '0'", "0xx")
	client.NewRequest("XXX", "/get/0xx").Do().CheckStatus(405).Out()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
