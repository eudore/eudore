/*
Package eudore golang http framework.

source: https://github.com/eudore/eudore

wiki: https://github.com/eudore/eudore/wiki
*/
package eudore // import "github.com/eudore/eudore"

// Application defines the basic Application object, and additional functions
// can be combined with App objects.

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
)

// App combines [context.Context] [Logger] [Config] [Router] [Client] [Server],
// and implements [net/http.Handler] and interfaces and some wrap methods.
//
// You can reassemble a new App and customize the components.
type App struct {
	context.Context    `alias:"context"`
	context.CancelFunc `alias:"cancelfunc"`
	Logger             `alias:"logger"`
	Config             `alias:"config"`
	Router             `alias:"router"`
	Client             `alias:"client"`
	Server             `alias:"server"`
	GetWrap            `alias:"getwrap"`
	HandlerFuncs       `alias:"handlerfuncs"`
	ContextPool        *sync.Pool `alias:"contextpool"`
	CancelError        error      `alias:"cancelerror"`
	Mutex              sync.Mutex `alias:"mutex"`
	Values             []any      `alias:"values"`
}

// The NewApp() function creates an App object, initializes various components
// of the application, and returns the App object.
func NewApp() *App {
	app := &App{}
	app.GetWrap = NewGetWrapWithApp(app)
	app.HandlerFuncs = HandlerFuncs{app.serveContext}
	app.Context, app.CancelFunc = context.WithCancel(context.Background())
	app.SetValue(ContextKeyLogger, NewLogger(nil))
	app.SetValue(ContextKeyConfig, NewConfig(nil))
	app.SetValue(ContextKeyRouter, NewRouter(nil))
	app.SetValue(ContextKeyClient, NewClient())
	app.SetValue(ContextKeyServer, NewServer(nil))
	app.SetValue(ContextKeyContextPool, NewContextBasePool(app))
	return app
}

// The Run() method starts the application and blocks it until it is finished.
func (app *App) Run() error {
	defer app.SetValue(ContextKeyLogger, nil)
	defer app.SetValue(ContextKeyConfig, nil)
	defer app.SetValue(ContextKeyRouter, nil)
	defer app.SetValue(ContextKeyClient, nil)
	defer app.SetValue(ContextKeyServer, nil)
	defer func() {
		for i := len(app.Values) - 2; i > -1; i -= 2 {
			app.SetValue(app.Values[i], nil)
		}

		log := app.WithField(ParamDepth, 2)
		if errors.Is(app.Err(), context.Canceled) {
			log.Info("eudore app context canceled")
		} else {
			log.Fatal("eudore app error:", app.Err())
		}
	}()
	<-app.Done()
	return app.Err()
}

// SetValue method sets the specified key value from the App.
//
// If the value implements the Mount/Unmount method,
// this method is automatically called when setting and unsetting.
func (app *App) SetValue(key, val any) {
	anyMount(app, val)
	defer anyUnmount(app, app.Value(key))
	app.Mutex.Lock()
	defer app.Mutex.Unlock()
	switch key {
	case ContextKeyLogger:
		app.Logger, _ = val.(Logger)
	case ContextKeyConfig:
		app.Config, _ = val.(Config)
	case ContextKeyClient:
		app.Client, _ = val.(Client)
	case ContextKeyServer:
		app.Server, _ = val.(Server)
	case ContextKeyRouter:
		app.Router, _ = val.(Router)
	case ContextKeyContextPool:
		app.ContextPool, _ = val.(*sync.Pool)
	case ContextKeyError:
		if val != nil {
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
				app.Values[i+1] = val
				return
			}
		}
		app.Values = append(app.Values, key, val)
	}
}

// Value method gets the specified key value from the App.
func (app *App) Value(key any) any {
	app.Mutex.Lock()
	defer app.Mutex.Unlock()
	switch key {
	case ContextKeyApp:
		return app
	case ContextKeyAppCancel:
		return app.CancelFunc
	case ContextKeyLogger:
		return app.Logger
	case ContextKeyConfig:
		return app.Config
	case ContextKeyClient:
		return app.Client
	case ContextKeyServer:
		return app.Server
	case ContextKeyRouter:
		return app.Router
	case ContextKeyAppValues:
		vals := make([]any, 0, 12+len(app.Values))
		vals = append(vals,
			ContextKeyApp, app,
			ContextKeyLogger, app.Logger,
			ContextKeyConfig, app.Config,
			ContextKeyRouter, app.Router,
			ContextKeyClient, app.Client,
			ContextKeyServer, app.Server,
		)
		vals = append(vals, app.Values...)
		return vals
	}
	for i := 0; i < len(app.Values); i += 2 {
		if app.Values[i] == key {
			return app.Values[i+1]
		}
	}
	return app.Context.Value(key)
}

// Err method returns an error at the end of the App Context.
func (app *App) Err() error {
	app.Mutex.Lock()
	defer app.Mutex.Unlock()
	if app.CancelError != nil {
		return app.CancelError
	}
	return app.Context.Err()
}

func anyMount(ctx context.Context, i any) {
	loader, ok := i.(interface{ Mount(ctx context.Context) })
	if ok {
		loader.Mount(ctx)
	}
}

func anyUnmount(ctx context.Context, i any) {
	closer, ok := i.(interface{ Unmount(ctx context.Context) })
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

// The ServeHTTP method implements the [http.Handler] interface to process
// http requests.
//
// Create and initialize a [Context], set [App.HandlerFuncs] as the
// global [HandlerFuncs] of [Context].
// When [App.HandlerFuncs] is last processed,
// the app.serveContext method is called,
// Use [App.Router] to match the route processing function
// of this request for secondary request processing.
func (app *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pool := app.ContextPool
	ctx := pool.Get().(Context)
	ctx.Reset(w, r)
	ctx.SetHandlers(-1, app.HandlerFuncs)
	ctx.Next()
	pool.Put(ctx)
}

// serveContext Implement the request context function.
func (app *App) serveContext(ctx Context) {
	ctx.SetHandlers(-1, app.Router.Match(ctx.Method(), ctx.Path(), ctx.Params()))
	ctx.Next()
}

// The AddMiddleware method implements registration of global Middleware,
// which runs before the [router.Match].
//
// If the first parameter of the method is the string "global",
// it will be added to the App as a global middleware,
// using [NewHandlerExtenderWithContext] to create request a [HandlerFuncs],
// otherwise it is calling the [Router.AddMiddleware] method.
//
// refer [Router].AddMiddleware.
func (app *App) AddMiddleware(hs ...any) error {
	if len(hs) > 1 {
		name, ok := hs[0].(string)
		if ok && name == "global" {
			hs := NewHandlerExtenderWithContext(app).CreateHandlers("", hs[1:])
			app.WithField(ParamDepth, 1).Info(
				"register app global middleware:", hs,
			)
			app.HandlerFuncs = NewHandlerFuncsCombine(
				app.HandlerFuncs[0:len(app.HandlerFuncs)-1],
				append(hs, app.HandlerFuncs[len(app.HandlerFuncs)-1]),
			)
			return nil
		}
	}
	return app.Router.AddMiddleware(hs...)
}

// Listen method listens to an http port.
func (app *App) Listen(addr string) error {
	conf := ServerListenConfig{
		Addr: addr,
	}
	ln, err := conf.Listen()
	if err != nil {
		app.Error(err)
		return err
	}
	app.WithField(ParamDepth, 1).Infof(
		"listen http in %s %s",
		ln.Addr().Network(), ln.Addr().String(),
	)
	app.Serve(ln)
	return nil
}

// ListenTLS method listens to an https port, if h2 is enabled by default.
func (app *App) ListenTLS(addr, cert, key string) error {
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
	app.WithField(ParamDepth, 1).Infof(
		"listen https in %s %s, host name: %v",
		ln.Addr().Network(), ln.Addr().String(), conf.Certificate.DNSNames,
	)
	app.Serve(ln)
	return nil
}

// Serve method starts a [Server] monitor non-blocking, and uses the app to
// process the monitor and return an error.
func (app *App) Serve(ln net.Listener) {
	srv := app.Server
	go func() {
		app.SetValue(ContextKeyError, srv.Serve(ln))
	}()
}

// The Parse method calls [Config].Parse and handles errors.
// Use the current [App] as [context.Context].
//
// The Parse method executes all [ConfigParseFunc].
// If the parsing function returns error, it stops parsing and returns error.
func (app *App) Parse() error {
	err := app.Config.Parse(app)
	if err != nil {
		app.SetValue(ContextKeyError, err)
	}
	return app.Err()
}
