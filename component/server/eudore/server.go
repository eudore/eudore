package eudore

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/protocol"
	"github.com/eudore/eudore/protocol/http"
	"github.com/eudore/eudore/protocol/http2"
	"github.com/eudore/eudore/protocol/server"
)

type (
	// ServerConfig 定义Server超时配置。
	ServerConfig struct {
		ReadTimeout  interface{} `set:"readtimeout" description:"Http server read timeout."`
		WriteTimeout interface{} `set:"writetimeout" description:"Http server write timeout."`
		IdleTimeout  interface{} `set:"idletimeout"`
	}
	// Server 定义eudore
	Server struct {
		http      *server.Server
		listeners []net.Listener
		handler   protocol.HandlerHttp
		mu        sync.Mutex
		wg        sync.WaitGroup
		Print     func(...interface{}) `set:"print" json:"-"`
	}
)

var _ eudore.Server = (*Server)(nil)

// NewServer 创建Server
func NewServer(arg interface{}) *Server {
	httpsrc := server.NewServer()
	if read, err := getTime(eudore.Get(arg, "ReadTimeout")); err == nil {
		httpsrc.ReadTimeout = read
	}
	if write, err := getTime(eudore.Get(arg, "WriteTimeout")); err == nil {
		httpsrc.WriteTimeout = write
	}
	if idle, err := getTime(eudore.Get(arg, "IdleTimeout")); err == nil {
		httpsrc.IdleTimeout = idle
	}
	return &Server{
		handler: protocol.HandlerHttpFunc(func(_ context.Context, w protocol.ResponseWriter, _ protocol.RequestReader) {
			w.Write([]byte("start eudore server, this default page."))
		}),
		http: httpsrc,
	}
}

// Start 启动Server
func (srv *Server) Start() error {
	if len(srv.listeners) == 0 {
		return errors.New("eudore server not found listen")
	}
	srv.mu.Lock()

	// 初始化服务连接处理者。
	srv.http.Print = srv.Print
	srv.http.SetHandler(http.NewHttpHandler(srv.handler))
	srv.http.SetNextHandler("h2", http2.NewServer(srv.handler))

	// 启动http
	errs := eudore.NewErrors()
	for i := range srv.listeners {
		srv.wg.Add(1)
		go func(ln net.Listener) {
			stopErr := fmt.Sprintf("accept %s %s: use of closed network connection", ln.Addr().Network(), ln.Addr().String())
			err := srv.http.Serve(ln)
			if stopErr != err.Error() {
				errs.HandleError(err)
			}
			srv.wg.Done()
		}(srv.listeners[i])
	}

	// 等待结束
	srv.mu.Unlock()
	srv.wg.Wait()
	if errs.GetError() != nil {
		return errs
	}
	return eudore.ErrApplicationStop
}

// Close 方法关闭Server。
func (srv *Server) Close() error {
	return srv.Shutdown(context.Background())
}

// Shutdown 方法关闭Server。
func (srv *Server) Shutdown(ctx context.Context) (err error) {
	return srv.http.Shutdown(ctx)
}

// AddHandler 方法设置http清楚处理。
func (srv *Server) AddHandler(h protocol.HandlerHttp) {
	srv.handler = h
}

// AddListener 方法增加一个监听。
func (srv *Server) AddListener(l net.Listener) {
	srv.listeners = append(srv.listeners, l)
}

// Set 方法允许设置输出函数和配置。
func (srv *Server) Set(key string, value interface{}) error {
	switch val := value.(type) {
	case func(...interface{}):
		srv.Print = val
	case ServerConfig, *ServerConfig:
		eudore.ConvertTo(value, srv.http)
	default:
		return eudore.ErrSeterNotSupportField
	}
	return nil
}

// getTime 获得时间。
func getTime(i interface{}) (time.Duration, error) {
	if t, ok := i.(string); ok {
		return time.ParseDuration(t)
	}
	if t, ok := i.(float64); ok {
		return time.Duration(t), nil
	}
	if t, ok := i.(int64); ok {
		return time.Duration(t), nil
	}
	if t, ok := i.(int); ok {
		return time.Duration(t), nil
	}
	return 0, fmt.Errorf("not parse time: %v", i)
}
