package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewApp()
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.Redirect(302, "/hello")
	})
	app.GetFunc("/hello", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore")
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/").Do().CheckStatus(200).Out()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
