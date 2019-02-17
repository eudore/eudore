package eudore

import (
	"os"
	"fmt"
	"sync"
	"sort"
	"context"
	"strings"
	"net/http"
)

type (
	// eudore 
	Eudore struct {
		*App
		pool			*pool
		Handlers		[]Handler
		reloads			map[string]ReloadInfo
	}
	// eudore reload funcs.
	ReloadFunc func(*Eudore) error
	// Save reloadhook name, index fn.
	ReloadInfo struct {
		name	string
		index	int
		fn		ReloadFunc
	}
	// Define pool.
	pool struct {
		httprequest  sync.Pool
		httpresponse sync.Pool
		httpcontext  sync.Pool
	}
)

var defaultEudore *Eudore

// Create a new Eudore.
func NewEudore() *Eudore {
	e := &Eudore{
		App:			NewApp(),
		reloads:		make(map[string]ReloadInfo),
	}
	// set eudore pool
	e.pool = &pool{
		httprequest: sync.Pool {
			New: func() interface{} {
				return &RequestReaderHttp{}
			},
		},
		httpresponse: sync.Pool {
			New: func() interface{} {
				return &ResponseWriterHttp{}
			},
		},
		httpcontext: sync.Pool {
			New: func() interface{} {
				return &ContextHttp{
					app:	e.App,
				}
			},
		},
	}
	// Register eudore default components
	e.HandleError(e.RegisterComponents(
		[]string{"config", "logger-init", "router", "cache", "view"}, 
		[]interface{}{nil, nil, nil, nil, nil},
	))
	// Register eudore default reload func
	// e.RegisterReload("eudore-keys", 0x008, ReloadKeys)
	e.RegisterReload("eudore-config", 0x009, ReloadConfig)
	e.RegisterReload("eudore-logger", 0x015 , ReloadLogger)
	e.RegisterReload("eudore-server", 0x016 , ReloadServer)
	e.RegisterReload("eudore-signal", 0x018 , ReloadSignal)
	e.RegisterReload("eudore-defaule-logger", 0xa15 , ReloadDefaultLogger)
	e.RegisterReload("eudore-defaule-server", 0xa16 , ReloadDefaultServer)
	e.RegisterReload("eudore-component-info", 0xc01 , ReloadListComponent)
	e.RegisterReload("eudore-test-stop", 0xfff, ReloadStop)
	return e
}

// Get the default eudore, if it is empty, create a new singleton.
//
// 获取默认的eudore，如果为空，创建一个新的单例。
func DefaultEudore() *Eudore {
	if defaultEudore == nil {
		defaultEudore = NewEudore()
	}
	return defaultEudore
}

// Parse the current command, if the command is 'start', start eudore.
//
// 解析当前命令，如果命令是启动，则启动eudore。
func (e *Eudore) Run() (err error) {
	var cmd, pid string
	defer func(){
		// if err != nil {
		// 	ReloadLogger(e)
		// }
		if _, ok := e.Logger.(LoggerInitHandler); ok && cmd == "start" {
			e.RegisterComponent("logger", nil)
		}
		// if err := recover();err != nil{
		// 	e.Fatal(err)
		// }
	}()

	// Parse config
	e.Debug("eudore start parse config")
	if err = e.Config.Parse(); err != nil {
		e.Error("eudore parse config error: ", err)
		return
	}

	// cmd := e.globalConfig.GetString("#command", DEFAULT_CONFIG_COMMAND)
	// pid := e.globalConfig.GetString("#pidfile", DEFAULT_CONFIG_PIDFILE)
	// Json(e.config, cmd, pid)
	cmd = e.Config.Get("#command").(string)
	pid = e.Config.Get("#pidfile").(string)
	fmt.Println(cmd, pid)
	err = NewCommand(cmd , pid, e.Start).Run()
	return
}

// Load all configurations and then start listening for all services.
//
// 加载全部配置，然后启动监听全部服务。
func (e *Eudore) Start() error {
	// Reload
	e.Info("eudore start reload all func")
	if err := e.Reload(); err != nil {
		e.Error(err)
		return err
	}

	// Start server
	if e.Server == nil {
		err := fmt.Errorf("Eudore can't start the service, the server is empty.")
		e.Error(err)
		return err
	}
	e.Info("eudore start all server.")
	return e.Server.Start()
}

// Execute the eudore reload function.
// names are a list of function names that need to be executed; if the list is empty, execute all reload functions.
//
// 执行eudore重新加载函数。
// names是需要执行的函数名称列表；如果列表为空，执行全部重新加载函数。
func (e *Eudore) Reload(names ...string) (err error) {
	// get names
	names = e.getReloadNames(names)
	// exec
	var i int
	var name string
	num := len(names)
	defer func() {
		if err1 := recover(); err1 != nil {
			if err2, ok := err1.(error);ok {
				err = err2
			}else {
				err = fmt.Errorf("eudore reload %s %d/%d recover error: %v", name, i + 1, num, err1)
			}
		}
	}()
	for i, name = range names {
		if err = e.reloads[name].fn(e);err != nil {
			return fmt.Errorf("eudore reload %d/%d %s error: %v",i + 1, num, name, err)
		}
		e.Infof("eudore reload %d/%d %s success.", i + 1, num, name)
	}
	e.Info("eudore reload all success.")
	return nil
}

// 处理名称并对reloads排序。
func (e *Eudore) getReloadNames(names []string) []string {
	// get all exec names
	if len(names) == 0 {
		names = make([]string, 0, len(e.reloads))
		for k := range e.reloads {
			names = append(names, k)
		}
	}else {
		for i, name := range names {
			if _, ok := e.reloads[name]; !ok {
				names[i] = ""
				e.Warning("Invalid overloaded function name: ", name)
			}
		}
		names = arrayclean(names)
	}
	// index
	sort.Slice(names, func(i, j int) bool {
		return e.reloads[names[i]].index < e.reloads[names[j]].index
	})
	return names
}


// Restart Eudore
// Invoke ServerManager Restart
func (e *Eudore) Restart() error {
	return e.Server.Restart()
}

// Eudore Stop immediately
func (e *Eudore) Stop() error {
	return e.Server.Close()
}

// Eudore Wait quit.
func (e *Eudore) Shutdown() error {
	return e.Server.Shutdown(context.Background())
}


// Register a Reload function, index determines the function loading order, and name is used for a specific load function.
//
// 注册一个Reload函数，index决定函数加载顺序，name用于特定加载函数。
func (e *Eudore) RegisterReload(name string, index int, fn ReloadFunc) {
	if name != "" && fn != nil {
		e.reloads[name] = ReloadInfo{name, index, fn}
	}
}

// Send a specific message to eudore to execute the corresponding signal should function.
//
// 给eudore发送一个特定信息，用于执行对应信号应该函数。
func (e *Eudore) HandleSignal(sig os.Signal) error {
	return SignalHandle(sig)
}

// Register Signal exec func.
// bf alise befor,if bf is ture add func to funcs first.
func (e *Eudore) RegisterSignal(sig os.Signal, bf bool, fn SignalFunc) {
	SignalRegister(sig, bf, fn)
}

// Set Pool new func.
// Type is context, request and response.
func (e *Eudore) RegisterPool(name string, fn func() interface{}) {
	switch name{
	case "httpcontext":
		e.pool.httpcontext.New = fn
	case "httprequest":
		e.pool.httprequest.New = fn
	case "httpresponse":
		e.pool.httpresponse.New = fn
	}
}

func (e *Eudore) RegisterComponents(names []string, args []interface{}) error {
	errs := NewErrors()
	for i, name := range names {
		errs.HandleError(e.RegisterComponent(name, args[i]))
	}
	return errs.GetError()
}

func (e *Eudore) RegisterComponent(name string,  arg interface{}) (err error) {
	err = e.App.RegisterComponent(name, arg)
	if err == nil {
		if strings.HasPrefix(name, ComponentServerName) {
			e.Server.SetErrorFunc(e.HandleError)
			e.Server.SetHandler(e)	
		}
	}
	return 
}


// http method

// Register a static file Handle.
func (e *Eudore) RegisterStatic(path , dir string) {
	e.Router.GetFunc(path, func(ctx Context){
			ctx.WriteFile(dir + ctx.Path())
			// ctx.End()
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
		e.Error(err)
	}
}

func (e *Eudore) Handle(ctx Context) {
	e.Router.Handle(ctx)
}

func (e *Eudore) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	pool := e.pool
	// get
	request := pool.httprequest.Get().(*RequestReaderHttp)
	response := pool.httpresponse.Get().(*ResponseWriterHttp)
	// init
	ResetRequestReaderHttp(request, req)
	ResetResponseWriterHttp(response, w)
	e.EudoreHTTP(req.Context(), response, request)
	// clean
	pool.httprequest.Put(request)
	pool.httpresponse.Put(response)
}


func (e *Eudore) EudoreHTTP(pctx context.Context,w ResponseWriter, req RequestReader) {
	// init
	pool := e.pool.httpcontext
	ctx := pool.Get().(Context)
	// handle
	ctx.Reset(pctx, w, req)
	e.Router.Handle(ctx)
	// release
	pool.Put(ctx)
}
