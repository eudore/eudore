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

type App struct {
	context.Context    `alias:"context"`
	context.CancelFunc `alias:"cancelfunc"`
	Logger             `alias:"logger"`
	Config             `alias:"config"`
	Router             `alias:"router"`
	Client             `alias:"client"`
	Server             `alias:"server"`
	GetWrap            `alias:"getwrap"`
	HandlerFuncs       `alias:"handlerfuncs"`
	ContextPool        *sync.Pool `alias:"contextpool"`
	CancelError        error      `alias:"cancelerror"`
	Mutex              sync.Mutex `alias:"mutex"`
	Values             []any      `alias:"values"`
}

func main() {
	_ = &App{}

	fmt.Println("contextBase", unsafe.Sizeof(contextBase{}), "256-288")
	fmt.Println("loggerStd", unsafe.Sizeof(loggerStd{}), "160")
	fmt.Println("LoggerEntry", unsafe.Sizeof(LoggerEntry{}), "112-128")
	fmt.Println("nodeMux", unsafe.Sizeof(nodeMux{}), "224-240")
}

type Logger interface {
	Debug(...any)
	Info(...any)
	Warning(...any)
	Error(...any)
	Fatal(...any)
	Debugf(string, ...any)
	Infof(string, ...any)
	Warningf(string, ...any)
	Errorf(string, ...any)
	Fatalf(string, ...any)
	WithField(string, any) Logger
	WithFields([]string, []any) Logger
	GetLevel() LoggerLevel
	SetLevel(LoggerLevel)
}

type LoggerLevel int

type loggerStd struct {
	LoggerEntry
	Handlers []LoggerHandler
	Pool     *sync.Pool
	Logger   bool
	Depth    int32
}

type LoggerEntry struct {
	Level   LoggerLevel
	Time    time.Time
	Message string
	Keys    []string
	Vals    []any
	Buffer  []byte
}

type LoggerHandler interface {
	HandlerPriority() int
	HandlerEntry(*LoggerEntry)
}

type Config interface {
	Get(string) any
	Set(string, any) error
	ParseOption(...ConfigParseFunc)
	Parse(context.Context) error
}

type ConfigParseFunc func(context.Context, Config) error

type Client interface {
	NewRequest(string, string, ...any) error
	WithOptions(...any) Client
	GetRequest(string, ...any) error
	PostRequest(string, ...any) error
	PutRequest(string, ...any) error
	DeleteRequest(string, ...any) error
	HeadRequest(string, ...any) error
	PatchRequest(string, ...any) error
}

type Server interface {
	SetHandler(http.Handler)
	Serve(net.Listener) error
	Shutdown(context.Context) error
}

type Router interface {
	RouterCore
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

type RouterCore interface {
	HandleFunc(string, string, HandlerFuncs)
	Match(string, string, *Params) HandlerFuncs
}

type nodeMux struct {
	path    string
	name    string
	route   string
	childc  []*nodeMux
	childpv []*nodeMux
	childp  []*nodeMux
	childwv []*nodeMux
	childw  *nodeMux
	check   func(string) bool
	// handlers
	handlers   []nodeMuxHandler
	anyHandler []HandlerFunc
	anyParams  Params
}

type nodeMuxHandler struct {
	method string
	params Params
	funcs  []HandlerFunc
}

type Controller interface {
	Inject(Controller, Router) error
}

type HandlerFunc func(Context)

type HandlerFuncs []HandlerFunc

type HandlerExtender interface {
	RegisterExtender(string, any) error
	CreateHandlers(string, any) HandlerFuncs
	List() []string
}

type Context interface {
	// context
	Reset(w http.ResponseWriter, r *http.Request)
	Context() context.Context
	Request() *http.Request
	Response() ResponseWriter
	Value(key any) any
	SetContext(c context.Context)
	SetRequest(r *http.Request)
	SetResponse(w ResponseWriter)
	SetValue(key any, val any)
	SetHandlers(index int, handlers []HandlerFunc)
	GetHandlers() (int, []HandlerFunc)
	Next()
	End()
	Err() error
	Done() <-chan struct{}

	// request
	Read(b []byte) (int, error)
	Host() string
	Method() string
	Path() string
	RealIP() string
	Body() []byte
	Bind(data any) error

	// param query header cookie form
	Params() *Params
	GetParam(key string) string
	SetParam(key string, val string)
	Querys() (url.Values, error)
	GetQuery(key string) string
	GetHeader(key string) string
	SetHeader(key string, val string)
	Cookies() []Cookie
	GetCookie(key string) string
	SetCookie(cookie *CookieSet)
	SetCookieValue(name string, value string, age int)
	FormValue(key string) string
	FormValues() (map[string][]string, error)
	FormFile(key string) *multipart.FileHeader
	FormFiles() map[string][]*multipart.FileHeader

	// response
	Write(b []byte) (int, error)
	WriteString(s string) (int, error)
	WriteStatus(code int)
	WriteHeader(code int)
	WriteFile(path string) error
	Redirect(code int, url string)
	Render(data any) error

	// Logger interface
	Debug(args ...any)
	Info(args ...any)
	Warning(args ...any)
	Error(args ...any)
	Fatal(args ...any)
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warningf(format string, args ...any)
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
	WithField(key string, val any) Logger
	WithFields(keys []string, vals []any) Logger
}

type ResponseWriter interface {
	Header() http.Header
	Write([]byte) (int, error)
	WriteHeader(int)
	WriteStatus(code int)
	Flush()
	Hijack() (net.Conn, *bufio.ReadWriter, error)
	Push(string, *http.PushOptions) error
	Size() int
	Status() int
}

type responseWriterHTTP struct {
	http.ResponseWriter
	code int
	size int
}

type CookieSet = http.Cookie

type Cookie struct {
	Name  string
	Value string
}

type Params []string

type contextBase struct {
	// context
	index          int
	handlers       []HandlerFunc
	params         Params
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
	Logger                 Logger
	Bind                   func(Context, any) error
	Render                 func(Context, any) error
	MaxApplicationFormSize int64
	MaxMultipartFormMemory int64
}

type contextBaseValue struct {
	sync.RWMutex
	context.Context
	Logger
	Error  error
	Values []any
}

type GetWrap func(string) interface{}
