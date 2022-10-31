// Package simple implementation of the http protocol as a learning sample.
package simple

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/textproto"
	"strings"
	"time"
)

func main() {
	ln, err := net.Listen("tcp", ":8085")
	if err != nil {
		fmt.Println(err)
		return
	}
	server := &Server{
		Handler: func(w *Response, r *Request) {
			w.Header().Add("Server", "simple server")
			w.Write([]byte("hello http server. your remote addr is " + r.RemoteAddr()))
		},
	}
	server.Serve(ln)
}

type (
	// Server 定义http服务。
	Server struct {
		ctx     context.Context
		Handler func(*Response, *Request)
	}
	// conn 定义http连接。
	conn struct {
		server *Server
		rwc    net.Conn
	}
	// Request 定义http请求
	Request struct {
		method     string
		requestURI string
		proto      string
		header     Header
		reader     io.Reader
		conn       net.Conn
	}
	// Response 定义http响应
	Response struct {
		request *Request
		iswrite bool
		status  int
		header  Header
		writer  *bufio.ReadWriter
	}
	// Header 定义http请求。
	Header map[string][]string
)

// TimeFormat 是Date header格式化时间。
const TimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

// Status 定义状态的描述。
var Status = map[int]string{
	200: "OK",
}

// ListenAndServe 方法监听一个tcp连接，并启动服务。
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

// Serve 方法服务处理监听
func (srv *Server) Serve(l net.Listener) error {
	for {
		// 读取连接
		rw, err := l.Accept()
		// 错误连接丢弃
		if err != nil {
			break
		}
		// Handle new connections
		// 处理新连接
		go srv.newConn(rw).serve(srv.ctx)
	}
	return nil
}

// newConn Encapsulate an http connection object
//
// newConn 封装一个http连接对象
func (srv *Server) newConn(rwc net.Conn) *conn {
	return &conn{
		server: srv,
		rwc:    rwc,
	}
}

// serve Handling http connections
//
// serve 处理http连接
func (c *conn) serve(ctx context.Context) {
	defer c.rwc.Close()
	var ok bool
	// Create the currently connected io buffer object.
	// 创建当前连接的io缓冲对象。
	rw := bufio.NewReadWriter(bufio.NewReader(c.rwc), bufio.NewWriter(c.rwc))
	// Create a text protocol parsing object.
	// 创建一个文本协议解析对象。
	reader := textproto.NewReader(rw.Reader)
	fmt.Println("conn serve:", c.rwc.RemoteAddr().String())
	for {
		// Initialize the request object.
		// 初始化请求对象。
		req := &Request{
			header: Header{},
			reader: rw,
			conn:   c.rwc,
		}
		resp := &Response{
			request: req,
			status:  200,
			header:  Header{},
			writer:  rw,
		}
		// Read the http request line.
		// 读取http请求行。
		line, err := reader.ReadLine()
		if err != nil {
			return
		}
		fmt.Println("read line:", line)
		// Split the http request line.
		// 拆分http请求行。
		req.method, req.requestURI, req.proto, ok = parseRequestLine(line)
		if !ok {
			break
		}
		// read http headers
		// 读取http headers
		for {
			// Read a line of content.
			// 读取一行内容。
			line, err := reader.ReadLine()
			if err != nil || len(line) == 0 {
				break
			}
			fmt.Println("read header:", line)
			// Split into headers and store them in the request.
			// 分割成header存储到请求中。
			req.header.Add(split2(line, ": "))
		}
		fmt.Println("handler start")
		// Call the handle object to handle the request.
		// 调用handle对象处理这个请求。
		c.server.Handler(resp, req)
		// Write the cached data and send it back to the client.
		// 将缓存数据写入，发送返回给客户端。
		resp.Flush()
		// // Close the connection and do not implement connection multiplexing.
		// 关闭连接，未实现连接复用。
		c.Close()
		fmt.Println("handler end")
	}
}

// Close 方法关闭http连接
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

// Method 方法获取请求的http方法。
func (r *Request) Method() string {
	return r.method
}

// Proto 方法获取http协议版本。
func (r *Request) Proto() string {
	return r.proto
}

// RequestURI 方法获取http请求uri。
func (r *Request) RequestURI() string {
	return r.requestURI
}

// Header 方法返回请求header。
func (r *Request) Header() Header {
	return r.header
}

// Read 方法读取http请求内容。
func (r *Request) Read(b []byte) (int, error) {
	return r.reader.Read(b)
}

// Host 方法获得请求Host。
func (r *Request) Host() string {
	return r.header.Get("Host")
}

// RemoteAddr 方法获取远程连接地址。
func (r *Request) RemoteAddr() string {
	return r.conn.RemoteAddr().String()
}

// TLS 方法获取TLS状态，未实现，应该net.Conn短语*tls.Conn，然后获得。
func (r *Request) TLS() *tls.ConnectionState {
	return nil
}

// Header 方法返回响应Header。
func (w *Response) Header() Header {
	return w.header
}

// Write 方法实现写入响应内容。
func (w *Response) Write(b []byte) (int, error) {
	// If it is the first time to write to the body, write the response line and headers before this.
	// 如果是第一次写入body，在此之前写入响应行和headers。
	if !w.iswrite {
		// Set default headers
		// 设置默认headers
		w.Header().Add("Date", time.Now().Format(TimeFormat))
		// Write response line
		// 写入响应行
		fmt.Fprintf(w.writer, "%s %d %s\r\n", w.request.Proto(), w.status, Status[w.status])
		// Write headers
		// 写入headers
		for k, v := range w.header {
			fmt.Fprintf(w.writer, "%s: %s\r\n", k, v[0])
		}
		// Write header separator
		// 写入header后分割符
		w.writer.Write([]byte("\r\n"))
		// Set the write standard to true.
		// 设置写入标准为true。
		w.iswrite = true
	}
	return w.writer.Write(b)
}

// WriteHeader 方法设置响应状态码。
func (w *Response) WriteHeader(codeCode int) {
	w.status = codeCode
}

// Flush 实现Flush内容。
func (w *Response) Flush() {
	w.writer.Flush()
}

// Hijack 方法实现获取连接对象。
func (w *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.request.conn, w.writer, nil
}

// Status 方法获取设置的状态码
func (w *Response) Status() int {
	return w.status
}

// Get 方法获得header的值
func (h Header) Get(key string) string {
	val, ok := h[key]
	if ok {
		return val[0]
	}
	return ""
}

// Add 方法添加一个Header。
func (h Header) Add(key, val string) {
	h[key] = append(h[key], val)
}
