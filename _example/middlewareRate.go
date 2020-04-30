package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewRateFunc(app, 1, 3))
	app.AnyFunc("/*", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("PUT", "/file/data/2").Do().CheckStatus(200)
	client.NewRequest("PUT", "/file/data/2").Do().CheckStatus(200)
	client.NewRequest("PUT", "/file/data/2").Do().CheckStatus(200)
	client.NewRequest("PUT", "/file/data/2").Do().CheckStatus(200)
	client.NewRequest("PUT", "/file/data/2").Do().CheckStatus(200)
	for client.Next() {
		app.Error(client.Error())
	}

	app.CancelFunc()
	app.Run()
}
