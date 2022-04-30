package eudore_test

import (
	"testing"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func init() {
	eudore.DefaultLoggerTimeFormat = "none"
}

func TestAppRun(*testing.T) {
	client := eudore.NewClientWarp()
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyClient, client)
	app.SetValue(eudore.ContextKeyRender, eudore.RenderJSON)
	app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))

	app.AddMiddleware(middleware.NewRecoverFunc())
	app.AddMiddleware("global", middleware.NewLoggerFunc(app, "route"))
	app.GetFunc("/hello", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore")
	})

	client.NewRequest("GET", "/hello").BodyString("trace").Do().Callback(eudore.NewResponseReaderCheckStatus(200))

	app.Value(eudore.ContextKeyLogger)
	app.Value(eudore.ContextKeyConfig)
	app.Value(eudore.ContextKeyDatabase)
	app.Value(eudore.ContextKeyClient)
	app.Value(eudore.ContextKeyRouter)

	app.SetValue(eudore.ContextKeyError, "stop app")
	app.CancelFunc()
	app.Run()
}

func TestAppListen(*testing.T) {
	app := eudore.NewApp()
	app.Listen(":8088")
	app.Listen(":8088")
	app.Listen(":8089")
	app.ListenTLS(":8088", "", "")
	app.ListenTLS(":8090", "", "")
	app.Listen("localhost")
	app.ListenTLS("localhost", "", "")

	app.CancelFunc()
	app.Run()
}
