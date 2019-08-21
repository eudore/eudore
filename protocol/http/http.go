package http

import (
	"bufio"
	"context"
	"errors"
	"github.com/eudore/eudore/protocol"
	"log"
	"net"
	"sync"
	"time"
)

var (
	crlf         = []byte("\r\n")
	colonSpace   = []byte(": ")
	constinueMsg = []byte("HTTP/1.1 100 Continue\r\n\r\n")
	rwPool       = sync.Pool{
		New: func() interface{} {
			return &Response{
				request: &Request{
					reader: bufio.NewReaderSize(nil, 2048),
				},
				writer: bufio.NewWriterSize(nil, 2048),
				buf:    make([]byte, 2048),
			}
		},
	}
	// ErrLineInvalid 定义http请求行无效的错误。
	ErrLineInvalid = errors.New("request line is invalid")
)

// HttpHandler 定义解析处理http连接。
type HttpHandler struct {
	Handler      protocol.HandlerHttp
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Print        func(...interface{}) `set:"print"`
}

// NewHttpHandler 函数创建一个http/1.1的http处理这
func NewHttpHandler(h protocol.HandlerHttp) *HttpHandler {
	return &HttpHandler{
		Handler:      h,
		ReadTimeout:  60 * time.Minute,
		WriteTimeout: 60 * time.Minute,
		IdleTimeout:  60 * time.Minute,
		Print:        log.Println,
	}
}

// EudoreConn 实现protocol.HandlerConn接口，处理http连接。
func (h *HttpHandler) EudoreConn(pctx context.Context, c net.Conn) {
	// Initialize the request object.
	// 初始化请求对象。
	resp := rwPool.Get().(*Response)
	for {
		c.SetReadDeadline(time.Now().Add(h.ReadTimeout))
		if err := resp.request.Reset(c); err != nil { // && err != io.EOF
			// handler error
			h.Print(err)
			break
		}
		resp.Reset(c)
		ctx, cancelCtx := context.WithCancel(pctx)
		resp.cancel = cancelCtx
		// 处理请求
		c.SetWriteDeadline(time.Now().Add(h.WriteTimeout))
		h.Handler.EudoreHTTP(ctx, resp, resp.request)
		if resp.ishjack {
			return
		}
		resp.finalFlush()
		if resp.request.isnotkeep {
			break
		}
		// c.SetDeadline(time.Now().Add(h.IdleTimeout))
	}
	c.Close()
	rwPool.Put(resp)
}

// SetIdleTimeout 设置http连接处理的IdleTimeout时间。
func (h *HttpHandler) SetIdleTimeout(t time.Duration) {
	h.IdleTimeout = t
}

// SetReadDeadline 设置http连接处理的ReadTimeout时间。
func (h *HttpHandler) SetReadTimeout(t time.Duration) {
	h.ReadTimeout = t

}

// SetWriteDeadline 设置http连接处理的WriteTimeout时间。
func (h *HttpHandler) SetWriteTimeout(t time.Duration) {
	h.WriteTimeout = t
}

func (h *HttpHandler) SetPrint(fn func(...interface{})) {
	h.Print = fn
}
