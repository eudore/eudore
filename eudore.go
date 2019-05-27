package eudore

/*
Eudore是组合App对象后的一种实例化，用于启动主程序。
*/

import (
	"os"
	"fmt"
	"time"
	"sync"
	"sort"
	"context"
	"net/http"
	"github.com/eudore/eudore/protocol"
)

type (
	// eudore 
	Eudore struct {
		*App
		Httprequest		sync.Pool
		Httpresponse	sync.Pool
		Httpcontext		sync.Pool
		inits			map[string]initInfo
		stop 			chan error
	}
	// eudore reload funcs.
	InitFunc func(*Eudore) error
	// Save reloadhook name, index fn.
	initInfo struct {
		name	string
		index	int
		fn		InitFunc
	}
)

var defaultEudore *Eudore

// Create a new Eudore.
func NewEudore(components ...ComponentConfig) *Eudore {
	app := &Eudore{
		App:			NewApp(),
		inits:			make(map[string]initInfo),
		Httprequest: sync.Pool {
			New: func() interface{} {
				return &RequestReaderHttp{}
			},
		},
		Httpresponse: sync.Pool {
			New: func() interface{} {
				return &ResponseWriterHttp{}
			},
		},
		stop: 			make(chan error, 10),
	}
	// set eudore context pool
	app.Httpcontext = sync.Pool {
		New: func() interface{} {
			return NewContextBase(app.App)
		},
	}

	// Register eudore components
	for _, config := range components {
		app.RegisterComponent(config.Name, config.Config)
	}
	app.HandleError(app.InitComponent())

	// Register eudore default reload func
	app.RegisterInit("eudore-config", 0x008, InitConfig)
	app.RegisterInit("eudore-workdir", 0x009, InitWorkdir)
	app.RegisterInit("eudore-command", 0x00a, InitCommand)
	app.RegisterInit("eudore-server", 0x015 , InitServer)
	app.RegisterInit("eudore-logger", 0x01f , InitLogger)
	app.RegisterInit("eudore-component-info", 0x54 , InitListComponent)
	app.RegisterInit("eudore-signal", 0x57 , InitSignal)
	app.RegisterInit("eudore-server-start", 0xff0 , InitServerStart)
	app.RegisterInit("eudore-test-stop", 0xfff, InitStop)
	return app
}

// Get the default eudore, if it is empty, create a new singleton.
//
// 获取默认的eudore，如果为空，创建一个新的单例。
func DefaultEudore(components ...ComponentConfig) *Eudore {
	if defaultEudore == nil {
		defaultEudore = NewEudore(components...)
	}
	return defaultEudore
}

// Parse the current command, if the command is 'start', start eudore.
//
// 解析当前命令，如果命令是启动，则启动eudore。
func (app *Eudore) Run() (err error) {
	return app.Start()
}

// Load all configurations and then start listening for all services.
//
// 加载全部配置，然后启动监听全部服务。
func (app *Eudore) Start() error {
	defer func(){
		if _, ok := app.Logger.(LoggerInitHandler); ok  {
			app.RegisterComponent("logger", nil)
		}
		time.Sleep(100 * time.Millisecond)
	}()

	// Reload
	go func(){
		app.Info("eudore start reload all func")
		app.HandleError(app.Init())
	}()
	
	// 阻塞主线程
	time.Sleep(100 * time.Millisecond)
	err := <- app.stop
	if err == nil {
		app.Info("eudore stop success.")
	}else {
		app.Error("eudore stop error: ", err)
	}
	return err
}

// Execute the eudore reload function.
// names are a list of function names that need to be executed; if the list is empty, execute all reload functions.
//
// 执行eudore重新加载函数。
// names是需要执行的函数名称列表；如果列表为空，执行全部重新加载函数。
func (app *Eudore) Init(names ...string) (err error) {
	// get names
	names = app.getInitNames(names)
	// exec
	var i int
	var name string
	num := len(names)
/*	defer func() {
		if err1 := recover(); err1 != nil {
			if err2, ok := err1.(error);ok {
				err = err2
			}else {
				err = fmt.Errorf("eudore init %s %d/%d recover error: %v", name, i + 1, num, err1)
			}
		}
	}()*/
	for i, name = range names {
		if err = app.inits[name].fn(app);err != nil {
			return fmt.Errorf("eudore init %d/%d %s error: %v",i + 1, num, name, err)
		}
		app.Infof("eudore init %d/%d %s success.", i + 1, num, name)
	}
	app.Info("eudore init all success.")
	return nil
}

// 处理名称并对reloads排序。
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
	return app.Server.Restart()
}

// Eudore Stop immediately
func (app *Eudore) Stop() error {
	return app.Server.Close()
}

// Eudore Wait quit.
func (app *Eudore) Shutdown() error {
	return app.Server.Shutdown(context.Background())
}


// Register a Reload function, index determines the function loading order, and name is used for a specific load function.
//
// 注册一个Reload函数，index决定函数加载顺序，name用于特定加载函数。
func (app *Eudore) RegisterInit(name string, index int, fn InitFunc) {
	if name != "" {
		if fn == nil {
			delete(app.inits, name)
		}else {
			app.inits[name] = initInfo{name, index, fn}
		}
	}
}

// Send a specific message to eudore to execute the corresponding signal should function.
//
// 给eudore发送一个特定信息，用于执行对应信号应该函数。
func (*Eudore) HandleSignal(sig os.Signal) error {
	return SignalHandle(sig)
}

// Register Signal exec func.
// bf alise befor,if bf is ture add func to funcs first.
func (*Eudore) RegisterSignal(sig os.Signal, bf bool, fn SignalFunc) {
	SignalRegister(sig, bf, fn)
}

// Set Pool new func.
// Type is context, request and response.
func (app *Eudore) RegisterPool(name string, fn func() interface{}) {
	switch name{
	case "Httpcontext":
		app.Httpcontext.New = fn
	case "Httprequest":
		app.Httprequest.New = fn
	case "Httpresponse":
		app.Httpresponse.New = fn
	}
}

/*
func (app *Eudore) RegisterComponents(names []string, args []interface{}) error {
	errs := NewErrors()
	for i, name := range names {
		errs.HandleError(app.RegisterComponent(name, args[i]))
	}
	return errs.GetError()
}


*/
func (app *Eudore) RegisterComponent(name string,  arg interface{}) (c Component,err error) {
	c, err = app.App.RegisterComponent(name, arg)
	app.HandleError(err)
	return 
}

// Register a static file Handle.
func (e *Eudore) RegisterStatic(path , dir string) {
	e.Router.GetFunc(path, func(ctx Context){
		ctx.WriteFile(dir + ctx.Path())
	})
}

// log out
func (e *Eudore) Debug(args ...interface{}) {
	e.logReset().Debug(args...)
}

func (e *Eudore) Info(args ...interface{}) {
	e.logReset().Info(args...)
}

func (e *Eudore) Warning(args ...interface{}) {
	e.logReset().Warning(args...)
}

func (e *Eudore) Error(args ...interface{}) {
	e.logReset().Error(args...)
}

func (e *Eudore) Debugf(format string, args ...interface{}) {
	e.logReset().Debug(fmt.Sprintf(format, args...))
}

func (e *Eudore) Infof(format string, args ...interface{}) {
	e.logReset().Info(fmt.Sprintf(format, args...))
}

func (e *Eudore) Warningf(format string, args ...interface{}) {
	e.logReset().Warning(fmt.Sprintf(format, args...))
}

func (e *Eudore) Errorf(format string, args ...interface{}) {
	e.logReset().Error(fmt.Sprintf(format, args...))
}

func (e *Eudore) logReset() LogOut {
	file, line := LogFormatFileLine(0)
	f := Fields{
		"file":				file,
		"line":				line,
	}
	return e.Logger.WithFields(f)
}




func (e *Eudore) HandleError(err error) {
	if err != nil {
		if err != ErrApplicationStop {
			e.Error(err)
			e.stop <- err
			return
		}
		e.stop <- nil
	}
}

func (e *Eudore) Handle(ctx Context) {
	ctx.SetHandler(e.Router.Match(ctx.Method(), ctx.Path(), ctx))
	ctx.Next()
}

func (e *Eudore) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// get
	request := e.Httprequest.Get().(*RequestReaderHttp)
	response := e.Httpresponse.Get().(*ResponseWriterHttp)
	// init
	ResetRequestReaderHttp(request, req)
	ResetResponseWriterHttp(response, w)
	e.EudoreHTTP(req.Context(), response, request)
	// clean
	e.Httprequest.Put(request)
	e.Httpresponse.Put(response)
}


func (e *Eudore) EudoreHTTP(pctx context.Context,w protocol.ResponseWriter, req protocol.RequestReader) {
	// init
	ctx := e.Httpcontext.Get().(Context)
	// handle
	ctx.Reset(pctx, w, req)
	ctx.SetHandler(e.Router.Match(ctx.Method(), ctx.Path(), ctx))
	ctx.Next()
	// release
	e.Httpcontext.Put(ctx)
}
