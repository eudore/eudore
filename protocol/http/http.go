package http

import (
	"bufio"
	"context"
	"errors"
	"github.com/eudore/eudore/protocol"
	"io"
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

// Handler 定义解析处理http连接。
type Handler struct {
	Handler      protocol.HandlerHTTP
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	Print        func(...interface{}) `set:"print"`
}

// NewHandler 函数创建一个http/1.1的http处理这
func NewHandler(h protocol.HandlerHTTP) *Handler {
	return &Handler{
		Handler:      h,
		ReadTimeout:  60 * time.Minute,
		WriteTimeout: 60 * time.Minute,
		IdleTimeout:  60 * time.Minute,
		Print:        log.Println,
	}
}

// EudoreConn 实现protocol.HandlerConn接口，处理http连接。
func (h *Handler) EudoreConn(pctx context.Context, c net.Conn) {
	// Initialize the request object.
	// 初始化请求对象。
	resp := rwPool.Get().(*Response)
	for {
		c.SetReadDeadline(time.Now().Add(h.ReadTimeout))
		err := resp.request.Reset(c)
		if err != nil {
			// handler error
			if isNotCommonNetReadError(err) {
				h.Print("eudore http request read error: ", err)
			}
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

// isNotCommonNetReadError 函数检查net读取错误是否未非通用错误。
func isNotCommonNetReadError(err error) bool {
	if err == io.EOF {
		return false
	}
	if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
		return false
	}
	if oe, ok := err.(*net.OpError); ok && oe.Op == "read" {
		return false
	}
	return true
}

// SetIdleTimeout 方法设置http连接处理的IdleTimeout时间。
func (h *Handler) SetIdleTimeout(t time.Duration) {
	h.IdleTimeout = t
}

// SetReadTimeout 方法设置http连接处理的ReadTimeout时间。
func (h *Handler) SetReadTimeout(t time.Duration) {
	h.ReadTimeout = t

}

// SetWriteTimeout 方法设置http连接处理的WriteTimeout时间。
func (h *Handler) SetWriteTimeout(t time.Duration) {
	h.WriteTimeout = t
}

// SetPrint 方法设置输出函数
func (h *Handler) SetPrint(fn func(...interface{})) {
	h.Print = fn
}
