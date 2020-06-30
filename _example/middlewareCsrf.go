package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AnyFunc("/query", middleware.NewCsrfFunc("query: csrf", "_csrf"), eudore.HandlerEmpty)
	app.AnyFunc("/header", middleware.NewCsrfFunc("header: "+eudore.HeaderXCSRFToken, eudore.SetCookie{Name: "_csrf", MaxAge: 86400}), eudore.HandlerEmpty)
	app.AnyFunc("/form", middleware.NewCsrfFunc("form: csrf", &eudore.SetCookie{Name: "_csrf", MaxAge: 86400}), eudore.HandlerEmpty)
	app.AnyFunc("/fn", middleware.NewCsrfFunc(func(ctx eudore.Context) string { return ctx.GetQuery("csrf") }, "_csrf"), eudore.HandlerEmpty)
	app.AnyFunc("/nil", middleware.NewCsrfFunc(nil, "_csrf"), eudore.HandlerEmpty)

	app.AddMiddleware(middleware.NewCsrfFunc("csrf", nil))
	app.AnyFunc("/*", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/1").Do().CheckStatus(200)
	csrfval := client.GetCookie("/", "_csrf")
	client.NewRequest("POST", "/2").Do().CheckStatus(400)
	client.NewRequest("POST", "/1").WithAddQuery("csrf", csrfval).Do().CheckStatus(200)
	client.NewRequest("POST", "/query").WithAddQuery("csrf", csrfval).Do().CheckStatus(200)
	client.NewRequest("POST", "/header").WithHeaderValue(eudore.HeaderXCSRFToken, csrfval).Do().CheckStatus(200)
	client.NewRequest("POST", "/form").WithBodyFormValue("csrf", csrfval).Do().CheckStatus(200)
	client.NewRequest("POST", "/form").WithBodyJSONValue("csrf", csrfval).Do().CheckStatus(400)
	client.NewRequest("POST", "/fn").WithAddQuery("csrf", csrfval).Do().CheckStatus(200)
	client.NewRequest("POST", "/nil").WithAddQuery("csrf", csrfval).Do().CheckStatus(200)

	for client.Next() {
		app.Error(client.Error())
	}

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
