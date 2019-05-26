# Application

app是全局对象的集合，App对象组合了Config、Logger、Server、Router、Cache、Session、Binder、Client、View、Renderer这些全局对象。

定义文件：app.go、core.go、eudore.go

App对象定义：

```golang
// app.go
type App struct {
	Config
	Logger
	Server
	Router
	Cache
	Session
	Client
	View
	Binder
	Renderer
}
```

```godoc
type App
	func NewApp() *App
	func (app *App) GetAllComponent() ([]string, []Component)
	func (app *App) InitComponent() error
	func (app *App) RegisterComponent(name string, arg interface{}) (Component, error)
	func (app *App) RegisterComponents(names []string, args []interface{}) error
```

app对象无法直接使用，仅定义了部分组件加载函数，需要额外封装一层启动程序相关函数。

## Core

Core组合App对象，额外添加了Listen、Run和EudoreHTTP三个函数。

Listen是添加一个监听端口信息，Run用来启动程序，启动Server监听端口。

EudoreHTTP是实现protocol.Handler接口，额外兼容实现了http.Handler接口，用于处理Server传统的请求。

```golang
// protocol/protocol.go
type Handler interface {
	EudoreHTTP(context.Context, ResponseWriter, RequestReader)
}

// core.go
type Core struct {
	*App
	Poolctx sync.Pool
	Poolreq	sync.Pool
	Poolresp sync.Pool
}
```

```godoc
type Core
	func NewCore() *Core
	func (app *Core) EudoreHTTP(pctx context.Context, w protocol.ResponseWriter, req protocol.RequestReader)
	func (app *Core) Listen(addr string) *Core
	func (app *Core) ListenTLS(addr, key, cert string) *Core
	func (app *Core) Run() (err error)
	func (app *Core) ServeHTTP(w http.ResponseWriter, req *http.Request)
```

## Eudore

Eudore是App对象的另外一种实现方式，主要添加了日志函数、初始化函数、阻塞chan。

基于Logger重新封装了部分方法，输出将带有文件位置信息。

初始化函数是保存执行时使用的初始化函数，会按照优先级依次执行。

阻塞chan由于异步启动服务等行为，在Start时启动一个goroutine来执行全部初始化函数，同时会阻塞char，HandleError处理一个非空错误时，会放入chan中，结束主进程的阻塞。

```golang
// eudore.go
type Eudore struct {
	*App
	Httprequest		sync.Pool
	Httpresponse	sync.Pool
	Httpcontext		sync.Pool
	inits			map[string]initInfo
	stop 			chan error
}
```

```godoc
type Eudore
	func DefaultEudore(components ...ComponentConfig) *Eudore
	func NewEudore(components ...ComponentConfig) *Eudore
	func (*Eudore) Debug(args ...interface{})
	func (*Eudore) Debugf(format string, args ...interface{})
	func (*Eudore) Error(args ...interface{})
	func (*Eudore) Errorf(format string, args ...interface{})
	func (*Eudore) EudoreHTTP(pctx context.Context, w protocol.ResponseWriter, req protocol.RequestReader)
	func (*Eudore) Handle(ctx Context)
	func (*Eudore) HandleError(err error)
	func (*Eudore) HandleSignal(sig os.Signal) error
	func (*Eudore) Info(args ...interface{})
	func (*Eudore) Infof(format string, args ...interface{})
	func (*Eudore) Init(names ...string) (err error)
	func (*Eudore) RegisterComponent(name string, arg interface{}) (c Component, err error)
	func (*Eudore) RegisterInit(name string, index int, fn InitFunc)
	func (*Eudore) RegisterPool(name string, fn func() interface{})
	func (*Eudore) RegisterSignal(sig os.Signal, bf bool, fn SignalFunc)
	func (*Eudore) RegisterStatic(path, dir string)
	func (*Eudore) Restart() error
	func (*Eudore) Run() (err error)
	func (*Eudore) ServeHTTP(w http.ResponseWriter, req *http.Request)
	func (*Eudore) Shutdown() error
	func (*Eudore) Start() error
	func (*Eudore) Stop() error
	func (*Eudore) Warning(args ...interface{})
	func (*Eudore) Warningf(format string, args ...interface{})
```