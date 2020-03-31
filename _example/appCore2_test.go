package eudore_test

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"testing"
)

func TestAppCoreListen2(*testing.T) {
	app := eudore.NewCore()
	httptest.NewClient(app).Stop(0)
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore core")
	})

	app.Listen(":8088")
	app.Listen(":8089")
	app.Listen(":8088")
	app.ListenTLS(":8088", "", "")
	app.Run()
}
