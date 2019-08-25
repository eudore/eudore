package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/session"
)

func main() {
	// 创建session，并注册转换函数。
	provider := session.NewSessionMap()
	eudore.RegisterHandlerFunc(func(fn func(session.ContextSession)) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			fn(session.ContextSession{
				Context: ctx,
				Session: provider,
			})
		}
	})

	app := eudore.NewCore()
	app.GetFunc("/set", func(ctx session.ContextSession) {
		// 读取会话
		data := ctx.GetSession()
		data["key1"] = 22
		ctx.Debugf("session set key1 value: %v", data["key1"])

		// 保存会话数据
		ctx.SetSession(data)
	})
	app.GetFunc("/get", func(ctx session.ContextSession) {
		data := ctx.GetSession()
		ctx.Debugf("session get key1 value: %v", data["key1"])
	})

	app.Listen(":8085")
	app.Run()
}
