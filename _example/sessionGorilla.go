package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/component/session"
	"github.com/eudore/eudore/component/session/gorilla"
	gorillasession "github.com/gorilla/sessions"
)

func main() {
	// 创建session，并注册转换函数。
	app := eudore.NewApp()
	app.AddHandlerExtend(gorilla.NewExtendContextSession(gorillasession.NewCookieStore([]byte("sessionid"))))

	app.GetFunc("/set", func(ctx session.Context) {
		// 读取会话
		ctx.SetSession("key1", 22)
		ctx.Debugf("session set key1 value: %v", ctx.GetSession("key1"))
		ctx.SessionRelease()
	})
	app.GetFunc("/get", func(ctx session.Context) {
		ctx.Debugf("session get key1 value: %v", ctx.GetSession("key1"))
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/get").Do()
	client.NewRequest("GET", "/set").Do()
	// 如果第二次get还是nil，是httptest库还未正确处理cookie，请使用阅览器测试。
	client.NewRequest("GET", "/get").Do()

	app.CancelFunc()
	app.Run()
}
