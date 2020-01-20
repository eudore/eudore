package main

/*
Context接口中日志相关方法。
type Context interface {
	...
	Debug(...interface{})
	Info(...interface{})
	Warning(...interface{})
	Error(...interface{})
	Fatal(...interface{})
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warningf(string, ...interface{})
	Errorf(string, ...interface{})
	Fatalf(string, ...interface{})
	WithField(key string, value interface{}) Logout
	WithFields(fields Fields) Logout
	Logger() Logout
}

ctx.Fatal会返回状态码500和err的内容。
例如： {"error":"fatal logger method: GET path: /err","status":"500","x-request-id":""}
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewCore()
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Info("hello")
		ctx.Infof("hello path is %s", ctx.GetParam("*"))
	})
	app.AnyFunc("/err", func(ctx eudore.Context) {
		ctx.Fatalf("fatal logger method: %s path: %s", ctx.Method(), ctx.Path())
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/ffile").WithHeaderValue(eudore.HeaderAccept, eudore.MimeApplicationJSON).Do().CheckStatus(200).Out()
	client.NewRequest("GET", "/err").WithHeaderValue(eudore.HeaderAccept, eudore.MimeApplicationJSON).Do().CheckStatus(200).Out()
	for client.Next() {
		app.Error(client.Error())
	}
	client.Stop(0)

	app.Listen(":8088")
	app.Run()
}
