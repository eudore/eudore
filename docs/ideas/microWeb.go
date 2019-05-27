package main

import (
	"io"
	"os"
	"fmt"
	"time"
	"strings"
	"net/http"
)

type (
	// 框架主体
	App struct {
		Addr 		string
		Logger
		Router
		Server
		Middlewares		[]MiddlewareFunc
	}
	// 请求处理函数
	HandleFunc func(*Context)
	// 中间件函数
	MiddlewareFunc  func(HandleFunc) HandleFunc
	// 日志输出接口
	Logger interface {
		Print(...interface{})
		Printf(string, ...interface{})
	}
	// 路由器接口
	Router interface{
		Match(string, string) HandleFunc
		RegisterFunc(string, string, HandleFunc)
	}
	// 服务启动接口
	Server interface {
		ListenAndServe() error
	}
	// 请求上下文，封装请求操作，未详细实现。
	Context struct {
		*http.Request
		http.ResponseWriter
		Logger
	}
	// 基于map和遍历实现的简化路由器
	MyRouter struct {
		RoutesConst	map[string]HandleFunc
		RoutesPath	[]string
		RoutesFunc	[]HandleFunc
	}
	// 输出到标准输出的日志接口实现
	MyLogger struct {
		out 	io.Writer
	}
)

// 创建一个app
func NewApp() *App {
	return &App{
		Addr:	":8088",
		Logger:	&MyLogger{},
		Router:	&MyRouter{},
	}
}

// 启动App
func (app *App) Run() error {
	// Server初始化
	if app.Server == nil {
		app.Server = &http.Server{
			Addr:		app.Addr,
			Handler:	app,
		}
	}
	app.Printf("start server: %s", app.Addr)
	return app.Server.ListenAndServe()
}

// 处理Http请求
func (app *App) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	ctx := &Context{
		Request: req,
		ResponseWriter:	resp,
		Logger:	app,
	}
	// 路由匹配
	h := app.Router.Match(ctx.Method(), ctx.Path())
	// 处理中间件
	for _, i := range app.Middlewares {
		h = i(h)
	}
	// 处理请求
	h(ctx)
}

// App增加一个处理中间件
func (app *App) AddMiddleware(m ...MiddlewareFunc) {
	app.Middlewares = append(app.Middlewares, m...)
}



// 日志输出
func (l *MyLogger) Print(args ...interface{}) {
	if l.out == nil {
		l.out = os.Stdout
	}
	fmt.Print(time.Now().Format("2006-01-02 15:04:05 - "))
	fmt.Fprintln(l.out, args...)
}

// 日志输出
func (l *MyLogger) Printf(format string, args ...interface{}) {
	l.Print(fmt.Sprintf(format, args...))
}



// 匹配一个Context的请求
func (r *MyRouter) Match(method, path string) HandleFunc {
	// 查找路由
	path = method + path
	h, ok := r.RoutesConst[path]
	if ok {
		return h
	}
	for i, p := range r.RoutesPath {
		if strings.HasPrefix(path, p) {
			return r.RoutesFunc[i]
		}
	}
	return Handle404
}

// 处理404响应，没有找到对应的资源。
func Handle404(ctx *Context) {
	ctx.ResponseWriter.WriteHeader(404)
	ctx.ResponseWriter.Write([]byte("404 Not Found"))
}

// 注册路由处理函数
func (r *MyRouter) RegisterFunc(method string, path string, handle HandleFunc) {
	if r.RoutesConst == nil {
		r.RoutesConst = make(map[string]HandleFunc)
	}
	path = method + path
	if path[len(path) - 1] == '*' {
		r.RoutesPath = append(r.RoutesPath, path)
		r.RoutesFunc = append(r.RoutesFunc, handle)
	}else {
		r.RoutesConst[path] = handle
	}
}



// 获取请求方法
func (ctx *Context) Method() string {
	return ctx.Request.Method
}

// 获取请求路径
func (ctx *Context) Path() string {
	return ctx.Request.URL.Path
}

// 获取客户端真实地址
func (ctx *Context) RemoteAddr() string {
	xforward := ctx.Request.Header.Get("X-Forwarded-For")
	if "" == xforward {
		return strings.SplitN(ctx.Request.RemoteAddr, ":", 2)[0]
	}
	return strings.SplitN(string(xforward), ",", 2)[0]
}

func (ctx *Context) WriteString(s string) {
	ctx.ResponseWriter.Write([]byte(s))
}


// 日志中间件函数
func MiddlewareLoggerFunc(h HandleFunc) HandleFunc {
	return func(ctx *Context) {
		ctx.Printf("%s %s %s", ctx.RemoteAddr(), ctx.Method(), ctx.Path())
		h(ctx)
	}
}

func main() {
	app := NewApp()
	app.AddMiddleware(MiddlewareLoggerFunc)
	app.RegisterFunc("GET", "/hello", func(ctx *Context) {
		ctx.WriteString("hello micro web")
	})
	app.Run()
}
