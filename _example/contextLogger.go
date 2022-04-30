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
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyLogger, eudore.NewLoggerStd(map[string]interface{}{"FileLine": true}))
	app.AddMiddleware(func(ctx eudore.Context) {
		ctx.Request().Header.Add(eudore.HeaderXRequestID, "requestid")
	})
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Info("hello")
		ctx.Infof("hello path is %s", ctx.GetParam("*"))
		ctx.Warning("warning")
		ctx.Warningf("warningf")
		ctx.Error(nil)
		ctx.Error("test error")
		ctx.Errorf("test error")
		ctx.Fatal(nil)
	})
	app.AnyFunc("/err", func(ctx eudore.Context) {
		ctx.Fatalf("fatal logger method: %s path: %s", ctx.Method(), ctx.Path())
		ctx.Debug("err:", ctx.Err())
	})
	app.AnyFunc("/field", func(ctx eudore.Context) {
		ctx.WithFields([]string{"key", "name"}, []interface{}{"ctx.WithFields", "eudore"}).Debug("hello fields")
		ctx.WithField("logger", true).Debug("hello empty fields")
		ctx.WithField("key", "test-firle").Debug("debug")
		ctx.WithField("key", "test-firle").Debugf("debugf")
		ctx.WithField("key", "test-firle").Info("hello")
		ctx.WithField("key", "test-firle").Infof("hello path is %s", ctx.GetParam("*"))
		ctx.WithField("key", "test-firle").Warning("warning")
		ctx.WithField("key", "test-firle").Warningf("warningf")
		ctx.WithField("key", "test-firle").Error(nil)
		ctx.WithField("key", "test-firle").Errorf("test error")
		ctx.WithField("key", "test-firle").Fatal(nil)
		ctx.WithField("key", "test-firle").WithField("hello", "haha").Fatalf("fatal logger method: %s path: %s", ctx.Method(), ctx.Path())
		ctx.WithField("method", "WithField").WithFields([]string{"key", "name"}, []interface{}{"ss", "eudore"}).Debug("hello fields")
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/ffile").WithHeaderValue(eudore.HeaderAccept, eudore.MimeApplicationJSON).Do().CheckStatus(200).Out()
	client.NewRequest("GET", "/err").WithHeaderValue(eudore.HeaderAccept, eudore.MimeApplicationJSON).Do().CheckStatus(200).Out()
	client.NewRequest("GET", "/field").WithHeaderValue(eudore.HeaderAccept, eudore.MimeApplicationJSON).Do().CheckStatus(200).Out()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
