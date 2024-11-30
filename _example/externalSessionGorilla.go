//go:build go1.23

package main

import (
	"fmt"

	"github.com/eudore/eudore"
	sessions "github.com/gorilla/sessions"
)

func main() {
	sessionName := "sessionid"
	// 需要设置cookie对称加密密钥
	globalSessions := sessions.NewCookieStore([]byte("cookie secret"))

	app := eudore.NewApp()
	app.GetFunc("/set", func(ctx eudore.Context) {
		// 读取会话
		data, err := globalSessions.Get(ctx.Request(), sessionName)
		if err != nil {
			ctx.Fatal(err)
			return
		}

		data.Values["key"] = 22
		ctx.Debugf("session set key value: %v", data.Values["key"])
		data.Save(ctx.Request(), ctx.Response())
	})
	app.GetFunc("/get", func(ctx eudore.Context) {
		data, err := globalSessions.Get(ctx.Request(), sessionName)
		if err != nil {
			ctx.Fatal(err)
			return
		}

		ctx.Debugf("session get key value: %v", data.Values["key"])
		ctx.WriteString(fmt.Sprintf("session get key value: %v", data.Values["key"]))
	})

	app.Listen(":8088")
	app.Run()
}
