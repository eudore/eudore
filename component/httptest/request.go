package httptest

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/url"
	"strings"

	"github.com/eudore/eudore/protocol"
)

type (
	// RequestReaderTest 实现protocol.RequestReader接口，用于执行测试请求。
	RequestReaderTest struct {
		//
		client *Client
		File   string
		Line   int
		// data
		method string
		path   string
		args   url.Values
		proto  string
		header HeaderMap
		body   io.Reader
	}
)

// NewRequestReaderTest 函数创建一个测试http请求。
func NewRequestReaderTest(client *Client, method, path string) *RequestReaderTest {
	r := &RequestReaderTest{
		client: client,
		method: method,
		header: make(HeaderMap),
	}
	pos := strings.IndexByte(path, '?')
	if pos == -1 {
		r.path = path
		r.args = make(url.Values)
	} else {
		r.path = path[:pos]
		r.args, _ = url.ParseQuery(path[pos+1:])
	}
	return r
}

func (r *RequestReaderTest) WithAddParam(key, val string) *RequestReaderTest {
	r.args.Add(key, val)
	return r
}
func (r *RequestReaderTest) WithHeader(headers protocol.Header) *RequestReaderTest {
	headers.Range(func(key, val string) {
		r.header.Add(key, val)
	})
	return r
}

func (r *RequestReaderTest) WithHeaderValue(key, val string) *RequestReaderTest {
	r.header.Add(key, val)
	return r
}

func (r *RequestReaderTest) WithBody(reader io.Reader) *RequestReaderTest {
	r.body = reader
	return r
}

func (r *RequestReaderTest) WithBodyString(s string) *RequestReaderTest {
	r.body = strings.NewReader(s)
	return r
}

func (r *RequestReaderTest) WithBodyByte(b []byte) *RequestReaderTest {
	r.body = bytes.NewReader(b)
	return r
}
func (r *RequestReaderTest) WithBodyJson(data interface{}) *RequestReaderTest {
	r.header.Add("Content-Type", "application/json")
	reader, writer := io.Pipe()
	r.body = reader
	json.NewEncoder(writer).Encode(data)
	return r
}

func (r *RequestReaderTest) WithBodyFrom() *RequestReaderTest {
	return r
}

func (r *RequestReaderTest) Do() *ResponseWriterTest {
	// 附加客户端公共参数
	for key, vals := range r.client.Args {
		for _, val := range vals {
			r.args.Add(key, val)
		}
	}
	r.client.Headers.Range(func(key, val string) {
		r.header.Add(key, val)
	})
	if r.body == nil {
		r.body = bytes.NewReader(nil)
	}
	// 创建响应并处理
	resp := NewResponseWriterTest(r.client, r)
	r.client.EudoreHTTP(context.Background(), resp, r)
	return resp
}

// Method 方法返回请求方法。
func (r *RequestReaderTest) Method() string {
	return r.method
}

// Proto 方法返回http协议版本，使用"HTTP/1.0"。
func (r *RequestReaderTest) Proto() string {
	return "HTTP/1.0"
}

// RequestURI 方法返回http请求的完整uri。
func (r *RequestReaderTest) RequestURI() string {
	raw := r.RawQuery()
	if raw == "" {
		return r.Path()
	}
	return r.path + "?" + raw
}

// Path 方法返回http请求的方法。
func (r *RequestReaderTest) Path() string {
	return r.path
}

// RawQuery 方法返回http请求的uri参数。
func (r *RequestReaderTest) RawQuery() string {
	return r.args.Encode()
}

// Header 方法返回http请求的header
func (r *RequestReaderTest) Header() protocol.Header {
	return r.header
}

// Read 方法实现io.Reader接口，用于读取body内容。
func (r *RequestReaderTest) Read(p []byte) (int, error) {
	return r.body.Read(p)
}

// Host 方法返回请求Host。
func (r *RequestReaderTest) Host() string {
	return "eudore-httptest"
}

// RemoteAddr 方法返回远程连接地址，使用默认值"192.0.2.1:1234"。
func (r *RequestReaderTest) RemoteAddr() string {
	return "192.0.2.1:1234"
}

// TLS 方法返回tls状态。
func (r *RequestReaderTest) TLS() *tls.ConnectionState {
	return nil
}
