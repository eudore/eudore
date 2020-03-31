package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewCore()
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Debugf("querys: %v", ctx.Querys())
		ctx.Debugf("name: %s", ctx.GetQuery("name"))
	})

	client := httptest.NewClient(app)
	client.NewRequest("PUT", "/file/22?name=eudore&type=2&size=722").WithHeaderValue(eudore.HeaderContentType, eudore.MimeApplicationForm).Do().CheckStatus(200).Out()
	client.NewRequest("PUT", "/file/22?%gh&%ij").WithHeaderValue(eudore.HeaderContentType, eudore.MimeApplicationForm).Do().CheckStatus(200).Out()
	for client.Next() {
		app.Error(client.Error())
	}

	app.Run()
}
