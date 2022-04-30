package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AnyFunc("/*", middleware.NewRateRequestFunc(1, 3, app.Context), eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)

	app.Listen(":8088")
	app.Run()
}
