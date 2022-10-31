package main

import (
	"bufio"
	"context"
	"fmt"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
	"unsafe"
)

// App 组合主要功能接口，实现简单的基本方法。
type App struct {
	context.Context    `alias:"context"`
	context.CancelFunc `alias:"cancelfunc"`
	Logger             `alias:"logger"`
	Config             `alias:"config"`
	Database           `alias:"database"`
	Client             `alias:"client"`
	Server             `alias:"server"`
	Router             `alias:"router"`
	GetWarp            `alias:"getwarp"`
	HandlerFuncs       HandlerFuncs `alias:"handlerfuncs"`
	ContextPool        *sync.Pool   `alias:"contextpool"`
	CancelError        error        `alias:"cancelerror"`
	cancelMutex        sync.Mutex
	Values             []interface{}
}

func main() {
	_ = &App{}

	fmt.Println("contextBase", unsafe.Sizeof(contextBase{}))
	fmt.Println("LoggerStd", unsafe.Sizeof(LoggerStd{}))
	fmt.Println("stdNode", unsafe.Sizeof(stdNode{}))
}

// Logger 日志输出接口
type Logger interface {
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
	WithField(string, interface{}) Logger
	WithFields([]string, []interface{}) Logger
	GetLevel() LoggerLevel
	SetLevel(LoggerLevel)
	Sync() error
}

// LoggerLevel 定义日志级别
type LoggerLevel int32

type LoggerStd struct {
	LoggerStdData
	// enrty data
	Time       time.Time
	Message    string
	Keys       []string
	Vals       []interface{}
	Buffer     []byte
	Timeformat string
	// 日志标识 true是Logger false是Entry
	Logger bool
	Level  LoggerLevel
	Depth  int
}

// LoggerStdData 定义loggerStd的数据存储
type LoggerStdData interface {
	GetLogger() *LoggerStd
	PutLogger(*LoggerStd)
	Sync() error
}

// Config 定义配置管理，使用配置读写和解析功能。
type Config interface {
	Get(string) interface{}
	Set(string, interface{}) error
	ParseOption([]ConfigParseFunc) []ConfigParseFunc
	Parse() error
}

// ConfigParseFunc 定义配置解析函数。
type ConfigParseFunc func(Config) error

type Database interface {
	AutoMigrate(interface{}) error
	Query(context.Context, interface{}, DatabaseStmt) error
	Exec(context.Context, DatabaseStmt) error
}

type DatabaseStmt interface {
	Build(DatabaseBuilder)
}
type DatabaseBuilder interface {
	Context() context.Context
	WriteStmts(...interface{})
	Result() (string, []interface{}, error)
}

type Client interface {
	Do(*http.Request) (*http.Response, error)
}

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
// RouterMethod 路由默认直接注册的接口，设置路由参数、组路由、中间件、函数扩展、控制器等行为。
//
// 任何时候请不要使用RouterCore的方法直接注册路由，应该使用RouterMethod的Add...方法。
type Router interface {
	RouterCore
	// RouterMethod method
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
}

// RouterCore接口，执行路由的注册和匹配一个请求并返回处理者。
//
// RouterCore主要实现路由匹配相关细节。
type RouterCore interface {
	HandleFunc(string, string, HandlerFuncs)
	Match(string, string, *Params) HandlerFuncs
}

type stdNode struct {
	isany uint16
	kind  uint16
	pnum  uint32
	check func(string) bool
	path  string
	name  string
	route string

	// 默认标签的名称和值
	params     [7]Params
	handlers   [7]HandlerFuncs
	others     map[string]stdOtherHandler
	Wchildren  *stdNode
	Cchildren  []*stdNode
	Pchildren  []*stdNode
	PVchildren []*stdNode
	WVchildren []*stdNode
}

type stdOtherHandler struct {
	any     bool
	params  Params
	handler HandlerFuncs
}

// Controller 定义控制器必要的接口。
type Controller interface {
	Inject(Controller, Router) error
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
	Reset(http.ResponseWriter, *http.Request)
	GetContext() context.Context
	Request() *http.Request
	Response() ResponseWriter
	Value(interface{}) interface{}
	SetContext(context.Context)
	SetRequest(*http.Request)
	SetResponse(ResponseWriter)
	SetValue(interface{}, interface{})
	// handles
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
	ContentType() string
	Istls() bool
	Body() []byte
	Bind(interface{}) error

	// param query header cookie form
	Params() *Params
	GetParam(string) string
	SetParam(string, string)
	Querys() url.Values
	GetQuery(string) string
	GetHeader(string) string
	SetHeader(string, string)
	Cookies() []Cookie
	GetCookie(string) string
	SetCookie(*SetCookie)
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
	WriteString(string) error
	WriteFile(string) error

	// Database interface
	Query(interface{}, DatabaseStmt) error
	Exec(DatabaseStmt) error
	NewRequest(string, string, ...interface{}) (*http.Response, error)
	// Logger interface
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
	WithField(string, interface{}) Logger
	WithFields([]string, []interface{}) Logger
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

// responseWriterHTTP 是对net/http.ResponseWriter接口封装
type responseWriterHTTP struct {
	http.ResponseWriter
	code int
	size int
}

// SetCookie 定义响应返回的set-cookie header的数据生成
type SetCookie = http.Cookie

// Cookie 定义请求读取的cookie header的键值对数据存储
type Cookie struct {
	Name  string
	Value string
}

// Params 定义用于保存一些键值数据。
type Params []string

// contextBase 实现Context接口。
type contextBase struct {
	// context
	index          int
	handler        HandlerFuncs
	httpParams     Params
	config         *contextBaseConfig
	RequestReader  *http.Request
	ResponseWriter ResponseWriter
	context        context.Context
	// data
	contextValues contextBaseValue
	httpResponse  responseWriterHTTP
	cookies       []Cookie
	bodyContent   []byte
}

type contextBaseConfig struct {
	Logger   Logger
	Database Database
	Client   Client
	Bind     func(Context, interface{}) error
	Validate func(Context, interface{}) error
	Filte    func(Context, interface{}) error
	Render   func(Context, interface{}) error
}

type contextBaseValue struct {
	context.Context
	Logger   Logger
	Database Database
	Client   Client
	Error    error
	Values   []interface{}
}

// GetWarp 对象封装Get函数提供类型转换功能。
type GetWarp func(string) interface{}
