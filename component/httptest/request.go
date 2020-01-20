package httptest

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"runtime"
	"strings"
)

type (
	// RequestReaderTest 实现protocol.RequestReader接口，用于执行测试请求。
	RequestReaderTest struct {
		//
		client *Client
		File   string
		Line   int
		// data
		*http.Request
	}
)

// NewRequestReaderTest 函数创建一个测试http请求。
func NewRequestReaderTest(client *Client, method, path string) *RequestReaderTest {
	r := &RequestReaderTest{
		client: client,
		Request: &http.Request{
			Method:     method,
			RequestURI: path,
			Header:     make(http.Header),
			Proto:      "HTTP/1.0",
			Host:       "eudore-httptest",
			RemoteAddr: "192.0.2.1:1234",
		},
	}
	r.Request.URL, _ = url.ParseRequestURI(path)
	r.Form, _ = url.ParseQuery(r.Request.URL.RawQuery)
	return r
}

// WithAddQuery 方法给请求添加一个url参数。
func (r *RequestReaderTest) WithAddQuery(key, val string) *RequestReaderTest {
	r.Form.Add(key, val)
	return r
}

// WithHeaders 方法给请求添加多个header。
func (r *RequestReaderTest) WithHeaders(headers http.Header) *RequestReaderTest {
	for key, vals := range headers {
		for _, val := range vals {
			r.Request.Header.Add(key, val)
		}
	}
	return r
}

// WithHeaderValue 方法给请求添加一个header的值。
func (r *RequestReaderTest) WithHeaderValue(key, val string) *RequestReaderTest {
	r.Request.Header.Add(key, val)
	return r
}

// WithBody 方法设置请求的body。
func (r *RequestReaderTest) WithBody(reader io.Reader) *RequestReaderTest {
	r.Request.Body = ioutil.NopCloser(reader)
	return r
}

// WithBodyString 方法设置请求的字符串body。
func (r *RequestReaderTest) WithBodyString(s string) *RequestReaderTest {
	r.Body = ioutil.NopCloser(strings.NewReader(s))
	r.ContentLength = int64(len(s))
	return r
}

// WithBodyByte 方法设置请的字节body。
func (r *RequestReaderTest) WithBodyByte(b []byte) *RequestReaderTest {
	r.Body = ioutil.NopCloser(bytes.NewReader(b))
	r.ContentLength = int64(len(b))
	return r
}

// WithBodyJSON 方法设置body为一个对象的json字符串。
func (r *RequestReaderTest) WithBodyJSON(data interface{}) *RequestReaderTest {
	r.Request.Header.Add("Content-Type", "application/json")
	reader, writer := io.Pipe()
	r.Body = reader
	go json.NewEncoder(writer).Encode(data)
	return r
}

// WithBodyFrom 方法设置body是一个From对象，未完成。
func (r *RequestReaderTest) WithBodyFrom() *RequestReaderTest {
	return r
}

// Do 方法发送这个请求，使用客户端处理这个请求返回响应。
func (r *RequestReaderTest) Do() *ResponseWriterTest {
	// 附加客户端公共参数
	for key, vals := range r.client.Args {
		for _, val := range vals {
			r.Request.Form.Add(key, val)
		}
	}
	r.Request.URL.RawQuery = r.Form.Encode()
	r.Form = nil

	for key, vals := range r.client.Headers {
		for _, val := range vals {
			r.Request.Header.Add(key, val)
		}
	}

	if r.Request.Body == nil {
		r.Request.Body = ioutil.NopCloser(bytes.NewReader(nil))
		r.Request.ContentLength = -1
	}
	defer r.Body.Close()
	// 创建响应并处理
	resp := NewResponseWriterTest(r.client, r)
	r.client.Handler.ServeHTTP(resp, r.Request)
	return resp
}

// logFormatFileLine 函数获得调用的文件位置，默认层数加三。
//
// 文件位置会从第一个src后开始截取，处理gopath下文件位置。
func logFormatFileLine(depth int) (string, int) {
	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		file = "???"
		line = 1
	} else {
		// slash := strings.LastIndex(file, "/")
		slash := strings.Index(file, "src")
		if slash >= 0 {
			file = file[slash+4:]
		}
	}
	return file, line
}
