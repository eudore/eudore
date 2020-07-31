package main

/*
ContextData额外增加了数据类型转换方法。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewApp()
	app.AnyFunc("/*", func(ctx eudore.ContextData) {
		var id int = ctx.GetQueryInt("id")
		ctx.WriteString("hello eudore core")
		ctx.Infof("id is %d", id)
	})
	app.GetFunc("/params/:key", func(ctx eudore.ContextData) {
		ctx.Debugf("bool: %#v", ctx.GetParamBool("key"))
		ctx.Debugf("int: %#v", ctx.GetParamInt("key"))
		ctx.Debugf("int: %#v", ctx.GetParamInt("key", -101))
		ctx.Debugf("int64: %#v", ctx.GetParamInt64("key"))
		ctx.Debugf("int64: %#v", ctx.GetParamInt64("key", -164))
		ctx.Debugf("float32: %#v", ctx.GetParamFloat32("key"))
		ctx.Debugf("float32: %#v", ctx.GetParamFloat32("key", -132))
		ctx.Debugf("float64: %#v", ctx.GetParamFloat64("key"))
		ctx.Debugf("float64: %#v", ctx.GetParamFloat64("key", 0))
		ctx.Debugf("string: %#v", ctx.GetParamString("keysss", "default string"))
	})
	app.GetFunc("/header", func(ctx eudore.ContextData) {
		ctx.Debugf("bool: %#v", ctx.GetHeaderBool("key"))
		ctx.Debugf("int: %#v", ctx.GetHeaderInt("key"))
		ctx.Debugf("int: %#v", ctx.GetHeaderInt("key", -101))
		ctx.Debugf("int64: %#v", ctx.GetHeaderInt64("key"))
		ctx.Debugf("int64: %#v", ctx.GetHeaderInt64("key", -164))
		ctx.Debugf("float32: %#v", ctx.GetHeaderFloat32("key"))
		ctx.Debugf("float32: %#v", ctx.GetHeaderFloat32("key", -132))
		ctx.Debugf("float64: %#v", ctx.GetHeaderFloat64("key"))
		ctx.Debugf("float64: %#v", ctx.GetHeaderFloat64("key", 0))
		ctx.Debugf("string: %#v", ctx.GetHeaderString("keysss", "default string"))
	})
	app.GetFunc("/query", func(ctx eudore.ContextData) {
		ctx.Debugf("bool: %#v", ctx.GetQueryBool("key"))
		ctx.Debugf("int: %#v", ctx.GetQueryInt("key"))
		ctx.Debugf("int: %#v", ctx.GetQueryInt("key", -101))
		ctx.Debugf("int64: %#v", ctx.GetQueryInt64("key"))
		ctx.Debugf("int64: %#v", ctx.GetQueryInt64("key", -164))
		ctx.Debugf("float32: %#v", ctx.GetQueryFloat32("key"))
		ctx.Debugf("float32: %#v", ctx.GetQueryFloat32("key", -132))
		ctx.Debugf("float64: %#v", ctx.GetQueryFloat64("key"))
		ctx.Debugf("float64: %#v", ctx.GetQueryFloat64("key", 0))
		ctx.Debugf("string: %#v", ctx.GetQueryString("keysss", "default string"))
	})
	app.GetFunc("/cookie", func(ctx eudore.ContextData) {
		ctx.Debugf("bool: %#v", ctx.GetCookieBool("key"))
		ctx.Debugf("int: %#v", ctx.GetCookieInt("key"))
		ctx.Debugf("int: %#v", ctx.GetCookieInt("key", -101))
		ctx.Debugf("int64: %#v", ctx.GetCookieInt64("key"))
		ctx.Debugf("int64: %#v", ctx.GetCookieInt64("key", -164))
		ctx.Debugf("float32: %#v", ctx.GetCookieFloat32("key"))
		ctx.Debugf("float32: %#v", ctx.GetCookieFloat32("key", -132))
		ctx.Debugf("float64: %#v", ctx.GetCookieFloat64("key"))
		ctx.Debugf("float64: %#v", ctx.GetCookieFloat64("key", 0))
		ctx.Debugf("string: %#v", ctx.GetCookieString("keysss", "default string"))
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/?id=333").Do().Out()
	client.NewRequest("GET", "/params/333").Do()
	client.NewRequest("GET", "/header").WithHeaderValue("key", "123").Do()
	client.NewRequest("GET", "/query?key=111").Do()
	client.NewRequest("GET", "/cookie").WithHeaderValue("Cookie", "key=1234").Do()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
