package main

/*
先设置bind调用validate app.Binder = eudore.NewValidateBinder(app.Binder)，然后bind时会调用Validate验证数据。
*/

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
	app.Binder = eudore.NewBinderHeader(app.Binder)

	// 上传文件信息
	app.AnyFunc("/*", func(ctx eudore.Context) {
		var h bindHeader
		ctx.Bind(&h)
	})

	client := httptest.NewClient(app)
	client.NewRequest("HEAD", "/2").WithHeaderValue("Content-Type", "application/json").Do().CheckStatus(200)
	for client.Next() {
		app.Error(client.Error())
	}

	app.CancelFunc()
	app.Run()
}
