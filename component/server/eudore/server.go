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
	ServerConfig struct {
		ReadTimeout  interface{} `set:"readtimeout" description:"Http server read timeout."`
		WriteTimeout interface{} `set:"writetimeout" description:"Http server write timeout."`
	}
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

func (srv *Server) Close() error {
	return srv.Shutdown(context.Background())
}

func (srv *Server) Shutdown(ctx context.Context) (err error) {
	return srv.http.Shutdown(ctx)
}

func (srv *Server) AddHandler(h protocol.HandlerHttp) {
	srv.handler = h
}

func (srv *Server) AddListener(l net.Listener) {
	srv.listeners = append(srv.listeners, l)
}

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
