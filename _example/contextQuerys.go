package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewApp()
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Debugf("querys: %v", ctx.Querys())
		ctx.Debugf("name: %s", ctx.GetQuery("name"))
	})
	app.AnyFunc("/err", func(ctx eudore.Context) {
		ctx.Request().URL.RawQuery = "%gh&%ij"
		ctx.Debugf("name: %s", ctx.GetQuery("name"))
		ctx.Debugf("querys: %v", ctx.Querys())
	})

	client := httptest.NewClient(app)
	client.NewRequest("PUT", "/file/22?name=eudore&type=2&size=722").WithHeaderValue(eudore.HeaderContentType, eudore.MimeApplicationForm).Do().CheckStatus(200)
	client.NewRequest("PUT", "/file/22?%gh&%ij").WithHeaderValue(eudore.HeaderContentType, eudore.MimeApplicationForm).Do().CheckStatus(200)
	client.NewRequest("PUT", "/err").WithHeaderValue(eudore.HeaderContentType, eudore.MimeApplicationForm).Do().CheckStatus(200)

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
