package http

import (
	"net"
	"fmt"
	"sync"
	"bufio"
	"errors"
	"context"
	"github.com/eudore/eudore/protocol"
)


var (
	crlf		= []byte("\r\n")
	colonSpace	= []byte(": ")
	constinueMsg	=	[]byte("HTTP/1.1 100 Continue\r\n\r\n")
	rwPool	=	sync.Pool {
		New:	func() interface{} {
			return &Response{
				request:	&Request{
								reader:	bufio.NewReaderSize(nil, 2048),
							},
				writer:		bufio.NewWriterSize(nil, 2048),
				buf:		make([]byte, 2048),
			}
		},
	}
	ErrLineInvalid	=	errors.New("request line is invalid")
)

type HttpHandler struct {
	Handler		protocol.Handler
	Errfunc		func(error)			`set:"errfunc`
}

func NewHttpHandler(h protocol.Handler) *HttpHandler {
	return &HttpHandler{h, printErr}
}

func printErr(err error) {
	fmt.Println("eudore http error:", err)
}

// Handling http connections
//
// 处理http连接
func (hh *HttpHandler) EudoreConn(pctx context.Context, c net.Conn) {
	// Initialize the request object.
	// 初始化请求对象。
	resp := rwPool.Get().(*Response)
	for {
		if err := resp.request.Reset(c); err != nil { // && err != io.EOF
			// handler error
			hh.Errfunc(err)
			break
		}
		resp.Reset(c)
		ctx, cancelCtx := context.WithCancel(pctx)
		resp.cancel = cancelCtx
		// 处理请求
		hh.Handler.EudoreHTTP(ctx, resp, resp.request)
		if resp.ishjack {
			return
		}
		resp.finalFlush()
		if resp.request.isnotkeep {
			break
		}
	}
	c.Close()
	rwPool.Put(resp)
}
