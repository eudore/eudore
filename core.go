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
		poolctx sync.Pool
		poolreq	sync.Pool
		poolresp sync.Pool
	}
)

func NewCore() *Core {
	app := &Core{
		App:		NewApp(),
		poolctx:	sync.Pool{},
		poolreq:	sync.Pool{
			New: 	func() interface{} {
				return &RequestReaderHttp{}
			},
		},
		poolresp:	sync.Pool{
			New:	func() interface{} {
				return &ResponseWriterHttp{}
			},
		},
	}
	
	app.poolctx.New = func() interface{} {
		return &ContextHttp{
			app:	app.App,
			fields:	make(Fields, 5),
		}
	}

	// 初始化组件
	app.RegisterComponents(
		[]string{"logger", "config", "router", "server", "cache", "view"}, 
		[]interface{}{nil, nil, nil, nil, nil, nil},
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
	// init sync.Pool
	if fn, ok := app.Pools["context"];ok {
		app.poolctx.New = fn
	}
	if fn, ok := app.Pools["request"];ok {
		app.poolreq.New = fn
	}
	if fn, ok := app.Pools["response"];ok {
		app.poolresp.New = fn
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
	ctx := app.poolctx.Get().(Context)
	request := app.poolreq.Get().(*RequestReaderHttp)
	response := app.poolresp.Get().(*ResponseWriterHttp)
	// init
	ResetRequestReaderHttp(request, req)
	ResetResponseWriterHttp(response, w)
	// handle
	ctx.Reset(req.Context(), response, request)
	ctx.SetHandler(app.Router.Match(ctx.Method(), ctx.Path(), ctx))
	ctx.Next()
	// clean
	app.poolreq.Put(request)
	app.poolresp.Put(response)
	app.poolctx.Put(ctx)
}


func (app *Core) EudoreHTTP(pctx context.Context,w protocol.ResponseWriter, req protocol.RequestReader) {
	// init
	ctx := app.poolctx.Get().(Context)
	// handle
	ctx.Reset(pctx, w, req)
	ctx.SetHandler(app.Router.Match(ctx.Method(), ctx.Path(), ctx))
	ctx.Next()
	// release
	app.poolctx.Put(ctx)
}
