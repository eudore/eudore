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
	HandlerFuncs       `alias:"handlerfuncs"`
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
	app.Server.SetHandler(app)
	app.GetWarp = NewGetWarpWithApp(app)
	app.HandlerFuncs = HandlerFuncs{app.serveContext}
	app.ContextPool.New = func() interface{} { return NewContextBase(app) }
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
			Set(app.Config, "print", NewPrintFunc(val))
			Set(app.Server, "print", NewPrintFunc(val))
			Set(app.Router, "print", NewPrintFunc(val))
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
			app.Logger.Warningf("eudore app invalid option: %v", i)
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

// serveContext 实现处理请求上下文函数。
func (app *App) serveContext(ctx Context) {
	ctx.SetHandler(-1, app.Router.Match(ctx.Method(), ctx.Path(), ctx.Params()))
	ctx.Next()
}

// ServeHTTP 方法实现http.Handler接口，处理http请求。
func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := app.ContextPool.Get().(Context)
	ctx.Reset(r.Context(), w, r)
	ctx.SetHandler(-1, app.HandlerFuncs)
	ctx.Next()
	ctx.End()
	app.ContextPool.Put(ctx)
}

// AddMiddleware 方法如果第一个参数为字符串"global",则使用DefaultHandlerExtend创建请求处理函数，并作为全局请求中间件添加给App。
func (app *App) AddMiddleware(hs ...interface{}) error {
	if len(hs) > 1 {
		name, ok := hs[0].(string)
		if ok && name == "global" {
			handler := DefaultHandlerExtend.NewHandlerFuncs("", hs[1:])
			app.Info("Register app global middleware:", handler)
			app.HandlerFuncs = HandlerFuncsCombine(app.HandlerFuncs[0:len(app.HandlerFuncs)-1], handler)
			app.HandlerFuncs = HandlerFuncsCombine(app.HandlerFuncs, HandlerFuncs{app.serveContext})
			return nil
		}
	}
	return app.Router.AddMiddleware(hs...)
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
	app.Logger.Infof("listen http in %s %s", ln.Addr().Network(), ln.Addr().String())
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
	app.Logger.Infof("listen https in %s %s", ln.Addr().Network(), ln.Addr().String())
	app.Serve(ln)
	return nil
}

// Serve 方法非阻塞启动一个Server监听，并处理结束返回err。
func (app *App) Serve(ln net.Listener) {
	go func() {
		app.Options(app.Server.Serve(ln))
	}()
}
