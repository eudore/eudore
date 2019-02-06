package eudore


import (
	"sync"
	"net/http"
)

type (
	Core struct {
		*App
		poolctx sync.Pool
		poolreq	sync.Pool
		poolresp sync.Pool
		ports []*ServerConfigGeneral
	}
)

func NewCore() *Core {
	app := &Core{
		App:		NewApp(),
		poolctx:	sync.Pool{},
		poolreq:	sync.Pool{
			New: 	func() interface{} {
				return NewRequestReaderHttp(nil)
			},
		},
		poolresp:	sync.Pool{
			New:	func() interface{} {
				return NewResponseWriterHttp(nil)
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
	// start server
	return app.Server.Start()
}


func (app *Core) Listen(addr string) *Core {
	app.RegisterComponent("server-std", &ServerConfigGeneral{
		Addr:		addr,
		Http2:		false,
		Handler:	app,
	})
	return app
}

func (app *Core) ListenTLS(addr, key, cert string) *Core {
	app.ports = append(app.ports, &ServerConfigGeneral{
		Addr:		addr,
		Http2:		true,
		Keyfile:	key,
		Certfile:	cert,
		Handler:	app,
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
	ctx.Reset(req.Context(), response, request)
	// match router
	fn, routepath := app.Router.Match(ctx.Method(), ctx.Path(), ctx.Params())
	ctx.SetParam("route", routepath)
	// handle
	ctx.Handles(fn)
	ctx.Next()
	// clean
	app.poolreq.Put(request)
	app.poolresp.Put(response)
	app.poolctx.Put(ctx)
}