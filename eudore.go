package eudore

/*
Eudore是组合App对象后的一种实例化，用于启动主程序。
*/

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"sync"
	"time"
)

type (
	// Eudore 定义Eudore App对象。
	Eudore struct {
		*App
		GetWarp
		Signaler
		err       error
		mu        sync.Mutex
		inits     map[string]initInfo
		Handlers  HandlerFuncs
		listeners []net.Listener
	}
	// EudoreFunc 定义Eudore app处理函数。
	EudoreFunc func(*Eudore) error
	// initInfo 保存初始化函数的信息。
	initInfo struct {
		name  string
		index int
		fn    EudoreFunc
	}
	// Signaler 定义一个信号处理对象。
	Signaler struct {
		signalChan  chan os.Signal
		signalFuncs map[os.Signal][]func() error
	}
)

// NewEudore Create a new Eudore app.
//
// 默认注册Eudore初始化函数，建议重新注册修改eudore-server为最后一个Init函数。
//
// app.RegisterInit("eudore-config", 0x003, InitConfig)
//
// app.RegisterInit("eudore-workdir", 0x006, InitWorkdir)
//
// app.RegisterInit("eudore-signal", 0x009, InitSignal)
//
// app.RegisterInit("eudore-logger", 0x00c, InitLogger)
//
// app.RegisterInit("eudore-server", 0x00f, InitServer)
//
func NewEudore(options ...interface{}) *Eudore {
	app := &Eudore{
		App:   NewApp(),
		inits: make(map[string]initInfo),
		Signaler: Signaler{
			signalChan:  make(chan os.Signal),
			signalFuncs: make(map[os.Signal][]func() error),
		},
	}
	app.Options(options...)
	app.GetWarp = NewGetWarpWithApp(app.App)
	app.Context = context.WithValue(app.Context, AppContextKey, app)
	app.Handlers = HandlerFuncs{app.HandleContext}

	Set(app.Config, "print", NewPrintFunc(app.App))
	Set(app.Router, "print", NewPrintFunc(app.App))
	Set(app.Server, "print", NewPrintFunc(app.App))

	// Register eudore default reload func
	app.RegisterInit("eudore-config", 0x003, InitConfig)
	app.RegisterInit("eudore-workdir", 0x006, InitWorkdir)
	app.RegisterInit("eudore-signal", 0x009, InitSignal)
	app.RegisterInit("eudore-logger", 0x00c, InitLogger)
	app.RegisterInit("eudore-server", 0x00f, InitServer)
	go app.handlerChannel()
	return app
}

func (app *Eudore) handlerChannel() {
	ticker := time.NewTicker(time.Millisecond * 40)
	defer ticker.Stop()
	for {
		select {
		case <-app.Done():
			return
		case sig := <-app.signalChan:
			err := app.HandleSignal(sig)
			if err != nil {
				app.Error(err)
			}
		case <-ticker.C:
			app.Logger.Sync()
		}
	}
}

// Options 方法加载eudore app组件。
func (app *Eudore) Options(options ...interface{}) {
	// init options
	for _, i := range options {
		switch val := i.(type) {
		case context.Context:
			app.Context = val
			app.Context, app.CancelFunc = context.WithCancel(app.Context)
		case Config:
			app.Config = val
			Set(app.Config, "print", NewPrintFunc(app.App))
		case Logger:
			app.Logger = val
		case Server:
			app.Server = val
			Set(app.Server, "print", NewPrintFunc(app.App))
		case Router:
			app.Router = val
			Set(app.Router, "print", NewPrintFunc(app.App))
		case Binder:
			app.Binder = val
		case Renderer:
			app.Renderer = val
		default:
			app.Logger.Warningf("eudore app init invalid option: %v", i)
		}
	}
}

// Run 方法加载配置，然后启动全部初始化函数，等待App结束。
func (app *Eudore) Run() error {
	go app.InitAll()
	<-app.Done()

	// 处理后续日志
	if initlog, ok := app.Logger.(loggerInitHandler); ok {
		app.Logger = NewLoggerStd(nil)
		initlog.NextHandler(app.Logger)
	}
	time.Sleep(100 * time.Millisecond)
	app.Logger.Sync()
	time.Sleep(50 * time.Millisecond)
	return app.Err()
}

// InitAll 方法调用全部初始化函数。
func (app *Eudore) InitAll() error {
	app.Logger.Info("eudore start init all func")
	err := app.Init()
	app.HandleError(err)
	if err == nil {
		app.Logger.Info("eudore init all success.")
	}
	return err
}

// Init execute the eudore init function.
//
// Init 方法执行eudore全部初始函数。
func (app *Eudore) Init() (err error) {
	app.mu.Lock()
	defer app.mu.Unlock()
	// get names and exec
	names := app.getInitNames()
	var (
		i    int
		name string
		num  = len(names)
	)
	defer func() {
		rerr := recover()
		if rerr != nil {
			err = fmt.Errorf("eudore init %d/%d %s recover error: %v", i+1, num, name, rerr)
			app.WithField("stack", getPanicStakc(3)).Error(err)
		}
	}()
	for i, name = range names {
		err = app.inits[name].fn(app)
		if err != nil {
			if err == ErrEudoreInitIgnore {
				app.Logger.Errorf("eudore init %d/%d %s ignore the remaining init function.", i+1, num, name)
				return nil
			}
			return fmt.Errorf("eudore init error: %v", err)
		}
		app.Logger.Infof("eudore init %d/%d %s success.", i+1, num, name)
	}
	return nil
}

// getInitNames 方法获得全部names并排序。
func (app *Eudore) getInitNames() []string {
	names := make([]string, 0, len(app.inits))
	for k := range app.inits {
		names = append(names, k)
	}
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
	err := startNewProcess(app.listeners)
	if err == nil {
		app.Logger.Info("eudore restart success.")
		app.Server.Shutdown(app.Context)
		app.HandleError(ErrApplicationStop)
	}
	return err
}

// Shutdown 方法正常退出关闭app。
func (app *Eudore) Shutdown() error {
	app.mu.Lock()
	defer app.mu.Unlock()
	defer app.HandleError(ErrApplicationStop)
	return app.Server.Shutdown(app.Context)
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

// HandleError 定义Eudore App处理一个error，如果err非空则结束app，当err为ErrApplicationStop正常退出。
func (app *Eudore) HandleError(err error) {
	if err != nil && app.err == nil {
		if err != ErrApplicationStop {
			app.err = err
			app.Logger.Errorf("eudore stop error: %s", err.Error())
		} else {
			app.Logger.Info("eudore stop success.")
		}
		app.CancelFunc()
	}
}

// Err 实现Context.Errr()返回error，如果app.err为空返回app.Context.Err()。
func (app *Eudore) Err() error {
	if app.err != nil {
		return app.err
	}
	return app.Context.Err()
}

// Listen 监听一个http端口
func (app *Eudore) Listen(addr string) error {
	conf := ServerListenConfig{
		Addr: addr,
	}
	ln, err := conf.Listen()
	if err != nil {
		app.Error(err)
		return err
	}
	app.Serve(ln)
	return nil
}

// ListenTLS 监听一个https端口，如果支持默认开启h2
func (app *Eudore) ListenTLS(addr, key, cert string) error {
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
	app.Serve(ln)
	return nil
}

// Serve 方式给Server添加一个net.Listener,同时会记录net.Listener对象，用于热重启传递fd。
func (app *Eudore) Serve(ln net.Listener) {
	app.Logger.Infof("listen %s %s", ln.Addr().Network(), ln.Addr().String())
	app.listeners = append(app.listeners, ln)
	go func() {
		app.HandleError(app.Server.Serve(ln))
	}()
}

// AddGlobalMiddleware 给eudore添加全局中间件，会在Router.Match前执行。
func (app *Eudore) AddGlobalMiddleware(hs ...HandlerFunc) {
	app.Handlers = HandlerFuncsCombine(app.Handlers[0:len(app.Handlers)-1], hs)
	app.Handlers = HandlerFuncsCombine(app.Handlers, HandlerFuncs{app.HandleContext})
}

// HandleContext 实现处理请求上下文函数。
func (app *Eudore) HandleContext(ctx Context) {
	ctx.SetHandler(-1, app.Router.Match(ctx.Method(), ctx.Path(), ctx.Params()))
	ctx.Next()
}

// ServeHTTP 实现http.Handler接口，处理http请求。
func (app *Eudore) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// init
	ctx := app.ContextPool.Get().(Context)
	response := responseWriterHTTPPool.Get().(*ResponseWriterHTTP)
	// handle
	response.Reset(w)
	ctx.Reset(r.Context(), response, r)
	ctx.SetHandler(-1, app.Handlers)
	ctx.Next()
	ctx.End()
	// release
	responseWriterHTTPPool.Put(response)
	app.ContextPool.Put(ctx)
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

// Fatal 方法输出Fatal级别日志。
func (app *Eudore) Fatal(args ...interface{}) {
	app.logReset().Fatal(args...)
	time.Sleep(90 * time.Millisecond)
	err := fmt.Sprintln(args...)
	panic(err[:len(err)-1])
}

// Debugf 方法输出Debug级别日志。
func (app *Eudore) Debugf(format string, args ...interface{}) {
	app.logReset().Debugf(format, args...)
}

// Infof 方法输出Info级别日志。
func (app *Eudore) Infof(format string, args ...interface{}) {
	app.logReset().Infof(format, args...)
}

// Warningf 方法输出Warning级别日志。
func (app *Eudore) Warningf(format string, args ...interface{}) {
	app.logReset().Warningf(format, args...)
}

// Errorf 方法输出Error级别日志。
func (app *Eudore) Errorf(format string, args ...interface{}) {
	app.logReset().Errorf(format, args...)
}

// Fatalf 方法输出Error级别日志。
func (app *Eudore) Fatalf(format string, args ...interface{}) {
	app.logReset().Errorf(format, args...)
	time.Sleep(90 * time.Millisecond)
	panic(fmt.Sprintf(format, args...))
}

func (app *Eudore) logReset() Logout {
	file, line := logFormatFileLine(3)
	f := Fields{
		"file": file,
		"line": line,
	}
	return app.Logger.WithFields(f)
}

// NewSignaler 函数创建一个信号处理者。
func NewSignaler() *Signaler {
	return &Signaler{
		signalChan:  make(chan os.Signal),
		signalFuncs: make(map[os.Signal][]func() error),
	}
}

// HandleSignal 方法执行对应信号应该函数。
func (s *Signaler) HandleSignal(sig os.Signal) error {
	fns, ok := s.signalFuncs[sig]
	if ok {
		for _, fn := range fns {
			err := fn()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// RegisterSignal 方法注册一个信号响应函数。
func (s *Signaler) RegisterSignal(sig os.Signal, fn func() error) {
	fns, ok := s.signalFuncs[sig]
	s.signalFuncs[sig] = append(fns, fn)
	if !ok {
		sigs := make([]os.Signal, 0, len(s.signalFuncs))
		for i := range s.signalFuncs {
			sigs = append(sigs, i)
		}

		signal.Stop(s.signalChan)
		signal.Notify(s.signalChan, sigs...)
	}
}

// Run 方法执行Signaler信号响应处理。
func (s *Signaler) Run(ctx context.Context) {
	defer signal.Stop(s.signalChan)
	for {
		select {
		case <-ctx.Done():
			return
		case sig := <-s.signalChan:
			s.HandleSignal(sig)
		}
	}
}
