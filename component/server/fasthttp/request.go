package fasthttp

import (
	"bytes"
	"crypto/tls"
	"github.com/eudore/eudore/protocol"
	"github.com/valyala/fasthttp"
)

type (
	// Request 实现protocol.RequestReader接口，适配fasthttp请求。
	Request struct {
		ctx     *fasthttp.RequestCtx
		Request *fasthttp.Request
		header  *Header
		body    []byte
		read    *bytes.Reader
	}
)

// Reset 方法重置对象。
func (req *Request) Reset(ctx *fasthttp.RequestCtx) {
	req.ctx = ctx
	req.Request = &ctx.Request
	req.header.header = &ctx.Request.Header
	req.body = req.Body()
	req.read.Reset(req.body)
}

// Method 方法获得http请求方法。
func (req *Request) Method() string {
	return string(req.ctx.Method())
}

// Proto 方法获得http协议版本。
func (req *Request) Proto() string {
	return "HTTP/1.1"
}

// RequestURI 方法获得http请求的uri。
func (req *Request) RequestURI() string {
	return string(req.Request.RequestURI())
}

// Path 方法返回http请求的方法。
func (req *Request) Path() string {
	return string(req.ctx.Path())
}

// RawQuery 方法返回http请求的uri参数。
func (req *Request) RawQuery() string {
	return string(req.Request.URI().QueryArgs().QueryString())
}

// Header 方法获得http请求的header。
func (req *Request) Header() protocol.Header {
	return req.header
}

// Read 方法实现io.Reader接口。
func (req *Request) Read(b []byte) (int, error) {
	return req.read.Read(b)
}

// Host 方法获取请求的Host。
func (req *Request) Host() string {
	return string(req.Request.Host())
}

// RemoteAddr 方法获得http连接的远程连接地址。
func (req *Request) RemoteAddr() string {
	return req.ctx.RemoteAddr().String()
}

// TLS 方法获得tls状态信息。
func (req *Request) TLS() *tls.ConnectionState {
	return req.ctx.TLSConnectionState()
}

// Body 方法返回全部body内容。
func (req *Request) Body() []byte {
	return req.body
}
