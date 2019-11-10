package eudore

import (
	"bufio"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"

	"crypto/tls"
	"golang.org/x/net/http2"
)

var (
	crlf         = []byte("\r\n")
	colonSpace   = []byte(": ")
	constinueMsg = []byte("HTTP/1.1 100 Continue\r\n\r\n")
	rwPool       = sync.Pool{
		New: func() interface{} {
			return &Response{
				request: Request{
					Request: http.Request{
						ProtoMajor: 1,
						ProtoMinor: 1,
					},
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

// HTTPHandler 函数处理http/1.1请求
func HTTPHandler(pctx context.Context, conn net.Conn, handler http.Handler) {
	// Initialize the request object.
	// 初始化请求对象。
	resp := rwPool.Get().(*Response)
	for {
		// c.SetReadDeadline(time.Now().Add(h.ReadTimeout))
		err := resp.request.Reset(conn)
		if err != nil {
			// handler error
			if isNotCommonNetReadError(err) {
				// h.Print("eudore http request read error: ", err)
			}
			break
		}
		resp.Reset(conn)
		ctx, cancelCtx := context.WithCancel(pctx)
		resp.cancel = cancelCtx
		// 处理请求
		// c.SetWriteDeadline(time.Now().Add(h.WriteTimeout))
		handler.ServeHTTP(resp, resp.request.Request.WithContext(ctx))
		if resp.ishjack {
			return
		}
		resp.finalFlush()
		if resp.request.isnotkeep {
			break
		}
		// c.SetDeadline(time.Now().Add(h.IdleTimeout))
	}
	conn.Close()
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

// NewHTTP2Handler 方法创建一个h2处理函数。
func NewHTTP2Handler() func(context.Context, *tls.Conn, http.Handler) {
	h2svc := &http2.Server{}
	return func(ctx context.Context, conn *tls.Conn, h http.Handler) {
		h2svc.ServeConn(conn, &http2.ServeConnOpts{
			Context: ctx,
			Handler: h,
		})
	}
}
