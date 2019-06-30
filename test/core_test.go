package test

import (
	"github.com/eudore/eudore"
	"testing"
	"time"
)

func TestCore(*testing.T) {
	app := eudore.NewCore()
	app.AnyFunc("/", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore core")
	})
	time.AfterFunc(5*time.Second, func() {
		app.Server.Close()
	})
	app.Listen(":8082")
	app.Info("start core result: ", app.Run())
}
