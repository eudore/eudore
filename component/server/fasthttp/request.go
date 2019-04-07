package fasthttp

import (
	"bytes"
	"crypto/tls"
	"github.com/eudore/eudore/protocol"
	"github.com/valyala/fasthttp"
)

type (
	Request struct {
		ctx			*fasthttp.RequestCtx
		Request		*fasthttp.Request
		header		*Header
		body		[]byte
		read		*bytes.Reader
	}
)

func (req *Request) Reset(ctx *fasthttp.RequestCtx) {
	req.ctx = ctx
	req.Request = &ctx.Request
	req.header.header = &ctx.Request.Header
	req.body = req.Body()
	req.read.Reset(req.body)
}

func (req *Request) Method() string {
	return string(req.ctx.Method())
}

func (req *Request) Proto() string {
	return "HTTP/1.1"
}

func (req *Request) RequestURI() string {
	return string(req.Request.RequestURI())
}

func (req *Request) Header() protocol.Header {
	return req.header
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