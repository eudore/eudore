package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

type bindHeader struct {
	Accept      string
	ContextType string `alias:"Content-Type"`
}

func main() {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyBind, eudore.NewBindWithHeader(eudore.NewBinds(nil)))
	app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))

	// 上传文件信息
	app.AnyFunc("/*", func(ctx eudore.Context) {
		var h bindHeader
		ctx.Bind(&h)
	})

	client := httptest.NewClient(app)
	client.NewRequest("HEAD", "/2").WithHeaderValue("Content-Type", "application/json").Do().CheckStatus(200)

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
