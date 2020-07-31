package main

import (
	"bufio"
	"context"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"sync"
)

// App 组合主要功能接口，实现简单的基本方法。
type App struct {
	context.Context    `alias:"context"`
	context.CancelFunc `alias:"cancelfunc"`
	Config             `alias:"config"`
	Logger             `alias:"logger"`
	Server             `alias:"server"`
	Router             `alias:"router"`
	Binder             `alias:"binder"`
	Renderer           `alias:"renderer"`
	Validater          `alias:"validater"`
	GetWarp            `alias:"getwarp"`
	HandlerFuncs       `alias:"handlerfuncs"`
	ContextPool        sync.Pool `alias:"contextpool"`
}

func main() {
	_ = &App{}
}

// Config 定义配置管理，使用配置读写和解析功能。
type Config interface {
	Get(string) interface{}
	Set(string, interface{}) error
	ParseOption(ConfigParseOption)
	Parse() error
}

// ConfigParseFunc 定义配置解析函数。
type ConfigParseFunc func(Config) error

// ConfigParseOption 定义配置解析选项，用于修改配置解析函数。
type ConfigParseOption func([]ConfigParseFunc) []ConfigParseFunc

// Logger 定义日志处理器定义
type Logger interface {
	Logout
	Sync() error
	SetLevel(LoggerLevel)
}

// Logout 日志输出接口
type Logout interface {
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

// LoggerLevel 定义日志级别
type LoggerLevel int32

// func (l LoggerLevel) MarshalText() (text []byte, err error)
// func (l LoggerLevel) String() string
// func (l *LoggerLevel) UnmarshalText(text []byte) error

// Fields 定义多个日志属性
type Fields map[string]interface{}

// Server 定义启动http服务的对象。
type Server interface {
	SetHandler(http.Handler)
	Serve(net.Listener) error
	Shutdown(context.Context) error
}

// Router 接口，需要实现路由器方法、路由器核心两个接口。
//
// RouterCore实现路由匹配细节，RouterMethod调用RouterCore提供对外使用的方法。
//
// 任何时候请不要使用RouterCore的方法直接注册路由，应该使用RouterMethod的Add...方法。
type Router interface {
	RouterCore
	RouterMethod
}

// RouterCore接口，执行路由的注册和匹配一个请求并返回处理者。
//
// RouterCore主要实现路由匹配相关细节。
type RouterCore interface {
	HandleFunc(string, string, HandlerFuncs)
	Match(string, string, Params) HandlerFuncs
}

// RouterMethod 路由默认直接注册的接口，设置路由参数、组路由、中间件、函数扩展、控制器等行为。
type RouterMethod interface {
	Group(string) Router
	Params() *Params
	AddHandler(string, string, ...interface{}) error
	AddController(...Controller) error
	AddMiddleware(...interface{}) error
	AddHandlerExtend(...interface{}) error
	AnyFunc(string, ...interface{})
	GetFunc(string, ...interface{})
	PostFunc(string, ...interface{})
	PutFunc(string, ...interface{})
	DeleteFunc(string, ...interface{})
	HeadFunc(string, ...interface{})
	PatchFunc(string, ...interface{})
	OptionsFunc(string, ...interface{})
}

// Controller 定义控制器必要的接口。
//
// 控制器默认具有Base、Data、Singleton、View四种实现。
type Controller interface {
	Init(Context) error
	Release(Context) error
	Inject(Controller, Router) error
}

// Binder 定义Bind函数处理请求。
type Binder func(Context, io.Reader, interface{}) error

// Renderer 接口定义根据请求接受的数据类型来序列化数据。
type Renderer func(Context, interface{}) error

// Validater 接口定义验证器。
type Validater interface {
	RegisterValidations(string, ...interface{})
	Validate(interface{}) error
	ValidateVar(interface{}, string) error
}

// HandlerFunc 是处理一个Context的函数
type HandlerFunc func(Context)

// func (h HandlerFunc) String() string

// HandlerFuncs 是HandlerFunc的集合，表示多个请求处理函数。
type HandlerFuncs []HandlerFunc

// HandlerExtender 定义函数扩展处理者的方法。
//
// HandlerExtender默认拥有Base、Warp、Tree三种实现，具体参数三种对象的文档。
type HandlerExtender interface {
	RegisterHandlerExtend(string, interface{}) error
	NewHandlerFuncs(string, interface{}) HandlerFuncs
	ListExtendHandlerNames() []string
}

// Context 定义请求上下文接口。
type Context interface {
	// context
	Reset(context.Context, http.ResponseWriter, *http.Request)
	GetContext() context.Context
	Request() *http.Request
	Response() ResponseWriter
	Logger() Logout
	WithContext(context.Context)
	SetRequest(*http.Request)
	SetResponse(ResponseWriter)
	SetLogger(Logout)
	SetHandler(int, HandlerFuncs)
	GetHandler() (int, HandlerFuncs)
	Next()
	End()
	Err() error

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
	Validate(interface{}) error

	// param query header cookie session
	Params() *Params
	GetParam(string) string
	SetParam(string, string)
	AddParam(string, string)
	Querys() url.Values
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
	Push(string, *http.PushOptions) error
	Render(interface{}) error
	RenderWith(interface{}, Renderer) error
	WriteString(string) error
	WriteJSON(interface{}) error
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

// ResponseWriter 接口用于写入http请求响应体status、header、body。
//
// net/http.response实现了flusher、hijacker、pusher接口。
type ResponseWriter interface {
	// http.ResponseWriter
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(int)
	// http.Flusher
	Flush()
	// http.Hijacker
	Hijack() (net.Conn, *bufio.ReadWriter, error)
	// http.Pusher
	Push(string, *http.PushOptions) error
	Size() int
	Status() int
}

// SetCookie 定义响应返回的set-cookie header的数据生成
type SetCookie = http.Cookie

// Cookie 定义请求读取的cookie header的键值对数据存储
type Cookie struct {
	Name  string
	Value string
}

// Params 定义用于保存一些键值数据。
type Params struct {
	Keys []string
	Vals []string
}
