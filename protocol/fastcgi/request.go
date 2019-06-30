package fastcgi

import (
	"crypto/tls"
	"errors"
	"github.com/eudore/eudore/protocol"
	"io"
	"net"
	"strconv"
	"strings"
)

type requestReader struct {
	method     string
	uri        string
	proto      string
	remoteAddr string
	length     int64
	tls        *tls.ConnectionState
	header     Header
	trailer    Header
	body       io.ReadCloser
}

// RequestFromMap creates an http.Request from CGI variables.
// The returned Request's Body field is not populated.
func newRequestReader(params map[string]string) (*requestReader, error) {
	r := new(requestReader)
	r.method = params["REQUEST_METHOD"]
	if r.method == "" {
		return nil, errors.New("cgi: no REQUEST_METHOD in environment")
	}

	// r.Close = true
	r.trailer = make(Header)
	r.header = make(Header)

	if lenstr := params["CONTENT_LENGTH"]; lenstr != "" {
		clen, err := strconv.ParseInt(lenstr, 10, 64)
		if err != nil {
			return nil, errors.New("cgi: bad CONTENT_LENGTH in environment: " + lenstr)
		}
		r.length = clen
	}

	if ct := params["CONTENT_TYPE"]; ct != "" {
		r.header.Set("Content-Type", ct)
	}

	// Copy "HTTP_FOO_BAR" variables to "Foo-Bar" Headers
	for k, v := range params {
		if !strings.HasPrefix(k, "HTTP_") {
			continue
		}
		r.header.Add(strings.Replace(k[5:], "_", "-", -1), v)
	}

	// There's apparently a de-facto standard for this.
	// http://docstore.mik.ua/orelly/linux/cgi/ch03_02.htm#ch03-35636
	if s := params["HTTPS"]; s == "on" || s == "ON" || s == "1" {
		r.tls = &tls.ConnectionState{HandshakeComplete: true}
	}

	r.proto = params["SERVER_PROTOCOL"]
	r.uri = params["REQUEST_URI"]

	// Request.RemoteAddr has its port set by Go's standard http
	// server, so we do here too.
	remotePort, _ := strconv.Atoi(params["REMOTE_PORT"]) // zero if unset or invalid
	r.remoteAddr = net.JoinHostPort(params["REMOTE_ADDR"], strconv.Itoa(remotePort))

	return r, nil
}

func (r *requestReader) Method() string {
	return r.method
}

func (r *requestReader) Proto() string {
	return r.proto
}

func (r *requestReader) RequestURI() string {
	return r.uri
}

func (r *requestReader) Header() protocol.Header {
	return r.header
}

func (r *requestReader) Read(b []byte) (int, error) {
	return r.body.Read(b)
}

func (r *requestReader) Host() string {
	return r.header.Get("Host")
}

// conn data
func (r *requestReader) RemoteAddr() string {
	return r.remoteAddr
}

func (r *requestReader) TLS() *tls.ConnectionState {
	return r.tls
}
