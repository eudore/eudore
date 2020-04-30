package main

/*
Context中关于请求响应的定义。
type Context interface{
	Response() ResponseWriter
	SetResponse(ResponseWriter)

	Write([]byte) (int, error)
	WriteHeader(int)
	Redirect(int, string)
	Push(string, *http.PushOptions) error
	Render(interface{}) error
	RenderWith(interface{}, Renderer) error
	// render writer
	WriteString(string) error
	WriteJSON(interface{}) error
	WriteFile(string) error
	...
}

type ResponseWriter interface {
	// http.ResponseWriter
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(int)
	// http.Flusher
	Flush()
	// http.Hijacker
	Hijack() (net.Conn, *bufio.ReadWriter, error)
	// http.Pusher
	Push(string, *http.PushOptions) error
	Size() int
	Status() int
}
*/

import (
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewApp()
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.WriteHeader(201)
		ctx.WriteString("host: " + ctx.Host())

		// 等待
		ctx.Response().Flush()
		time.Sleep(1 * time.Second)
		ctx.WriteString("end")
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/").Do().Out()

	app.CancelFunc()
	app.Run()
}
