package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewSingleFlightFunc(), middleware.NewGzipFunc(5))
	app.AnyFunc("/sf", func(ctx eudore.Context) {
		ctx.Redirect(301, "/")
		ctx.Debug(ctx.Response().Status(), ctx.Response().Size())
		ctx.Response().Hijack()
		ctx.Push("/js", nil)
		ctx.Response().Flush()
	})
	app.AnyFunc("/*", func(ctx eudore.Context) {
		time.Sleep(time.Second / 3)
		ctx.WriteString("hello eudore")
	})

	client := httptest.NewClient(app)
	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(i int) {
			client.NewRequest("GET", "/?c="+fmt.Sprint(i)).Do().CheckBodyString("hello eudore")
			wg.Done()
		}(i)
	}
	wg.Wait()
	client.NewRequest("GET", "/sf").Do()
	client.NewRequest("POST", "/sf").Do()
	client.NewRequest("GET", "/s").Do().CheckBodyString("hello eudore")

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
