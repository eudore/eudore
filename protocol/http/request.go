package http

import (
	"io"
	"net"
	"sync"
	"bufio"
	"bytes"
	"strconv"
	"crypto/tls"
	"github.com/eudore/eudore/protocol"
)

type Request struct {
	conn		net.Conn
	reader		*bufio.Reader
	method		string
	requestURI	string
	proto		string
	header		Header
	//
	mu			sync.Mutex
	length		int
	sawEOF		bool
	expect		bool
}


func(r *Request) Reset(conn net.Conn) error {
	r.conn = conn
	r.reader.Reset(conn)
	r.header.Reset()
	// Read the http request line.
	// 读取http请求行。
	line, err := r.readLine()
	if err != nil {
		return err
	}
	// 读取请求行，sawEOF作为临时变量
	r.method, r.requestURI, r.proto, r.sawEOF = parseRequestLine(line)
	if !r.sawEOF {
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
		// Split into headers and store them in the request.
		// 分割成header存储到请求中。
		r.header.Add(splitHeader(line))
	}
	// Read the request body length from the header, if no body direct length is 0
	// 从header中读取请求body长度，如果无body直接长度为0
	r.length, _ = strconv.Atoi(r.header.Get("Content-Length"))
	// When the body length is zero, the body is read directly to return EOF.
	// body长度为零时，读取body直接返回EOF。
	r.sawEOF = r.length == 0
	r.expect = r.header.Get("Expect") == "100-continue"
	return err
}

func(r *Request) Read(p []byte) (int, error) {
	r.mu.Lock()
	// First judge whether it has been read
	// 先判断是否已经读取完毕
	if r.sawEOF {
		r.mu.Unlock()
		return 0, io.EOF
	}
	// If Expect is 100-continue, return 100 first and continue reading data.
	// 如果Expect为100-continue，先返回100然后继续读取数据。
	if r.expect {
		r.expect = false
		r.conn.Write(constinueMsg)
	}
	// read data from the connection
	// 从连接读取数据
	n, err := r.reader.Read(p)
	if err == io.EOF {
		// read return EOF
		// 读取返回EOF
		r.sawEOF = true
	}else if err == nil && n > 0 {
		// Reduce the length of unread data
		// 减少未读数据长度
		r.length -= n
		// set EOF
		// 设置EOF
		r.sawEOF = r.length == 0
		if r.sawEOF {
			err = io.EOF
		}
	}
	r.mu.Unlock()
	return n, err
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
	return &r.header
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

// 从请求读取一行数据
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

// 将header的键值切分
func splitHeader(line []byte) (string, string) {
	i := bytes.Index(line, colonSpace)
	if i != -1 {
		return string(line[:i]), string(line[i + 2:])
	}
	return "", ""
}