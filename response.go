package eudore

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/eudore/eudore/protocol"
)

type (
	// ResponseWriterTest is an implementation of protocol.ResponseWriter that
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

// NewResponseWriterTest 方法返回一个测试使用的响应写入对象*ResponseWriterTest。
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

// Hijack 方法返回劫持的连接，该方法始终返回空连接和不支持该方法的错误。
func (rw *ResponseWriterTest) Hijack() (net.Conn, error) {
	return nil, fmt.Errorf("ResponseWriterTest no support hijack")
}

// Push 方法实现http2 push操作，改方法始终为空操作。
func (rw *ResponseWriterTest) Push(string, *protocol.PushOptions) error {
	return nil
}

// Size 方法返回写入的body的长度。
func (rw *ResponseWriterTest) Size() int {
	return 0
}

// Status 方法返回响应状态码。
func (rw *ResponseWriterTest) Status() int {
	return rw.Code
}

func (rw *ResponseWriterTest) CheckHeader() *ResponseWriterTest {
	return rw
}

func (rw *ResponseWriterTest) CheckStatus(status ...int) *ResponseWriterTest {
	for _, i := range status {
		if i == rw.Code {
			fmt.Printf("response status succeeds. status is %d", rw.Code)
			return rw
		}
	}
	fmt.Printf("response status is invalid %d,check status: %v", rw.Code, status)
	return rw
}

func (rw *ResponseWriterTest) Show() {
	fmt.Println("status:", rw.Code)
	for k, v := range rw.HeaderMap {
		fmt.Printf("%s: %s\n", k, strings.Join(v, ", "))
	}
	fmt.Println(rw.Body.String())
}

func TestAppRequest(handler protocol.HandlerHttp, method, path string, body interface{}) *ResponseWriterTest {
	req, _ := NewRequestReaderTest(method, path, body)
	resp := NewResponseWriterTest()
	handler.EudoreHTTP(context.Background(), resp, req)
	return resp
}
