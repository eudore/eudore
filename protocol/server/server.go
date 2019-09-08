package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/eudore/eudore/protocol"
	"github.com/eudore/eudore/protocol/http"
	"net"
	"sync"
	"time"
)

type (
	// Server 定义一个http server
	Server struct {
		ReadTimeout   time.Duration
		WriteTimeout  time.Duration
		IdleTimeout   time.Duration
		ctx           context.Context
		mu            sync.Mutex
		listeners     []net.Listener
		proto         string
		nextHandle    protocol.HandlerConn
		defaultHandle protocol.HandlerConn
		Print         func(...interface{}) `set:"print"`
	}
	// SetTimeouter 定义设置超时的接口
	SetTimeouter interface {
		SetIdleTimeout(time.Duration)
		SetReadTimeout(time.Duration)
		SetWriteTimeout(time.Duration)
	}
	// SetPrinter 定义设置输出函数的接口
	SetPrinter interface {
		SetPrint(func(...interface{}))
	}
)

// NextProtoTLS 定义TLS Next Proto的值
var NextProtoTLS = "h2"

// LocalAddrContextKey 定义context获得远程地址的key
var LocalAddrContextKey = struct{}{}

// NewServer 方法创建一个server
func NewServer() *Server {
	return &Server{
		ReadTimeout:   60 * time.Second,
		WriteTimeout:  60 * time.Second,
		IdleTimeout:   60 * time.Second,
		ctx:           context.Background(),
		defaultHandle: &http.Handler{},
	}
}

// ListenAndServe 方法监听一个tcp连接，并启动服务。
func (srv *Server) ListenAndServe(addr string, handle protocol.HandlerHTTP) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return srv.Serve(ln)
}

// ListenAndServeTLS 方法监听一个tcp连接，并启动服务。
func (srv *Server) ListenAndServeTLS(addr, certFile, keyFile string, handle protocol.HandlerHTTP) error {
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

func (srv *Server) newConnServe(c net.Conn) {
	remoteAddr := c.RemoteAddr().String()
	ctx := context.WithValue(srv.ctx, LocalAddrContextKey, remoteAddr)
	if tlsConn, ok := c.(*tls.Conn); ok {
		if err := tlsConn.Handshake(); err != nil {
			// Gol.12 version
			// If the handshake failed due to the client not speaking
			// TLS, assume they're speaking plaintext HTTP and write a
			// 400 response on the TLS conn's underlying net.Conn.
			// if re, ok := err.(tls.RecordHeaderError); ok && re.Conn != nil && tlsRecordHeaderLooksLikeHTTP(re.RecordHeader) {
			// 	io.WriteString(re.Conn, "HTTP/1.0 400 Bad Request\r\n\r\nClient sent an HTTP request to an HTTPS server.\n")
			// 	re.Conn.Close()
			// 	return
			// }
			// c.server.logf("http: TLS handshake error from %s: %v", c.rwc.RemoteAddr(), err)
			srv.Print(fmt.Errorf("TLS handshake error from %s: %v", c.RemoteAddr(), err))
			return
		}

		if proto := tlsConn.ConnectionState().NegotiatedProtocol; validNPN(proto) && proto == srv.proto {
			srv.nextHandle.EudoreConn(ctx, c)
			return
		}
	}
	srv.defaultHandle.EudoreConn(ctx, c)
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

// SetHandler 方法设置serve的连接处理者
func (srv *Server) SetHandler(h protocol.HandlerConn) {
	srv.defaultHandle = h
	srv.SetHandlerTimeouter(h)
	srv.SetHandlerPrinter(h)
}

// SetNextHandler 方法设置serve的next连接处理者
func (srv *Server) SetNextHandler(proto string, h protocol.HandlerConn) error {
	switch proto {
	case "h2":
		srv.proto, srv.nextHandle = proto, h
		srv.SetHandlerTimeouter(h)
		srv.SetHandlerPrinter(h)
		return nil
	}
	return fmt.Errorf("tls nosuppered npn proto")
}

// SetHandlerTimeouter 方法设置连接处理者超时
func (srv *Server) SetHandlerTimeouter(h protocol.HandlerConn) {
	if timer, ok := h.(SetTimeouter); ok {
		timer.SetIdleTimeout(srv.IdleTimeout)
		timer.SetReadTimeout(srv.ReadTimeout)
		timer.SetWriteTimeout(srv.WriteTimeout)
	}
}

// SetHandlerPrinter 方法设置连接处理者输出函数
func (srv *Server) SetHandlerPrinter(h protocol.HandlerConn) {
	if printer, ok := h.(SetPrinter); ok {
		printer.SetPrint(srv.Print)
	}
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

// tlsRecordHeaderLooksLikeHTTP reports whether a TLS record header
// looks like it might've been a misdirected plaintext HTTP request.
func tlsRecordHeaderLooksLikeHTTP(hdr [5]byte) bool {
	switch string(hdr[:]) {
	case "GET /", "HEAD ", "POST ", "PUT /", "OPTIO":
		return true
	}
	return false
}
