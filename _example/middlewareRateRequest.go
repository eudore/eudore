package main

import (
	"context"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AnyFunc("/*", middleware.NewRateRequestFunc(1, 3, app.Context), eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()

	middlewareRate2()
	middlewareRate3()
	middlewareRate4()
	middlewareRate5()
}

func middlewareRate2() {
	app := eudore.NewApp()
	app.AddMiddleware("/out", func(ctx eudore.Context) {
		cctx, cannel := context.WithTimeout(ctx.GetContext(), time.Millisecond*20)
		go func() {
			<-cctx.Done()
			cannel()
		}()
		ctx.WithContext(cctx)
	})
	app.AddMiddleware(middleware.NewRateRequestFunc(1, 3, app.Context, time.Millisecond*10, func(ctx eudore.Context) string {
		return ctx.RealIP()
	}))
	app.AnyFunc("/out", eudore.HandlerEmpty)
	app.AnyFunc("/*", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	time.Sleep(50 * time.Millisecond)
	client.NewRequest("PUT", "/out").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/out").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

func middlewareRate3() {
	app := eudore.NewApp()
	app.AddMiddleware("/out", func(ctx eudore.Context) {
		cctx, cannel := context.WithTimeout(ctx.GetContext(), time.Millisecond*2)
		cannel()
		ctx.WithContext(cctx)
	})
	app.AddMiddleware(middleware.NewRateRequestFunc(1, 3, app.Context, time.Millisecond*10, func(ctx eudore.Context) string {
		return ctx.RealIP()
	}))
	app.AnyFunc("/out", func(ctx eudore.Context) {
		time.Sleep(time.Millisecond * 5)
	})
	app.AnyFunc("/*", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/out").Do().CheckStatus(200)
	client.NewRequest("PUT", "/out").Do().CheckStatus(200)

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

func middlewareRate4() {
	app := eudore.NewApp()
	app.AnyFunc("/*", middleware.NewRateRequestFunc(1, 3, app.Context, time.Millisecond*100), eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	time.Sleep(time.Second / 2)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

func middlewareRate5() {
	app := eudore.NewApp()
	app.AnyFunc("/*", middleware.NewRateRequestFunc(1, 2, app.Context, time.Microsecond*49), eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	client.NewRequest("PUT", "/").Do().CheckStatus(200)
	time.Sleep(time.Second)

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
