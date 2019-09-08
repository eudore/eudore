package fasthttp

import (
	"bytes"
	"context"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/protocol"
	"github.com/valyala/fasthttp"
	"net"
	"sync"
)

type (
	// Server 定义适配fasthttp.Server
	Server struct {
		Config    interface{} `set:"config"`
		srvs      []*fasthttp.Server
		handler   protocol.HandlerHTTP
		listeners []net.Listener `set:"listeners"`
		wg        sync.WaitGroup
	}
)

var (
	_       eudore.Server = (*Server)(nil)
	poolreq               = sync.Pool{
		New: func() interface{} {
			return &Request{
				read:   bytes.NewReader(nil),
				header: &Header{},
			}
		},
	}
	poolresp = sync.Pool{
		New: func() interface{} {
			return &Response{
				header: &Header{},
			}
		},
	}
)

// NewServer 创建一个Server，参数为fasthttp的配置。
func NewServer(arg interface{}) *Server {
	return &Server{
		Config: arg,
	}
}

// Start 启动fasthttp。
func (srv *Server) Start() error {
	errs := eudore.NewErrors()
	for _, ln := range srv.listeners {
		srv.wg.Add(1)
		go func(ln net.Listener) {
			http := &fasthttp.Server{}
			eudore.ConvertTo(srv.Config, http)
			http.Handler = srv.HandlerFasthttp
			srv.srvs = append(srv.srvs, http)
			errs.HandleError(http.Serve(ln))
			srv.wg.Done()
		}(ln)
	}

	srv.wg.Wait()
	return errs.GetError()
}

// Close 方法关闭Server。
func (srv *Server) Close() error {
	return srv.Shutdown(nil)
}

// Shutdown 方法关闭Server。
func (srv *Server) Shutdown(context.Context) error {
	errs := eudore.NewErrors()
	for _, srv := range srv.srvs {
		errs.HandleError(srv.Shutdown())
	}
	return errs
}

// AddHandler 方法设置http处理者。
func (srv *Server) AddHandler(h protocol.HandlerHTTP) {
	srv.handler = h
}

// AddListener 添加一个监听。
func (srv *Server) AddListener(l net.Listener) {
	srv.listeners = append(srv.listeners, l)
}

// HandlerFasthttp 实现fasthttp.Handler。
func (srv *Server) HandlerFasthttp(ctx *fasthttp.RequestCtx) {
	req := poolreq.Get().(*Request)
	resp := poolresp.Get().(*Response)
	req.Reset(ctx)
	resp.Reset(ctx)
	srv.handler.EudoreHTTP(ctx, resp, req)
	poolreq.Put(req)
	poolresp.Put(resp)
}
