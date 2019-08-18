// Package protocol 定义通用http接口。
package protocol

import (
	"context"
	"crypto/tls"
	"net"
)

type (
	// HandlerConn 接口定义eudore处理net.Conn
	HandlerConn interface {
		EudoreConn(context.Context, net.Conn)
	}
	// HandlerHttp 接口定义eudore处理http请求
	HandlerHttp interface {
		EudoreHTTP(context.Context, ResponseWriter, RequestReader)
	}
	// HandlerFunc 定义http处理函数
	HandlerHttpFunc func(context.Context, ResponseWriter, RequestReader)
	// Header 定义http header
	Header interface {
		Get(string) string
		Set(string, string)
		Add(string, string)
		Del(string)
		Range(func(string, string))
	}
	// RequestReader 接口根据http协议请求体，从RequestReader获取方法、版本、uri、header、body。(golang net/http库header中没有host)
	//
	// 从net.Conn连接读取远程连接地址和TLS信息。
	RequestReader interface {
		// http protocol data
		Method() string
		Proto() string
		RequestURI() string
		Path() string
		RawQuery() string
		Header() Header
		Read([]byte) (int, error)
		Host() string
		// conn data
		RemoteAddr() string
		TLS() *tls.ConnectionState
	}

	// ResponseWriter 接口用于写入http请求响应体status、header、body。
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
		Hijack() (net.Conn, error)
		// http.Pusher
		Push(string, *PushOptions) error
		Size() int
		Status() int
	}

	RequestWriter interface {
		Header() Header
		Do() (ResponseReader, error)
	}
	// ResponseReader is used to read the http protocol response message information.
	//
	// ResponseReader用于读取http协议响应报文信息。
	ResponseReader interface {
		Proto() string
		Statue() int
		Code() string
		Header() Header
		Read([]byte) (int, error)
		TLS() *tls.ConnectionState
		Close() error
	}
	// PushOptions 定义http2 push的选项
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

func (fn HandlerHttpFunc) EudoreHTTP(ctx context.Context, w ResponseWriter, r RequestReader) {
	fn(ctx, w, r)
}

var (
	HeaderTransferEncoding = "Transfer-Encoding"
)
