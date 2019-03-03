package http2

import (
	"net"
	"context"
	"github.com/eudore/eudore/protocol"
)


type (
	HandlerConn = protocol.HandlerConn 
	Handler = protocol.Handler
	RequestReader = protocol.RequestReader
	ResponseWriter = protocol.ResponseWriter
	Header = protocol.Header
	PushOptions = protocol.PushOptions

	HandlerFunc func(context.Context, ResponseWriter, RequestReader)
)


func (fn HandlerFunc) EudoreHTTP(ctx context.Context, w ResponseWriter, r RequestReader) {
	fn(ctx, w, r)
}


type Http2Handler struct {
}

// Handling http connections
//
// 处理http连接
func (hh *Http2Handler) EudoreConn(ctx context.Context, c net.Conn, h Handler) {
}


func strSliceContains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}