package http

import (
	"net"
	"bufio"
	"bytes"
	"crypto/tls"
	"github.com/eudore/eudore/protocol"
	"github.com/eudore/eudore/protocol/header"
)

type Request struct {
	conn		net.Conn
	reader		*bufio.Reader
	method		string
	requestURI	string
	proto		string
	header		protocol.Header
	ok			bool
}


func(r *Request) Reset(conn net.Conn) error {
	r.conn = conn
	r.reader.Reset(conn)
	r.header = make(header.HeaderMap)
	// Read the http request line.
	// 读取http请求行。
	line, err := r.readLine()
	if err != nil {
		return err
	}
	r.method, r.requestURI, r.proto, r.ok = parseRequestLine(line)
	if !r.ok {
		return ErrLineInvalid
	}
	// read http headers
	// 读取http headers
	for {
		// Read a line of content.
		// 读取一行内容。
		line, err = r.readLine()
		if err != nil || len(line) == 0 {
			break
		}
		// fmt.Println("read header:", line)
		// Split into headers and store them in the request.
		// 分割成header存储到请求中。
		r.header.Add(splitHeader(line))
	}
	return nil
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



func (r *Request) readLine() ([]byte, error) {
	// r.closeDot()
	var line []byte
	for {
		l, more, err := r.reader.ReadLine()
		if err != nil {
			return nil, err
		}
		// Avoid the copy if the first call produced a full line.
		if line == nil && !more {
			return l, nil
		}
		line = append(line, l...)
		if !more {
			break
		}
	}
	return line, nil
}

// parseRequestLine parses "GET /foo HTTP/1.1" into its three parts.
func parseRequestLine(line []byte) (method, requestURI, proto string, ok bool) {
	s1 := bytes.IndexByte(line, ' ')
	s2 := bytes.IndexByte(line[s1+1:], ' ')
	if s1 < 0 || s2 < 0 {
		return
	}
	s2 += s1 + 1
	return string(line[:s1]), string(line[s1+1 : s2]), string(line[s2+1:]), true
}


func splitHeader(line []byte) (string, string) {
	i := bytes.Index(line, colonSpace)
	if i != -1 {
		return string(line[:i]), string(line[i + 2:])
	}
	return "", ""
}