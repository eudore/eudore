/*
Package eudore golang http framework, less is more.

source: https://github.com/eudore/eudore

wiki: https://github.com/eudore/eudore/wiki

exapmle: https://github.com/eudore/eudore/tree/master/_example

godev: https://pkg.go.dev/github.com/eudore/eudore
*/
package eudore // import "github.com/eudore/eudore"

// Application 定义基本的Application对象，额外功能对App对象组合App即可。

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"sync"
)

/*
The App struct is defined as the main object for the application,
which combines various functional interfaces and implements basic methods.
It provides additional features such as

	Manage Object Lifecycle
	Store global data
	Register global middleware
	Start port monitoring
	Block running service
	Get configuration value and convert type
*/
type App struct {
	context.Context    `alias:"context"`
	context.CancelFunc `alias:"cancelfunc"`
	Logger             `alias:"logger"`
	Config             `alias:"config"`
	Database           `alias:"database"`
	Client             `alias:"client"`
	Server             `alias:"server"`
	Router             `alias:"router"`
	GetWarp            `alias:"getwarp"`
	HandlerFuncs       HandlerFuncs `alias:"handlerfuncs"`
	ContextPool        *sync.Pool   `alias:"contextpool"`
	CancelError        error        `alias:"cancelerror"`
	cancelMutex        sync.Mutex   `alias:"cancelmutex"`
	Values             []any        `alias:"values"`
}

// The NewApp() function creates an App object, initializes various components of the application, and returns the App object.
func NewApp() *App {
	app := &App{}
	app.GetWarp = NewGetWarpWithApp(app)
	app.HandlerFuncs = HandlerFuncs{app.serveContext}
	app.Context, app.CancelFunc = context.WithCancel(context.Background())
	app.SetValue(ContextKeyLogger, NewLogger(nil))
	app.SetValue(ContextKeyConfig, NewConfig(nil))
	app.SetValue(ContextKeyDatabase, NewDatabase(nil))
	app.SetValue(ContextKeyClient, NewClient())
	app.SetValue(ContextKeyServer, NewServer(nil))
	app.SetValue(ContextKeyRouter, NewRouter(nil))
	app.SetValue(ContextKeyBind, NewBinds(nil))
	app.SetValue(ContextKeyRender, NewRenders(nil))
	app.SetValue(ContextKeyTemplate, template.Must(template.New("").Parse(DefaultTemplateInit)))
	app.ContextPool = NewContextBasePool(app)
	return app
}

// The Run() method starts the application and blocks it until it is finished.
//
// Run 方法启动App阻塞等待App结束。
func (app *App) Run() error {
	defer app.SetValue(ContextKeyLogger, nil)
	defer app.SetValue(ContextKeyConfig, nil)
	defer app.SetValue(ContextKeyDatabase, nil)
	defer app.SetValue(ContextKeyClient, nil)
	defer app.SetValue(ContextKeyServer, nil)
	defer app.SetValue(ContextKeyRouter, nil)
	defer func() {
		for i := len(app.Values) - 2; i > -1; i -= 2 {
			app.SetValue(app.Values[i], nil)
		}

		log := app.WithField(ParamDepth, 2)
		if errors.Is(app.Err(), context.Canceled) {
			log.Info("eudore app", app.Err())
		} else {
			log.Fatal("eudore app error:", app.Err())
		}
	}()
	<-app.Done()
	return app.Err()
}

// SetValue method sets the specified key value from the App.
// If the value implements the Mount/Unmount method,
// this method is automatically called when setting and unsetting.
//
// SetValue 方法从App设置指定键值，如果值实现Mount/Unmount方法在设置和取消设置时自动调用该方法。
func (app *App) SetValue(key, val any) {
	anyMount(app, val)
	switch key {
	case ContextKeyLogger:
		defer anyUnmount(app, app.Logger)
		app.Logger, _ = val.(Logger)
	case ContextKeyConfig:
		defer anyUnmount(app, app.Config)
		app.Config, _ = val.(Config)
	case ContextKeyDatabase:
		defer anyUnmount(app, app.Database)
		app.Database, _ = val.(Database)
	case ContextKeyClient:
		defer anyUnmount(app, app.Client)
		app.Client, _ = val.(Client)
	case ContextKeyServer:
		defer anyUnmount(app, app.Server)
		app.Server, _ = val.(Server)
	case ContextKeyRouter:
		defer anyUnmount(app, app.Router)
		app.Router, _ = val.(Router)
	case ContextKeyContextPool:
		app.ContextPool, _ = val.(*sync.Pool)
	case ContextKeyError:
		if val != nil {
			app.cancelMutex.Lock()
			defer app.cancelMutex.Unlock()
			err, ok := val.(error)
			if !ok {
				err = fmt.Errorf("%v", val)
			}
			app.CancelError = err
			app.CancelFunc()
		}
	default:
		for i := 0; i < len(app.Values); i += 2 {
			if app.Values[i] == key {
				defer anyUnmount(app, app.Values[i+1])
				app.Values[i+1] = val
				return
			}
		}
		app.Values = append(app.Values, key, val)
	}
}

// Value method gets the specified key value from the App.
//
// Value 方法从App获取指定键值。
func (app *App) Value(key any) any {
	switch key {
	case ContextKeyApp:
		return app
	case ContextKeyLogger:
		return app.Logger
	case ContextKeyConfig:
		return app.Config
	case ContextKeyDatabase:
		return app.Database
	case ContextKeyClient:
		return app.Client
	case ContextKeyServer:
		return app.Server
	case ContextKeyRouter:
		return app.Router
	case ContextKeyAppKeys:
		keys := make([]any, 0, 6+len(app.Values)/2)
		keys = append(keys,
			ContextKeyLogger, ContextKeyConfig, ContextKeyDatabase,
			ContextKeyClient, ContextKeyServer, ContextKeyRouter,
		)
		for i := 0; i < len(app.Values); i += 2 {
			keys = append(keys, app.Values[i])
		}
		return keys
	}
	for i := 0; i < len(app.Values); i += 2 {
		if app.Values[i] == key {
			return app.Values[i+1]
		}
	}
	return app.Context.Value(key)
}

// Err method returns an error at the end of the App Context.
//
// Err 方法返回App Context结束时错误。
func (app *App) Err() error {
	app.cancelMutex.Lock()
	defer app.cancelMutex.Unlock()
	if app.CancelError != nil {
		return app.CancelError
	}
	return app.Context.Err()
}

func anyMount(ctx context.Context, i any) {
	loader, ok := i.(interface{ Mount(context.Context) })
	if ok {
		loader.Mount(ctx)
	}
}

func anyUnmount(ctx context.Context, i any) {
	closer, ok := i.(interface{ Unmount(context.Context) })
	if ok {
		closer.Unmount(ctx)
	}
}

func anyMetadata(i any) any {
	metaer, ok := i.(interface{ Metadata() any })
	if ok {
		return metaer.Metadata()
	}
	return nil
}

// serveContext Implement the request context function.
// serveContext 实现处理请求上下文函数。
func (app *App) serveContext(ctx Context) {
	ctx.SetHandler(-1, app.Router.Match(ctx.Method(), ctx.Path(), ctx.Params()))
	ctx.Next()
}

// The ServeHTTP method implements the http.Handler interface to process http requests.
//
// Create and initialize a Context, set app.HandlerFuncs as the global request handler function of Context.
// When app.HandlerFuncs is last processed, the app.serveContext method is called,
// Use app.Router to match the route processing function of this request for secondary request processing.
//
// ServeHTTP 方法实现http.Handler接口，处理http请求。
//
// 创建并初始化一个Context，设置app.HandlerFuncs为Context的全局请求处理函数。
// 在app.HandlerFuncs最后一次处理时，调用了app.serveContext方法，
// 使用app.Router匹配出这个请求的路由处理函数进行二次请求处理。
func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pool := app.ContextPool
	ctx := pool.Get().(Context)
	ctx.Reset(w, r)
	ctx.SetHandler(-1, app.HandlerFuncs)
	ctx.Next()
	ctx.End()
	pool.Put(ctx)
}

// AddMiddleware If the first parameter of the AddMiddleware method is the string "global",
// it will be added to the App as a global request middleware,
// using NewHandlerExtenderWithContext to create a request processing function,
// otherwise it is equivalent to calling the app.Rputer.AddMiddleware method.
//
// AddMiddleware 方法如果第一个参数为字符串"global",
// 为全局请求中间件添加给App(使用NewHandlerExtenderWithContext创建请求处理函数),
// 否则等同于调用app.Rputer.AddMiddleware方法。
func (app *App) AddMiddleware(hs ...any) error {
	if len(hs) > 1 {
		name, ok := hs[0].(string)
		if ok && name == "global" {
			handler := NewHandlerExtenderWithContext(app).CreateHandler("", hs[1:])
			app.WithField(ParamDepth, 1).
				Info("Register app global middleware:", handler)
			last := app.HandlerFuncs[len(app.HandlerFuncs)-1]
			app.HandlerFuncs = NewHandlerFuncsCombine(app.HandlerFuncs[0:len(app.HandlerFuncs)-1], handler)
			app.HandlerFuncs = NewHandlerFuncsCombine(app.HandlerFuncs, HandlerFuncs{last})
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
	app.WithField(ParamDepth, 1).Infof("listen http in %s %s",
		ln.Addr().Network(), ln.Addr().String())
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
	app.WithField(ParamDepth, 1).Infof("listen https in %s %s, host name: %v",
		ln.Addr().Network(), ln.Addr().String(), conf.Certificate.DNSNames)
	app.Serve(ln)
	return nil
}

// Serve method starts a Server monitor non-blocking, and uses the app to process the monitor and return an error.
//
// Serve 方法非阻塞启动一个Server监听，并使用app处理监听结束返回错误。
func (app *App) Serve(ln net.Listener) {
	srv := app.Server
	go func() {
		app.SetValue(ContextKeyError, srv.Serve(ln))
	}()
}

func (app *App) Parse() error {
	err := app.Config.Parse(app)
	if err != nil {
		app.SetValue(ContextKeyError, err)
	}
	return err
}
