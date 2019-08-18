package eudore

/*
Core是组合App对象后的一种实例化，用于启动主程序。
*/

import (
	"context"

	"github.com/eudore/eudore/protocol"
)

type (
	// Core 定义Core对象，是App对象的一种实例化。
	Core struct {
		*App
	}
)

// NewCore 创建一个Core对象，并使用默认组件初始化。
func NewCore() *Core {
	return &Core{App: NewApp()}
}

// Run 加载配置然后启动Core。
func (app *Core) Run() (err error) {
	defer app.Logger.Sync()
	if initlog, ok := app.Logger.(LoggerInitHandler); ok {
		app.Logger, _ = NewLoggerStd(nil)
		initlog.NextHandler(app.Logger)
	}
	// start server
	app.Server.AddHandler(app)
	defer app.Logger.Sync()
	return app.Server.Start()
}

// Listen 监听一个http端口
func (app *Core) Listen(addr string) *Core {
	conf := ServerListenConfig{
		Addr: addr,
	}
	ln, err := conf.Listen()
	if err != nil {
		app.Error(err)
	}
	app.Server.AddListener(ln)
	return app
}

// ListenTLS 监听一个https端口，如果支持默认开启h2
func (app *Core) ListenTLS(addr, key, cert string) *Core {
	conf := ServerListenConfig{
		Addr:     addr,
		Https:    true,
		Http2:    true,
		Keyfile:  key,
		Certfile: cert,
	}
	ln, err := conf.Listen()
	if err != nil {
		app.Error(err)
	}
	app.Server.AddListener(ln)
	return app
}

// EudoreHTTP 方法实现protocol.HandlerHttp接口，处理http请求。
func (app *Core) EudoreHTTP(pctx context.Context, w protocol.ResponseWriter, req protocol.RequestReader) {
	// init
	ctx := app.ContextPool.Get().(Context)
	// handle
	ctx.Reset(pctx, w, req)
	ctx.SetHandler(app.Router.Match(ctx.Method(), ctx.Path(), ctx))
	ctx.Next()
	ctx.End()
	// release
	app.ContextPool.Put(ctx)
}
