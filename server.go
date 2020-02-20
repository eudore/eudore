package eudore

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"time"
)

type (
	// Server 定义启动http服务的对象。
	Server interface {
		SetHandler(http.Handler)
		Serve(net.Listener) error
		Shutdown(ctx context.Context) error
	}
	// ServerStdConfig 定义ServerStd使用的配置
	ServerStdConfig struct {
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
	}

	// ServerFcgi 定义fastcgi server
	ServerFcgi struct {
		http.Handler
		Listeners []net.Listener
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
		srv.Server.ErrorLog = newNetHTTPLogger(srv.Print)
	}
}

// Set 方法允许Server设置输出函数和配置
func (srv *ServerStd) Set(key string, value interface{}) error {
	switch val := value.(type) {
	case func(...interface{}):
		srv.Print = val
		srv.Server.ErrorLog = newNetHTTPLogger(srv.Print)
	case ServerStdConfig, *ServerStdConfig:
		ConvertTo(value, srv.Server)
	default:
		return ErrSeterNotSupportField
	}
	return nil
}

// newNetHTTPLog 实现将一个日志处理函数适配成log.Logger对象。
func newNetHTTPLogger(fn func(...interface{})) *log.Logger {
	e := &netHTTPLog{
		print: fn,
	}
	return log.New(e, "", 0)
}

func (e *netHTTPLog) Write(p []byte) (n int, err error) {
	e.print(string(p))
	return 0, nil
}

// NewServerFcgi 函数创建一个fcgi server。
func NewServerFcgi() Server {
	return &ServerFcgi{Handler: http.NotFoundHandler()}
}

// SetHandler 方法设置fcgi处理对象。
func (srv *ServerFcgi) SetHandler(h http.Handler) {
	srv.Handler = h
}

// Serve 方法启动一个新的fcgi监听。
func (srv *ServerFcgi) Serve(ln net.Listener) error {
	srv.Listeners = append(srv.Listeners, ln)
	return fcgi.Serve(ln, srv.Handler)
}

// Shutdown 方法关闭fcgi关闭监听。
func (srv *ServerFcgi) Shutdown(ctx context.Context) error {
	var errs Errors
	for _, ln := range srv.Listeners {
		err := ln.Close()
		errs.HandleError(err)
	}
	return errs.GetError()
}
