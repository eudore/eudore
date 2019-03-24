package http

import (
	"net"
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
	requestPool		sync.Pool
	responsePool	sync.Pool
	ErrLineInvalid	=	errors.New("request line is invalid")
)

func init() {
	requestPool.New = func() interface{} {
		return &Request{
			reader:	bufio.NewReaderSize(nil, 2048),
		}
	}
	responsePool.New = func() interface{} {
		return &Response{
			writer:	bufio.NewWriterSize(nil, 2048),
			buf:	make([]byte, 2048),
		}
	}
}

type HttpHandler struct {
}

func NewHttpHandler() *HttpHandler {
	return &HttpHandler{}
}

// Handling http connections
//
// 处理http连接
func (hh *HttpHandler) EudoreConn(ctx context.Context, c net.Conn, h protocol.Handler) {
	// fmt.Println("conn serve:", c.RemoteAddr().String())
	// Initialize the request object.
	// 初始化请求对象。
	req := requestPool.Get().(*Request)
	resp := responsePool.Get().(*Response)
	resp.request = req
	for {
		if err := req.Reset(c); err != nil {
			// handler error
			// fmt.Println(err)
			return
		}
		resp.Reset(c)
		// 处理请求
		h.EudoreHTTP(ctx, resp, req)
		resp.flushend()
		if req.header.Get("Connection") != "keep-alive" {
			c.Close()
			break
		}
	}
	requestPool.Put(req)
	responsePool.Put(resp)
}