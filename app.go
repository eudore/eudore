package eudore // import "github.com/eudore/eudore"

// Application 定义基本的Application对象，额外功能对App对象组合App即可。

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
	CancelError        error     `alias:"cancelerror"`
	cancelMutex        sync.Mutex
}

// NewApp function creates an App object.
//
// NewApp 函数创建一个App对象。
func NewApp(options ...interface{}) *App {
	app := &App{
		Config:    NewConfigMap(nil),
		Logger:    NewLoggerStd(nil),
		Server:    NewServerStd(nil),
		Router:    NewRouterStd(nil),
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

// Options method loads the app component. When the option type is context.Context, Logger, Config, Server, Router, Binder, Renderer, Validater, the app property will be set,
// and the print property of the component will be set. If the type is error, it will be the app end error Return to the Run method.
//
// Options 方法加载app组件，option类型为context.Context、Logger、Config、Server、Router、Binder、Renderer、Validater时会设置app属性，
// 并设置组件的print属性，如果类型为error将作为app结束错误返回给Run方法。
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
			app.Logger = val
			Set(app.Config, "print", NewPrintFunc(val))
			Set(app.Server, "print", NewPrintFunc(val))
			Set(app.Router, "print", NewPrintFunc(val))
			if ok {
				initlog.NextHandler(val)
			}
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
			app.cancelMutex.Lock()
			defer app.cancelMutex.Unlock()
			// 记录第一个错误
			if app.CancelError == nil {
				app.CancelError = val
			}
		default:
			app.Logger.Warningf("eudore app invalid option: %v", i)
		}
	}
}

// Run method starts the App and blocks and waits for the end of the App, and periodically calls app.Logger.Sync() to output the log.
//
// Run 方法启动App阻塞等待App结束，并周期调用app.Logger.Sync()将日志输出。
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
	app.cancelMutex.Lock()
	defer app.cancelMutex.Unlock()
	return app.CancelError
}

// serveContext Implement the request context function.
// serveContext 实现处理请求上下文函数。
func (app *App) serveContext(ctx Context) {
	ctx.SetHandler(-1, app.Router.Match(ctx.Method(), ctx.Path(), ctx.Params()))
	ctx.Next()
}

// The ServeHTTP method implements the http.Handler interface to process http requests.
//
// Create and initialize a Context, then set app.HandlerFuncs as the handler of the Context to handle the global middleware chain.
// When app.HandlerFuncs is processed for the last time, the app.serveContext method is called,
// and the route of this request is matched using app.Router The middleware and routing processing functions perform secondary request processing.
//
// ServeHTTP 方法实现http.Handler接口，处理http请求。
//
// 创建并初始化一个Context，然后设置app.HandlerFuncs为Context的处理者处理全局中间件链，
// 在app.HandlerFuncs最后一次处理时，调用了app.serveContext方法，使用app.Router匹配出这个请求的路由中间件和路由处理函数进行二次请求处理。
func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := app.ContextPool.Get().(Context)
	ctx.Reset(r.Context(), w, r)
	ctx.SetHandler(-1, app.HandlerFuncs)
	ctx.Next()
	ctx.End()
	app.ContextPool.Put(ctx)
}

// AddMiddleware If the first parameter of the AddMiddleware method is the string "global",
// it will be added to the App as a global request middleware (using DefaultHandlerExtend to create a request processing function),
// otherwise it is equivalent to calling the app.Rputer.AddMiddleware method.
//
// AddMiddleware 方法如果第一个参数为字符串"global",则作为全局请求中间件添加给App(使用DefaultHandlerExtend创建请求处理函数),
// 否则等同于调用app.Rputer.AddMiddleware方法。
func (app *App) AddMiddleware(hs ...interface{}) error {
	if len(hs) > 1 {
		name, ok := hs[0].(string)
		if ok && name == "global" {
			handler := DefaultHandlerExtend.NewHandlerFuncs("", hs[1:])
			app.Info("Register app global middleware:", handler)
			last := app.HandlerFuncs[len(app.HandlerFuncs)-1]
			app.HandlerFuncs = HandlerFuncsCombine(app.HandlerFuncs[0:len(app.HandlerFuncs)-1], handler)
			app.HandlerFuncs = HandlerFuncsCombine(app.HandlerFuncs, HandlerFuncs{last})
			return nil
		}
	}
	return app.Router.AddMiddleware(hs...)
}

// Listen method listens to an http port.
//
// Listen 方法监听一个http端口。
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

// ListenTLS method listens to an https port, if h2 is enabled by default.
//
// ListenTLS 方法监听一个https端口，如果默认开启h2。
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
	app.Logger.Infof("listen https in %s %s,host name: %v", ln.Addr().Network(), ln.Addr().String(), conf.Certificate.DNSNames)
	app.Serve(ln)
	return nil
}

// Serve method starts a Server monitor non-blocking, and uses the app to process the monitor and return an error.
//
// Serve 方法非阻塞启动一个Server监听，并使用app处理监听结束返回错误。
func (app *App) Serve(ln net.Listener) {
	go func() {
		app.Options(app.Server.Serve(ln))
	}()
}
