package main

/*
App全局中间件会在路由匹配前执行，可以影响路由匹配数据。

全局中间件会在路由匹配前执行，不会存在路由详细、也可以修改基础信息影响路由匹配。

ServeHTTP时先设置请求上下文的处理函数是全部全局中间件函数处理请求。
最后一个全局中间件函数是app.HandleContext，该方法才会匹配路由请求，然后重新调用ctx.SetHandler方法设置多个请求处理函数是路由匹配后的结果。
*/

import (
	"github.com/eudore/eudore"
	"net/http"
)

type (
	// App 定义一个简单app
	App struct {
		*eudore.App
		Handlers eudore.HandlerFuncs
	}
)

func main() {
	app := NewApp()
	app.AddGlobalMiddleware(func(ctx eudore.Context) {
		ctx.Request().Method = "PUT"
		ctx.Request().URL.Path = "/"
	})
	app.AnyFunc("/*", "Hello, 世界")
	app.Info("hello")
	app.Listen(":8088")
	app.CancelFunc()
	app.Run()
}

// NewApp 方法创建一个自定义app
func NewApp() *App {
	app := &App{App: eudore.NewApp()}
	app.Handlers = eudore.HandlerFuncs{app.HandleContext}
	return app
}

// AddGlobalMiddleware 给app添加全局中间件，会在Router.Match前执行。
func (app *App) AddGlobalMiddleware(hs ...eudore.HandlerFunc) {
	app.Handlers = eudore.HandlerFuncsCombine(app.Handlers[0:len(app.Handlers)-1], hs)
	app.Handlers = eudore.HandlerFuncsCombine(app.Handlers, eudore.HandlerFuncs{app.HandleContext})
}

// HandleContext 实现处理请求上下文函数。
func (app *App) HandleContext(ctx eudore.Context) {
	ctx.SetHandler(-1, app.Router.Match(ctx.Method(), ctx.Path(), ctx.Params()))
	ctx.Next()
}

// ServeHTTP 方法实现http.Handler接口，处理http请求。
func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// init
	ctx := app.ContextPool.Get().(eudore.Context)
	response := new(eudore.ResponseWriterHTTP)
	// handle
	response.Reset(w)
	ctx.Reset(r.Context(), response, r)
	ctx.SetHandler(-1, app.Handlers)
	ctx.Next()
	ctx.End()
	// release
	app.ContextPool.Put(ctx)
}
