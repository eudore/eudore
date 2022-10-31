package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/http2"
)

func main() {
	srv := NewServerEudore()
	srv.SetHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.RequestURI)
		w.Write([]byte("eudore server"))
	}))
	srv.ListenAndServe(":8088")
}

var (
	crlf         = []byte("\r\n")
	colonSpace   = []byte(": ")
	constinueMsg = []byte("HTTP/1.1 100 Continue\r\n\r\n")
	rwPool       = sync.Pool{
		New: func() interface{} {
			return &Response{
				request: Request{
					Request: http.Request{
						ProtoMajor: 1,
						ProtoMinor: 1,
					},
					reader: bufio.NewReaderSize(nil, 2048),
				},
				writer: bufio.NewWriterSize(nil, 2048),
				buf:    make([]byte, 2048),
			}
		},
	}
	// ErrLineInvalid 定义http请求行无效的错误。
	ErrLineInvalid = errors.New("request line is invalid")
)

// Server 定义http server。
type Server struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	ctx          context.Context
	mu           sync.Mutex
	wg           sync.WaitGroup
	listeners    []net.Listener
	proto        string
	// nextHandler    protocol.HandlerConn
	httpHandler   http.Handler
	serverHandler func(context.Context, net.Conn, http.Handler)
	nextHandler   func(context.Context, *tls.Conn, http.Handler)
	Print         func(...interface{}) `alias:"print"`
}

// NewServerEudore 方法创建一个server
func NewServerEudore() *Server {
	return &Server{
		ReadTimeout:   60 * time.Second,
		WriteTimeout:  60 * time.Second,
		IdleTimeout:   60 * time.Second,
		ctx:           context.Background(),
		httpHandler:   http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		serverHandler: httpHandlerr,
		nextHandler:   NewHTTP2Handler(),
		proto:         "h2",
	}
}

// ListenAndServe 方法监听一个tcp连接，并启动服务。
func (srv *Server) ListenAndServe(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return srv.Serve(ln)
}

// ListenAndServeTLS 方法监听一个tcp连接，并启动服务。
func (srv *Server) ListenAndServeTLS(addr, certFile, keyFile string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	config := &tls.Config{
		Certificates:             make([]tls.Certificate, 1),
		PreferServerCipherSuites: true,
	}

	if config.NextProtos == nil && len(srv.proto) > 0 {
		config.NextProtos = []string{srv.proto}
	}

	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	return srv.Serve(tls.NewListener(ln, config))
}

// Serve 方法服务处理监听
func (srv *Server) Serve(ln net.Listener) error {
	srv.mu.Lock()
	for _, i := range srv.listeners {
		if i == ln {
			return fmt.Errorf("ln is serve status")
		}
	}
	srv.listeners = append(srv.listeners, ln)
	srv.mu.Unlock()
	for {
		// 读取连接
		c, err := ln.Accept()
		// 错误连接丢弃
		if err != nil {
			// 等待重试
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				time.Sleep(5 * time.Millisecond)
				continue
			}
			return err
		}
		// Handle new connections
		// 处理新连接
		go srv.newConnServe(c)
	}
}

func (srv *Server) newConnServe(conn net.Conn) {
	remoteAddr := conn.RemoteAddr().String()
	ctx := context.WithValue(srv.ctx, http.LocalAddrContextKey, remoteAddr)
	if tlsConn, ok := conn.(*tls.Conn); ok {
		if err := tlsConn.Handshake(); err != nil {
			srv.Print(fmt.Errorf("TLS handshake error from %s: %v", conn.RemoteAddr(), err))
			return
		}

		if proto := tlsConn.ConnectionState().NegotiatedProtocol; validNPN(proto) && proto == srv.proto && srv.nextHandler != nil {
			srv.nextHandler(ctx, tlsConn, srv.httpHandler)
			return
		}
	}
	srv.serverHandler(ctx, conn, srv.httpHandler)
}

// Shutdown 方法关闭Server
func (srv *Server) Shutdown(ctx context.Context) error {
	var stop = make(chan error)
	go func() {
		stop <- srv.Close()
	}()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-stop:
			return err
		}
	}
}

// Close 方法关闭Server
func (srv *Server) Close() (err error) {
	srv.mu.Lock()
	for _, ln := range srv.listeners {
		if e := ln.Close(); e != nil && err == nil {
			err = e
		}
	}
	srv.listeners = nil
	srv.mu.Unlock()
	return err
}

// validNPN reports whether the proto is not a blacklisted Next
// Protocol Negotiation protocol. Empty and built-in protocol types
// are blacklisted and can't be overridden with alternate
// implementations.
func validNPN(proto string) bool {
	switch proto {
	case "", "http/1.1", "http/1.0":
		return false
	}
	return true
}

// SetHandler 方法设置server的请求处理者
func (srv *Server) SetHandler(h http.Handler) {
	srv.httpHandler = h
}

// SetDefaulteHandler 方法设置默认http处理函数。
func (srv *Server) SetServerHandler(h func(context.Context, net.Conn, http.Handler)) {
	srv.serverHandler = h
}

// SetnextHandlerr 方法设置serve的tls处理函数。
func (srv *Server) SetnextHandlerr(proto string, h func(context.Context, *tls.Conn, http.Handler)) error {
	switch proto {
	case "h2":
		srv.proto, srv.nextHandler = proto, h
		return nil
	}
	return fmt.Errorf("tls nosuppered npn proto")
}

// httpHandlerr 函数处理http/1.1请求
func httpHandlerr(pctx context.Context, conn net.Conn, handler http.Handler) {
	// Initialize the request object.
	// 初始化请求对象。
	resp := rwPool.Get().(*Response)
	for {
		// c.SetReadDeadline(time.Now().Add(h.ReadTimeout))
		err := resp.request.Reset(conn)
		if err != nil {
			// handler error
			if isNotCommonNetReadError(err) {
				// h.Print("eudore http request read error: ", err)
			}
			break
		}
		resp.Reset(conn)
		ctx, cancelCtx := context.WithCancel(pctx)
		resp.cancel = cancelCtx
		// 处理请求
		// c.SetWriteDeadline(time.Now().Add(h.WriteTimeout))
		handler.ServeHTTP(resp, resp.request.Request.WithContext(ctx))
		if resp.ishjack {
			return
		}
		resp.finalFlush()
		if resp.request.Header.Get("Connection") != "keep-alive" {
			break
		}
		// c.SetDeadline(time.Now().Add(h.IdleTimeout))
	}
	conn.Close()
	rwPool.Put(resp)
}

// isNotCommonNetReadError 函数检查net读取错误是否未非通用错误。
func isNotCommonNetReadError(err error) bool {
	if err == io.EOF {
		return false
	}
	if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
		return false
	}
	if oe, ok := err.(*net.OpError); ok && oe.Op == "read" {
		return false
	}
	return true
}

// NewHTTP2Handler 方法创建一个h2处理函数。
func NewHTTP2Handler() func(context.Context, *tls.Conn, http.Handler) {
	h2svc := &http2.Server{}
	return func(ctx context.Context, conn *tls.Conn, h http.Handler) {
		h2svc.ServeConn(conn, &http2.ServeConnOpts{
			Context: ctx,
			Handler: h,
		})
	}
}

// Request 定义一个http请求。
type Request struct {
	http.Request
	conn net.Conn
	// read body
	reader     *bufio.Reader
	nextLength int64
	expect     bool
	mu         sync.Mutex
}

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
		r.Body = http.NoBody
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

// TimeFormat 定义响应header写入Date的时间格式。
const TimeFormat = "Mon, 02 Jan 2006 15:04:05 GMT"

// Response 定义http响应对象
type Response struct {
	request Request
	writer  *bufio.Writer
	header  http.Header
	status  int
	size    int
	iswrite bool
	chunked bool
	ishjack bool
	// buffer
	buf []byte
	n   int
	err error
	//
	cancel context.CancelFunc
}

// cancelConn 定义Conn在Close时执行Context cancel
type cancelConn struct {
	net.Conn
	cancel context.CancelFunc
}

// Reset 方法重置http响应状态
func (w *Response) Reset(conn net.Conn) {
	w.writer.Reset(conn)
	w.header = make(http.Header)
	w.status = 200
	w.size = 0
	w.iswrite = false
	w.chunked = false
	w.ishjack = false
	w.err = nil
	w.n = 0
}

// Header 方法获得http响应header对象。
func (w *Response) Header() http.Header {
	return w.header
}

// WriteHeader 方法写入状态码
func (w *Response) WriteHeader(codeCode int) {
	w.status = codeCode
}

// Write 方法写入数据，如果写入数据长度小于缓冲，不会立刻返回，也不会写入状态行。
func (w *Response) Write(p []byte) (int, error) {
	// 数据大于缓冲，发送数据
	if w.n+len(p) > len(w.buf) {
		// 写入数据
		n, _ := w.writeDate(p, len(p))
		// 更新数据长度
		w.size += n
		return n, w.err
	}
	// 数据小于缓存，保存
	n := copy(w.buf[w.n:], p)
	w.n += n
	// 更新数据长度
	w.size += n
	return n, nil
}

// writeDate 方法写入数据并返回。
//
// 会先写入缓冲数据，然后将当前数据写入
//
// 提升分块效率，会将大小两块合并发送。
func (w *Response) writeDate(p []byte, length int) (n int, err error) {
	// 写入状态行
	w.writerResponseLine()
	// 如果有写入错误，或者数据长度为0则返回。
	if w.err != nil || (length+w.n) == 0 {
		return 0, w.err
	}
	// 数据写入
	if w.chunked {
		// 分块写入
		fmt.Fprintf(w.writer, "%x\r\n", length+w.n)
		// 写入缓冲数据和当前数据
		w.writer.Write(w.buf[0:w.n])
		n, err = w.writer.Write(p)
		// 分块结束
		w.writer.Write([]byte{13, 10})
	} else {
		w.writer.Write(w.buf[0:w.n])
		n, err = w.writer.Write(p)
	}
	w.n = 0
	// 检测写入的长度
	if n < length {
		err = io.ErrShortWrite
	}
	w.err = err
	return
}

// writerResponseLine 方法写入状态行
func (w *Response) writerResponseLine() {
	// 已经写入则返回
	if w.iswrite {
		return
	}
	// 设置写入标志为true。
	w.iswrite = true
	// Write response line
	// 写入响应行
	fmt.Fprintf(w.writer, "%s %d %s\r\n", w.request.Proto, w.status, http.StatusText(w.status))
	// Write headers
	// 写入headers
	for key, vals := range w.header {
		for _, val := range vals {
			fmt.Fprintf(w.writer, "%s: %s\r\n", key, val)
		}
	}
	// 写入时间和Server
	fmt.Fprintf(w.writer, "Date: %s\r\nServer: eudore\r\n", time.Now().UTC().Format(TimeFormat))
	// 检测是否有写入长度，没有则进行分块传输。
	// 未检测Content-Length值是否合法
	w.chunked = len(w.header.Get("Content-Length")) == 0 && w.header.Get("Upgrade") == ""
	if w.chunked {
		fmt.Fprintf(w.writer, "Transfer-Encoding: chunked\r\n")
	}
	// Write header separator
	// 写入header后分割符
	w.writer.Write([]byte("\r\n"))
}

// Flush 方法数据写入
func (w *Response) Flush() {
	// 将缓冲数据写入
	w.writeDate(nil, 0)
	w.n = 0
	// 发送writer的全部数据
	w.writer.Flush()
}

// finalFlush 方法请求结束时flush写入数据。
func (w *Response) finalFlush() (err error) {
	// 如果没有写入状态行，并且没有指定内容长度。
	// 设置内容长度为当前缓冲数据。
	if !w.iswrite && len(w.header.Get("Content-Length")) == 0 {
		w.header.Set("Content-Length", fmt.Sprint(w.n))
	}
	// 将缓冲数据写入
	w.writeDate(nil, 0)
	// 处理分段传输
	if w.chunked {
		// 处理Trailer header
		tr := w.header.Get("Trailer")
		if len(tr) == 0 {
			// 没有Trailer,直接写入结束
			w.writer.Write([]byte{0x30, 0x0d, 0x0a, 0x0d, 0x0a})
		} else {
			// 写入结尾
			w.writer.Write([]byte{0x30, 0x0d, 0x0a})
			// 写入Trailer的值
			for _, k := range strings.Split(tr, ",") {
				fmt.Fprintf(w.writer, "%s: %s\r\n", k, w.header.Get(k))
			}
			w.writer.Write([]byte{0x0d, 0x0a})
		}
	}
	// 发送数据
	err = w.writer.Flush()
	w.cancel()
	return
}

// Hijack 方法劫持http连接。
func (w *Response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.ishjack = true
	// w.request.conn.SetDeadline(time.Time{})
	// return &cancelConn{w.request.conn, w.cancel}, nil
	return w.request.conn, bufio.NewReadWriter(w.request.reader, w.writer), nil
}

// Push 方法http协议不支持push方法。
func (*Response) Push(string, *http.PushOptions) error {
	return nil
}

// Status 方法返回当前状态码。
func (w *Response) Status() int {
	return w.status
}

// Size 方法返回写入的数据长度。
func (w *Response) Size() int {
	return w.size
}

// Close 方法在net.Conn关闭时，执行context cancel。
func (c *cancelConn) Close() (err error) {
	err = c.Conn.Close()
	c.cancel()
	return
}
