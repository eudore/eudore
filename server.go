package eudore

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"
)

type (
	// Server 定义启动http服务的对象。
	Server interface {
		SetHandler(http.Handler)
		Serve(net.Listener) error
		Shutdown(ctx context.Context) error
	}
	// ServerConfigStd 定义ServerStd使用的配置
	ServerConfigStd struct {
		// ReadTimeout is the maximum duration for reading the entire
		// request, including the body.
		//
		// Because ReadTimeout does not let Handlers make per-request
		// decisions on each request body's acceptable deadline or
		// upload rate, most users will prefer to use
		// ReadHeaderTimeout. It is valid to use them both.
		ReadTimeout time.Duration `set:"readtimeout" description:"Http server read timeout."`

		// ReadHeaderTimeout is the amount of time allowed to read
		// request headers. The connection's read deadline is reset
		// after reading the headers and the Handler can decide what
		// is considered too slow for the body.
		ReadHeaderTimeout time.Duration // Go 1.8

		// WriteTimeout is the maximum duration before timing out
		// writes of the response. It is reset whenever a new
		// request's header is read. Like ReadTimeout, it does not
		// let Handlers make decisions on a per-request basis.
		WriteTimeout time.Duration `set:"writetimeout" description:"Http server write timeout."`

		// IdleTimeout is the maximum amount of time to wait for the
		// next request when keep-alives are enabled. If IdleTimeout
		// is zero, the value of ReadTimeout is used. If both are
		// zero, ReadHeaderTimeout is used.
		IdleTimeout time.Duration // Go 1.8

		// MaxHeaderBytes controls the maximum number of bytes the
		// server will read parsing the request header's keys and
		// values, including the request line. It does not limit the
		// size of the request body.
		// If zero, DefaultMaxHeaderBytes is used.
		MaxHeaderBytes int
	}
	// ServerStd 定义使用net/http启动http server。
	ServerStd struct {
		*http.Server
		Print func(...interface{}) `set:"print"`
	}
	// netHTTPLog 实现一个函数处理log.Logger的内容，用于捕捉net/http.Server输出的error内容。
	netHTTPLog struct {
		print func(...interface{})
		log   *log.Logger
	}
)

// NewServerStd 创建一个标准server。
func NewServerStd(arg interface{}) Server {
	httpserver := &http.Server{
		ReadTimeout:  4 * time.Second,
		WriteTimeout: 4 * time.Second,
		IdleTimeout:  60 * time.Second,
		TLSNextProto: nil,
	}
	ConvertTo(arg, httpserver)
	return &ServerStd{
		Server: httpserver,
	}
}

// SetHandler 方法设置server的http处理者。
func (srv *ServerStd) SetHandler(h http.Handler) {
	srv.Server.Handler = h
}

// SetPrint 设置server输出函数。
func (srv *ServerStd) SetPrint(fn func(...interface{})) {
	srv.Print = fn
	if srv.Print != nil {
		srv.Server.ErrorLog = newNetHTTPLog(srv.Print).Logger()
	}
}

// Set 方法允许Server设置输出函数和配置
func (srv *ServerStd) Set(key string, value interface{}) error {
	switch val := value.(type) {
	case func(...interface{}):
		srv.Print = val
		srv.Server.ErrorLog = newNetHTTPLog(srv.Print).Logger()
	case ServerConfigStd, *ServerConfigStd:
		ConvertTo(value, srv.Server)
	default:
		return ErrSeterNotSupportField
	}
	return nil
}

// newNetHTTPLog 实现将一个日志处理函数适配成log.Logger对象。
func newNetHTTPLog(fn func(...interface{})) *netHTTPLog {
	e := &netHTTPLog{
		print: fn,
	}
	e.log = log.New(e, "", 0)
	return e
}

func (e *netHTTPLog) Write(p []byte) (n int, err error) {
	e.print(string(p))
	return 0, nil
}

func (e *netHTTPLog) Logger() *log.Logger {
	return e.log
}
