package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/component/pprof"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp(eudore.NewRouterStd(eudore.NewRouterCoreDebug(nil)))

	admin := app.Group("/eudore/debug")
	admin.AddMiddleware("global", middleware.NewBasicAuthFunc(map[string]string{"user": "pw"}))
	pprof.Init(admin)
	admin.AnyFunc("/pprof/look/* godoc=/eudore/debug/pprof/godoc", pprof.NewLook(app))
	admin.AnyFunc("/pprof/expvar godoc=/eudore/debug/pprof/godoc", pprof.Expvar)

	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.AddMiddleware(middleware.NewDumpFunc(admin))
	app.AddMiddleware(middleware.NewBlackFunc(map[string]bool{"0.0.0.0/0": true, "10.0.0.0/8": false}, admin))
	app.AddMiddleware(middleware.NewCorsFunc(nil, map[string]string{
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Headers":     "Authorization,DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,X-Parent-Id",
		"Access-Control-Expose-Headers":    "X-Request-Id",
		"access-control-allow-methods":     "GET, POST, PUT, DELETE, HEAD",
		"access-control-max-age":           "1000",
	}))
	app.AddMiddleware(middleware.NewBreakerFunc(admin))
	app.AddHandler("404", "", eudore.HandlerRouter404)
	app.AddHandler("405", "", eudore.HandlerRouter405)
	app.AnyFunc("/echo", func(ctx eudore.Context) {
		ctx.Write(ctx.Body())
	})
	app.AnyFunc("/eudore/debug/admin/ui", middleware.HandlerAdmin)
	app.AnyFunc("/", func(ctx eudore.Context) {
		ctx.Redirect(301,"/eudore/debug/admin/ui")	
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/1").Do()
	client.NewRequest("GET", "/eudore/debug/admin/ui").Do()

	app.Listen(":8088")
	app.ListenTLS(":8089", "", "")
	// app.CancelFunc()
	app.Run()
}
