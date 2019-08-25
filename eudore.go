package eudore

/*
Eudore是组合App对象后的一种实例化，用于启动主程序。
*/

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sort"
	"sync"
	"time"

	"github.com/eudore/eudore/protocol"
)

type (
	// Eudore 定义Eudore App对象。
	Eudore struct {
		*App
		cancel      context.CancelFunc
		err         error
		mu          sync.Mutex
		inits       map[string]initInfo
		handlers    HandlerFuncs
		listeners   []net.Listener
		signalChan  chan os.Signal
		signalFuncs map[os.Signal][]EudoreFunc
	}
	// EudoreFunc 定义Eudore app处理函数。
	EudoreFunc func(*Eudore) error
	// initInfo 保存初始化函数的信息。
	initInfo struct {
		name  string
		index int
		fn    EudoreFunc
	}
)

// NewEudore Create a new Eudore.
func NewEudore(options ...interface{}) *Eudore {
	app := &Eudore{
		App:         NewApp(),
		inits:       make(map[string]initInfo),
		signalChan:  make(chan os.Signal),
		signalFuncs: make(map[os.Signal][]EudoreFunc),
	}
	app.Context, app.cancel = context.WithCancel(app.Context)
	app.handlers = HandlerFuncs{app.HandleContext}

	// init options
	for _, i := range options {
		switch val := i.(type) {
		case Config:
			app.Config = val
		case Logger:
			app.Logger = val
		case Server:
			app.Server = val
		case Router:
			app.Router = val
		case Binder:
			app.Binder = val
		case Renderer:
			app.Renderer = val
		case PoolGetFunc:
			app.ContextPool.New = val
		case error:
			app.HandleError(val)
		default:
			app.Logger.Warningf("eudore app unid option: %v", i)

		}
	}
	Set(app.Router, "print", NewLoggerPrintFunc(app.Logger))
	Set(app.Server, "print", NewLoggerPrintFunc(app.Logger))

	// Register eudore default reload func
	app.RegisterInit("eudore-config", 0x008, InitConfig)
	app.RegisterInit("eudore-workdir", 0x009, InitWorkdir)
	app.RegisterInit("eudore-signal", 0x57, InitSignal)
	app.RegisterInit("eudore-server-start", 0xff0, InitServerStart)
	go func(app *Eudore) {
		ticker := time.NewTicker(time.Millisecond * 40)
		defer ticker.Stop()
		for {
			select {
			case <-app.Done():
				return
			case sig := <-app.signalChan:
				app.HandleSignal(sig)
			case <-ticker.C:
				app.Logger.Sync()
			}
		}
	}(app)
	return app
}

// Run method parse the current command, if the command is 'start', start eudore.
//
// Run 解析当前命令，如果命令是启动，则启动eudore。
func (app *Eudore) Run() error {
	return app.Start()
}

// Start 方法加载配置，然后启动全部初始化函数，等待App结束。
func (app *Eudore) Start() error {
	// Reload
	go func() {
		app.Logger.Info("eudore start reload all func")
		app.HandleError(app.Init())
	}()

	defer app.Logger.Sync()
	<-app.Done()
	return app.Err()
}

// Init execute the eudore reload function.
// names are a list of function names that need to be executed; if the list is empty, execute all reload functions.
//
// Init 执行eudore重新加载函数。
// names是需要执行的函数名称列表；如果列表为空，执行全部重新加载函数。
func (app *Eudore) Init(names ...string) (err error) {
	app.mu.Lock()
	// get names
	names = app.getInitNames(names)
	// exec
	num := len(names)
	for i, name := range names {
		if err = app.inits[name].fn(app); err != nil {
			return fmt.Errorf("eudore init %d/%d %s error: %v", i+1, num, name, err)
		}
		app.Logger.Infof("eudore init %d/%d %s success.", i+1, num, name)
	}
	app.Logger.Info("eudore init all success.")
	app.mu.Unlock()
	return nil
}

// getInitNames 处理名称并对reloads排序。
func (app *Eudore) getInitNames(names []string) []string {
	// get not names
	notnames := eachstring(names, func(name string) string {
		if len(name) > 0 && name[0] == '!' {
			return name[1:]
		}
		return ""
	})

	// get names
	names = eachstring(names, func(name string) string {
		if len(name) == 0 || name[0] == '!' {
			return ""
		}
		return name
	})

	// set default name
	if len(names) == 0 {
		names = make([]string, 0, len(app.inits))
		for k := range app.inits {
			names = append(names, k)
		}
	}

	// filter
	names = eachstring(names, func(name string) string {
		// filter not name
		for _, i := range notnames {
			if i == name {
				return ""
			}
		}
		// filter invalid name
		if _, ok := app.inits[name]; !ok {
			app.Warning("Invalid overloaded function name: ", name)
			return ""
		}
		return name
	})

	// sort index
	sort.Slice(names, func(i, j int) bool {
		return app.inits[names[i]].index < app.inits[names[j]].index
	})
	return names
}

// Restart Eudore
// Invoke ServerManager Restart
func (app *Eudore) Restart() error {
	app.mu.Lock()
	defer app.mu.Unlock()
	err := StartNewProcess(app.listeners)
	if err == nil {
		app.Server.Shutdown(context.Background())
	}
	return err
}

// Close 方法关闭app。
func (app *Eudore) Close() error {
	app.mu.Lock()
	defer app.mu.Unlock()
	return app.Server.Close()
}

// Shutdown 方法正常退出关闭app。
func (app *Eudore) Shutdown() error {
	app.mu.Lock()
	defer app.mu.Unlock()
	return app.Server.Shutdown(context.Background())
}

// RegisterInit method register a Init function, index determines the function loading order, and name is used for a specific load function.
//
// RegisterInit 注册一个初始化函数，index决定函数加载顺序，name用于特定加载函数。
func (app *Eudore) RegisterInit(name string, index int, fn EudoreFunc) {
	if name != "" {
		if fn == nil {
			delete(app.inits, name)
		} else {
			app.inits[name] = initInfo{name, index, fn}
		}
	}
}

// HandleSignal 方法执行对应信号应该函数。
func (app *Eudore) HandleSignal(sig os.Signal) {
	fns, ok := app.signalFuncs[sig]
	if ok {
		for _, fn := range fns {
			err := fn(app)
			if err != nil {
				app.Logger.Error(err)
			}
		}
	}
}

// RegisterSignal 方法给Eudore app注册一个信号响应函数。
func (app *Eudore) RegisterSignal(sig os.Signal, fn EudoreFunc) {
	fns, ok := app.signalFuncs[sig]
	if !ok {
		sigs := make([]os.Signal, 0, len(app.signalFuncs))
		for i := range app.signalFuncs {
			sigs = append(sigs, i)
		}

		signal.Stop(app.signalChan)
		signal.Notify(app.signalChan, sigs...)
	}
	app.signalFuncs[sig] = append(fns, fn)
}

// Listen 监听一个http端口
func (app *Eudore) Listen(addr string) *Eudore {
	conf := ServerListenConfig{
		Addr: addr,
	}
	ln, err := conf.Listen()
	if err != nil {
		app.Error(err)
	}
	app.AddListener(ln)
	return app
}

// ListenTLS 监听一个https端口，如果支持默认开启h2
func (app *Eudore) ListenTLS(addr, key, cert string) *Eudore {
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
	app.AddListener(ln)
	return app
}

// AddListener 方式给Server添加一个net.Listener,同时会记录net.Listener对象，用于热重启传递fd。
func (app *Eudore) AddListener(l net.Listener) {
	app.listeners = append(app.listeners, l)
	app.Server.AddListener(l)
}

// AddStatic method register a static file Handle.
func (app *Eudore) AddStatic(path, dir string) {
	app.Router.GetFunc(path, func(ctx Context) {
		ctx.WriteFile(dir + ctx.Path())
	})
}

// AddGlobalMiddleware 给eudore添加全局中间件，会在Router.Match前执行。
func (app *Eudore) AddGlobalMiddleware(hs ...HandlerFunc) {
	app.handlers = CombineHandlerFuncs(app.handlers[0:len(app.handlers)-1], hs)
	app.handlers = CombineHandlerFuncs(app.handlers, HandlerFuncs{app.HandleContext})
}

// HandleContext 实现处理请求上下文函数。
func (app *Eudore) HandleContext(ctx Context) {
	ctx.SetHandler(app.Router.Match(ctx.Method(), ctx.Path(), ctx))
	ctx.Next()
	ctx.End()
}

// Debug 方法输出Debug级别日志。
func (app *Eudore) Debug(args ...interface{}) {
	app.logReset().Debug(args...)
}

// Info 方法输出Info级别日志。
func (app *Eudore) Info(args ...interface{}) {
	app.logReset().Info(args...)
}

// Warning 方法输出Warning级别日志。
func (app *Eudore) Warning(args ...interface{}) {
	app.logReset().Warning(args...)
}

// Error 方法输出Error级别日志。
func (app *Eudore) Error(args ...interface{}) {
	app.logReset().Error(args...)
}

// Debugf 方法输出Debug级别日志。
func (app *Eudore) Debugf(format string, args ...interface{}) {
	app.logReset().Debug(fmt.Sprintf(format, args...))
}

// Infof 方法输出Info级别日志。
func (app *Eudore) Infof(format string, args ...interface{}) {
	app.logReset().Info(fmt.Sprintf(format, args...))
}

// Warningf 方法输出Warning级别日志。
func (app *Eudore) Warningf(format string, args ...interface{}) {
	app.logReset().Warning(fmt.Sprintf(format, args...))
}

// Errorf 方法输出Error级别日志。
func (app *Eudore) Errorf(format string, args ...interface{}) {
	app.logReset().Error(fmt.Sprintf(format, args...))
}

func (app *Eudore) logReset() Logout {
	file, line := LogFormatFileLine(0)
	f := Fields{
		"file": file,
		"line": line,
	}
	return app.Logger.WithFields(f)
}

// HandleError 定义Eudore App处理一个error，如果err非空则结束app，当err为ErrApplicationStop正常退出。
func (app *Eudore) HandleError(err error) {
	if err != nil {
		if err != ErrApplicationStop {
			app.err = err
			app.Logger.Error("eudore stop error: ", err)
		} else {
			app.Logger.Info("eudore stop success.")
		}
		app.cancel()
	}
}

// Err 实现Context.Errr()返回error，如果app.err为空返回app.Context.Err()。
func (app *Eudore) Err() error {
	if app.err != nil {
		return app.err
	}
	return app.Context.Err()
}

// EudoreHTTP 方法处理一个http请求。
func (app *Eudore) EudoreHTTP(pctx context.Context, w protocol.ResponseWriter, req protocol.RequestReader) {
	// init
	ctx := app.ContextPool.Get().(Context)
	// handle
	ctx.Reset(pctx, w, req)
	ctx.SetHandler(app.handlers)
	ctx.Next()
	ctx.End()
	// release
	app.ContextPool.Put(ctx)
}
