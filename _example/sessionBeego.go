package main

import (
	beegosession "github.com/astaxie/beego/session"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/component/session"
	"github.com/eudore/eudore/component/session/beego"
)

func main() {
	// 创建session，并注册转换函数。
	sessionConfig := &beegosession.ManagerConfig{
		CookieName:      "gosessionid",
		EnableSetCookie: true,
		Gclifetime:      3600,
		Maxlifetime:     3600,
		Secure:          false,
		CookieLifeTime:  3600,
		ProviderConfig:  "./tmp",
	}
	globalSessions, _ := beegosession.NewManager("memory", sessionConfig)
	go globalSessions.GC()
	app := eudore.NewApp()
	app.AddHandlerExtend(beego.NewExtendContextSession(globalSessions))

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
