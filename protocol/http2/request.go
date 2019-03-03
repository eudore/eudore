package http2

import (
	"crypto/tls"
)
type (
	requestReader struct {
		requestBody
		method		string
		uri			string
		remoteAddr	string
		header		Header
		proto		string
		host		string
		length		int64
		trailer		Header
		tls			*tls.ConnectionState
	}
)

func (r *requestReader) Method() string {
	return r.method
}

func (r *requestReader) Proto() string {
	return r.proto
}

func (r *requestReader) RequestURI() string {
	return r.uri
}

func (r *requestReader) Header() Header {
	return r.header
}

func (r *requestReader) Read(b []byte) (int, error) {
	return r.requestBody.Read(b)
}

func (r *requestReader) Host() string {
	return r.host
}

// conn data
func (r *requestReader) RemoteAddr() string {
	return r.remoteAddr
}

func (r *requestReader) TLS() *tls.ConnectionState {
	return r.tls
}

