package eudore


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
		// ports []*ServerConfigGeneral
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
		}
	}

	app.RegisterComponents(
		[]string{"logger", "config", "router", "server", "cache"}, 
		[]interface{}{nil, nil, nil, nil, nil},
	)
	return app
}

func (app *Core) Run() (err error) {
	// parse config
	err = app.Config.Parse()
	if err != nil {
		return
	}
	// read and set server config
/*	server := app.Config.Get("#component.server")
	err = app.Server.Register(server)
	for _, p := range app.ports {
		app.Server.Register(p)
	}
	if err != nil {
		return
	}*/
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
/*	switch len(app.ports) {
	case 0:
		// 未注册Server信息
		return fmt.Errorf("Undefined Server component, Please Listen or ListenTLS.")
	case 1:
		// 单端口启动
		err = app.RegisterComponent("server-std", app.ports[0])
	default:
		// 多端口启动
		err = app.RegisterComponent(ComponentServerMultiName, app.ports)
	}*/
	ComponentSet(app.Server, "handler", app)
	if err != nil {
		return
	}
	return app.Server.Start()
}


func (app *Core) Listen(addr string) *Core {
	ComponentSet(app.Server, "config.listeners.+", 	&ServerListenConfig{
		Addr:		addr,
	})

/*	app.ports = append(app.ports, &ServerConfigGeneral{
		Addr:		addr,
		Http2:		false,
		Handler:	app,
	})*/
	return app
}

func (app *Core) ListenTLS(addr, key, cert string) *Core {
/*	app.ports = append(app.ports, &ServerConfigGeneral{
		Addr:		addr,
		Http2:		true,
		Keyfile:	key,
		Certfile:	cert,
		Handler:	app,
	})*/	
	ComponentSet(app.Server, "config.listeners.+", 	&ServerListenConfig{
		Addr:		addr,
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
