# golf

LICENSE：MIT

github: [https://github.com/dinever/golf][1]

golf框架实现简单，对应功能和扩展性相对弱一些，具有一定研究价值。

## example

```golang
package main

import "github.com/dinever/golf"

func mainHandler(ctx *golf.Context) {
  ctx.Send("Hello World!")
}

func pageHandler(ctx *golf.Context) {
  ctx.Send("Page: " + ctx.Param("page"))
}

func main() {
  app := golf.New()
  app.Get("/", mainHandler)
  app.Get("/p/:page/", pageHandler)
  app.Run(":9000")
}
```

main创建框架对象，注册两个路由然后启动，基本的web框架操作。

## Application

Application部分框架主体的定义

```golang
// https://github.com/dinever/golf/blob/master/app.go#L13
// Application is an abstraction of a Golf application, can be used for
// configuration, etc.
type Application struct {
	router *router

	// A map of string slices as value to indicate the static files.
	staticRouter map[string][]string

	// The View model of the application. View handles the templating and page
	// rendering.
	View *View

	// Config provides configuration management.
	Config *Config

	SessionManager SessionManager

	// NotFoundHandler handles requests when no route is matched.
	NotFoundHandler HandlerFunc

	// MiddlewareChain is the middlewares that Golf uses.
	middlewareChain *Chain

	pool sync.Pool

	errorHandler map[int]ErrorHandlerFunc

	// The default error handler, if the corresponding error code is not specified
	// in the `errorHandler` map, this handler will be called.
	DefaultErrorHandler ErrorHandlerFunc

	handlerChain HandlerFunc
}

// by godoc
func New() *Application
func (app *Application) Delete(pattern string, handler HandlerFunc)
func (app *Application) Error(statusCode int, handler ErrorHandlerFunc)
func (app *Application) Get(pattern string, handler HandlerFunc)
func (app *Application) Head(pattern string, handler HandlerFunc)
func (app *Application) Options(pattern string, handler HandlerFunc)
func (app *Application) Patch(pattern string, handler HandlerFunc)
func (app *Application) Post(pattern string, handler HandlerFunc)
func (app *Application) Put(pattern string, handler HandlerFunc)
func (app *Application) Run(addr string)
func (app *Application) RunTLS(addr, certFile, keyFile string)
func (app *Application) ServeHTTP(res http.ResponseWriter, req *http.Request)
func (app *Application) Static(url string, path string)
func (app *Application) Use(m ...MiddlewareHandlerFunc)
```

Application 对象是框架启动服务对象，里面有router、view、config、Session、Middleware五个对象，即该框架支持这些功能；

其他对象中、staticRouter处理静态文件路由，router是不公开（小写）的对象类型，还将static独立出来，就并没有提供自定义实现的可能，框架自己实现了一套路由；

pool是用来回收Context对象，减少GC，剩余属性就是各种处理函数

```golang
// https://github.com/dinever/golf/blob/master/app.go#L97
// Basic entrance of an `http.ResponseWriter` and an `http.Request`.
func (app *Application) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	ctx := app.pool.Get().(*Context)
	ctx.reset()
	ctx.Request = req
	ctx.Response = res
	ctx.App = app
	app.handlerChain(ctx)
	app.pool.Put(ctx)
}

// https://github.com/dinever/golf/blob/master/app.go#L113
// Run the Golf Application.
func (app *Application) Run(addr string) {
	err := http.ListenAndServe(addr, app)
	if err != nil {
		panic(err)
	}
}
```

Application的两个基本方法，ServeHTTP和Start，都是标准流程。

其他方法就是路由方法(Get/Post...)和一个中间件方法(Use)，详情自己翻[函数文档][2]。

### Start

golf对应实现的Start函数是`func (app *Application) Run(addr string)`和`func (app *Application) RunTLS(addr, certFile, keyFile string)`两个，功能一致。

在Run函数中直接调用标准库net/http的`func ListenAndServe(addr string, handler Handler) error`方法，给个函数监听地址和http.Handler就启动了，简单粗暴但是有效，RunTLS一样。

### ServeHTTP

```golang
// https://golang.org/src/net/http/server.go?s=2730:2793#L74
type Handler interface {
	ServeHTTP(ResponseWriter, *Request)
}
```

ServeHTTP函数是用来实现了http.Handler 接口，就是一个http.Handler对象，在Start的时候启动标准库Server，把App当作Handler传递进去就可以。

当http.Server处理一个请求的时，调用了传入的http.Handler对象的`ServeHTTP(ResponseWriter, *Request)`方法，然后对应函数处理，*Request是标准库反序列化好的请求对象，ResponseWriter是准备返回的对象。

而自己需要做的就是读取Request然后操作，将返回写入http.ResponseWriter就完成请求了，对应的就是ServeHTTP方法的实现。

在golf的ServeHTTP方法中，头尾两行是sync.Pool的操作，分配回收Context对象；接着4行初始化Context对象；第五行调用app.handlerChain对象把这个Context对象处理掉。

app.handlerChain是golf的函数指针对象，就保存请求处理链，和命名一样的。

app.handlerChain赋值就是`app.handlerChain = app.middlewareChain.Final(app.handler)`，在New()和Use()函数中才调用，New就是新建框架对象的初始化一下，Use是新增中间件以后刷新处理链。

### handler

```golang
// https://github.com/dinever/golf/blob/master/app.go#L72
// First search if any of the static route matches the request.
// If not, look up the URL in the router.
func (app *Application) handler(ctx *Context) {
	for prefix, staticPathSlice := range app.staticRouter {
		if strings.HasPrefix(ctx.Request.URL.Path, prefix) {
			for _, staticPath := range staticPathSlice {
				filePath := path.Join(staticPath, ctx.Request.URL.Path[len(prefix):])
				fileInfo, err := os.Stat(filePath)
				if err == nil && !fileInfo.IsDir() {
					staticHandler(ctx, filePath)
					return
				}
			}
		}
	}

	handler, params, err := app.router.FindRoute(ctx.Request.Method, ctx.Request.URL.Path)
	if err != nil {
		app.handleError(ctx, 404)
	} else {
		ctx.Params = params
		handler(ctx)
	}
	ctx.IsSent = true
}
```

handler就是golf框架标准的默认的Context处理方法。

先匹配一下是不是静态请求，然后调用路由器匹配一下，匹配到就返回参数和处理函数，然后处理请求。ctx.IsSent 就是处理标志位。

## Context

```golang
// https://github.com/dinever/golf/blob/master/context.go#L16
// Context is a wrapper of http.Request and http.ResponseWriter.
type Context struct {
	// http.Request
	Request *http.Request

	// http.ResponseWriter
	Response http.ResponseWriter

	// URL Parameter
	Params Parameter

	// HTTP status code
	statusCode int

	// The application
	App *Application

	// Session instance for the current context.
	Session Session

	// Indicating if the response is already sent.
	IsSent bool

	// Indicating loader of the template
	templateLoader string
}


// by godoc
func NewContext(req *http.Request, res http.ResponseWriter, app *Application) *Context
func (ctx *Context) Abort(statusCode int, data ...map[string]interface{})
func (ctx *Context) AddHeader(key, value string)
func (ctx *Context) ClientIP() string
func (ctx *Context) Cookie(key string) (string, error)
func (ctx *Context) Header(key string) string
func (ctx *Context) JSON(obj interface{})
func (ctx *Context) JSONIndent(obj interface{}, prefix, indent string)
func (ctx *Context) Loader(name string) *Context
func (ctx *Context) Param(key string) string
func (ctx *Context) Query(key string, index ...int) (string, error)
func (ctx *Context) Redirect(url string)
func (ctx *Context) Redirect301(url string)
func (ctx *Context) Render(file string, data ...map[string]interface{})
func (ctx *Context) RenderFromString(tplSrc string, data ...map[string]interface{})
func (ctx *Context) Send(body interface{})
func (ctx *Context) SendStatus(statusCode int)
func (ctx *Context) SetCookie(key string, value string, expire int)
func (ctx *Context) SetHeader(key, value string)
func (ctx *Context) StatusCode() int
```

Context就是一个请求的大小问，保存一次请求的数据和返回，并提供了基本操作。

从Api list和Context定义中可以看到函数基本功能就是: Cookie、Header、Query、Param、Session、Render、Binder(JSON)、Writer(Send、Statue)操作,简单的基本封装自己查看[源码][3]。

## router

source： https://github.com/dinever/golf/blob/master/router.go#L13


```golang

https://github.com/dinever/golf/blob/master/router.go#L13

type router struct {
	trees map[string]*node
}

// https://github.com/dinever/golf/blob/master/tree.go#L9

type node struct {
	text    string
	names   map[string]int
	handler HandlerFunc

	parent *node
	colon  *node

	children nodes
	start    byte
	max      byte
	indices  []uint8
}
```

## view

```golang

// https://github.com/dinever/golf/blob/master/view.go#L9

// View handles templates rendering
type View struct {
	FuncMap template.FuncMap

	// A view may have multiple template managers, e.g., one for the admin panel,
	// another one for the user end.
	templateLoader map[string]*TemplateManager
}

// https://github.com/dinever/golf/blob/master/template.go#L20

// TemplateLoader is the loader interface for templates.
type TemplateLoader interface {
	LoadTemplate(string) (string, error)
}
```

golf的view相当于独立实现or封装了一个模板渲染器。


## Config

```golang
// https://github.com/dinever/golf/blob/master/config.go#L36

// Config control for the application.
type Config struct {
	mapping map[string]interface{}
}
```

golf Config简单实现了一些Get/Set方法，没有学习价值，不分析。

## Session

```golang
// https://github.com/dinever/golf/blob/master/session.go#L16
// SessionManager manages a map of sessions.
type SessionManager interface {
	sessionID() (string, error)
	NewSession() (Session, error)
	Session(string) (Session, error)
	GarbageCollection()
	Count() int
}

// https://github.com/dinever/golf/blob/master/session.go#L89
// Session is an interface for session instance, a session instance contains
// data needed.
type Session interface {
	Set(key string, value interface{}) error
	Get(key string) (interface{}, error)
	Delete(key string) error
	SessionID() string
	isExpired() bool
}
```

Session通用实现，没什么特殊的,请参考独立的Session实现分析。


## middleware

```golang
// https://github.com/dinever/golf/blob/master/middleware.go#L12
// MiddlewareHandlerFunc defines the middleware function type that Golf uses.
type MiddlewareHandlerFunc func(next HandlerFunc) HandlerFunc

// Chain contains a sequence of middlewares.
type Chain struct {
	middlewareHandlers []MiddlewareHandlerFunc
}

// ......

// Final indicates a final Handler, chain the multiple middlewares together with the
// handler, and return them together as a handler.
func (c Chain) Final(fn HandlerFunc) HandlerFunc {
	for i := len(c.middlewareHandlers) - 1; i >= 0; i-- {
		fn = c.middlewareHandlers[i](fn)
	}
	return fn
}
```

HandlerFunc是一个Context处理函数。Chain中文是链，大概意思就是链式处理，Chain对象保存多个MiddlewareHandlerFunc函数，而MiddlewareHandlerFun
c是传入一个HandlerFunc然后返回一个HandlerFunc，在Final()函数中，传入了基本的处理，然后一层层构造然后一个最后HandlerFunc函数去处理请求，最后的处理函数就是多次嵌套构造的。和echo处理方法一致。


[1]: https://github.com/dinever/golf
[2]: https://godoc.org/github.com/dinever/golf#pkg-index
[3]: https://github.com/dinever/golf/blob/master/context.go#L16
