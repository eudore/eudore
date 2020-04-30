package httptest

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

type (
	// ResponseWriterTest is an implementation of protocol.ResponseWriter that
	// records its mutations for later inspection in tests.
	ResponseWriterTest struct {
		Client  *Client
		Request *RequestReaderTest

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
		HeaderMap http.Header

		// Body is the buffer to which the Handler's Write calls are sent.
		// If nil, the Writes are silently discarded.
		Body *bytes.Buffer

		// Flushed is whether the Handler called Flush.
		Flushed bool

		//		result      *http.Response // cache of Result's return value
		snapHeader  http.Header // snapshot of HeaderMap at first Write
		wroteHeader bool
	}
)

// NewResponseWriterTest 方法返回一个测试使用的响应写入对象*ResponseWriterTest。
func NewResponseWriterTest(client *Client, req *RequestReaderTest) *ResponseWriterTest {
	return &ResponseWriterTest{
		Client:    client,
		Request:   req,
		HeaderMap: make(http.Header),
		Body:      new(bytes.Buffer),
		Code:      200,
	}
}

// DefaultRemoteAddr is the default remote address to return in RemoteAddr if
// an explicit DefaultRemoteAddr isn't set on ResponseWriterTest.
const DefaultRemoteAddr = "1.2.3.4"

// Header returns the response headers.
func (rw *ResponseWriterTest) Header() http.Header {
	m := rw.HeaderMap
	if m == nil {
		m = make(http.Header)
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
		rw.HeaderMap = make(http.Header)
	}
	rw.snapHeader = cloneHeaderMap(rw.HeaderMap)
}

func cloneHeaderMap(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
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

// Hijack 方法返回劫持的连接。
func (rw *ResponseWriterTest) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if rw.Request.websocketServer != nil {
		go func() {
			resp, err := http.ReadResponse(bufio.NewReader(rw.Request.websocketClient), rw.Request.Request)
			if err != nil {
				rw.Request.websocketClient.Close()
				rw.Request.Error(err)
				return
			}
			rw.HandleRespone(resp)
			rw.Request.websocketHandle(rw.Request.websocketClient)
		}()
		return rw.Request.websocketServer, bufio.NewReadWriter(bufio.NewReader(rw.Request.websocketServer), bufio.NewWriter(rw.Request.websocketServer)), nil
	}
	return nil, nil, ErrResponseWriterTestNotSupportHijack
}

// Size 方法返回写入的body的长度。
func (rw *ResponseWriterTest) Size() int {
	return rw.Body.Len()
}

// Status 方法返回响应状态码。
func (rw *ResponseWriterTest) Status() int {
	return rw.Code
}

// HandleRespone 方法处理一个http.Response对象数据。
func (rw *ResponseWriterTest) HandleRespone(resp *http.Response) *ResponseWriterTest {
	rw.Code = resp.StatusCode
	rw.HeaderMap = resp.Header
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		rw.Request.Error(err)
		return rw
	}
	rw.Body = bytes.NewBuffer(body)
	return rw
}

// CheckStatus 方法检查状态码。
func (rw *ResponseWriterTest) CheckStatus(status ...int) *ResponseWriterTest {
	for _, i := range status {
		if i == rw.Code {
			return rw
		}
	}
	rw.Request.Errorf("CheckStatus response status is invalid %d,check status: %v", rw.Code, status)
	return rw
}

// CheckHeader 方法检查多个header的值
func (rw *ResponseWriterTest) CheckHeader(h ...string) *ResponseWriterTest {
	for i := 0; i < len(h)/2; i++ {
		if rw.HeaderMap.Get(h[i]) != h[i+1] {
			rw.Request.Errorf("CheckHeader response header %s value is %s,not is %s", h[i], rw.HeaderMap.Get(h[i]), h[i+1])
		}
	}
	return rw
}

// CheckBodyContainString 方法检查响应的字符串body是否包含指定多个字符串。
func (rw *ResponseWriterTest) CheckBodyContainString(strs ...string) *ResponseWriterTest {
	body := rw.Body.String()
	for _, str := range strs {
		if !strings.Contains(body, str) {
			rw.Request.Errorf("CheckBodyContainString response body not contains string: %s", str)
		}
	}
	return rw
}

// CheckBodyString 方法检查body是否为指定字符串。
func (rw *ResponseWriterTest) CheckBodyString(s string) *ResponseWriterTest {
	if s != rw.Body.String() {
		rw.Request.Errorf("CheckBodyString response body size %d not is check string", rw.Body.Len())
	}
	return rw
}

// CheckBodyJSON 方法检查body是否是指定对象的json， 未实现。
func (rw *ResponseWriterTest) CheckBodyJSON(data interface{}) *ResponseWriterTest {
	return rw
}

// Out 方法输出完整响应。
func (rw *ResponseWriterTest) Out() *ResponseWriterTest {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("httptest request %s %s status: %d", rw.Request.Method, rw.Request.RequestURI, rw.Code))
	for k, v := range rw.HeaderMap {
		b.WriteString(fmt.Sprintf("\n%s: %s", k, v))
	}
	b.WriteString("\n\n" + rw.Body.String())
	rw.Client.Println(b.String())
	return rw
}

// OutStatus 方法输出状态码。
func (rw *ResponseWriterTest) OutStatus() *ResponseWriterTest {
	rw.Client.Printf("httptest request %s %s status: %d\n", rw.Request.Method, rw.Request.RequestURI, rw.Code)
	return rw
}

// OutHeader 方法输出全部header。
func (rw *ResponseWriterTest) OutHeader() *ResponseWriterTest {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("httptest request %s %s status: %d\n", rw.Request.Method, rw.Request.RequestURI, rw.Code))
	for k, v := range rw.HeaderMap {
		b.WriteString(fmt.Sprintf("\n%s: %s", k, v))
	}
	rw.Client.Println(b.String())
	return rw
}

// OutBody 方法输出body字符串信息。
func (rw *ResponseWriterTest) OutBody() *ResponseWriterTest {
	rw.Client.Printf("httptest request %s %s body: %s\n", rw.Request.Method, rw.Request.RequestURI, rw.Body.String())
	return rw
}
