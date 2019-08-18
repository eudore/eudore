package test

import (
	"github.com/eudore/eudore"
	"net/http"
	"testing"
	"time"
)

func TestServerStd(t *testing.T) {
	app := eudore.NewCore()
	app.Listen(":8088")
	app.AnyFunc("/*", func(ctx eudore.Context) {
		t.Log(ctx.Path())
	})
	time.AfterFunc(1*time.Second, func() {
		http.Get("http://localhost:8088/index")
	})
	time.AfterFunc(2*time.Second, func() {
		app.Close()
	})
	t.Log(app.Run())
}

func TestServerMulti(t *testing.T) {
	app := eudore.NewCore()
	app.Listen(":8088")
	app.Listen(":8089")
	time.AfterFunc(1*time.Second, func() {
		http.Get("http://localhost:8088/")
	})
	time.AfterFunc(2*time.Second, func() {
		app.Close()
	})
	app.Run()
}
