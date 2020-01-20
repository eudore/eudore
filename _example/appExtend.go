package main

/*
为什么没有日志？  app默认Logger是LoggerInit只会保存日志并未处理,参考core实现。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"net"
	"net/http"
)

type (
	// App 定义一个简单app
	App struct {
		*eudore.App
	}
)

func main() {
	app := NewApp()
	httptest.NewClient(app).Stop(0)
	app.AnyFunc("/*", "Hello, 世界")
	app.Info("hello")
	app.ListenAndServe(":8088")
}

// NewApp 方法创建一个自定义app
func NewApp() *App {
	return &App{eudore.NewApp()}
}

// ListenAndServe 方法监听一个tcp地址并启动server
func (app *App) ListenAndServe(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return app.Serve(ln)
}

// ServeHTTP 方法实现http.Handler接口，处理http请求。
func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// init
	ctx := app.ContextPool.Get().(eudore.Context)
	response := new(eudore.ResponseWriterHTTP)
	// handle
	response.Reset(w)
	ctx.Reset(r.Context(), response, r)
	ctx.SetHandler(app.Router.Match(ctx.Method(), ctx.Path(), ctx.Params()))
	ctx.Next()
	ctx.End()
	// release
	app.ContextPool.Put(ctx)

}
