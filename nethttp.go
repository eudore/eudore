package eudore

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/textproto"
	"sync"

	"github.com/eudore/eudore/protocol"
)

type (
	// RequestReaderHttp Convert net.http.Request to protocol.RequestReader.
	//
	// RequestReaderHttp 将net/http.Request转换成RequestReader。
	RequestReaderHttp struct {
		*http.Request
		header protocol.Header
	}
	// ResponseWriterHttp 是对net/http.ResponseWriter接口封装
	ResponseWriterHttp struct {
		http.ResponseWriter
		header protocol.Header
		code   int
		size   int
	}
	// HeaderMap 使用map实现protocol.Header接口
	HeaderMap map[string][]string
)

var (
	_                     protocol.RequestReader  = (*RequestReaderHttp)(nil)
	_                     protocol.ResponseWriter = (*ResponseWriterHttp)(nil)
	requestReaderHttpPool                         = sync.Pool{
		New: func() interface{} {
			return &RequestReaderHttp{}
		},
	}
	responseWriterHttpPool = sync.Pool{
		New: func() interface{} {
			return &ResponseWriterHttp{}
		},
	}
)

// GetNetHttpHandler 函数实现将protocol.HandlerHttp转换成http.Handler对象。
func GetNetHttpHandler(ctx context.Context, h protocol.HandlerHttp) http.Handler {
	if ctx == nil {
		ctx = context.Background()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request := requestReaderHttpPool.Get().(*RequestReaderHttp)
		response := responseWriterHttpPool.Get().(*ResponseWriterHttp)
		request.Reset(r)
		response.Reset(w)
		h.EudoreHTTP(ctx, response, request)
		requestReaderHttpPool.Put(request)
		responseWriterHttpPool.Put(response)
	})
}

// Reset 方法重置RequestReaderHttp。
func (r *RequestReaderHttp) Reset(req *http.Request) {
	r.Request = req
	r.header = HeaderMap(req.Header)
}

// Read 方法实现io.Reader接口。
func (r *RequestReaderHttp) Read(p []byte) (int, error) {
	return r.Request.Body.Read(p)
}

// Method 方法获得http请求方法。
func (r *RequestReaderHttp) Method() string {
	return r.Request.Method
}

// Proto 方法获得http协议版本。
func (r *RequestReaderHttp) Proto() string {
	return r.Request.Proto
}

// Host 方法获取请求的Host。
func (r *RequestReaderHttp) Host() string {
	return r.Request.Host
}

// RequestURI 方法获得http请求的uri。
func (r *RequestReaderHttp) RequestURI() string {
	return r.Request.RequestURI
}

// Path 方法返回http请求的方法。
func (r *RequestReaderHttp) Path() string {
	return r.URL.Path
}

// RawQuery 方法返回http请求的uri参数。
func (r *RequestReaderHttp) RawQuery() string {
	return r.URL.RawQuery
}

// Header 方法获得http请求的header。
func (r *RequestReaderHttp) Header() protocol.Header {
	return r.header
}

// RemoteAddr 方法获得http连接的远程连接地址。
func (r *RequestReaderHttp) RemoteAddr() string {
	return r.Request.RemoteAddr
}

// TLS 方法获得tls状态信息，
func (r *RequestReaderHttp) TLS() *tls.ConnectionState {
	return r.Request.TLS
}

// GetNetHttpRequest 方法返回*http.Request对象。
func (r *RequestReaderHttp) GetNetHttpRequest() *http.Request {
	return r.Request
}

// Reset 方法重置ResponseWriterHttp对象。
func (w *ResponseWriterHttp) Reset(writer http.ResponseWriter) {
	w.ResponseWriter = writer
	w.header = HeaderMap(writer.Header())
	w.code = http.StatusOK
	w.size = 0
}

// Header 方法获得响应的Header。
func (w *ResponseWriterHttp) Header() protocol.Header {
	return w.header
}

// Write 方法实现io.Writer接口。
func (w *ResponseWriterHttp) Write(data []byte) (int, error) {
	n, err := w.ResponseWriter.Write(data)
	w.size = w.size + n
	return n, err
}

// WriteHeader 方法实现写入http请求状态码。
func (w *ResponseWriterHttp) WriteHeader(codeCode int) {
	w.code = codeCode
	w.ResponseWriter.WriteHeader(w.code)
}

// Flush 方法实现刷新缓冲，将缓冲的请求发送给客户端。
func (w *ResponseWriterHttp) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

// Hijack 方法实现劫持http连接。
func (w *ResponseWriterHttp) Hijack() (conn net.Conn, err error) {
	if hj, ok := w.ResponseWriter.(http.Hijacker); ok {
		conn, _, err = hj.Hijack()
		return
	}
	err = fmt.Errorf("http.Hijacker interface is not supported")
	return
}

// Push 方法实现http Psuh，如果ResponseWriterHttp实现http.Push接口，则Push资源。
func (w *ResponseWriterHttp) Push(target string, opts *protocol.PushOptions) error {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		// TODO: add con
		return pusher.Push(target, &http.PushOptions{})
	}
	return nil
}

// Size 方法获得写入的数据长度。
func (w *ResponseWriterHttp) Size() int {
	return w.size
}

// Status 方法获得设置的http状态码。
func (w *ResponseWriterHttp) Status() int {
	return w.code
}

// Get 方法获得一个Header值。
func (h HeaderMap) Get(key string) string {
	return textproto.MIMEHeader(h).Get(key)
}

// Set 方法设置一个Header值。
func (h HeaderMap) Set(key, value string) {
	textproto.MIMEHeader(h).Set(key, value)
}

// Add 方法添加一个Header值。
func (h HeaderMap) Add(key, value string) {
	textproto.MIMEHeader(h).Add(key, value)
}

// Del 方法删除一个Header值。
func (h HeaderMap) Del(key string) {
	textproto.MIMEHeader(h).Del(key)
}

// Range 方法遍历Header全部键值。
func (h HeaderMap) Range(fn func(string, string)) {
	for k, v := range h {
		for _, vv := range v {
			fn(k, vv)
		}
	}
}
