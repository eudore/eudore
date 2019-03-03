package eudore


import (
	"io"
	"bytes"
	"io/ioutil"
	"net/http"
	"crypto/tls"
	"github.com/eudore/eudore/protocol"
)

type (
/*	protocol.RequestReader interface {
		// http protocol data
		Method() string
		Proto() string
		RequestURI() string
		Header() Header
		Read([]byte) (int, error)
		Host() string
		// conn data
		RemoteAddr() string
		TLS() *tls.ConnectionState
	}*/
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
	RequestReaderEudore struct {
		method string
		proto string
		requestURI string
		remoteAddr string
		header http.Header
		body []byte
		tls *tls.ConnectionState
	}
)


var _ protocol.RequestReader		=	&RequestReaderHttp{}

func NewRequestReaderHttp(r *http.Request) protocol.RequestReader {
	return &RequestReaderHttp{
		Request:	*r,
		header:	httpHeader(r.Header),
	}
}

func ResetRequestReaderHttp(r *RequestReaderHttp, req *http.Request) protocol.RequestReader {
	r.Request = *req
	r.header = httpHeader(req.Header)
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
