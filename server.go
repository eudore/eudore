package eudore

import (
	"context"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/eudore/eudore/protocol"
)

// 定义ServerState的值。
const (
	ServerStateInit ServerState = iota
	ServerStateRun
	ServerStateClose
	ServerStateUnknown
	// EnvEudoreGracefulAddrs 按顺序记录fork多端口fd对应的地址。
	EnvEudoreGracefulAddrs = "EnvEudoreGracefulAddrs"
)

type (
	// ServerState 定义Server的状态。
	ServerState int
	// Server 定义启动http服务的对象。
	Server interface {
		AddHandler(protocol.HandlerHTTP)
		AddListener(net.Listener)
		Start() error
		Close() error
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
		handler   protocol.HandlerHTTP
		listeners []net.Listener       `set:"listeners"`
		Print     func(...interface{}) `set:"print"`
		mu        sync.Mutex           `set:"-"`
		wg        sync.WaitGroup       `set:"-"`
		state     ServerState          `set:"-"`
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
		state:  ServerStateInit,
	}
}

// Start 方法使用全部注册的net.Listener启动服务监听。
func (srv *ServerStd) Start() error {
	if len(srv.listeners) == 0 {
		return ErrServerNotAddListener
	}
	// update server state
	srv.mu.Lock()
	if srv.state != ServerStateInit {
		return ErrServerStdStateException
	}
	srv.state = ServerStateRun
	srv.mu.Unlock()

	// setting server
	srv.Server.Handler = srv
	if h, ok := srv.handler.(http.Handler); ok {
		srv.Server.Handler = h
	}
	if srv.Print != nil {
		srv.Server.ErrorLog = newNetHTTPLog(srv.Print).Logger()
	}

	// start server
	errs := NewErrors()
	for i := range srv.listeners {
		srv.wg.Add(1)
		go func(ln net.Listener) {
			err := srv.Server.Serve(ln)
			if err != http.ErrServerClosed && err != nil {
				errs.HandleError(err)
			}
			srv.wg.Done()
		}(srv.listeners[i])
	}

	// wait over
	srv.wg.Wait()
	return errs.GetError()
}

// Close 方法关闭server。
func (srv *ServerStd) Close() (err error) {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.state == ServerStateRun {
		srv.state = ServerStateClose
		return srv.Server.Close()
	}
	return nil
}

// Shutdown 方法关闭server
func (srv *ServerStd) Shutdown(ctx context.Context) error {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	srv.state = ServerStateClose
	return srv.Server.Shutdown(ctx)
}

// AddHandler 方法设置server的http处理者。
//
// 如果处理者同时实现了http.Handler接口，会使用处理者的http.Handler的接口。
func (srv *ServerStd) AddHandler(h protocol.HandlerHTTP) {
	srv.handler = h
}

// AddListener 方法给server新增一个监听者。
func (srv *ServerStd) AddListener(l net.Listener) {
	srv.listeners = append(srv.listeners, l)
}

// ServeHTTP 实现http.Handler接口，处理net/http Server的请求。
func (srv *ServerStd) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	request := requestReaderHTTPPool.Get().(*RequestReaderHTTP)
	response := responseWriterHTTPPool.Get().(*ResponseWriterHTTP)

	request.Reset(r)
	response.Reset(w)
	srv.handler.EudoreHTTP(r.Context(), response, request)

	requestReaderHTTPPool.Put(request)
	responseWriterHTTPPool.Put(response)
}

// Set 方法允许Server设置输出函数和配置
func (srv *ServerStd) Set(key string, value interface{}) error {
	switch val := value.(type) {
	case func(...interface{}):
		srv.Print = val
	case ServerConfigStd, *ServerConfigStd:
		ConvertTo(value, srv.Server)
	default:
		return ErrSeterNotSupportField
	}
	return nil
}

// SetPrint 设置server输出函数。
func (srv *ServerStd) SetPrint(fn func(...interface{})) {
	srv.Print = fn
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
