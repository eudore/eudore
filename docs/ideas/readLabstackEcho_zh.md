# echo

github: [https://github.com/labstack/echo][1]

# Example

echo的example过程分析。

```golang
package main

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", hello)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}

// Handler
func hello(c echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
```

## New

先[godoc][2]打开echo的源码,地址就是域名`godoc.org`加`/`再加报名`github.com/labstack/echo`。

![image](https://raw.githubusercontent.com/eudore/eudore/master/docs/resource/img/echo01.png)

然后选择`Index`，就会跳过前面的介绍，直接到文档的函数和类型，在选择`type Echo`下的`func New() (e *Echo)`函数跳转到函数介绍。

![image](https://raw.githubusercontent.com/eudore/eudore/master/docs/resource/img/echo02.png)

选择蓝色的New函数，就会跳转到函数定义。

![image](https://raw.githubusercontent.com/eudore/eudore/master/docs/resource/img/echo03.png)

会看见的github定义的函数New内容如下：

```golang
// New creates an instance of Echo.
func New() (e *Echo) {
	e = &Echo{
		Server:    new(http.Server),
		TLSServer: new(http.Server),
		AutoTLSManager: autocert.Manager{
			Prompt: autocert.AcceptTOS,
		},
		Logger:   log.New("echo"),
		colorer:  color.New(),
		maxParam: new(int),
	}
	e.Server.Handler = e
	e.TLSServer.Handler = e
	e.HTTPErrorHandler = e.DefaultHTTPErrorHandler
	e.Binder = &DefaultBinder{}
	e.Logger.SetLevel(log.ERROR)
	e.StdLogger = stdLog.New(e.Logger.Output(), e.Logger.Prefix()+": ", 0)
	e.pool.New = func() interface{} {
		return e.NewContext(nil, nil)
	}
	e.router = NewRouter(e)
	return
}
```

New函数中创建了Echo对象，设置Server、Logger、Router等Echo的属性。

## Use

在godoc列表里面像Echo.New函数一样，向下找到定义的`func (e *Echo) Use(middleware ...MiddlewareFunc)`函数，然后跳转到定义内容如下：

```golang
// Use adds middleware to the chain which is run after router.
func (e *Echo) Use(middleware ...MiddlewareFunc) {
	e.middleware = append(e.middleware, middleware...)
}
```

Use方法给echo的中间件追加中间件处理函数。

## Get

Get方法调用Add方法注册一个新路由，Add方法让Echo的router注册方法，并保存该路由的信息，在启动时Echo会输出注册的路由信息。

```golang
// GET registers a new GET route for a path with matching handler in the router
// with optional route-level middleware.
func (e *Echo) GET(path string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return e.Add(http.MethodGet, path, h, m...)
}

// Add registers a new route for an HTTP method and path with matching handler
// in the router with optional route-level middleware.
func (e *Echo) Add(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) *Route {
	name := handlerName(handler)
	e.router.Add(method, path, func(c Context) error {
		h := handler
		// Chain middleware
		for i := len(middleware) - 1; i >= 0; i-- {
			h = middleware[i](h)
		}
		return h(c)
	})
	r := &Route{
		Method: method,
		Path:   path,
		Name:   name,
	}
	e.router.routes[method+path] = r
	return r
}
```

## Start

Start方法启动一个http.Server

```golang
// Start starts an HTTP server.
func (e *Echo) Start(address string) error {
	e.Server.Addr = address
	return e.StartServer(e.Server)
}
```

# Application

```golang
// https://github.com/labstack/echo/blob/master/echo.go#L64
// Echo is the top-level framework instance.
Echo struct {
	StdLogger        *stdLog.Logger
	colorer          *color.Color
	premiddleware    []MiddlewareFunc
	middleware       []MiddlewareFunc
	maxParam         *int
	router           *Router
	notFoundHandler  HandlerFunc
	pool             sync.Pool
	Server           *http.Server
	TLSServer        *http.Server
	Listener         net.Listener
	TLSListener      net.Listener
	AutoTLSManager   autocert.Manager
	DisableHTTP2     bool
	Debug            bool
	HideBanner       bool
	HidePort         bool
	HTTPErrorHandler HTTPErrorHandler
	Binder           Binder
	Validator        Validator
	Renderer         Renderer
	Logger           Logger
}

// by godoc
func New() (e *Echo)
func (e *Echo) AcquireContext() Context
func (e *Echo) Add(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) *Route
func (e *Echo) Any(path string, handler HandlerFunc, middleware ...MiddlewareFunc) []*Route
func (e *Echo) CONNECT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
func (e *Echo) Close() error
func (e *Echo) DELETE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
func (e *Echo) DefaultHTTPErrorHandler(err error, c Context)
func (e *Echo) File(path, file string, m ...MiddlewareFunc) *Route
func (e *Echo) GET(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
func (e *Echo) Group(prefix string, m ...MiddlewareFunc) (g *Group)
func (e *Echo) HEAD(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
func (e *Echo) Match(methods []string, path string, handler HandlerFunc, middleware ...MiddlewareFunc) []*Route
func (e *Echo) NewContext(r *http.Request, w http.ResponseWriter) Context
func (e *Echo) OPTIONS(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
func (e *Echo) PATCH(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
func (e *Echo) POST(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
func (e *Echo) PUT(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
func (e *Echo) Pre(middleware ...MiddlewareFunc)
func (e *Echo) ReleaseContext(c Context)
func (e *Echo) Reverse(name string, params ...interface{}) string
func (e *Echo) Router() *Router
func (e *Echo) Routes() []*Route
func (e *Echo) ServeHTTP(w http.ResponseWriter, r *http.Request)
func (e *Echo) Shutdown(ctx stdContext.Context) error
func (e *Echo) Start(address string) error
func (e *Echo) StartAutoTLS(address string) error
func (e *Echo) StartServer(s *http.Server) (err error)
func (e *Echo) StartTLS(address string, certFile, keyFile string) (err error)
func (e *Echo) Static(prefix, root string) *Route
func (e *Echo) TRACE(path string, h HandlerFunc, m ...MiddlewareFunc) *Route
func (e *Echo) URI(handler HandlerFunc, params ...interface{}) string
func (e *Echo) URL(h HandlerFunc, params ...interface{}) string
func (e *Echo) Use(middleware ...MiddlewareFunc)
```


## Start

```golang
https://github.com/labstack/echo/blob/master/echo.go#L642
// StartServer starts a custom http server.
func (e *Echo) StartServer(s *http.Server) (err error) {
	// Setup
	e.colorer.SetOutput(e.Logger.Output())
	s.ErrorLog = e.StdLogger
	s.Handler = e
	if e.Debug {
		e.Logger.SetLevel(log.DEBUG)
	}

	if !e.HideBanner {
		e.colorer.Printf(banner, e.colorer.Red("v"+Version), e.colorer.Blue(website))
	}

	if s.TLSConfig == nil {
		if e.Listener == nil {
			e.Listener, err = newListener(s.Addr)
			if err != nil {
				return err
			}
		}
		if !e.HidePort {
			e.colorer.Printf("⇨ http server started on %s\n", e.colorer.Green(e.Listener.Addr()))
		}
		return s.Serve(e.Listener)
	}
	if e.TLSListener == nil {
		l, err := newListener(s.Addr)
		if err != nil {
			return err
		}
		e.TLSListener = tls.NewListener(l, s.TLSConfig)
	}
	if !e.HidePort {
		e.colorer.Printf("⇨ https server started on %s\n", e.colorer.Green(e.TLSListener.Addr()))
	}
	return s.Serve(e.TLSListener)
}
```

## ServeHTTP

```golang
// https://github.com/labstack/echo/blob/master/echo.go#L563
// ServeHTTP implements `http.Handler` interface, which serves HTTP requests.
func (e *Echo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Acquire context
	c := e.pool.Get().(*context)
	c.Reset(r, w)

	h := NotFoundHandler

	if e.premiddleware == nil {
		e.router.Find(r.Method, getPath(r), c)
		h = c.Handler()
		for i := len(e.middleware) - 1; i >= 0; i-- {
			h = e.middleware[i](h)
		}
	} else {
		h = func(c Context) error {
			e.router.Find(r.Method, getPath(r), c)
			h := c.Handler()
			for i := len(e.middleware) - 1; i >= 0; i-- {
				h = e.middleware[i](h)
			}
			return h(c)
		}
		for i := len(e.premiddleware) - 1; i >= 0; i-- {
			h = e.premiddleware[i](h)
		}
	}
	// Execute chain
	if err := h(c); err != nil {
		e.HTTPErrorHandler(err, c)
	}

	// Release context
	e.pool.Put(c)
}
```

ServeHTTP函数先分配Context对象,然后Reset初始化，再初始化处理函数。

h处理函数先给一个初始值就是NotFound的处理，再检查是否有之前执行的中间件，

最后执行中间件执行链封装的函数h，然后h函数处理Context对象。

释放Context。

# middleware

```golang
// https://github.com/labstack/echo/blob/master/echo.go#L104
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// https://github.com/labstack/echo/blob/master/echo.go#L362
// Pre adds middleware to the chain which is run before router.
func (e *Echo) Pre(middleware ...MiddlewareFunc) {
	e.premiddleware = append(e.premiddleware, middleware...)
}

// Use adds middleware to the chain which is run after router.
func (e *Echo) Use(middleware ...MiddlewareFunc) {
	e.middleware = append(e.middleware, middleware...)
}

```

echo中间件使用HandlerFunc进行一层层装饰，最后返回一个HandlerFunc处理Context。

基于echo性能测试用例，发现增加echo中间件会有相对明显性能下降。


[1]: https://github.com/labstack/echo
[2]: https://godoc.org/github.com/labstack/echo