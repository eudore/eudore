package main

/*
RouterStd是eudore的默认路由器，使用基数树算法独立实现，性能与httprouter相识。

具有路由匹配优先级： 常量匹配 > 变量校验匹配 >变量匹配 > 通配符校验匹配 > 通配符匹配
方法优先级： 具体方法 > Any方法

用法：在正常变量和通配符后，使用'|'符号分割，后为校验规则，num是校验函数；{min:100}为动态检验函数，min是动态校验函数名称，':'后为参数；如果为'^'开头为正则校验,并且要使用'$'作为结尾。

在路径中使用'{}'包裹的一段字符串为块模式，切分时将整块紧跟上一个字符串，这样允许在校验规则内使用任何字符,
字符空格、冒号、星号、前花括号、后花括号、斜杠均为特殊符号（' '、':'、'*'、'{'、'}'、'/'），一定需要使用块模式包裹字符串。

可以使用app.AddHandler("TEST","/api/v:v/user/*name", eudore.HandlerEmpty)查看debug信息。
例如路径切割的切片，首字符为':'是变量匹配，首字符为'*'是通配符匹配，其他都是常量字符串匹配。
变量匹配从当前到下一个斜杠('/')处或结尾，通配符匹配当前位置到结尾，常量匹配对应的字符串。
```
:num|num
:num|{min:100}
:num|{^0.*$}
*num|num
*num|{min:100}
*num|{^0.*$}
```
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app))
	// 添加路由中间件后 必须重新注册404 405，使404 405具有中间件效果
	app.AddHandler("404", "", eudore.HandlerRouter404)
	app.AddHandler("405", "", eudore.HandlerRouter405)

	// ---------- 基础路由 ----------
	app.AnyFunc("/eudore/debug/meta/:name", middleware.NewMetadataFunc(app))
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

	// ---------- 校验路由 ----------
	// 正则校验，相当于 regexp:{^0.*$}，是一个动态校验函数。
	app.GetFunc("/get/:num|{^0.*$}", func(ctx eudore.Context) {
		ctx.WriteString("first char is '0', num is: " + ctx.GetParam("num") + "\n")
	})
	// 动态校验函数，min闭包生成校验函数。
	app.GetFunc("/get/:num|{min:100}", func(ctx eudore.Context) {
		ctx.WriteString("num great 100, num is: " + ctx.GetParam("num") + "\n")
	})
	// 校验函数，使用校验函数num。
	app.GetFunc("/get/:num|num", func(ctx eudore.Context) {
		ctx.WriteString("num num is: " + ctx.GetParam("num") + "\n")
	})

	// 通配符研究不写了，和变量校验相同。
	app.GetFunc("/*path|{^0.*$}", func(ctx eudore.Context) {
		ctx.WriteString("get path first char is '0', path is: " + ctx.GetParam("path") + "\n")
	})
	app.GetFunc("/:path|enum=1,2,3", eudore.HandlerRouter404)

	// ---------- 路由Debug ----------
	api := app.Group("/api/{v 1} version=v1")
	app.AddHandler("TEST", "/:path|{^0.*$}/*path|{^0.*$}", eudore.HandlerEmpty)
	api.AddHandler("test", "/get/:name action=GetName", eudore.HandlerEmpty)
	api.AddHandler("test", "/get/{{}} action=GetName", eudore.HandlerEmpty)
	app.AddHandler("TEST", "/api/v:v/user/*name", eudore.HandlerEmpty)

	// ---------- 分割线 运行测试请求 ----------

	// 测试
	status := eudore.NewClientCheckStatus
	body := eudore.NewClientCheckBody
	app.NewRequest("GET", "/get", status(200), body("get"))
	app.NewRequest("GET", "/get/ha", status(200), body("get name"))
	app.NewRequest("GET", "/get/eudore", status(200), body("get eudore"))
	app.NewRequest("PUT", "/get/eudore", status(405))

	app.NewRequest("GET", "/get/2", status(200), body("num"))
	app.NewRequest("GET", "/get/22", status(200), body("num"))
	app.NewRequest("GET", "/get/222", status(200), body("num great 100"))
	app.NewRequest("GET", "/get/0xx", status(200), body("first char is '0'"))
	app.NewRequest("XXX", "/get/0xx", status(405))

	app.Listen(":8088")
	app.Run()
}
