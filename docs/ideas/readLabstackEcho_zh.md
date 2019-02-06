# echo

github: [https://github.com/labstack/echo][1]



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



[1]: https://github.com/labstack/echo
