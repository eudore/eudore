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
	app.AddController(middleware.NewPprofController())

	app.AddMiddleware(middleware.NewCacheFunc(time.Second/10, app.Context))
	app.AnyFunc("/sf", func(ctx eudore.Context) {
		ctx.Redirect(301, "/")
		ctx.Debug(ctx.Response().Status(), ctx.Response().Size())
	})
	app.AnyFunc("/*", func(ctx eudore.Context) {
		time.Sleep(time.Second / 3)
		ctx.WriteString("hello eudore")
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/sf").Do()
	wg := sync.WaitGroup{}
	wg.Add(5)
	for n := 0; n < 5; n++ {
		go func() {
			for i := 0; i < 3; i++ {
				client.NewRequest("GET", "/?c="+fmt.Sprint(i)).Do().CheckBodyString("hello eudore")
				client.NewRequest("GET", "/?c="+fmt.Sprint(i)).Do().CheckBodyString("hello eudore")
				time.Sleep(time.Millisecond * 200)
				client.NewRequest("GET", "/?c="+fmt.Sprint(i)).Do().CheckBodyString("hello eudore")
			}
			wg.Done()
		}()
	}
	wg.Wait()

	client.NewRequest("GET", "/sf").Do()
	client.NewRequest("POST", "/sf").Do()
	client.NewRequest("GET", "/s").Do().CheckBodyString("hello eudore")

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
