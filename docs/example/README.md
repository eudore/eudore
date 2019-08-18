# example

本部分主要演示文档，详细文档查看[框架文档](../frame)或者源码。

## new app

```golang
func main() {
	app := eudore.NewCore()
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore")
	})
	app.Listen(":8080")
	app.Run()
}
```

## Context

Context是请求上下文定义了主要方法，额外方法需要扩展，已有方法修改可以使用接口重写实现。
请求对象使用Context的Request和SetRequest方法读写。

Context主要分为请求上下文数据、请求、参数、响应、日志输出五部分。

### 请求上下文部分数据

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

### 请求

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

### 参数

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
#### uri参数

```golang
app.GetFunc(func(ctx eudore.Context) {
	ctx.GetQuery("val")
})
```

### 响应

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


## Handler

eudore几乎支持任意处理函数，只需要额外注册一下转换函数，将任意函数转换成eudore.HanderFunc对象即可。

可以使用eudore.ListExtendHandlerFun()函数查看内置支持的任意函数类型，如果是注册了不支持的处理函数类型会触发panic。

内置处理函数类型：

```
func(eudore.Context) error
func(eudore.Context) (interface {}, error)
func(eudore.ContextData)
func(eudore.Context, map[string]interface {}) (map[string]interface {}, error)
```

```golang

func main() {
	app := eudore.NewCore()
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore")
	})
	app.GetFunc("/check", func(ctx eudore.Context) error {
		if len(ctx.GetQuery("value")) > 3 {
			return fmt.Errorf("value is %s len great 3", ctx.GetQuery("value"))
		}
		return nil
	})
	app.GetFunc("/data", func(ctx eudore.Context) (interface{}, error) {
		return map[string]string{
			"a": "1",
			"b": "2",
		}, nil
	})
	app.Listen(":8080")
	app.Run()
}
```

### 实现一个扩展函数

MyContext额外实现了一个Hello方法，然后使用eudore.RegisterHandlerFunc注册一个转换函数，转换函数要求参数是一个函数，返回参数是一个eudore.HandlerFunc。

闭包一个`func(fn func(MyContext)) eudore.HandlerFunc`转换函数，就将MyContext类型处理函数转换成了eudore.HandlerFunc，然后就可以使用路由注册自己定义的处理函数。

```golang
type MyContext eudore.Context

func (ctx MyContext) Hello() {
	ctx.WriteString("hellp")
}

func main() {
	eudore.RegisterHandlerFunc(func(fn func(MyContext)) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			fn(MyContext(ctx))
		}
	}) 

	app := eudore.NewCore()
	app.GetFunc("/*", func(ctx MyContext) {
		ctx.Hello()
	})
}
```

### Middleware

eudore的Middleware是一个函数，类型是eudore.HandlerFunc,可以在中间件类型使用ctx.Next来调用先后续处理函数，然后再继续执行定义的内容，ctx.End直接忽略后续处理。

ctx.Fatal默认会调用ctx.End，也是和ctx.Error的区别。

```golang
func main() {
	app := eudore.NewCore()
	app.AddMiddleware(func(ctx eudore.Context) {
		...
		ctx.Next()
		...
	})
}
```

