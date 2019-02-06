package simple

import (
	"io"
	"fmt"
	"time"
	"bufio"
	"context"
	"strings"
	"net"
	"net/textproto"
	"crypto/tls"
)

type (
	Server struct {
		ctx			context.Context
		Handler		func(*Response, *Request)
	}
	conn struct {
		server		*Server
		rwc			net.Conn
	}
	Request struct {
		method		string
		requestURI	string
		proto		string
		header		Header
		reader		io.Reader
		conn		net.Conn
	}
	Response struct {
		request		*Request
		iswrite		bool
		status		int
		header		Header
		writer		*bufio.ReadWriter

	}
	Header map[string][]string
)

const TimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"
var Status = map[int]string{
	200:	"OK",
}

// 监听一个tcp连接，并启动服务。
func (srv *Server) ListenAndServe(addr string, handle func(*Response, *Request)) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	if handle != nil {
		srv.Handler = handle
	}
	return srv.Serve(ln)
}

func (srv *Server) Serve(l net.Listener) error {
	for {
		rw, err := l.Accept()
		if err != nil {
			break
		}
		srv.newConn(rw).serve(srv.ctx)
	}
	return nil
}

func (srv *Server) newConn(rwc net.Conn) *conn {
	return &conn{
		server:	srv,
		rwc:    rwc,
	}
}

func (c *conn) serve(ctx context.Context) {
	defer c.rwc.Close()
	var ok bool
	rw := bufio.NewReadWriter(bufio.NewReader(c.rwc), bufio.NewWriter(c.rwc))
	reader := textproto.NewReader(rw.Reader)
	fmt.Println("conn serve:", c.rwc.RemoteAddr().String())
	for {
		req := &Request{
			header:	Header{},
			reader: rw,
			conn:	c.rwc,
		}
		resp := &Response{
			request:req,
			status: 200,
			header:	Header{},
			writer:	rw,
		}
		// head line
		line, err := reader.ReadLine()
		if err != nil {
			return
		}
		fmt.Println("read line:", line)
		req.method, req.requestURI, req.proto, ok = parseRequestLine(line)
		if !ok {
			break
		}
		// header
		for {
			line, err := reader.ReadLine()
			if err != nil || len(line) == 0 {
				break
			}
			fmt.Println("read header:", line)
			req.header.Add(split2(line, ": "))
		}
		fmt.Println("handler start")
		// handler request
		c.server.Handler(resp, req)
		// writer
		resp.Flush()
		c.Close()
		fmt.Println("handler end")
	}
}

func (c *conn) Close() error {
	return c.rwc.Close()
}

// parseRequestLine parses "GET /foo HTTP/1.1" into its three parts.
func parseRequestLine(line string) (method, requestURI, proto string, ok bool) {
	s1 := strings.Index(line, " ")
	s2 := strings.Index(line[s1+1:], " ")
	if s1 < 0 || s2 < 0 {
		return
	}
	s2 += s1 + 1
	return line[:s1], line[s1+1 : s2], line[s2+1:], true
}

func split2(str string, s string) (string, string) {
	ss := strings.SplitN(str, s, 2)
	if len(ss) == 2 {
		return ss[0], ss[1]
	}
	return "", ""
}


func (r *Request) Method() string {
	return r.method
}
func (r *Request) Proto() string {
	return r.proto
}
func (r *Request) RequestURI() string {
	return r.requestURI
}
func (r *Request) Header() Header {
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


func (w *Response) Header() Header {
	return w.header
}

func (w *Response) Write(b []byte) (int, error) {
	if !w.iswrite {
		// set default header
		w.Header().Add("Date", time.Now().Format(TimeFormat))
		// write line and header
		// w.writer.Write([]byte("HTTP/1.1 200 OK\r\n"))
		fmt.Fprintf(w.writer, "%s: %d OK\r\n", w.request.Proto(), w.status, Status[w.status])
		for k, v := range w.header {
			fmt.Fprintf(w.writer, "%s: %s\r\n", k, v)
		}
		w.writer.Write([]byte("\r\n"))
		// set write flag is true.
		w.iswrite = true
	}
	return w.writer.Write(b)
}

func (w *Response) WriteHeader(codeCode int) {
	w.status = codeCode
}

func (w *Response) Flush() {
	w.writer.Flush()
}

func (w *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.request.conn, w.writer, nil
}

func (w *Response) Status() int {
	return w.status
}

func (h Header) Get(key string) string {
	val, ok := h[key]
	if ok {
		return val[0]
	}
	return ""
}

func (h Header) Add(key, val string) {
	h[key] = append(h[key], val)
}
