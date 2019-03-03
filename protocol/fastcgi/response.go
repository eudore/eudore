package fastcgi

import (
	"fmt"
	"time"
	"net"
	"net/http"
	"bufio"
)

// response implements http.ResponseWriter.
type response struct {
	req         *request
	header      Header
	w           *bufWriter
	wroteHeader bool
}

func newResponse(c *child, req *request) *response {
	return &response{
		req:    req,
		header: Header{},
		w:      newWriter(c.conn, typeStdout, req.reqId),
	}
}

func (r *response) Header() Header {
	return r.header
}

func (r *response) Write(data []byte) (int, error) {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	return r.w.Write(data)
}

func (r *response) WriteHeader(code int) {
	if r.wroteHeader {
		return
	}
	r.wroteHeader = true
	if code == http.StatusNotModified {
		// Must not have body.
		r.header.Del("Content-Type")
		r.header.Del("Content-Length")
		r.header.Del("Transfer-Encoding")
	} else if r.header.Get("Content-Type") == "" {
		r.header.Set("Content-Type", "text/html; charset=utf-8")
	}

	if r.header.Get("Date") == "" {
		r.header.Set("Date", time.Now().UTC().Format(http.TimeFormat))
	}

	fmt.Fprintf(r.w, "Status: %d %s\r\n", code, http.StatusText(code))
	
	for k, v := range r.header {
		fmt.Fprintf(r.w, "%s: %s\r\n", k, v[0])
	}

	r.w.WriteString("\r\n")
}

func (r *response) Flush() {
	if !r.wroteHeader {
		r.WriteHeader(http.StatusOK)
	}
	r.w.Flush()
}

func (r *response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, nil
}


func (r *response) Size() int {
	return 0
}

func (r *response) Status() int {
	return 0
}

func (r *response) Close() error {
	r.Flush()
	return r.w.Close()
}
