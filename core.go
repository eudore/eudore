package eudore

/*
Core是组合App对象后的一种实例化，用于启动主程序。
Core的特点是简单。
*/

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type (
	// Core 定义Core对象，是App对象的一种实例化。
	Core struct {
		*App
		GetWarp
		wg sync.WaitGroup
	}
)

// NewCore 创建一个Core对象，并使用默认组件初始化。
func NewCore() *Core {
	app := NewApp()
	return &Core{App: app, GetWarp: NewGetWarpWithApp(app)}
}

// Run 方法初始化日志输出然后启动Core，Config需要手动调用Parse方法。
func (app *Core) Run() (err error) {
	go func(app *App) {
		ticker := time.NewTicker(time.Millisecond * 40)
		for range ticker.C {
			app.Logger.Sync()
		}
	}(app.App)

	defer app.Logger.Sync()
	if initlog, ok := app.Logger.(LoggerInitHandler); ok {
		app.Logger, _ = NewLoggerStd(nil)
		initlog.NextHandler(app.Logger)
		app.Logger.Sync()
	}
	// 解析配置
	err = app.Config.Parse()
	if err != nil {
		app.Error(err)
		return err
	}

	app.Server.SetHandler(app)
	// 等一下让go Serve启动
	time.Sleep(time.Millisecond * 100)
	app.wg.Wait()
	return nil
}

// Listen 方法监听一个http端口
func (app *Core) Listen(addr string) error {
	conf := ServerListenConfig{
		Addr: addr,
	}
	ln, err := conf.Listen()
	if err != nil {
		app.Error(err)
		return err
	}
	app.Logger.Infof("listen %s %s", ln.Addr().Network(), ln.Addr().String())
	go app.Serve(ln)
	return nil
}

// ListenTLS 方法监听一个https端口，如果支持默认开启h2
func (app *Core) ListenTLS(addr, key, cert string) error {
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
	go app.Serve(ln)
	return nil
}

// Serve 方法阻塞启动一个监听服务。
func (app *Core) Serve(ln net.Listener) error {
	app.wg.Add(1)
	err := app.Server.Serve(ln)
	if err != nil {
		app.Error(err)
	}
	app.wg.Done()
	return err
}

// ServeHTTP 方法实现http.Handler接口，处理http请求。
func (app *Core) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// init
	ctx := app.ContextPool.Get().(Context)
	response := responseWriterHTTPPool.Get().(*ResponseWriterHTTP)
	// handle
	response.Reset(w)
	ctx.Reset(r.Context(), response, r)
	ctx.SetHandler(app.Router.Match(ctx.Method(), ctx.Path(), ctx.Params()))
	ctx.Next()
	ctx.End()
	// release
	responseWriterHTTPPool.Put(response)
	app.ContextPool.Put(ctx)

}
