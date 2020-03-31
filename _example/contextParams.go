package main

/*
Params来源与自行设置和路由设置

Params相关方法定义。
type Context interface {
	Params() Params
	GetParam(string) string
	SetParam(string, string)
	AddParam(string, string)
	...
}

type Params interface {
	GetParam(string) string
	AddParam(string, string)
	SetParam(string, string)
}
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewCore()
	httptest.NewClient(app).Stop(0)
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.Debug("route:", ctx.GetParam("route"))
	})
	app.Run()
}
