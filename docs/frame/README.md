# Eudore

eudore具有以下对象，除Application以为均为接口，每个对象都具有明确语义，Application是最顶级对象可以通过组合方式实现重写，其他对象为接口定义直接重新实现，或组合接口实现部分重写。

| 名称 | 作用 | 定义 |
| ------------ | ------------ | ------------ |
| Application | 运行对象主体 | app.go core.go eudore.go |
| Context | 请求处理上下文 | context.go |
| Request | Http请求数据 | request.go |
| Response | http响应写入 | response.go |
| Router | 请求路由选择 | router.go routerRadix.go routerFull.go |
| Middleware | 多Handler组合运行 | handler.go |
| Logger | App和Ctx日志输出 | logger.go |
| Server | http Server启动 | server.go |
| Config | 配置数据管理 | config.go configparse.go |
| Cache | 全局缓存对象 | cache.go |
| View | 模板渲染 | view.go |
| Client | http客户端 | client.go |
| Session | 回话数据管理 | session.go |
| Controller | 解析执行控制器 | controller.go |
| Bind | 请求数据反序列化 | bind.go |
| Render | 响应数据序列化 | render.go |
| Websocket | websocket协议读写 | websocket.go |

其他文件定义内容

| 文件 | 作用 |
| ------------ | ------------ |
| command.go | 启动命令解析 |
| component.go | 组件定义 |
| const.go | 定义常量 |
| doc.go | godoc内容 |
| error.go | 定义错误 |
| listener.go | 全局监听管理 |
| reflect.go | 各类反射辅助函数 |
| setting.go | 配置化启动程序 |
| signal.go | 全局信号管理 |
| util.go | 辅助函数 |
| version.go | 版本信息常量 |

# Application

app是全局对象的集合，App对象组合了Config、Server、Logger、Router、Cache、Binder、Renderer、View这些全局对象。

App.Pools保存全部sync.Pool使用的函数，未来可能会改为App.Config存储。

App对象极简单，但是无法使用需要根据情况进一步封装然后使用。

App对象定义：

```golang
type (
	PoolGetFunc func() interface{}
	// The App combines the main functional interfaces, and the instantiation operations such as startup require additional packaging.
	//
	// App组合主要功能接口，启动等实例化操作需要额外封装。
	App struct {
		Config
		Server
		Logger
		Router
		Cache
		Binder
		Renderer
		View
		// pools存储各种Context、、构造函数，用于sync.pool Get一个新对象。
		Pools map[string]PoolGetFunc
	}
)
```

func list:

```golang
type App
	func NewApp() *App
	func (app *App) RegisterComponent(name string, arg interface{}) error
	func (app *App) RegisterComponents(names []string, args []interface{}) error
	func (app *App) RegisterPoolFunc(name string, fn PoolGetFunc)
```

## Core

App封装对象之一Core，
```golang
type (
	Core struct {
		*App
		poolctx sync.Pool
		poolreq	sync.Pool
		poolresp sync.Pool
		ports []*ServerConfigGeneral
	}
)
```

Core一个只有五个方法，Listen和ListenTLS添加Server监听端口，Run启动程序，EudoreHTTP和ServeHTTP实现Server Handler需要的接口，ServeHTTP未来可能移除。

```golang
type Core
    func NewCore() *Core
    func (app *Core) EudoreHTTP(pctx context.Context, w protocol.ResponseWriter, req protocol.RequestReader)
    func (app *Core) Listen(addr string) *Core
    func (app *Core) ListenTLS(addr, key, cert string) *Core
    func (app *Core) Run() (err error)
    func (app *Core) ServeHTTP(w http.ResponseWriter, req *http.Request)
```

## Eudore

```golang
type Eudore struct {
	*App
	pool			*pool
	reloads			map[string]ReloadInfo
}
```

Eudore相对Core主要添加了RegisterInit方法添加启动函数，其他则是在各个组件基础上封装。

```golang
type Eudore
    func DefaultEudore() *Eudore
    func NewEudore() *Eudore
    func (e *Eudore) Debug(args ...interface{})
    func (e *Eudore) Debugf(format string, args ...interface{})
    func (e *Eudore) Error(args ...interface{})
    func (e *Eudore) Errorf(format string, args ...interface{})
    func (e *Eudore) EudoreHTTP(pctx context.Context, w protocol.ResponseWriter, req protocol.RequestReader)
    func (e *Eudore) Handle(ctx Context)
    func (e *Eudore) HandleError(err error)
    func (*Eudore) HandleSignal(sig os.Signal) error
    func (e *Eudore) Info(args ...interface{})
    func (e *Eudore) Infof(format string, args ...interface{})
    func (app *Eudore) Init(names ...string) (err error)
    func (app *Eudore) RegisterComponents(names []string, args []interface{}) error
    func (app *Eudore) RegisterInit(name string, index int, fn InitFunc)
    func (app *Eudore) RegisterPool(name string, fn func() interface{})
    func (*Eudore) RegisterSignal(sig os.Signal, bf bool, fn SignalFunc)
    func (e *Eudore) RegisterStatic(path, dir string)
    func (app *Eudore) Restart() error
    func (app *Eudore) Run() (err error)
    func (e *Eudore) ServeHTTP(w http.ResponseWriter, req *http.Request)
    func (app *Eudore) Shutdown() error
    func (app *Eudore) Start() error
    func (app *Eudore) Stop() error
    func (e *Eudore) Warning(args ...interface{})
    func (e *Eudore) Warningf(format string, args ...interface{})
```

# Context

Context是一次请求的上下文环境，接口大概分类为：context设置、请求数据读取，响应写入、日志输出这四类。

context.Context接口实现未完善。

Context的定义：

```golang
type Context interface {
    // context
    Reset(context.Context, protocol.ResponseWriter, protocol.RequestReader)
    Request() protocol.RequestReader
    Response() protocol.ResponseWriter
    SetRequest(protocol.RequestReader)
    SetResponse(protocol.ResponseWriter)
    SetHandler(HandlerFuncs)
    Next()
    End()
    NewRequest(string, string, io.Reader) (protocol.ResponseReader, error)
    // context
    Deadline() (time.Time, bool)
    Done() <-chan struct{}
    Err() error
    Value(key interface{}) interface{}
    SetValue(interface{}, interface{})

    // request info
    Read([]byte) (int, error)
    Host() string
    Method() string
    Path() string
    RemoteAddr() string
    RequestID() string
    Referer() string
    ContentType() string
    Istls() bool
    Body() []byte

    // param header cookie
    Params() Params
    GetParam(string) string
    SetParam(string, string)
    AddParam(string, string)
    GetQuery(string) string
    GetHeader(name string) string
    SetHeader(string, string)
    Cookies() []*Cookie
    GetCookie(name string) string
    SetCookie(cookie *SetCookie)
    SetCookieValue(string, string, int)

    // response
    Write([]byte) (int, error)
    WriteHeader(int)
    Redirect(int, string)
    Push(string, *protocol.PushOptions) error
    // render writer
    WriteString(string) error
    WriteView(string, interface{}) error
    WriteJson(interface{}) error
    WriteFile(string) error

    // binder and renderer
    ReadBind(interface{}) error
    WriteRender(interface{}) error

    // log LogOut interface
    Debug(...interface{})
    Info(...interface{})
    Warning(...interface{})
    Error(...interface{})
    Fatal(...interface{})
    WithField(key string, value interface{}) LogOut
    WithFields(fields Fields) LogOut
    // app
    App() *App
}
```
# RequestReader & ResponseWriter

```golang
type (
	// Get the method, version, uri, header, body from the RequestReader according to the http protocol request body. (There is no host in the golang net/http library header)
	//
	// Read the remote connection address and TLS information from the net.Conn connection.
	//
	// 根据http协议请求体，从RequestReader获取方法、版本、uri、header、body。(golang net/http库header中没有host)
	//
	// 从net.Conn连接读取远程连接地址和TLS信息。
	RequestReader interface {
		// http protocol data
		Method() string
		Proto() string
		RequestURI() string
		Header() Header
		Read([]byte) (int, error)
		Host() string
		// conn data
		RemoteAddr() string
		TLS() *tls.ConnectionState
	}
	// ResponseWriter接口用于写入http请求响应体status、header、body。
	//
	// net/http.response实现了flusher、hijacker、pusher接口。
	ResponseWriter interface {
		// http.ResponseWriter
		Header() http.Header
		Write([]byte) (int, error)
		WriteHeader(codeCode int)
		// http.Flusher 
		Flush()
		// http.Hijacker
		Hijack() (net.Conn, *bufio.ReadWriter, error)
		// http.Pusher
		Push(string, *PushOptions) error
		Size() int
		Status() int
	}
)
```

# Handler & Middleware

Handler接口定义了处理Context的方法。

```golang
type (
	// Context handle func
	HandlerFunc func(Context)
	// Handler interface
	Handler interface {
		Handle(Context)
	}
	HandlerFuncs	[]HandlerFunc
)

```

# Router

Router对象由RouterCore和RouterMethod组合，RouterMethod实现各种路由注册封装，RouterCore用于实现路由器的注册和匹配。

当前实现RouterRadix、RouterFull、RouterInit、RouterHost三种Router。

RouterRadix基于基数树实现，实现路径参数、通配符参数、默认参数、参数校验三项基本功能。

RouterFull基于基数树实现，实现全部路由器相关特性,实现路径参数、通配符参数、默认参数、参数校验、通配符校验，未实现多参数正则捕捉。

RouterInit用于初始化处理或者作为一个空路由。

RouterHost使用基于Host匹配选择多个组合路由，Host匹配可自己实现。

路由器接口定义：

```golang
type (
	// Router method
	// 路由默认直接注册的方法，其他方法可以使用RouterRegister接口直接注册。
	RouterMethod interface {
		Group(string) RouterMethod
		AddHandler(string, string, ...HandlerFunc) RouterMethod
		AddMiddleware(...HandlerFunc) RouterMethod
		AddController(...Controller) RouterMethod
		Any(string, ...Handler)
		AnyFunc(string, ...HandlerFunc)
		Delete(string, ...Handler)
		DeleteFunc(string, ...HandlerFunc)
		Get(string, ...Handler)
		GetFunc(string, ...HandlerFunc)
		Head(string, ...Handler)
		HeadFunc(string, ...HandlerFunc)
		Options(string, ...Handler)
		OptionsFunc(string, ...HandlerFunc)
		Patch(string, ...Handler)
		PatchFunc(string, ...HandlerFunc)
		Post(string, ...Handler)
		PostFunc(string, ...HandlerFunc)
		Put(string, ...Handler)
		PutFunc(string, ...HandlerFunc)
	}
	// Router Core
	RouterCore interface {
		RegisterMiddleware(string, string, HandlerFuncs)
		RegisterHandler(string, string, HandlerFuncs)
		Match(string, string, Params) HandlerFuncs
	}
	// router
	Router interface {
		Component
		RouterCore
		RouterMethod
	}
)
```

## RouterMethod

RouterMethod分三种方法，Group、AddMiddleware、Any。

Group会创建一个组路由，之后注册的路由都会附加Group的参数的前缀和默认参数，路径结尾不可为'/*'和'/'，会知道删除。

AddMiddleware会给RouterCore注册中间件，注册Any方法，路径是组前缀的中间件，如果需要单独注册一种方法的中间件需要直接调用RegisterMiddleware。

Any等方法是注册restful请求。

## RouterCore

RouterCore拥有三个RegisterMiddleware、RegisterHandler、Match三个方法。

RegisterMiddleware直接记录一个方法路径下的中间件的数据，在注册路由的时候使用。

RegisterHandler注册一个路由请求，注册时会更具路径匹配使用的中间件附加到HandlerFuncs前方。

Match会更具方法和路径匹配路由，并附加相关的参数。

## Radix

RouterRadix和RouterFull都是基于Radix tree算法实现。

# Config

# Logger

```golang
```

# Server

```golang
```

# Cache

```golang
type Cache interface {
	Component
	// get cached value by key.
	Get(string) interface{}
	// set cached value with key and expire time.
	Set(string, interface{}, time.Duration) error
	// delete cached value by key.
	Delete(string) error
	// check if cached value exists or not.
	IsExist(string) bool
	// get all keys
	GetAllKeys() []string
	// get keys size
	Count() int
	// clean all cache.
	CleanAll() error
}
```
