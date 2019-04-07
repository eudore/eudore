package fasthttp

import (
	"net"
	"fmt"
	"sync"
	"bytes"
	"context"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/protocol"
	"github.com/valyala/fasthttp"
)

type (
	Server struct {
		Config		*ServerConfig		`set:"config"`
		Fasthttp	*fasthttp.Server 	`set:"fasthttp"`
		handler		protocol.Handler
		pool		sync.Pool
		wg			sync.WaitGroup
	}
	ServerConfig struct {
		Handler		interface{}			`set:"handler" json:"-"`
		Http		[]*eudore.ServerListenConfig `set:"http"`

	}
)

var (
	poolreq			= sync.Pool{
		New: func() interface{} {
			return &Request{
				read:	bytes.NewReader(nil),
				header:	&Header{},
			}
		},
	}
	poolresp			= sync.Pool{
		New: func() interface{} {
			return &Response{
				header:	&Header{},
			}
		},
	}
)

func init() {
	eudore.RegisterComponent(eudore.ComponentServerFasthttpName, func(arg interface{}) (eudore.Component, error) {
		return NewServer(arg)
	})
}

func NewServer(arg interface{}) (*Server, error) {
	config, ok := arg.(*ServerConfig)
	if !ok {
		config = &ServerConfig{}
	}
	return &Server{
		Config:		config,
		Fasthttp:	&fasthttp.Server{},
	}, nil
}

func (srv *Server) Start() error {
	// 初始化数据
	if h, ok := srv.Config.Handler.(protocol.Handler); ok {
		srv.handler = h
	}
	srv.Fasthttp.Handler = srv.HandlerFasthttp
	// 启动fasthttp
	errs := eudore.NewErrors()
	for _, http := range srv.Config.Http {
		ln, err := http.Listen()		
		if err != nil {
			errs.HandleError(err)
			continue
		}
		srv.wg.Add(1)
		go func(ln net.Listener){
			errs.HandleError(srv.Fasthttp.Serve(ln))
			srv.wg.Done()
		}(ln)
	}

	srv.wg.Wait()
	return errs.GetError()
}


func (srv *Server) Restart() error {
	err := eudore.StartNewProcess()
	if err == nil {
		srv.Fasthttp.Shutdown()
	}
	return err
}

func (srv *Server) Close() error {
	return srv.Fasthttp.Shutdown()
}

func (srv *Server) Shutdown(context.Context) error {
	return srv.Fasthttp.Shutdown()
}

func (srv *Server) HandlerFasthttp(ctx *fasthttp.RequestCtx) {
	req := poolreq.Get().(*Request)
	resp := poolresp.Get().(*Response)
	req.Reset(ctx)
	resp.Reset(ctx)
	fmt.Println("---")
	srv.handler.EudoreHTTP(ctx, resp, req)
	poolreq.Put(req)
	poolresp.Put(resp)
}

func (srv *Server) Set(key string, val interface{}) (err error) {
	_, err = eudore.Set(srv, key, val)
	return
}





func (*ServerConfig)  GetName() string {
	return eudore.ComponentServerFasthttpName
}
func (*Server)  GetName() string {
	return eudore.ComponentServerFasthttpName
}
func (*Server)  Version() string {
	return eudore.ComponentServerFasthttpVersion
}