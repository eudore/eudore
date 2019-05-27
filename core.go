package eudore

/*
Core是组合App对象后的一种实例化，用于启动主程序。
*/

import (
	// "fmt"
	"sync"
	"net/http"
	"context"
	"github.com/eudore/eudore/protocol"
)

type (
	Core struct {
		*App
		Poolctx sync.Pool
		Poolreq	sync.Pool
		Poolresp sync.Pool
	}
)

func NewCore() *Core {
	app := &Core{
		App:		NewApp(),
		Poolctx:	sync.Pool{},
		Poolreq:	sync.Pool{
			New: 	func() interface{} {
				return &RequestReaderHttp{}
			},
		},
		Poolresp:	sync.Pool{
			New:	func() interface{} {
				return &ResponseWriterHttp{}
			},
		},
	}
	
	app.Poolctx.New = func() interface{} {
		return NewContextBase(app.App)
	}

	// 初始化组件
	app.RegisterComponents(
		[]string{"logger", "config", "router", "server", "cache", "session", "view"}, 
		[]interface{}{nil, nil, nil, nil, nil, nil, nil},
	)
	return app
}

// 加载配置然后启动Core。
func (app *Core) Run() (err error) {
	// parse config
	err = app.Config.Parse()
	if err != nil {
		return
	}
	
	// start serverv
	ComponentSet(app.Server, "handler", app)
	if err != nil {
		return
	}
	return app.Server.Start()
}

// 监听一个http端口
func (app *Core) Listen(addr string) *Core {
	ComponentSet(app.Server, "config.listeners.+", 	&ServerListenConfig{
		Addr:		addr,
	})
	return app
}


// 监听一个https端口，如果支持默认开启h2
func (app *Core) ListenTLS(addr, key, cert string) *Core {
	ComponentSet(app.Server, "config.listeners.+", 	&ServerListenConfig{
		Addr:		addr,
		Https:		true,
		Http2:		true,
		Keyfile:	key,
		Certfile:	cert,
	})
	return app
}

func (app *Core) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := app.Poolctx.Get().(Context)
	request := app.Poolreq.Get().(*RequestReaderHttp)
	response := app.Poolresp.Get().(*ResponseWriterHttp)
	// init
	ResetRequestReaderHttp(request, req)
	ResetResponseWriterHttp(response, w)
	// handle
	ctx.Reset(req.Context(), response, request)
	ctx.SetHandler(app.Router.Match(ctx.Method(), ctx.Path(), ctx))
	ctx.Next()
	// clean
	app.Poolreq.Put(request)
	app.Poolresp.Put(response)
	app.Poolctx.Put(ctx)
}


func (app *Core) EudoreHTTP(pctx context.Context,w protocol.ResponseWriter, req protocol.RequestReader) {
	// init
	ctx := app.Poolctx.Get().(Context)
	// handle
	ctx.Reset(pctx, w, req)
	ctx.SetHandler(app.Router.Match(ctx.Method(), ctx.Path(), ctx))
	ctx.Next()
	// release
	app.Poolctx.Put(ctx)
}
