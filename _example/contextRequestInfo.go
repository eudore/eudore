package main

/*
Context中关于请求信息的定义。
type Context interface{
	Request() *RequestReader
	SetRequest(*RequestReader)

	Read([]byte) (int, error)
	Host() string
	Method() string
	Path() string
	RealIP() string
	RequestID() string
	Referer() string
	ContentType() string
	Istls() bool
	Body() []byte
	...
}

type RequestReader = http.Request
*/

import (
	"fmt"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("host: " + ctx.Host())
		ctx.WriteString("\nmethod: " + ctx.Method())
		ctx.WriteString("\npath: " + ctx.Path())
		ctx.WriteString("\nreal ip: " + ctx.RealIP())
		ctx.WriteString("\ncontext type: " + ctx.ContentType())
		ctx.WriteString("\nistls: " + fmt.Sprint(ctx.Istls()))
		body := ctx.Body()
		if len(body) > 0 {
			ctx.WriteString("\nbody: " + string(body))
		}
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/").WithHeaderValue(eudore.HeaderXForwardedFor, "192.168.1.4 192.168.1.1").Do().Out()
	client.NewRequest("GET", "/").WithHeaderValue(eudore.HeaderXRealIP, "192.168.1.4").Do().Out()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
