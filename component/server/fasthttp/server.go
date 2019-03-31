package fasthttp

import (
	"net"
	"sync"
	"bytes"
	"errors"
	"crypto/tls"
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
	Request struct {
		ctx			*fasthttp.RequestCtx
		Request		*fasthttp.Request
		body		[]byte
		read		*bytes.Reader
	}
	Response struct {
		ctx			*fasthttp.RequestCtx
		Response	*fasthttp.Response
	}
)

var (
	poolreq			= sync.Pool{
		New: func() interface{} {
			return &Request{
				read:	bytes.NewReader(nil),
			}
		},
	}
	poolresp			= sync.Pool{
		New: func() interface{} {
			return &Response{}
		},
	}
)

func NewServer() (*Server, error) {
	return &Server{
		Config:		&ServerConfig{},
		Fasthttp:		&fasthttp.Server{},
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

func (srv *Server) HandlerFasthttp(ctx *fasthttp.RequestCtx) {
	req := poolreq.Get().(*Request)
	resp := poolresp.Get().(*Response)
	req.Reset(ctx)
	resp.Reset(ctx)
	srv.handler.EudoreHTTP(ctx, resp, req)
	poolreq.Put(req)
	poolresp.Put(resp)
}

func (srv *Server) Set(key string, val interface{}) (err error) {
	_, err = eudore.Set(srv, key, val)
	return
}



func (req *Request) Reset(ctx *fasthttp.RequestCtx) {
	req.ctx = ctx
	req.Request = &ctx.Request
	req.body = req.Body()
	req.read.Reset(req.body)
}

func (req *Request) Method() string {
	return string(req.ctx.Method())
}

func (req *Request) Proto() string {
	return "http"
}

func (req *Request) RequestURI() string {
	return string(req.Request.RequestURI())
}

func (req *Request) Header() protocol.Header {
	return nil
}

func (req *Request) Read(b []byte) (int, error) {
	return req.read.Read(b)
}

func (req *Request) Host() string {
	return string(req.Request.Host())
}

func (req *Request) RemoteAddr() string {
	return req.ctx.RemoteAddr().String()
}

func (req *Request) TLS() *tls.ConnectionState {
	return req.ctx.TLSConnectionState()
}

func (req *Request) Body() []byte {
	return req.body
}



func (req *Response) Reset(ctx *fasthttp.RequestCtx) {
	req.ctx = ctx
	req.Response = &ctx.Response
}

func (resp *Response) Write(b []byte) (int, error) {
	return resp.ctx.Write(b)
}

func (resp *Response) Header() protocol.Header {
	return nil
}
func (resp *Response) WriteHeader(code int) {
	resp.ctx.SetStatusCode(code)
}
func (resp *Response) Flush() {
	
}
func (resp *Response) Hijack() (net.Conn, error) {
	return nil, nil
}
func (resp *Response) Push(string, *protocol.PushOptions) error {
	return errors.New("not supperd push")
}

func (resp *Response) Size() int {
	return resp.Response.Header.ContentLength()
}

func (resp *Response) Status() int {
	return resp.Response.Header.StatusCode()
}
