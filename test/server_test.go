package test

import (
	"time"
	"testing"
	"github.com/eudore/eudore"
)

func TestServerStd(t *testing.T) {
	app := eudore.NewCore()
	app.Listen(":8088")
	app.AnyFunc("/*", func(ctx eudore.Context){
		t.Log(ctx.Path())
	})
	time.AfterFunc(1 * time.Second, func() {
		testPort("http://localhost:8088/index")
	})
	time.AfterFunc(2 * time.Second, func() {
		// t.Log(app.Restart())
		app.Close()
	})
	t.Log(app.Run())
}


func TestServerMulti(t *testing.T) {
	app := eudore.NewCore()
	app.Listen(":8088")
	app.Listen(":8089")
	time.AfterFunc(1 * time.Second, func() {
		testPort("http://localhost:8088/")
	})
	time.AfterFunc(2 * time.Second, func() {
		app.Close()
	})
	app.Run()
}

func testPort(url string) {
	eudore.NewClientHttp().NewRequest("GET" ,url ,nil).Do()
}