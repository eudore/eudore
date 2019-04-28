# golang基于context的web范式

适用框架:golf、echo、gin、dotweb、iris、beego。

golang大部分框架都是基于标准库net/http包现实，[fasthttp][1]框架就是自己解析http协议，从新实现了类似net/http包的功能。

通常框架包含的部分有Application、Context、Request、Response、Router、Middleware、Logger、Binder、render、View、Session、Cache这些部分，一般都是有部分，不过前五个是一定存在。

以下列出了各框架主要部分定义位置:

|   |  golf  |  echo  |  gin  |  dotweb  |  iris  |  beego |
| ------------ | ------------ | ------------ | ------------ | ------------ | ------------ | ------------ |
|  Application  |  [app.go][10]  |  [echo.go][11]  |  [gin.go][12]  | [dotweb.go][13]  |  [iris.go][14] |  [app.go][15]  |
|  Context  |  [context.go][16]  |  [context.go][17]  |  [context.go][18]  |  [context.go][19]  |  [context.go][20]  |  [context.go][21]  |
|  Request  |  http.Request  |  http.Request  |  http.Request  |  [request.go][35]  |  http.Request  |  [input.go][50]  |
|  Response  |  http.ResponseWriter  |  [response.go][25]  |  [response_writer_1.8.go][30]  |  [response.go][36]  |  [response_writer.go][43]  |  [output.go][51]  |
|  Router  |  [router.go][22]  |  [router.go][26]  |  [routergroup.go][31]  |  [router.go][37]  |  [router.go][44]  |  [router.go][52]  |
|  Middleware  |  [middleware.go][62]  |  [echo.go][27]  |  [gin.go][32]  |  [middleware.go][38]  |  [handler.go][45]  |  [app.go][53]  |
|  Logger  |    |  [log.go][28]  |    |  [logger.go][39]  |  [logger.go][46]  |  [log.go][54]  |
|  Binder  |    |  [bind.go][29]  |  [binding.go][33]  |  [bind.go][40]  |    |    |
|  Render  |    |    |  [render.go][34]  |  [render.go][41]  |    |    |
|  View  |  [view.go][23]  |    |    |    |  [engine.go][47]  |    |
|  Session  |  [session.go][24]  |    |  [session.go][63]  |  [session.go][42]  |  [session.go][48]  |  [session.go][55]  |
|  Cache  |    |    |    |  [cache.go][61]  |  [cache.go][49]  |  [cache.go][56]  |
|  Websocket  |    |    |    |  [websocket.go][57]  |  [server.go][58]  |    |
|  MVC  |    |    |    |    |  [controller.go][59]  |  [controller.go][60]  |


源码解析：[golf][2]


# Application

application一般都是框架的主体，通常XXX框架的主体叫做XXX，当然也有叫App、Application、Server的实现，具体情况不你一致，不过一般都是叫XXX，源码就是XXX.go。

这部分一般都会实现两个方法`Start() error`和`ServeHTTP(ResponseWriter, *Request)`

`Start() error`一般是框架的启动函数，用来启动服务，名称可能会是Run，创建一个http.Server对象，设置TLS相关等配置，然后启动服务，当然也会出现Shutdown方法。

`ServeHTTP(ResponseWriter, *Request)`函数实现http.Handler接口，一般框架都是使用http.Server对象来启动的服务，所以需要实现此方法。

此本方法大概就三步，Init、Handle、Release。

第一步Init，一般就是初始化Context对象，其中包括Request和Response的初始化Reset，使用`ResponseWriter`和`*Request`对象来初始化，通常会使用Sync.Pool来回收释放减少GC。

第二步Handle，通常就是框架处理请求，其中一定包含路由处理，使用路由匹配出对应的Handler来处理当前请求。

第三步释放Context等对象。

简单实现：

[Router简单实现](#Router)

```golang
// 定义Application
type Application struct {
	mux		*Router
	pool	sync.Pool
}

func NewApplication() *Application{
	return &Application{
		mux:	new(Router),
		pool:	sync.Pool{
			New:	func() interface{} {
				return &Context{}
			},
		}
	}
}

// 注册一个GET请求方法，其他类型
func (app *Application) Get(path string, handle HandleFunc) {
	app.RegisterFunc("GET", path, handle)
}

// 调用路由注册一个请求
func (app *Application) RegisterFunc(method string, path string, handle HandleFunc)  {
	app.router.RegisterFunc(method, path, handle)
}

// 启动Application
func (app *Application) Start(addr string) error {
	// 创建一个http.Server并启动
	return http.Server{
		Addr:		addr,
		Handler:	app,
	}.ListenAndServe()
}

// 实现http.Handler接口，并出去net/http请求。
func (app *Application) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 创建一个ctx对象
	ctx := app.pool.Get().(*Context)
	// 初始化ctx
	ctx.Reset(w, r)	
	// 路由器匹配请求并处理ctx
	app.router.Match(r.Method, r.URL.Path)(ctx)
	// 回收ctx
	app.pool.Put(ctx)
}
```

# Context

Context包含Request和Response两部分，Request部分是请求，Response是响应。Context的各种方法基本都是围绕Request和Response来实现来，通常就是各种请求信息的获取和写入的封装。

简单实现:

```golang
// Context简单实现使用结构体，不使用接口，如有其他需要可继续增加方法。
type Context struct {
	http.ResponseWriter
	req *http.Request
}

// 初始化ctx
func (ctx *Context) Reset(w http.ResponseWriter, r *http.Request) {
	ctx.ResponseWriter, ctx.req = w, r
}

// Context实现获取方法
func (ctx *Context) Method() string {
	return ctx.req.Method
}

```
# RequestReader & ResponseWriter

http协议解析[文档][my_proto_http_zh]。

实现RequestReader和ResponseWriter接口。

根据http协议请求报文和响应报文RequestReader和ResponseWriter大概定义如下：

```golang
type (
	Header map[string][]string
	RequestReader interface {
		Method() string
		RequestURI() string
		Proto() string
		Header() Header
		Read([]byte) (int, error)
	}
	ResponseWriter interface {
		WriteHeader(int)
		Header() Header
		Write([]byte) (int, error)
	}
)

```

RequestReader用于读取http协议请求的请求行(Request Line)、请求头(Request Header)、body。

ResponseWriter用于返回http写法响应的状态行(Statue Line)、响应头(Response Header)、body这些数据。

在实际过程还会加入net.Conn对象的tcp连接信息。

通常net/http库下的RequestReader和ResponseWriter定义为http.Request和http.ResponseWriter，请求是一个结构体，拥有请求信息，不同情况下可能会有不同封装，或者直接使用net/http定义的读写对象。

# Router

Router是请求匹配的路由，并不复杂，但是每个框架都是必备的。

通常实现两个方法Match和RegisterFunc，给路由器注册新路由，匹配一个请求的路由，然后处理请求。

```golang
type (
	HandleFunc func(*Context)
	Router interface{
		Match(string, string) HandleFunc
		RegisterFunc(string, string, HandleFunc)
	}
)
```

定义一个非常非常简单的路由器。

```golang
type Router struct {
	Routes	map[string]map[string]HandleFunc
}

// 匹配一个Context的请求
func (r *Router) Match(path ,method string) HandleFunc {
	// 查找方法定义的路由
	rs, ok := r.Routes[method]
	if !ok {
		return Handle405
	}
	// 查找路由
	h, ok := rs[path]
	if !ok {
		return Handle404
	}
	return h
}

// 注册路由处理函数
func (r *Router) RegisterFunc(method string, path string, handle HandleFunc) {
	rs, ok := r.Routes[ctx.Method()]
	if !ok {
		rs = make(map[string]HandleFunc)
		r.Routes[ctx.Method()] = rs
	}
	rs[path] = handle
}

// 处理405响应，方法不允许；Allow Header返回允许的方法。
func Handle405(ctx Context) {
	ctx.Response().WriteHeader(405)
	ctx.Response().Header().Add("Allow", "GET, POST, HEAD")
}


// 处理404响应，没有找到对应的资源。
func Handle404(ctx Context) {
	ctx.Response().WriteHeader(404)
}
```

这个简单路由仅支持了rsetful风格，连通配符匹配都没有实现；但是体现了路由器的作用，输出一个参数，返回一个对应的处理者。

至于如何对应一个处理路由，就是路由器规则的设定了；例如通配符、参数、正则、数据校验等功能。

# Middleware

通常是多个Handler函数组合，在handler之前之后增一些处理函数。

## echo

```golang
type MiddlewareFunc func(HandlerFunc) HandlerFunc

// echo.ServeHTTP
h := NotFoundHandler
for i := len(e.premiddleware) - 1; i >= 0; i-- {
	h = e.premiddleware[i](h)
}
if err := h(c); err != nil {
	e.HTTPErrorHandler(err, c)
}
```

echo中间件使用装饰器模式。

echo中间件使用HandlerFunc进行一层层装饰，最后返回一个HandlerFunc处理Context

## gin

gin在路由注册的会中间件和route合并成一个handlers对象，然后httprouter返回匹配返回handlrs，在context reset时设置ctx的handlers为路由匹配出现的，handlers是一个HanderFunc数组，Next方法执行下一个索引的HandlerFunc，如果在一个HandlerFunc中使用ctx.Next()就先将后续的HandlerFunc执行，后续执行完才会继续那个HandlerFunc，调用ctx.End() 执行索引直接修改为最大值，应该是64以上，毕竟Handlers合并时的数据长度限制是64，执行索引成最大值了，那么后面就没有HandlerFunc，就完整了一次ctx的处理。

```golang
type HandlerFunc func(*Context)

// https://github.com/gin-gonic/gin/blob/master/context.go#L105
func (c *Context) Next() {
	c.index++
	for s := int8(len(c.handlers)); c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}
```

echo通过路由匹配返回一个[]HandlerFunc对象，并保存到Context里面，执行时按ctx.index一个个执行。

如果HandlerFunc里面调用ctx.Next(),就会提前将后序HandlerFunc执行，返回执行ctx.Next()后的内容，可以简单的指定调用顺序，ctx.Next()之前的在Handler前执行，ctx.Next()之后的在Handler后执行。

Context.handlers存储本次请求所有HandlerFunc，然后使用c.index标记当然处理中HandlerFunc，

## dotweb

dotweb Middleware接口中Exclude()、HasExclude()、ExistsExcludeRouter()用来排除路由级调用。

BaseMiddlware一般是Middleware的基础类，通常只重写Handle()方法，同时可以调用Next()方法执行写一个Middleware。

在Next()方法中，如果next为空就调用Context的RouterNode()方法，获得RouterNode对象，然后使用AppMiddlewares()方法读取app级别[]Middleware，对其中第一个Middleware调用Handle()就会开始执行app级别Middleware。

同级多个Middleware是单链的形式存储，那么只有最后一个Middleware的next为空，如果执行到这个Middleware，那么表示这一级Middleware全部处理完毕，就需要Context的middleware级别降级，然后开始执行下一级。

ServerHttp()时，调用
```golang
//middleware执行优先级：
//优先级1：app级别middleware
//优先级2：group级别middleware
//优先级3：router级别middleware

// Middleware middleware interface
type Middleware interface {
	Handle(ctx Context) error
	SetNext(m Middleware)
	Next(ctx Context) error
	Exclude(routers ...string)
	HasExclude() bool
	ExistsExcludeRouter(router string) bool
}


type RouterNode interface {
    Use(m ...Middleware) *Node
    AppMiddlewares() []Middleware
    GroupMiddlewares() []Middleware
    Middlewares() []Middleware
    Node() *Node
}

// Use registers a middleware
func (app *DotWeb) Use(m ...Middleware) {
	step := len(app.Middlewares) - 1
	for i := range m {
		if m[i] != nil {
			if step >= 0 {
				app.Middlewares[step].SetNext(m[i])
			}
			app.Middlewares = append(app.Middlewares, m[i])
			step++
		}
	}
}


func (bm *BaseMiddlware) Next(ctx Context) error {
	httpCtx := ctx.(*HttpContext)
	if httpCtx.middlewareStep == "" {
		httpCtx.middlewareStep = middleware_App
	}
	if bm.next == nil {
		if httpCtx.middlewareStep == middleware_App {
			httpCtx.middlewareStep = middleware_Group
			if len(httpCtx.RouterNode().GroupMiddlewares()) > 0 {
				return httpCtx.RouterNode().GroupMiddlewares()[0].Handle(ctx)
			}
		}
		if httpCtx.middlewareStep == middleware_Group {
			httpCtx.middlewareStep = middleware_Router
			if len(httpCtx.RouterNode().Middlewares()) > 0 {
				return httpCtx.RouterNode().Middlewares()[0].Handle(ctx)
			}
		}

		if httpCtx.middlewareStep == middleware_Router {
			return httpCtx.Handler()(ctx)
		}
	} else {
		//check exclude config
		if ctx.RouterNode().Node().hasExcludeMiddleware && bm.next.HasExclude() {
			if bm.next.ExistsExcludeRouter(ctx.RouterNode().Node().fullPath) {
				return bm.next.Next(ctx)
			}
		}
		return bm.next.Handle(ctx)
	}
	return nil
}

func (x *xMiddleware) Handle(ctx Context) error {
	httpCtx := ctx.(*HttpContext)
	if httpCtx.middlewareStep == "" {
		httpCtx.middlewareStep = middleware_App
	}
	if x.IsEnd {
		return httpCtx.Handler()(ctx)
	}
	return x.Next(ctx)
}
```

## iris

```golang
// If Handler panics, the server (the caller of Handler) assumes that the effect of the panic was isolated to the active request.
// It recovers the panic, logs a stack trace to the server error log, and hangs up the connection.
type Handler func(Context)

// Handlers is just a type of slice of []Handler.
//
// See `Handler` for more.
type Handlers []Handler
```

# Logger

框架基础日志接口

```golang
type Logger interface {
	Debug(...interface{})
	Info(...interface{})
	Warning(...interface{})
	Error(...interface{})
	Fatal(...interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields Fields) Logger
}
```

# Binder & Render & View

Binder的作用以各种格式方法解析Request请求数据，然后赋值给一个interface{}。

Render的作用和Binder相反，会把数据安装选择的格式序列化，然后写回给Response部分。

View和Render的区别是Render输入的数据，View是使用对应的模板渲染引擎渲染html页面。


# Session & Cache

seesion是一种服务端数据存储方案

## 持久化

Cache持久化就需要把数据存到其他地方，避免程序关闭时缓存丢失。

存储介质一般各种DB、file等都可以，实现一般都是下面几步操作。

1、getid：从请求读取sessionid。

2、initSession：用sessionid再存储中取得对应的数据，一般底层就是[]byte

3、newSession：反序列化成一个map[string]interface{}这样类似结构Session对象。

4、Set and Get：用户对Session对象各种读写操作。

5、Release：将Session对象序列化成[]byte，然后写会到存储。


## 存储介质

- 内存：使用简单，供测试使用，无法持久化。

- 文件：存储简单。

- sqlite：和文件差不多，使用sql方式操作。

- mysql等：数据共享、数据库持久化。

- redis：数据共享、协议对缓存支持好。

- memcache：协议简单、方法少、效率好。

- etcd：少量数据缓存，可用性高。

## golang session

一组核心的session接口定义。

```go
type (
	Session interface {
		ID() string							// back current sessionID
		Set(key, value interface{}) error	// set session value
		Get(key interface{}) interface{}	// get session value
		Del(key interface{}) error			// delete session value
		Release(w http.ResponseWriter)		// release value, save seesion to store
	}
	type Provider interface {
		SessionRead(sid string) (Session, error)
	}
)
```

`Provider.SessionRead(sid string) (Session, error)`用sid来从Provider读取一个Session返回,sid就是sessionid一般存储与cookie中,也可以使用url参数值,Session对象会更具sid从对应存储中读取数据,然后将数据反序列化来初始化Session对象。

`Session.Release(w http.ResponseWriter)`从名称上是释放这个Seesion,但是一般实际作用是将对应Session对象序列化,然后存储到对应的存储实现中,如果只是读取Session可以不Release。

简单的Seession实现可以使用一个map，那么你的操作就是操作这个map。

在初始化Session对象的时候，使用sessionId去存储里面取数据，数据不在内存中，你们通常不是map，比较常用的是[]byte，例如memcache就是[]byte，[]byte可以map之间就需要序列化和反序列化了。

在初始化时，从存储读取[]map，反序列化成一个map，然后返回给用户操作；最后释放Session对象时，就要将map序列化成[]byte，然后再回写到存储之中，保存新修改的数据。

### Beego.Seesion

[源码github][2]

使用例子：

```go
func login(w http.ResponseWriter, r *http.Request) {
    sess, _ := globalSessions.SessionStart(w, r)
    defer sess.SessionRelease(w)
    username := sess.Get("username")
    if r.Method == "GET" {
        t, _ := template.ParseFiles("login.gtpl")
        t.Execute(w, nil)
    } else {
        sess.Set("username", r.Form["username"])
    }
}
```

在beego.session中Store就是一个Session。

1、SessionStart定义在[session.go#L193][3]，用户获取Store对象，在194line获取sessionid，在200line用sid从存储读取sessio.Store对象。

2、memcache存储在[124line][4]定义SessionRead函数来初始化对象。

3、其中在[sess_memcache.go#L139][5]使用memcache获取的数据，使用gob编码反序列化生成了map[interface{}]interface{}类型的值kv，然后144line把kv赋值给了store。

4、在57、65 [set&get][6]操作的rs.values，就是第三部序列化生成的对象。

5、最后释放store对象，定义在[94line][7]，先把rs.values使用gob编码序列化成[]byte，然后使用memcache的set方法存储到memcache里面了，完成了持久化存储。

源码分析总结：

- session在多线程读写是不安全的，数据可能冲突，init和release中间不要有耗时操作，参考其他思路一样，暂无解决方案，实现读写对存储压力大。

- 对session只读就不用释放session，只读释放是无效操作，因为set值和原值一样。

- beego.session可以适配一个beego.cache的后端，实现模块复用，不过也多封装了一层。



## golang cache

实现Get and Set接口的一种实现。

### 简单实现

```go
type Cache interface {
	Delete(key interface{})
	Load(key interface{}) (value interface{}, ok bool)
	Store(key, value interface{})
}
```

这组接口[sync.Map][]简化出来的，这种简单的实现了get&set操作，数据存储于内存中，map类型也可以直接实现存储。

### 复杂实现

封装各种DB存储实现接口即可，这样就可以实现共享缓存和持久化存储。

# Websocket

协议见[文档][my_proto_websocket_zh]

[my_proto_http_zh]: ../webname/proto_http_zh.md
[my_proto_websocket_zh]: ../webname/proto_websocket_zh.md

[1]: https://github.com/valyala/fasthttp
[2]: readDineverGolf_zh.md
[3]: readLabstackEcho_zh.md

[10]: https://github.com/dinever/golf/blob/master/app.go#L13
[11]: https://github.com/labstack/echo/blob/master/echo.go#L64
[12]: https://github.com/gin-gonic/gin/blob/master/gin.go#L51
[13]: https://github.com/devfeel/dotweb/blob/master/dotweb.go#L28
[14]: https://github.com/kataras/iris/blob/master/iris.go#L131
[15]: https://github.com/astaxie/beego/blob/master/app.go#L47
[16]: https://github.com/dinever/golf/blob/master/context.go#L16
[17]: https://github.com/labstack/echo/blob/master/context.go#L21
[18]: https://github.com/gin-gonic/gin/blob/master/context.go#L41
[19]: https://github.com/devfeel/dotweb/blob/master/context.go#L34
[20]: https://github.com/kataras/iris/blob/master/context/context.go#L215
[21]: https://github.com/astaxie/beego/blob/master/context/context.go#L59
[22]: https://github.com/dinever/golf/blob/master/router.go#L13
[23]: https://github.com/dinever/golf/blob/master/view.go#L10
[24]: https://github.com/dinever/golf/blob/master/session.go#L16
[25]: https://github.com/labstack/echo/blob/master/response.go#L13
[26]: https://github.com/labstack/echo/blob/master/router.go#L6
[27]: https://github.com/labstack/echo/blob/master/echo.go#L104
[28]: https://github.com/labstack/echo/blob/master/log.go#L11
[29]: https://github.com/labstack/echo/blob/master/bind.go#L16
[30]: https://github.com/gin-gonic/gin/blob/master/response_writer_1.8.go#L14
[31]: https://github.com/gin-gonic/gin/blob/master/routergroup.go#L41
[32]: https://github.com/gin-gonic/gin/blob/master/gin.go#L26
[33]: https://github.com/gin-gonic/gin/blob/master/binding/binding.go#L26
[34]: https://github.com/gin-gonic/gin/blob/master/render/render.go#L10
[35]: https://github.com/devfeel/dotweb/blob/master/request.go#L11
[36]: https://github.com/devfeel/dotweb/blob/master/response.go#L13
[37]: https://github.com/devfeel/dotweb/blob/master/router.go#L66
[38]: https://github.com/devfeel/dotweb/blob/master/middleware.go#L23
[39]: https://github.com/devfeel/dotweb/blob/master/logger/logger.go#L22
[40]: https://github.com/devfeel/dotweb/blob/master/bind.go#L18
[41]: https://github.com/devfeel/dotweb/blob/master/render.go#L14
[42]: https://github.com/devfeel/dotweb/blob/master/session/session.go#L25
[43]: https://github.com/kataras/iris/blob/master/context/response_writer.go#L24
[44]: https://github.com/kataras/iris/blob/master/core/router/router.go#L17
[45]: https://github.com/kataras/iris/blob/master/context/handler.go#L22
[46]: https://github.com/kataras/golog/blob/master/logger.go#L28
[47]: https://github.com/kataras/iris/blob/master/view/engine.go#L24
[48]: https://github.com/kataras/iris/blob/master/sessions/session.go#L16
[49]: https://github.com/kataras/iris/blob/master/cache/cache.go#L47
[50]: https://github.com/astaxie/beego/blob/master/context/input.go#L46
[51]: https://github.com/astaxie/beego/blob/master/context/output.go#L37
[52]: https://github.com/astaxie/beego/blob/master/router.go#L125
[53]: https://github.com/astaxie/beego/blob/master/app.go#L60
[54]: https://github.com/astaxie/beego/blob/master/logs/log.go#L87
[55]: https://github.com/astaxie/beego/blob/master/session/session.go#L56
[56]: https://github.com/astaxie/beego/blob/master/cache/cache.go#L49
[57]: https://github.com/devfeel/dotweb/blob/master/websocket.go#L8
[58]: https://github.com/kataras/iris/blob/master/websocket/server.go#L36
[59]: https://github.com/kataras/iris/blob/master/mvc/controller.go#L66
[60]: https://github.com/astaxie/beego/blob/master/controller.go#L68
[61]: https://github.com/devfeel/dotweb/blob/master/cache/cache.go#L8
[62]: https://github.com/dinever/golf/blob/master/middleware.go#L15
[63]: https://github.com/gin-gonic/contrib/blob/master/sessions/sessions.go#L37
