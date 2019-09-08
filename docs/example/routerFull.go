package main

/*
RouterRadix是eudore默认路由，使用基数数算法实现。

具有路由匹配优先级： 常量匹配 > 变量校验匹配 >变量匹配 > 通配符校验匹配 > 通配符匹配
方法优先级： 具体方法 > Any方法

用法：在正常变量和通配符后，使用'|'符号分割，后为校验规则，isnum是校验函数；min:100为动态检验函数，min是动态校验函数名称，':'后为参数；如果为'^'开头为正则校验,并且要使用'$'作为结尾。

**注意: 正则表达式不要使用空格，会导致参数切割错误。**

```
:num|isnum
:num|min:100
:num|^0.*$
*num|isnum
*num|min:100
*num|^0.*$
```


测试命令：

curl -XGET localhost:8088/get
curl -XGET localhost:8088/get/ha
curl -XGET localhost:8088/get/eudore
curl -XPUT localhost:8088/get/eudore

curl -XGET localhost:8088/get/2
curl -XGET localhost:8088/get/22
curl -XGET localhost:8088/get/222
curl -XGET localhost:8088/get/0xx
*/
import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewCore()

	// 修改路由
	app.Router = eudore.NewRouterFull()
	eudore.Set(app.Router, "print", eudore.NewLoggerPrintFunc(app.Logger))

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

	// ---------- 分割线 -----上面是routerRadix.go例子复制的路由 下面注册RouterFull路由 ----------

	// 正则校验，相当于 regexp:^0.*$，是一个动态校验函数。
	app.GetFunc("/get/:num|^0.*$", func(ctx eudore.Context) {
		ctx.WriteString("first char is '0', num is: " + ctx.GetParam("num") + "\n")
	})
	// 动态校验函数，min闭包生成校验函数。
	app.GetFunc("/get/:num|min:100", func(ctx eudore.Context) {
		ctx.WriteString("num great 100, num is: " + ctx.GetParam("num") + "\n")
	})
	// 校验函数，使用校验函数isnum。
	app.GetFunc("/get/:num|isnum", func(ctx eudore.Context) {
		ctx.WriteString("isnum num is: " + ctx.GetParam("num") + "\n")
	})

	// 通配符研究不写了，和变量校验相同。
	app.GetFunc("/*path|^0.*$", func(ctx eudore.Context) {
		ctx.WriteString("get path first char is '0', path is: " + ctx.GetParam("path") + "\n")
	})

	// ---------- 分割线 运行测试请求 ----------

	// 测试
	client := httptest.NewClient(app)
	client.NewRequest("GET", "/get").Do().CheckStatus(200).CheckBodyContainString("get", "/*path")
	client.NewRequest("GET", "/get/ha").Do().CheckStatus(200).CheckBodyContainString("/get/:name")
	client.NewRequest("GET", "/get/eudore").Do().CheckStatus(200).CheckBodyContainString("/get/eudore")
	client.NewRequest("PUT", "/get/eudore").Do().CheckStatus(200).CheckBodyContainString("any", "/*path")

	client.NewRequest("GET", "/get/2").Do().CheckStatus(200).CheckBodyContainString("isnum")
	client.NewRequest("GET", "/get/22").Do().CheckStatus(200).CheckBodyContainString("isnum")
	client.NewRequest("GET", "/get/222").Do().CheckStatus(200).CheckBodyContainString("num great 100", "222")
	client.NewRequest("GET", "/get/0xx").Do().CheckStatus(200).CheckBodyContainString("first char is '0'", "0xx")
	client.NewRequest("XXX", "/get/0xx").Do().CheckStatus(200).Out()
	for client.Next() {
		app.Error(client.Error())
	}

	// 启动server
	app.Listen(":8088")
	app.Run()
}
