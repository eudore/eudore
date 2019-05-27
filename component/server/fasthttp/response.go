package fasthttp

import (
	"net"
	"fmt"
	"errors"
	"github.com/eudore/eudore/protocol"
	"github.com/valyala/fasthttp"
)

type (
	Response struct {
		ctx			*fasthttp.RequestCtx
		Response	*fasthttp.Response
		header		*Header
	}
)


func (resp *Response) Reset(ctx *fasthttp.RequestCtx) {
	resp.ctx = ctx
	resp.Response = &ctx.Response
	resp.header.header = &ctx.Response.Header
	ctx.Response.Header.SetContentType("")
	resp.header.Range(func(key, val string){
		fmt.Println(key,val)
		})
}

func (resp *Response) Write(b []byte) (int, error) {
	resp.Response.AppendBody(b)
	return len(b), nil
}

func (resp *Response) Header() protocol.Header {
	return resp.header
}

func (resp *Response) WriteHeader(code int) {
	resp.ctx.SetStatusCode(code)
}

func (resp *Response) Flush() {
	// Do nothing because nosuppert	
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

