// The current library defines a generic http interface.
//
// 当前库定义通用http接口。
package protocol

import (
	"net"
	"bufio"
	"context"
	"crypto/tls"
	// "net/textproto"
)

type (
	HandlerConn interface {
		EudoreConn(context.Context, net.Conn, Handler)
	}
	Handler interface {
		EudoreHTTP(context.Context, ResponseWriter, RequestReader)
	}
	HandlerFunc func(context.Context, ResponseWriter, RequestReader)
	// Header = textproto.MIMEHeader
	Header interface {
		Get(string) string
		Set(string, string)
		Add(string, string)
		Range(func(string, string))
	}
	// Get the method, version, uri, header, body from the RequestReader according to the http protocol request body. (There is no host in the golang net/http library header)
	//
	// Read the remote connection address and TLS information from the net.Conn connection.
	//
	// 根据http协议请求体，从RequestReader获取方法、版本、uri、header、body。(golang net/http库header中没有host)
	//
	// 从net.Conn连接读取远程连接地址和TLS信息。
	RequestReader interface {
		// http protocol data
		Method() string
		Proto() string
		RequestURI() string
		Header() Header
		Read([]byte) (int, error)
		Host() string
		// conn data
		RemoteAddr() string
		TLS() *tls.ConnectionState
	}

	// ResponseWriter接口用于写入http请求响应体status、header、body。
	//
	// net/http.response实现了flusher、hijacker、pusher接口。
	ResponseWriter interface {
		// http.ResponseWriter
		Header() Header
		Write([]byte) (int, error)
		WriteHeader(int)
		// http.Flusher
		Flush()
		// http.Hijacker
		Hijack() (net.Conn, *bufio.ReadWriter, error)
		// http.Pusher
		Push(string, *PushOptions) error
		Size() int
		Status() int
	}
	PushOptions struct {
		// Method specifies the HTTP method for the promised request.
		// If set, it must be "GET" or "HEAD". Empty means "GET".
		Method string

		// Header specifies additional promised request headers. This cannot
		// include HTTP/2 pseudo header fields like ":path" and ":scheme",
		// which will be added automatically.
		Header Header
	}
)

func (fn HandlerFunc) EudoreHTTP(ctx context.Context, w ResponseWriter, r RequestReader) {
	fn(ctx, w, r)
}

