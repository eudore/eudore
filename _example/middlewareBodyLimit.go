package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.AddMiddleware(middleware.NewBodyLimitFunc(32))
	// app.AddMiddleware(middleware.NewBodyLimitFunc(32 << 20))
	app.AnyFunc("/body1", func(ctx eudore.Context) {
		ctx.Body()
	})
	app.AnyFunc("/body2", func(ctx eudore.Context) {
		body := make([]byte, 8, 8)
		for {
			_, err := ctx.Read(body)
			if err != nil {
				break
			}
		}
		ctx.Read(body)
	})
	app.AnyFunc("/*", eudore.HandlerEmpty)

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/1").Do().Out()
	client.NewRequest("POST", "/body1").WithBody("1234567890").Do().Out()
	client.NewRequest("POST", "/body2").WithBody("1234567890").Do().Out()
	client.NewRequest("POST", "/body2").WithBody(bodycheck{}).Do().Out()
	client.NewRequest("POST", "/body2").WithBody("12345678901234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY").Do().Out()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

// bodycheck 对象实现io.Reader不实现Close方法，模拟分段传输的数据和长度。
type bodycheck struct{}

func (bodycheck) Read(p []byte) (int, error) {
	return cap(p), nil
}
