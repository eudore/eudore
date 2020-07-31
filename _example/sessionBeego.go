package main

import (
	beegosession "github.com/astaxie/beego/session"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	// 创建session，并注册转换函数。
	globalSessions, _ := beegosession.NewManager("memory", &beegosession.ManagerConfig{
		CookieName:      "gosessionid",
		EnableSetCookie: true,
		Gclifetime:      3600,
		Maxlifetime:     3600,
		Secure:          false,
		CookieLifeTime:  3600,
		ProviderConfig:  "./tmp",
	})
	go globalSessions.GC()

	app := eudore.NewApp()
	app.GetFunc("/set", func(ctx eudore.Context) {
		// 读取会话
		data, err := globalSessions.SessionStart(ctx.Response(), ctx.Request())
		if err != nil {
			ctx.Fatal(err)
		}

		data.Set("key1", 22)
		ctx.Debugf("session set key1 value: %v", data.Get("key1"))
		data.SessionRelease(ctx.Response())
	})
	app.GetFunc("/get", func(ctx eudore.Context) {
		data, err := globalSessions.SessionStart(ctx.Response(), ctx.Request())
		if err != nil {
			ctx.Fatal(err)
		}

		ctx.Debugf("session get key1 value: %v", data.Get("key1"))
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/get").Do()
	client.NewRequest("GET", "/set").Do()
	// 如果第二次get还是nil，是httptest库还未正确处理cookie，请使用阅览器测试。
	client.NewRequest("GET", "/get").Do()

	app.CancelFunc()
	app.Run()
}
