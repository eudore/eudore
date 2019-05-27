package eudore


import (
	"io"
	"fmt"
	"bytes"
	"strings"
	"io/ioutil"
	"net/http"
	"net/url"
	"crypto/tls"
	"github.com/eudore/eudore/protocol"
)

type (
	RequestReadSeeker interface {
		protocol.RequestReader
		io.Seeker
	}
	// Convert protocol.RequestReader to the net.http.Request object interface.
	//
	// 将RequestReader转换成net.http.Request对象接口。
	RequestConvertNetHttp interface {
		GetNetHttpRequest() *http.Request
	}
	// Convert net.http.Request to protocol.RequestReader.
	//
	// 将net/http.Request转换成RequestReader。
	RequestReaderHttp struct {
		http.Request
		header	protocol.Header
	}
	// Modify the protocol.RequestReader method and request uri inside the internal redirect.
	//
	// 内部重定向内修改RequestReader的方法和请求uri。
	RequestReaderRedirect struct {
		protocol.RequestReader
		method string
		uri	string
	}
	RequestReaderSeeker struct {
		protocol.RequestReader
		reader *bytes.Reader
	}
	RequestWriterHttp struct {
		*http.Client
		*http.Request
		err error
	}
	RequestReaderTest struct {
		method		string
		url			*url.URL
		proto		string
		header		HeaderMap
		body		io.Reader
	}
)


var _ protocol.RequestReader		=	&RequestReaderHttp{}

func NewRequestReaderHttp(r *http.Request) protocol.RequestReader {
	return &RequestReaderHttp{
		Request:	*r,
		header:	HeaderMap(r.Header),
	}
}

func ResetRequestReaderHttp(r *RequestReaderHttp, req *http.Request) protocol.RequestReader {
	r.Request = *req
	r.header = HeaderMap(req.Header)
	return r
}

func (r *RequestReaderHttp) Read(p []byte) (int, error) {
	return r.Request.Body.Read(p)
}

func (r *RequestReaderHttp) Method() string {
	return r.Request.Method 
} 

func (r *RequestReaderHttp) Proto() string {
	return r.Request.Proto
}

func (r *RequestReaderHttp) Host() string {
	return r.Request.Host	
}

func (r *RequestReaderHttp) RequestURI() string {
	return r.Request.RequestURI
}

func (r *RequestReaderHttp) Header() protocol.Header {
	return r.header
} 

func (r *RequestReaderHttp) RemoteAddr() string {
	return r.Request.RemoteAddr
}

func (r *RequestReaderHttp) TLS() *tls.ConnectionState {
	return r.Request.TLS
}

func (r *RequestReaderHttp) GetNetHttpRequest() *http.Request {
	return &r.Request
}


func NewRequestReaderRedirect(r protocol.RequestReader, method, uri string) (protocol.RequestReader) {
	return &RequestReaderRedirect{
		RequestReader:	r,
		method:			method,
		uri:			uri,
	}
}

func (r *RequestReaderRedirect) Method() string {
	return r.method
}

func (r *RequestReaderRedirect) RemoteAddr() string {
	return r.uri
}

func NewRequestReaderSeeker(r protocol.RequestReader) (RequestReadSeeker) {
	rs, ok := r.(RequestReadSeeker)
	if ok {
		return rs
	}
	bts, _ := ioutil.ReadAll(r)
	return &RequestReaderSeeker{
		RequestReader:	r,
		reader:			bytes.NewReader(bts),
	}
}

func (r *RequestReaderSeeker) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}

func (r *RequestReaderSeeker) Seek(offset int64, whence int) (int64, error) {
	return r.reader.Seek(offset, whence)
}

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
	if body == nil {
		r.body = nil
		return r, nil
	}
	switch t := body.(type) {
	case string:
		data, err := ioutil.ReadFile(t)
		if err != nil {
			return nil, err
		}
		r.body = bytes.NewReader(data)
	case []byte:
		r.body = bytes.NewReader(t)
	case io.Reader:
		r.body = t
	default:
		return nil, fmt.Errorf("unknown type used for body: %+v", body)
	}
	return r, nil
}

func (r *RequestReaderTest) Method() string {
	return r.method
}

func (r *RequestReaderTest) Proto() string {
	return "HTTP/1.1"
}

func (r *RequestReaderTest) RequestURI() string {
	return r.url.EscapedPath()
}

func (r *RequestReaderTest) Header() protocol.Header {
	return r.header
}

func (r *RequestReaderTest) Read(p []byte) (int, error) {
	return r.body.Read(p)
}

func (r *RequestReaderTest) Host() string {
	return r.url.Host
}

// conn data
func (r *RequestReaderTest) RemoteAddr() string {
	return "192.0.2.1:1234"
}

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

func (r *RequestWriterHttp) Header() protocol.Header {
	return HeaderMap(r.Request.Header)	
}


func (r *RequestWriterHttp) Do() (protocol.ResponseReader, error) {
	resp, err := r.Client.Do(r.Request)
	if err != nil {
		return nil, err
	}
	return NewResponseReaderHttp(resp), nil
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
		return nil, fmt.Errorf("unknown type used for body: %+v", body)
	}
}

// func (r *RequestWriterHttp)

/*
// Body makes the request use obj as the body. Optional.
// If obj is a string, try to read a file of that name.
// If obj is a []byte, send it directly.
// If obj is an io.Reader, use it directly.
// If obj is a runtime.Object, marshal it correctly, and set Content-Type header.
// If obj is a runtime.Object and nil, do nothing.
// Otherwise, set an error.
func (r *Request) Body(obj interface{}) *Request {
	if r.err != nil {
		return r
	}
	switch t := obj.(type) {
	case string:
		data, err := ioutil.ReadFile(t)
		if err != nil {
			r.err = err
			return r
		}
		glogBody("Request Body", data)
		r.body = bytes.NewReader(data)
	case []byte:
		glogBody("Request Body", t)
		r.body = bytes.NewReader(t)
	case io.Reader:
		r.body = t
	case runtime.Object:
		// callers may pass typed interface pointers, therefore we must check nil with reflection
		if reflect.ValueOf(t).IsNil() {
			return r
		}
		data, err := runtime.Encode(r.serializers.Encoder, t)
		if err != nil {
			r.err = err
			return r
		}
		glogBody("Request Body", data)
		r.body = bytes.NewReader(data)
		r.SetHeader("Content-Type", r.content.ContentType)
	default:
		r.err = fmt.Errorf("unknown type used for body: %+v", obj)
	}
	return r
}*/
