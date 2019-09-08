package eudore

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/textproto"
	"sync"

	"github.com/eudore/eudore/protocol"
)

type (
	// RequestReaderHTTP Convert net.http.Request to protocol.RequestReader.
	//
	// RequestReaderHTTP 将net/http.Request转换成RequestReader。
	RequestReaderHTTP struct {
		*http.Request
		header protocol.Header
	}
	// ResponseWriterHTTP 是对net/http.ResponseWriter接口封装
	ResponseWriterHTTP struct {
		http.ResponseWriter
		header protocol.Header
		code   int
		size   int
	}
	// HeaderMap 使用map实现protocol.Header接口
	HeaderMap map[string][]string
)

var (
	_                     protocol.RequestReader  = (*RequestReaderHTTP)(nil)
	_                     protocol.ResponseWriter = (*ResponseWriterHTTP)(nil)
	requestReaderHTTPPool                         = sync.Pool{
		New: func() interface{} {
			return &RequestReaderHTTP{}
		},
	}
	responseWriterHTTPPool = sync.Pool{
		New: func() interface{} {
			return &ResponseWriterHTTP{}
		},
	}
)

// GetNetHTTPHandler 函数实现将protocol.HandlerHTTP转换成http.Handler对象。
func GetNetHTTPHandler(ctx context.Context, h protocol.HandlerHTTP) http.Handler {
	if ctx == nil {
		ctx = context.Background()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		request := requestReaderHTTPPool.Get().(*RequestReaderHTTP)
		response := responseWriterHTTPPool.Get().(*ResponseWriterHTTP)
		request.Reset(r)
		response.Reset(w)
		h.EudoreHTTP(ctx, response, request)
		requestReaderHTTPPool.Put(request)
		responseWriterHTTPPool.Put(response)
	})
}

// Reset 方法重置RequestReaderHTTP。
func (r *RequestReaderHTTP) Reset(req *http.Request) {
	r.Request = req
	r.header = HeaderMap(req.Header)
}

// Read 方法实现io.Reader接口。
func (r *RequestReaderHTTP) Read(p []byte) (int, error) {
	return r.Request.Body.Read(p)
}

// Method 方法获得http请求方法。
func (r *RequestReaderHTTP) Method() string {
	return r.Request.Method
}

// Proto 方法获得http协议版本。
func (r *RequestReaderHTTP) Proto() string {
	return r.Request.Proto
}

// Host 方法获取请求的Host。
func (r *RequestReaderHTTP) Host() string {
	return r.Request.Host
}

// RequestURI 方法获得http请求的uri。
func (r *RequestReaderHTTP) RequestURI() string {
	return r.Request.RequestURI
}

// Path 方法返回http请求的方法。
func (r *RequestReaderHTTP) Path() string {
	return r.URL.Path
}

// RawQuery 方法返回http请求的uri参数。
func (r *RequestReaderHTTP) RawQuery() string {
	return r.URL.RawQuery
}

// Header 方法获得http请求的header。
func (r *RequestReaderHTTP) Header() protocol.Header {
	return r.header
}

// RemoteAddr 方法获得http连接的远程连接地址。
func (r *RequestReaderHTTP) RemoteAddr() string {
	return r.Request.RemoteAddr
}

// TLS 方法获得tls状态信息，
func (r *RequestReaderHTTP) TLS() *tls.ConnectionState {
	return r.Request.TLS
}

// GetNetHTTPRequest 方法返回*http.Request对象。
func (r *RequestReaderHTTP) GetNetHTTPRequest() *http.Request {
	return r.Request
}

// Reset 方法重置ResponseWriterHTTP对象。
func (w *ResponseWriterHTTP) Reset(writer http.ResponseWriter) {
	w.ResponseWriter = writer
	w.header = HeaderMap(writer.Header())
	w.code = http.StatusOK
	w.size = 0
}

// Header 方法获得响应的Header。
func (w *ResponseWriterHTTP) Header() protocol.Header {
	return w.header
}

// Write 方法实现io.Writer接口。
func (w *ResponseWriterHTTP) Write(data []byte) (int, error) {
	n, err := w.ResponseWriter.Write(data)
	w.size = w.size + n
	return n, err
}

// WriteHeader 方法实现写入http请求状态码。
func (w *ResponseWriterHTTP) WriteHeader(codeCode int) {
	w.code = codeCode
	w.ResponseWriter.WriteHeader(w.code)
}

// Flush 方法实现刷新缓冲，将缓冲的请求发送给客户端。
func (w *ResponseWriterHTTP) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

// Hijack 方法实现劫持http连接。
func (w *ResponseWriterHTTP) Hijack() (conn net.Conn, err error) {
	if hj, ok := w.ResponseWriter.(http.Hijacker); ok {
		conn, _, err = hj.Hijack()
		return
	}
	return nil, ErrResponseWriterHTTPNotHijacker
}

// Push 方法实现http Psuh，如果ResponseWriterHTTP实现http.Push接口，则Push资源。
func (w *ResponseWriterHTTP) Push(target string, opts *protocol.PushOptions) error {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, &http.PushOptions{})
	}
	return nil
}

// Size 方法获得写入的数据长度。
func (w *ResponseWriterHTTP) Size() int {
	return w.size
}

// Status 方法获得设置的http状态码。
func (w *ResponseWriterHTTP) Status() int {
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
