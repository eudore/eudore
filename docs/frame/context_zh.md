# Context

Context是一次请求的上下文环境，接口大概分类为：context设置、请求数据读取、上下文数据、响应写入、数据解析、日志输出这四类。

Context的生命周期就是一个请求开始到结束，里面记录整个请求的数据。

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

	// param header cookie session
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
	GetSession() SessionData
	SetSession(SessionData)


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
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warningf(string, ...interface{})
	Errorf(string, ...interface{})
	Fatalf(string, ...interface{})
	WithField(key string, value interface{}) LogOut
	WithFields(fields Fields) LogOut
	// app
	App() *App
}
```

## 接口详解

### context设置

`Reset(context.Context, protocol.ResponseWriter, protocol.RequestReader)`

Reset方法在EudoreHTTP中来使用http请求数据初始化ctx对象。

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

`NewRequest(string, string, io.Reader) (protocol.ResponseReader, error)`

使用客户端发起一次http请求。

`// context`
`Deadline() (time.Time, bool)`
`Done() <-chan struct{}`
`Err() error`
`Value(key interface{}) interface{}`
`SetValue(interface{}, interface{})`

未实现

### 请求信息

`Read([]byte) (int, error)`

实现io.Reader接口，可以直接读取请求body，和RequestReader.Read()方法一样。

`Host() string`

获取请求的Host

`Method() string`

获取请求的方法

`Path() string`

获取请求的路径

`RemoteAddr() string`

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

### 上下文数据

#### Params

`Params() Params`
`GetParam(string) string`
`SetParam(string, string)`
`AddParam(string, string)`

#### Query

`GetQuery(string) string`

获得请求uri中的参数，可以使用RequestReader.RequestURI()获得请求行中的uri。

#### Header

`GetHeader(name string) string`

获取请求Header

相当于ctx.Request().Header().Get()

`SetHeader(string, string)`

设置响应Header

相对于ctx.Response().Header().Set()

#### Cookie

`Cookies() []*Cookie`

获取全部请求Cookies

`GetCookie(name string) string`

获取指定请求Cookie的值

`SetCookie(cookie *SetCookie)`

设置响应Cookie，实现为给响应设置一个`Set-Cookie` header。

`SetCookieValue(string, string, int)`

设置响应Cookie

#### Session

`GetSession() SessionData`

获取请求的会话数据

`SetSession(SessionData)`

给请求设置会话数据

### 写入响应

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

### 数据解析

`ReadBind(interface{}) error`

调用Binder对象解析请求，并给对象绑定数据，根据请求Context-Type Header来决定Bind方法。

`WriteRender(interface{}) error`

调用Renderer渲染数据成对于格式，根据请求Accept Header来决定Render方法。

### 日志输出

实现请求上下文日志输出，封装Logger对象的处理，可以附件ctx的Field。

```golang
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
WithField(key string, value interface{}) LogOut
WithFields(fields Fields) LogOut
```

### App

`App() *App`

获得请求对象的App，**用途未知，可能移除**，推荐传入App然后闭包返回eudore.HanderFunc对象，