# Application

app是全局对象的集合，App对象组合了context.Context、Config、Logger、Server、Router、Binder、Renderer、ContextPool这些全局对象。

app对象无法直接使用,需要额外实现EudoreHTTP方法，然后给Server对象AddHandler和AddListener，之后启动Server。

**app内置两个实例化对象Core和Eudore，Core特点是代码少内容简单,Eudore特点是额外增加了些功能**

定义文件：app.go、core.go、eudore.go

App对象定义：

```golang
// app.go
type (
	// PoolGetFunc 定义sync.Pool对象使用的构造函数。
	PoolGetFunc func() interface{}
	// The App combines the main functional interfaces, and the instantiation operations such as startup require additional packaging.
	//
	// App 组合主要功能接口，启动等实例化操作需要额外封装。
	//
	// App初始化顺序请按照，Logger-Init、Config、Logger、...
	App struct {
		context.Context
		Config `set:"config"`
		Logger `set:"logger"`
		Server `set:"server"`
		Router `set:"router"`
		Binder
		Renderer
		ContextPool sync.Pool
	}
)

```

## Core

Core组合App对象，额外添加了Run、Listen、ListenTLS和EudoreHTTP四个函数，实现最简app。

Listen和ListenTLS是添加一个监听端口信息，Run用来启动程序，启动Server监听端口。

EudoreHTTP是实现protocol.Handler接口，额外兼容实现了http.Handler接口，用于处理Server传统的请求。

```golang
// protocol/protocol.go
type Handler interface {
	EudoreHTTP(context.Context, ResponseWriter, RequestReader)
}

// core.go
// Core 定义Core对象，是App对象的一种实例化。
type Core struct {
	*App
}
```

```godoc
type Core
    func NewCore() *Core
    func (app *Core) EudoreHTTP(pctx context.Context, w protocol.ResponseWriter, req protocol.RequestReader)
    func (app *Core) Listen(addr string) *Core
    func (app *Core) ListenTLS(addr, key, cert string) *Core
    func (app *Core) Run() (err error)
```

## Eudore

Eudore是App对象的另外一种实现方式，主要添加了日志函数、初始化函数、阻塞chan、信号监听和处理、全局中间件。

基于Logger重新封装了部分方法，输出将带有文件位置信息。

初始化函数是保存执行时使用的初始化函数，会按照优先级依次执行。

阻塞chan由于异步启动服务等行为，在Start时启动一个goroutine来执行全部初始化函数，同时会阻塞chan，HandleError处理一个非空错误时，会放入chan中，结束主进程的阻塞。

```golang
// eudore.go
type Eudore struct {
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
```

```godoc
type Eudore
    func NewEudore(options ...interface{}) *Eudore
    func (app *Eudore) AddGlobalMiddleware(hs ...HandlerFunc)
    func (app *Eudore) AddListener(l net.Listener)
    func (app *Eudore) AddStatic(path, dir string)
    func (app *Eudore) Close() error
    func (app *Eudore) Debug(args ...interface{})
    func (app *Eudore) Debugf(format string, args ...interface{})
    func (app *Eudore) Err() error
    func (app *Eudore) Error(args ...interface{})
    func (app *Eudore) Errorf(format string, args ...interface{})
    func (app *Eudore) EudoreHTTP(pctx context.Context, w protocol.ResponseWriter, req protocol.RequestReader)
    func (app *Eudore) HandleContext(ctx Context)
    func (app *Eudore) HandleError(err error)
    func (app *Eudore) HandleSignal(sig os.Signal)
    func (app *Eudore) Info(args ...interface{})
    func (app *Eudore) Infof(format string, args ...interface{})
    func (app *Eudore) Init(names ...string) (err error)
    func (app *Eudore) Listen(addr string) *Eudore
    func (app *Eudore) ListenTLS(addr, key, cert string) *Eudore
    func (app *Eudore) RegisterInit(name string, index int, fn EudoreFunc)
    func (app *Eudore) RegisterSignal(sig os.Signal, fn EudoreFunc)
    func (app *Eudore) Restart() error
    func (app *Eudore) Run() error
    func (app *Eudore) Shutdown() error
    func (app *Eudore) Start() error
    func (app *Eudore) Warning(args ...interface{})
    func (app *Eudore) Warningf(format string, args ...interface{})
```