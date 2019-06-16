package main

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewCore()
	// session数据保存到cache中
	app.RegisterComponent("session-cache", &eudore.SessionCacheConfig{
		Cache:	app.Cache,
	})


	app.GetFunc("/set", func(ctx eudore.Context){
		// 读取会话
		sess := ctx.GetSession()
		sess.Set("key1", 1)

		// 保存会话数据
		ctx.SetSession(sess)
	})
	app.GetFunc("/get", func(ctx eudore.Context){
		sess := ctx.GetSession()
		ctx.Debugf("session get key1 value: %v", sess.Get("key1"))
	})

	app.Listen(":8088")
	app.Run()
}