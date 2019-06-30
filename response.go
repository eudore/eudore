package eudore

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/eudore/eudore/protocol"
	"io"
	"net"
	"net/http"
)

type (
	// Encapsulate the net/http.Response response message and convert it to the ResponseReader interface.
	//
	// 封装net/http.Response响应报文，转换成ResponseReader接口
	ResponseReaderHttp struct {
		io.ReadCloser
		Data   *http.Response
		header protocol.Header
	}
	ResponseReaderHttpWithWiter struct {
		ResponseReaderHttp
		io.Writer
	}
	// net/http.ResponseWriter接口封装
	ResponseWriterHttp struct {
		http.ResponseWriter
		header protocol.Header
		code   int
		size   int
	}
	// ResponseWriterTest is an implementation of http.ResponseWriter that
	// records its mutations for later inspection in tests.
	ResponseWriterTest struct {
		// Code is the HTTP response code set by WriteHeader.
		//
		// Note that if a Handler never calls WriteHeader or Write,
		// this might end up being 0, rather than the implicit
		// http.StatusOK. To get the implicit value, use the Result
		// method.
		Code int

		// HeaderMap contains the headers explicitly set by the Handler.
		// It is an internal detail.
		//
		// Deprecated: HeaderMap exists for historical compatibility
		// and should not be used. To access the headers returned by a handler,
		// use the Response.Header map as returned by the Result method.
		HeaderMap HeaderMap

		// Body is the buffer to which the Handler's Write calls are sent.
		// If nil, the Writes are silently discarded.
		Body *bytes.Buffer

		// Flushed is whether the Handler called Flush.
		Flushed bool

		//		result      *http.Response // cache of Result's return value
		snapHeader  HeaderMap // snapshot of HeaderMap at first Write
		wroteHeader bool
	}
)

var _ protocol.ResponseWriter = &ResponseWriterHttp{}

func NewResponseWriterHttp(w http.ResponseWriter) protocol.ResponseWriter {
	return &ResponseWriterHttp{
		ResponseWriter: w,
		header:         HeaderMap(w.Header()),
	}
}

func ResetResponseWriterHttp(hw *ResponseWriterHttp, w http.ResponseWriter) protocol.ResponseWriter {
	hw.ResponseWriter = w
	hw.header = HeaderMap(w.Header())
	hw.code = http.StatusOK
	hw.size = 0
	return hw
}

func (w *ResponseWriterHttp) Header() protocol.Header {
	return w.header
}

func (w *ResponseWriterHttp) Write(data []byte) (int, error) {
	n, err := w.ResponseWriter.Write(data)
	w.size = w.size + n
	return n, err
}

func (w *ResponseWriterHttp) WriteHeader(codeCode int) {
	w.code = codeCode
	w.ResponseWriter.WriteHeader(w.code)
}

func (w *ResponseWriterHttp) Flush() {
	w.ResponseWriter.(http.Flusher).Flush()
}

func (w *ResponseWriterHttp) Hijack() (conn net.Conn, err error) {
	if hj, ok := w.ResponseWriter.(http.Hijacker); ok {
		conn, _, err = hj.Hijack()
		return
	}
	err = fmt.Errorf("http.Hijacker interface is not supported")
	return
}

// 如果ResponseWriterHttp实现http.Push接口，则Push资源。
func (w *ResponseWriterHttp) Push(target string, opts *protocol.PushOptions) error {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		// TODO: add con
		return pusher.Push(target, &http.PushOptions{})
	}
	return nil
}

func (w *ResponseWriterHttp) Size() int {
	return w.size
}

func (w *ResponseWriterHttp) Status() int {
	return w.code
}

func NewResponseReaderHttp(resp *http.Response) protocol.ResponseReader {
	var r = ResponseReaderHttp{
		ReadCloser: resp.Body,
		Data:       resp,
		header:     HeaderMap(resp.Header),
	}

	w, ok := resp.Body.(io.Writer)
	if ok {
		return &ResponseReaderHttpWithWiter{
			ResponseReaderHttp: r,
			Writer:             w,
		}
	}
	return &r
}

func (r *ResponseReaderHttpWithWiter) Write(p []byte) (n int, err error) {
	return r.Writer.Write(p)
}

func (r *ResponseReaderHttp) Proto() string {
	return r.Data.Proto
}

func (r *ResponseReaderHttp) Statue() int {
	return r.Data.StatusCode
}

func (r *ResponseReaderHttp) Code() string {
	return r.Data.Status
}

func (r *ResponseReaderHttp) Header() protocol.Header {
	return r.header
}

func (r *ResponseReaderHttp) TLS() *tls.ConnectionState {
	return r.Data.TLS
}

func NewResponseWriterTest() *ResponseWriterTest {
	return &ResponseWriterTest{
		HeaderMap: make(HeaderMap),
		Body:      new(bytes.Buffer),
		Code:      200,
	}
}

// DefaultRemoteAddr is the default remote address to return in RemoteAddr if
// an explicit DefaultRemoteAddr isn't set on ResponseWriterTest.
const DefaultRemoteAddr = "1.2.3.4"

// Header returns the response headers.
func (rw *ResponseWriterTest) Header() protocol.Header {
	m := rw.HeaderMap
	if m == nil {
		m = make(HeaderMap)
		rw.HeaderMap = m
	}
	return m
}

// writeHeader writes a header if it was not written yet and
// detects Content-Type if needed.
//
// bytes or str are the beginning of the response body.
// We pass both to avoid unnecessarily generate garbage
// in rw.WriteString which was created for performance reasons.
// Non-nil bytes win.
func (rw *ResponseWriterTest) writeHeader(b []byte, str string) {
	if rw.wroteHeader {
		return
	}
	if len(str) > 512 {
		str = str[:512]
	}

	m := rw.Header()

	hasType := m.Get("Content-Type") != ""
	hasTE := m.Get("Transfer-Encoding") != ""
	if !hasType && !hasTE {
		if b == nil {
			b = []byte(str)
		}
		m.Set("Content-Type", http.DetectContentType(b))
	}

	rw.WriteHeader(200)
}

// Write always succeeds and writes to rw.Body, if not nil.
func (rw *ResponseWriterTest) Write(buf []byte) (int, error) {
	rw.writeHeader(buf, "")
	if rw.Body != nil {
		rw.Body.Write(buf)
	}
	return len(buf), nil
}

// WriteString always succeeds and writes to rw.Body, if not nil.
func (rw *ResponseWriterTest) WriteString(str string) (int, error) {
	rw.writeHeader(nil, str)
	if rw.Body != nil {
		rw.Body.WriteString(str)
	}
	return len(str), nil
}

// WriteHeader sets rw.Code. After it is called, changing rw.Header
// will not affect rw.HeaderMap.
func (rw *ResponseWriterTest) WriteHeader(code int) {
	if rw.wroteHeader {
		return
	}
	rw.Code = code
	rw.wroteHeader = true
	if rw.HeaderMap == nil {
		rw.HeaderMap = make(HeaderMap)
	}
	rw.snapHeader = cloneHeaderMap(rw.HeaderMap)
}

func cloneHeaderMap(h HeaderMap) HeaderMap {
	h2 := make(HeaderMap, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}

// Flush sets rw.Flushed to true.
func (rw *ResponseWriterTest) Flush() {
	if !rw.wroteHeader {
		rw.WriteHeader(200)
	}
	rw.Flushed = true
}

func (rw *ResponseWriterTest) Hijack() (net.Conn, error) {
	return nil, nil
}
func (rw *ResponseWriterTest) Push(string, *protocol.PushOptions) error {
	return nil
}
func (rw *ResponseWriterTest) Size() int {
	return 0
}
func (rw *ResponseWriterTest) Status() int {
	return rw.Code
}
