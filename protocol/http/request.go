package http


import (
	"io"
	"net"
	"crypto/tls"
	"github.com/eudore/eudore/protocol"
)

type Request struct {
		method		string
		requestURI	string
		proto		string
		header		protocol.Header
		reader		io.Reader
		conn		net.Conn
	}


func(r *Request) Method() string {
	return r.method
}

func (r *Request) Proto() string {
	return r.proto
}

func (r *Request) RequestURI() string {
	return r.requestURI
}

func (r *Request) Header() protocol.Header {
	return r.header
}

func (r *Request) Read(b []byte) (int, error) {
	return r.reader.Read(b)
}

func (r *Request) Host() string {
	return r.header.Get("Host")
}

// conn data
func (r *Request) RemoteAddr() string {
	return r.conn.RemoteAddr().String()
}

func (r *Request) TLS() *tls.ConnectionState {
	return nil
}


