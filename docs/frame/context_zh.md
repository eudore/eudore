# Context

Context是请求上下文定义了主要方法，额外方法需要扩展，已有方法修改可以使用接口重写实现。
请求对象使用Context的Request和SetRequest方法读写。

Context主要分为请求上下文数据、请求、参数、响应、日志输出五部分。

Context的生命周期就是一个请求开始到结束，里面记录整个请求的数据。

Context的定义：

```golang
type Context interface {
	// context
	Reset(context.Context, protocol.ResponseWriter, protocol.RequestReader)
	Context() context.Context
	Request() protocol.RequestReader
	Response() protocol.ResponseWriter
	SetRequest(protocol.RequestReader)
	SetResponse(protocol.ResponseWriter)
	SetHandler(HandlerFuncs)
	Next()
	End()

	// request info
	Read([]byte) (int, error)
	Host() string
	Method() string
	Path() string
	RealIP() string
	RequestID() string
	Referer() string
	ContentType() string
	Istls() bool
	Body() []byte
	Bind(interface{}) error
	BindWith(interface{}, Binder) error

	// param query header cookie session
	Params() Params
	GetParam(string) string
	SetParam(string, string)
	AddParam(string, string)
	Querys() Querys
	GetQuery(string) string
	GetHeader(name string) string
	SetHeader(string, string)
	Cookies() []Cookie
	GetCookie(name string) string
	SetCookie(cookie *SetCookie)
	SetCookieValue(string, string, int)
	FormValue(string) string
	FormValues() map[string][]string
	FormFile(string) *multipart.FileHeader
	FormFiles() map[string][]*multipart.FileHeader

	// response
	Write([]byte) (int, error)
	WriteHeader(int)
	Redirect(int, string)
	Push(string, *protocol.PushOptions) error
	Render(interface{}) error
	RenderWith(interface{}, Renderer) error
	// render writer
	WriteString(string) error
	WriteJson(interface{}) error
	WriteFile(string) error

	// log Logout interface
	Debug(...interface{})
	Info(...interface{})
	Warning(...interface{})
	Error(...interface{})
	Fatal(...interface{})
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warningf(string, ...interface{})
	Errorf(string, ...interface{})
	Fatalf(string, ...interface{})
	WithField(key string, value interface{}) Logout
	WithFields(fields Fields) Logout
}
```

## 请求上下文部分数据

该部分主要是读写一些基本数据和中间件机制操作。

```golang
// context.go
type Context interface {
	// context
	Reset(context.Context, protocol.ResponseWriter, protocol.RequestReader)
	Context() context.Context
	Request() protocol.RequestReader
	Response() protocol.ResponseWriter
	SetRequest(protocol.RequestReader)
	SetResponse(protocol.ResponseWriter)
	SetHandler(HandlerFuncs)
	Next()
	End()
	...
}

```

`Reset(context.Context, protocol.ResponseWriter, protocol.RequestReader)`

Reset方法在EudoreHTTP中来使用http请求数据初始化ctx对象。

`Context() context.Context`

获得当前ctx的context.Context

`Request() protocol.RequestReader` 和 `Response() protocol.ResponseWriter`

获取ctx的请求和响应对象，允许直接操作ctx的底层请求对象。

`SetRequest(protocol.RequestReader)` 和 `SetResponse(protocol.ResponseWriter)`

设置请求和响应对象,可以用来重写ctx的请求和响应，例如gzip响应中间件实现。

`SetHandler(HandlerFuncs)`

设置上下文处理者，通常在ctx初始化后，然后调用Router.Match匹配得到多个请求处理者，设置给上下文，最后调用Next开始处理；未决定是否合并到Reset方法中，

`Next()`

调用下一个请求处理者开始处理，通过Next方法可以巧妙实现请求后处理请求。

```golang
ctx.Println("前执行")
ctx.Next()
fmt.Println("后执行")
```

`End()`

结束ctx的处理，忽略全部剩余的请求处理者，未实现获取是否结束处理状态。

同时结束Conext的生命周期

## 请求信息

请求部分主要是请求行、部分header、body的读取，header使用`ctx.Request().Header()`读取，`io.Reader`接口直接读取body的数据，Body方法读取全部内容，其他方法见方法名称。

```golang
type Context interface {
	// request info
	Read([]byte) (int, error)
	Host() string
	Method() string
	Path() string
	RealIP() string
	RequestID() string
	Referer() string
	ContentType() string
	Istls() bool
	Body() []byte
	Bind(interface{}) error
	BindWith(interface{}, Binder) error
	...
}
```

`Read([]byte) (int, error)`

实现io.Reader接口，可以直接读取请求body，和RequestReader.Read()方法一样。

`Host() string`

获取请求的Host

`Method() string`

获取请求的方法

`Path() string`

获取请求的路径

`RealIP() string`

获取请求远程真实ip地址，http连接的地址通过RequestReader.RemoteAddr()方法获取。

`RequestID() string`

获取`X-Request-ID` http header

`Referer() string`

获取`Referer` http header

`ContentType() string`

获取`Content-Type` http header

`Istls() bool`

获取是否是Tls连接(是否为https)，可以使用RequestReader.TLS()活动tls连接状态。

`Body() []byte`

获取请求Body

`Bind(interface{}) error`

调用Binder对象解析请求，并给对象绑定数据，根据请求Context-Type Header来决定Bind方法。

`BindWith(interface{}, Binder) error`

使用指定的Binder解析请求。

## 参数

参数部分是对param query header cookie form五部分读写，Params是Context的参数都是字符串类型，query是http uri参数，header分为请求header和响应header，cookie和header类似，form是对form请求解析的数据读取。

```golang
type Context interface {
	// param query header cookie form
	Params() Params
	GetParam(string) string
	SetParam(string, string)
	AddParam(string, string)
	Querys() Querys
	GetQuery(string) string
	GetHeader(name string) string
	SetHeader(string, string)
	Cookies() []Cookie
	GetCookie(name string) string
	SetCookie(cookie *SetCookie)
	SetCookieValue(string, string, int)
	FormValue(string) string
	FormValues() map[string][]string
	FormFile(string) *multipart.FileHeader
	FormFiles() map[string][]*multipart.FileHeader
	...
}
```
### Params

`Params() Params`
`GetParam(string) string`
`SetParam(string, string)`
`AddParam(string, string)`

### Query

`GetQuery(string) string`

获得请求uri中的参数，可以使用RequestReader.RequestURI()获得请求行中的uri。

### Header

`GetHeader(name string) string`

获取请求Header

相当于ctx.Request().Header().Get()

`SetHeader(string, string)`

设置响应Header

相对于ctx.Response().Header().Set()

### Cookie

`Cookies() []*Cookie`

获取全部请求Cookies

`GetCookie(name string) string`

获取指定请求Cookie的值

`SetCookie(cookie *SetCookie)`

设置响应Cookie，实现为给响应设置一个`Set-Cookie` header。

`SetCookieValue(string, string, int)`

设置响应Cookie

### Form

	FormValue(string) string
	FormValues() map[string][]string
	FormFile(string) *multipart.FileHeader
	FormFiles() map[string][]*multipart.FileHeader

## 响应

响应部分Write和WriteHeader是主要部分，其他方法就是封装写入不同的类型，header可以使用ctx.SetHeader和ctx.Response().Header()来操作。

```golang
type Context interface {
	// response
	Write([]byte) (int, error)
	WriteHeader(int)
	Redirect(int, string)
	Push(string, *protocol.PushOptions) error
	Render(interface{}) error
	RenderWith(interface{}, Renderer) error
	// render writer
	WriteString(string) error
	WriteJson(interface{}) error
	WriteFile(string) error
	...
}
```

`Write([]byte) (int, error)`

写入响应数据，和ResponseWriter.Write()一样。

net/http实现中，在第一次数据后无法写入Header和Status。

当前eudore/protocol/http中，在写入2k数据后无法写入Header和Status。

`WriteHeader(int)`

写入响应状态码

`Redirect(int, string)`

返回重定向，实现通过返回30x状态码和Location header记录重定向地址。

`Push(string, *protocol.PushOptions) error`

h2 Push资源,调用ResponseWriter.Push()。

`WriteString(string) error`

写入字符串

`WriteView(string, interface{}) error`

写入渲染模板，调用View对象渲染。

`WriteJson(interface{}) error`

写入Json数据

`WriteFile(string) error`

写入文件内容

`Render(interface{}) error`

调用Renderer渲染数据成对于格式，根据请求Accept Header来决定Render方法。

`RenderWith(interface{}, Renderer) error`

使用指定Renderer处理数据渲染。

### 日志

日志和Logout接口相同，定义输出日志信息，默认Context会输出调用文件行和请求id信息。

```golang
type Context interface {
	// log Logout interface
	Debug(...interface{})
	Info(...interface{})
	Warning(...interface{})
	Error(...interface{})
	Fatal(...interface{})
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warningf(string, ...interface{})
	Errorf(string, ...interface{})
	Fatalf(string, ...interface{})
	WithField(key string, value interface{}) Logout
	WithFields(fields Fields) Logout
	...
}
```

## 其他

曾经Context实现了context.Context接口，后来删除了实现改为`Context() context.Context`方法返回对象，因为自己实现的Context对象在context中会出现一些无法处理的问题，例如cancel取消子ctx，在cancelCtx对象的parentCancelCtx方法实现中使用了断言，自己实现的ctx无法被断言到会出现问题，具体参考context包源码。