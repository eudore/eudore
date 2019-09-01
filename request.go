package eudore

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/eudore/eudore/protocol"
	"io"
	"net/url"
	"strings"
)

type (
	// RequestReaderTest 实现protocol.RequestReader接口，用于执行测试请求。
	RequestReaderTest struct {
		method string
		url    *url.URL
		proto  string
		header HeaderMap
		body   io.Reader
	}
)

// NewRequestReaderTest 函数创建一个测试http请求。
func NewRequestReaderTest(method, addr string, body interface{}) (protocol.RequestReader, error) {
	r := &RequestReaderTest{
		method: method,
		header: make(HeaderMap),
	}
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	r.url = u
	r.body, err = transbody(body)
	return r, err
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
	return r.url.EscapedPath()
}

// Path 方法返回http请求的方法。
func (r *RequestReaderTest) Path() string {
	return r.url.Path
}

// RawQuery 方法返回http请求的uri参数。
func (r *RequestReaderTest) RawQuery() string {
	return r.url.RawQuery
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
	return r.url.Host
}

// RemoteAddr 方法返回远程连接地址，使用默认值"192.0.2.1:1234"。
func (r *RequestReaderTest) RemoteAddr() string {
	return "192.0.2.1:1234"
}

// TLS 方法返回tls状态，如果请求url协议是http返回空，否在返回tls1.2版本完成握手的状态。
func (r *RequestReaderTest) TLS() *tls.ConnectionState {
	if r.url.Scheme == "http" {
		return nil
	}
	return &tls.ConnectionState{
		Version:           tls.VersionTLS12,
		HandshakeComplete: true,
		ServerName:        r.Host(),
	}
}

func transbody(body interface{}) (io.Reader, error) {
	if body == nil {
		return nil, nil
	}
	switch t := body.(type) {
	case string:
		return strings.NewReader(t), nil
	case []byte:
		return bytes.NewReader(t), nil
	case io.Reader:
		return t, nil
	default:
		return nil, fmt.Errorf(ErrFormatUnknownTypeBody, body)
	}
}
