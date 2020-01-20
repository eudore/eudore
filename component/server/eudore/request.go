package eudore

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"sync"
)

type (
	// Request 定义一个http请求。
	Request struct {
		http.Request
		conn net.Conn
		// read body
		reader     *bufio.Reader
		nextLength int64
		expect     bool
		mu         sync.Mutex
	}
)

// Reset 方法重置请求对象
func (r *Request) Reset(conn net.Conn) error {
	r.conn = conn
	r.reader.Reset(conn)
	r.Header = make(http.Header)
	// Read the http request line.
	// 读取http请求行。
	line, err := r.readLine()
	if err != nil {
		return err
	}
	// 读取请求行，sawEOF作为临时变量
	r.Method, r.RequestURI, r.Proto, err = parseRequestLine(line)
	if err != nil {
		return err
	}
	// 初始化path和uri参数。
	r.URL, err = url.ParseRequestURI(r.RequestURI)
	if err != nil {
		return err
	}
	// 读取http headers
	for {
		// 读取一行内容。
		line, err = r.readLine()
		if err != nil || len(line) == 0 {
			break
		}
		// 分割成header存储到请求中。
		r.Header.Add(splitHeader(line))
	}
	r.Host = r.Header.Get("Host")
	r.RemoteAddr = conn.RemoteAddr().String()

	// 从header中读取请求body长度，如果无body直接长度为0,未处理分段传输。
	lenstr := r.Header.Get("Content-Length")
	if len(lenstr) > 0 {
		r.ContentLength, err = strconv.ParseInt(lenstr, 10, 64)
		if err != nil {
			return err
		}
		r.expect = r.Header.Get("Expect") == "100-continue"
		r.nextLength = r.ContentLength
		r.Body = r
	} else {
		r.ContentLength = 0
		r.Body = NoBody
	}
	return nil
}

// Read 方法读取数据，实现io.Reader。
func (r *Request) Read(p []byte) (int, error) {
	r.mu.Lock()
	if r.expect {
		r.expect = false
		r.conn.Write(constinueMsg)
	}
	if r.nextLength <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > r.nextLength {
		p = p[0:r.nextLength]
	}
	n, err := r.reader.Read(p)
	r.nextLength -= int64(n)
	r.mu.Unlock()
	return n, err
}

// Close 方法关闭read。
func (r *Request) Close() error {
	r.mu.Lock()
	io.CopyN(ioutil.Discard, r.reader, r.nextLength)
	r.mu.Unlock()
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
func parseRequestLine(line []byte) (method, requestURI, proto string, err error) {
	s1 := bytes.IndexByte(line, ' ')
	s2 := bytes.IndexByte(line[s1+1:], ' ')
	if s1 < 0 || s2 < 0 {
		return method, requestURI, proto, ErrLineInvalid
	}
	s2 += s1 + 1
	return string(line[:s1]), string(line[s1+1 : s2]), string(line[s2+1:]), nil
}

// 将header的键值切分
func splitHeader(line []byte) (string, string) {
	i := bytes.Index(line, colonSpace)
	if i != -1 {
		return textproto.CanonicalMIMEHeaderKey(string(line[:i])), string(line[i+2:])
	}
	return "", ""
}

// NoBody 定义没有数据的body。
var NoBody = noBody{}

type noBody struct{}

func (noBody) Read([]byte) (int, error)         { return 0, io.EOF }
func (noBody) Close() error                     { return nil }
func (noBody) WriteTo(io.Writer) (int64, error) { return 0, nil }
