package httptest

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"sync"
)

type (
	// ResponseWriterTest is an implementation of protocol.ResponseWriter that
	// records its mutations for later inspection in tests.
	ResponseWriterTest struct {
		Client  *Client
		Request *RequestReaderTest

		sync.WaitGroup
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
		Client:  client,
		Request: req,
		// HeaderMap: make(http.Header),
		Body: new(bytes.Buffer),
		Code: 200,
	}
}

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
func (rw *ResponseWriterTest) writeHeader(b []byte) {
	if rw.wroteHeader {
		return
	}

	m := rw.Header()
	hasType := m.Get("Content-Type") != ""
	hasTE := m.Get("Transfer-Encoding") != ""
	if !hasType && !hasTE {
		if b != nil {
			m.Set("Content-Type", http.DetectContentType(b))
		}
	}
}

// Write always succeeds and writes to rw.Body, if not nil.
func (rw *ResponseWriterTest) Write(buf []byte) (int, error) {
	rw.writeHeader(buf)
	if rw.Body != nil {
		rw.Body.Write(buf)
	}
	return len(buf), nil
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
	if rw.Request.websocketHandle != nil {
		serverConn, clientConn := net.Pipe()
		rw.Add(1)
		go func() {
			resp, err := http.ReadResponse(bufio.NewReader(clientConn), rw.Request.Request)
			if err != nil {
				clientConn.Close()
				rw.Client.Print(err)
				rw.Done()
				return
			}
			rw.HandleRespone(resp)
			rw.Done()
			rw.Request.websocketHandle(clientConn)
		}()
		return serverConn, bufio.NewReadWriter(bufio.NewReader(serverConn), bufio.NewWriter(serverConn)), nil
	}
	return nil, nil, ErrResponseWriterTestNotSupportHijack
}

// HandleRespone 方法处理一个http.Response对象数据。
func (rw *ResponseWriterTest) HandleRespone(resp *http.Response) *ResponseWriterTest {
	rw.Code = resp.StatusCode
	rw.HeaderMap = resp.Header
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		rw.Client.Print(err)
		return rw
	}
	resp.Body.Close()
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
	rw.Client.Printf("CheckStatus response status is invalid %d,check status: %v", rw.Code, status)
	return rw
}

// CheckHeader 方法检查多个header的值
func (rw *ResponseWriterTest) CheckHeader(h ...string) *ResponseWriterTest {
	for i := 0; i < len(h)/2; i++ {
		if rw.HeaderMap.Get(h[i]) != h[i+1] {
			rw.Client.Printf("CheckHeader response header %s value is %s,not is %s", h[i], rw.HeaderMap.Get(h[i]), h[i+1])
		}
	}
	return rw
}

func (rw *ResponseWriterTest) getBodyString() string {
	if rw.HeaderMap.Get("Content-Encoding") == "gzip" {
		r, err := gzip.NewReader(rw.Body)
		if err == nil {
			body, _ := ioutil.ReadAll(r)
			return string(body)
		}
	}
	return rw.Body.String()
}

// CheckBodyContainString 方法检查响应的字符串body是否包含指定多个字符串。
func (rw *ResponseWriterTest) CheckBodyContainString(strs ...string) *ResponseWriterTest {
	body := rw.getBodyString()
	for _, str := range strs {
		if !strings.Contains(body, str) {
			rw.Client.Printf("CheckBodyContainString response body not contains string: %s", str)
		}
	}
	return rw
}

// CheckBodyString 方法检查body是否为指定字符串。
func (rw *ResponseWriterTest) CheckBodyString(s string) *ResponseWriterTest {
	if s != rw.getBodyString() {
		rw.Client.Printf("CheckBodyString response body size %d not is check string", rw.Body.Len())
	}
	return rw
}

// CheckBodyJSON 方法检查body是否是指定对象的json， 未实现。
func (rw *ResponseWriterTest) CheckBodyJSON(data interface{}) *ResponseWriterTest {
	jsonbody, err := json.Marshal(data)
	if err != nil {
		rw.Client.Printf("CheckBodyJSON json marshal err: %s", err.Error())
		return rw
	}
	jsonbody = append(jsonbody, '\n')
	if bytes.Equal(jsonbody, rw.Body.Bytes()) {
		return rw
	}
	rw.Client.Printf("CheckBodyJSON not equal")
	return rw
}

// Out 方法输出完整响应。
func (rw *ResponseWriterTest) Out() *ResponseWriterTest {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf("httptest request %s %s status: %d", rw.Request.Method, rw.Request.RequestURI, rw.Code))
	for k, v := range rw.HeaderMap {
		b.WriteString(fmt.Sprintf("\n%s: %s", k, v))
	}
	b.WriteString("\n\n" + rw.getBodyString())
	rw.Client.Print(b.String())
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
	rw.Client.Print(b.String())
	return rw
}

// OutBody 方法输出body字符串信息。
func (rw *ResponseWriterTest) OutBody() *ResponseWriterTest {
	rw.Client.Printf("httptest request %s %s body: %s", rw.Request.Method, rw.Request.RequestURI, rw.getBodyString())
	return rw
}
