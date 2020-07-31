package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	gorillasession "github.com/gorilla/sessions"
)

func main() {
	sessionName := "sessionid"
	globalSessions := gorillasession.NewCookieStore([]byte(sessionName))

	app := eudore.NewApp()
	app.GetFunc("/set", func(ctx eudore.Context) {
		// 读取会话
		data, err := globalSessions.Get(ctx.Request(), sessionName)
		if err != nil {
			ctx.Fatal(err)
			return
		}

		data.Values["key1"] = 22
		ctx.Debugf("session set key1 value: %v", data.Values["key1"])
		data.Save(ctx.Request(), ctx.Response())
	})
	app.GetFunc("/get", func(ctx eudore.Context) {
		data, err := globalSessions.Get(ctx.Request(), sessionName)
		if err != nil {
			ctx.Fatal(err)
			return
		}

		ctx.Debugf("session get key1 value: %v", data.Values["key1"])
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/get").Do()
	client.NewRequest("GET", "/set").Do()
	// 如果第二次get还是nil，是httptest库还未正确处理cookie，请使用阅览器测试。
	client.NewRequest("GET", "/get").Do()

	app.CancelFunc()
	app.Run()
}
