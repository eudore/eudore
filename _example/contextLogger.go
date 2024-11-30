package main

/*
Context接口中日志相关方法，与Logger接口方法几乎完全一致。
type Context interface {
	...
	Debug(...any)
	Info(...any)
	Warning(...any)
	Error(...any)
	Fatal(...any)
	Debugf(string, ...any)
	Infof(string, ...any)
	Warningf(string, ...any)
	Errorf(string, ...any)
	Fatalf(string, ...any)
	WithField(string, any) Logger
	WithFields([]string, []any) Logger
}

middleware.NewRequestIDFunc(nil) 会自动将X-Request-Id写入日志Field。
行为区别在于ctx.Fatal()方法会调用ctx.End()方法结束处理和返回错误内容数据。
例如：
{
	"time": "2024-06-03 10:12:51.245",
	"host": "",
	"method": "GET",
	"path": "/err",
	"route": "/err",
	"status": 500,
	"x-request-id": "1717380771245684456-05c419",
	"error": "fatal logger method: GET path: /err"
}
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyLogger, eudore.NewLogger(&eudore.LoggerConfig{
		Stdout:   true,
		StdColor: true,
		Caller:   true,
	}))
	app.AddMiddleware(middleware.NewRequestIDFunc(nil))
	app.AnyFunc("/log", func(ctx eudore.Context) {
		ctx.WithFields([]string{"key", "name"}, []interface{}{"ctx.WithFields", "eudore"}).Debug("hello fields")
		ctx.Infof("hello path is %s", ctx.GetParam("*"))
	})

	app.AnyFunc("/err", func(ctx eudore.Context) {
		ctx.Fatalf("fatal logger method: %s path: %s", ctx.Method(), ctx.Path())
		ctx.Debug("err:", ctx.Err())
	})

	app.Listen(":8088")
	app.Run()
}
