package http

import (
	"io"
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
	requestPool		=	sync.Pool {
		New:	func() interface{} {
			return &Request{
				reader:	bufio.NewReaderSize(nil, 2048),
			}
		},
	}
	responsePool	=	sync.Pool {
		New:	func() interface{} {
			return &Response{
				writer:	bufio.NewWriterSize(nil, 2048),
				buf:	make([]byte, 2048),
			}
		},
	}
	ErrLineInvalid	=	errors.New("request line is invalid")
)

type HttpHandler struct {
	Handler		protocol.Handler
	ErrFunc		func(error)			`set:"errfunc`
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
func (hh *HttpHandler) EudoreConn(ctx context.Context, c net.Conn) {
	// Initialize the request object.
	// 初始化请求对象。
	req := requestPool.Get().(*Request)
	resp := responsePool.Get().(*Response)
	resp.request = req
	for {
		if err := req.Reset(c); err != nil && err != io.EOF {
			// handler error
			hh.ErrFunc(err)
			c.Close()
			return
		}
		resp.Reset(c)
		// 处理请求
		hh.Handler.EudoreHTTP(ctx, resp, req)
		resp.finalFlush()
		if resp.ishjack {
			break
		}
		if req.header.Get("Connection") != "keep-alive" || resp.ishjack {
			c.Close()
			break
		}
	}
	requestPool.Put(req)
	responsePool.Put(resp)
}
