package eudore // import "github.com/eudore/eudore"

// Application 定义基本的Application对象，如果需要复杂的App对象组合App即可。

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"
)

// App combines the main functional interfaces to implement a simple basic method.
//
// App 组合主要功能接口，实现简单的基本方法。
type App struct {
	context.Context    `alias:"context"`
	context.CancelFunc `alias:"cancelfunc"`
	Config             `alias:"config"`
	Logger             `alias:"logger"`
	Server             `alias:"server"`
	Router             `alias:"router"`
	Binder             `alias:"binder"`
	Renderer           `alias:"renderer"`
	Validater          `alias:"validater"`
	GetWarp            `alias:"getwarp"`
	ContextPool        sync.Pool `alias:"contextpool"`
}

// NewApp 函数创建一个App对象。
func NewApp(options ...interface{}) *App {
	app := &App{
		Config:    NewConfigMap(nil),
		Logger:    NewLoggerStd(nil),
		Server:    NewServerStd(nil),
		Router:    NewRouterRadix(),
		Binder:    BindDefault,
		Renderer:  RenderDefault,
		Validater: DefaultValidater,
	}
	app.Context, app.CancelFunc = context.WithCancel(context.WithValue(context.Background(), AppContextKey, app))
	app.ContextPool.New = func() interface{} {
		return NewContextBase(app)
	}
	app.Server.SetHandler(app)
	app.GetWarp = NewGetWarpWithApp(app)
	Set(app.Config, "print", NewPrintFunc(app))
	Set(app.Server, "print", NewPrintFunc(app))
	Set(app.Router, "print", NewPrintFunc(app))
	app.Options(options...)
	return app
}

// Options 方法加载app组件。
func (app *App) Options(options ...interface{}) {
	for _, i := range options {
		if i == nil {
			continue
		}
		switch val := i.(type) {
		case context.Context:
			app.Context = val
			app.Context, app.CancelFunc = context.WithCancel(app.Context)
		case Logger:
			initlog, ok := app.Logger.(loggerInitHandler)
			if ok {
				initlog.NextHandler(val)
			}
			app.Logger = val
		case Config:
			app.Config = val
			Set(app.Config, "print", NewPrintFunc(app))
		case Server:
			app.Server = val
			app.Server.SetHandler(app)
			Set(app.Server, "print", NewPrintFunc(app))
		case Router:
			app.Router = val
			Set(app.Router, "print", NewPrintFunc(app))
		case Binder:
			app.Binder = val
		case Renderer:
			app.Renderer = val
		case Validater:
			app.Validater = val
		case error:
			app.Error("eudore app cannel context on handler error: " + val.Error())
			app.CancelFunc()
		default:
			app.Logger.Warningf("eudore app  invalid option: %v", i)
		}
	}
}

// Run 方法阻塞等待App结束。
func (app *App) Run() error {
	ticker := time.NewTicker(time.Millisecond * 80)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			app.Logger.Sync()
		}
	}()

	<-app.Done()
	time.Sleep(time.Millisecond * 100)
	app.Shutdown(context.Background())
	time.Sleep(time.Millisecond * 100)
	return app.Err()
}

// ServeHTTP 方法实现http.Handler接口，处理http请求。
func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// init
	ctx := app.ContextPool.Get().(Context)
	response := responseWriterHTTPPool.Get().(*ResponseWriterHTTP)
	// handle
	response.Reset(w)
	ctx.Reset(r.Context(), response, r)
	ctx.SetHandler(-1, app.Router.Match(ctx.Method(), ctx.Path(), ctx.Params()))
	ctx.Next()
	ctx.End()
	// release
	responseWriterHTTPPool.Put(response)
	app.ContextPool.Put(ctx)
}

// Listen 方法监听一个http端口
func (app *App) Listen(addr string) error {
	conf := ServerListenConfig{
		Addr: addr,
	}
	ln, err := conf.Listen()
	if err != nil {
		app.Error(err)
		return err
	}
	app.Logger.Infof("listen %s %s", ln.Addr().Network(), ln.Addr().String())
	app.Serve(ln)
	return nil
}

// ListenTLS 方法监听一个https端口，如果支持默认开启h2
func (app *App) ListenTLS(addr, key, cert string) error {
	conf := ServerListenConfig{
		Addr:     addr,
		HTTPS:    true,
		HTTP2:    true,
		Keyfile:  key,
		Certfile: cert,
	}
	ln, err := conf.Listen()
	if err != nil {
		app.Error(err)
		return err
	}
	app.Logger.Infof("listen tls %s %s", ln.Addr().Network(), ln.Addr().String())
	app.Serve(ln)
	return nil
}

// Serve 方法非阻塞启动一个Server监听，并处理结束返回err。
func (app *App) Serve(ln net.Listener) {
	go func() {
		app.Options(app.Server.Serve(ln))
	}()
}
