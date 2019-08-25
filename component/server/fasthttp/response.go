package fasthttp

import (
	"errors"
	"github.com/eudore/eudore/protocol"
	"github.com/valyala/fasthttp"
	"net"
)

type (
	// Response 实现protocol.ResponseWriter接口
	Response struct {
		ctx      *fasthttp.RequestCtx
		Response *fasthttp.Response
		header   *Header
	}
)

// Reset 方法重置Response对象。
func (resp *Response) Reset(ctx *fasthttp.RequestCtx) {
	resp.ctx = ctx
	resp.Response = &ctx.Response
	resp.header.header = &ctx.Response.Header
	ctx.Response.Header.SetContentType("")
}

// Write 方法实现io.Writer接口。
func (resp *Response) Write(b []byte) (int, error) {
	resp.Response.AppendBody(b)
	return len(b), nil
}

// Header 方法获得响应的Header。
func (resp *Response) Header() protocol.Header {
	return resp.header
}

// WriteHeader 方法实现写入http请求状态码。
func (resp *Response) WriteHeader(code int) {
	resp.ctx.SetStatusCode(code)
}

// Flush 方法实现刷新缓冲，将缓冲的请求发送给客户端。
func (resp *Response) Flush() {
	// Do nothing because nosuppert
}

// Hijack 方法实现劫持http连接,该方法未实现。
func (resp *Response) Hijack() (net.Conn, error) {
	return nil, nil
}

// Push 方法实现接口，fasthttp不支持该方法。
func (resp *Response) Push(string, *protocol.PushOptions) error {
	return errors.New("fasthttp not supperd push")
}

// Size 方法获得写入的数据长度。
func (resp *Response) Size() int {
	return resp.Response.Header.ContentLength()
}

// Status 方法获得设置的http状态码。
func (resp *Response) Status() int {
	return resp.Response.Header.StatusCode()
}
