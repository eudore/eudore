# Application

app是全局对象的集合，App对象组合了Config、Server、Logger、Router、Cache、Binder、Renderer、View这些全局对象。

App.Pools保存全部sync.Pool使用的函数，未来可能会改为App.Config存储。

App对象极简单，但是无法使用需要根据情况进一步封装然后使用。

App对象定义：

```golang
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
```

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

```golang
type Core
	func NewCore() *Core
	func (app *Core) Listen(addr string) *Core
	func (app *Core) ListenTLS(addr, key, cert string) *Core
	func (app *Core) Run() (err error)
	func (app *Core) ServeHTTP(w http.ResponseWriter, req *http.Request)
```
## Eudore

# Context

```golang
type Context interface {
	// context
	Reset(context.Context, ResponseWriter, RequestReader)
	Request() RequestReader
	Response() ResponseWriter
	SetRequest(RequestReader)
	SetResponse(ResponseWriter)
	SetHandler(Middleware)
	Next()
	End()
	NewRequest(string, string, io.Reader) (ResponseReader, error)
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
	GetHeader(name string) string
	SetHeader(string, string)
	Cookies() []*CookieRead
	GetCookie(name string) string
	SetCookie(cookie *CookieWrite)
	SetCookieValue(string, string, int)


	// response
	Write([]byte) (int, error)
	WriteHeader(int)
	Redirect(int, string)
	Push(string, *PushOptions) error
	// render writer 
	WriteString(string) error
	WriteView(string, interface{}) error
	WriteJson(interface{}) error
	WriteFile(string) (int, error)
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

而Middleware组合Handler接口并新增SetNext和GetNext方法，用于读写下一个Middleware。

Middleware处理Context是链式的，先使用Handler处理Context，然后获得Next Middleware，然后循环。

```golang
type (
	// Context handle func
	HandlerFunc func(Context)
	// Handler interface
	Handler interface {
		Handle(Context)
	}

	// Middleware interface
	Middleware interface {
		Handler
		GetNext() Middleware
		SetNext(Middleware)
	}

	MiddlewareBase struct {
		Handler
		Next Middleware
	}
)

// Convert the HandlerFunc function to a Handler interface.
//
// 转换HandlerFunc函数成Handler接口。
func (f HandlerFunc) Handle(ctx Context) {
	f(ctx)
}
```

在Context对象中，需要先设置一个Middleware为处理者，然后使用Next开始处理Context。

Context部分定义：

```golang
type Context interface {
	...
	SetHandler(Middleware)
	Next()
	End()
	...
}
```

ContextHttp对这三个方法的实现。

SetHandler方法设置Context的处理Handler。

Next方法先获取当前Handler，然后设置当前Handler为Next，再使用Handler处理Context；若Handler对象中再次调用了Context.Next方法，就会调用后续Handler然后依次结束。

End方法设置Context的isrun标志为否。

在整个处理过程中Middleware为无状态的，而Context仅保存链式Middleware处理的一个结点，然后依次向后执行。

ctx.Middleware是当前处理者，ctx.handler为一个临时变量。

```golang
func (ctx *ContextHttp) SetHandler(m Middleware) {
	ctx.Middleware = m
}

func (ctx *ContextHttp) Next() {
	for ctx.Middleware != nil && ctx.isrun {
		ctx.handler = ctx.Middleware
		ctx.Middleware = ctx.Middleware.GetNext()
		ctx.handler.Handle(ctx)
	}
}

func (ctx *ContextHttp) End() {
	ctx.isrun = false
}
```

# Router

Router对象由RouterCore和RouterMethod组合，RouterMethod实现各种路由注册封装，RouterCore用于实现路由器的注册和匹配。

在Router中有Handler、Middleware、Router三种处理对象，三者依次组合

```golang
type (
	// Router method
	// 路由默认直接注册的方法，其他方法可以使用RouterRegister接口直接注册。
	RouterMethod interface {
		SubRoute(path string, router Router)
		AddHandler(...Handler)
		Any(string, Handler)
		AnyFunc(string, HandlerFunc)
		Delete(string, Handler)
		DeleteFunc(string, HandlerFunc)
		Get(string, Handler)
		GetFunc(string, HandlerFunc)
		Head(string, Handler)
		HeadFunc(string, HandlerFunc)
		Options(string, Handler)
		OptionsFunc(string, HandlerFunc)
		Patch(string, Handler)
		PatchFunc(string, HandlerFunc)
		Post(string, Handler)
		PostFunc(string, HandlerFunc)
		Put(string, Handler)
		PutFunc(string, HandlerFunc)
	}
	// Router Core
	RouterCore interface {
		Middleware
		RegisterMiddleware(...Handler)
		RegisterHandler(method string, path string, handler Handler)
		Match(Params) Middleware
	}
	// router
	Router interface {
		Component
		RouterCore
		RouterMethod
	}
)
```

RouterCore需要实现Middleware接口。

MiddlewareRouter是一个Router使用的Middleware，通过Handler调用Router.Match匹配出对应的Middleware，然后动态设置给Context，然后让Context继续处理。

```golang
type MiddlewareRouter struct {
	RouterCore
	Next Middleware
}

// Match a handler and directly use it with the Context object.
//
// Then set the tail handler appended by the SetNext method to be the follower.
//
// 匹配出一个处理者，并直接给Context对象并使用。
//
// 然后设置SetNext方法追加的尾处理者为后续处理者。
func (m *MiddlewareRouter) Handle(ctx Context) {
	ctx.SetHandler(m.Match(ctx.Params()))
	ctx.Next()
	ctx.SetHandler(m.Next)
	ctx.Next()
}

// The return processing middleware is nil.
//
// The router is stateless and cannot return directly to the next handler.
//
// When the router processes it, it will match the next handler and directly use it for the Context object.
//
// 返回处理中间件为空。
//
// 路由器是无状态的，无法直接返回下一个处理者。
//
// 在路由器处理时会匹配出下一个处理者，并直接给Context对象并使用。
func (m *MiddlewareRouter) GetNext() Middleware {
	return nil
}

// Set the post-processing chain after the route is processed.
//
// 设置路由处理完后的后序处理链。
func (m *MiddlewareRouter) SetNext(nm Middleware) {
	// 请求尾处理
	if nm == nil {
		m.Next = nil
		return
	}
	// 尾追加处理中间件
	link := m.Next
	n := link.GetNext();
	for n != nil {
		link = n
		n = link.GetNext();
	}
	link.SetNext(nm)
}

```

# Config

# Logger

```golang
```

# Server

```golang
```

# Cache


```golang
```