package server

import (
	"net"
	"fmt"
	"time"
	"sync"
	"context"
	"crypto/tls"
	"github.com/eudore/eudore/protocol"
	"github.com/eudore/eudore/protocol/http"
)


type (
	Server struct {
		ctx			context.Context
		Handler		protocol.Handler
		mu			sync.Mutex
		listeners	[]net.Listener
		proto		string
		nextHandle		protocol.HandlerConn
		defaultHandle	protocol.HandlerConn
	}
)

var NextProtoTLS = "h2"

// 监听一个tcp连接，并启动服务。
func (srv *Server) ListenAndServe(addr string, handle protocol.Handler) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	if handle != nil {
		srv.Handler = handle
	}
	return srv.Serve(ln)
}

// 监听一个tcp连接，并启动服务。
func (srv *Server) ListenAndServeTls(addr , certFile, keyFile string, handle protocol.Handler) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	if handle != nil {
		srv.Handler = handle
	}
	config := &tls.Config{
		Certificates: make([]tls.Certificate, 1),
		PreferServerCipherSuites:	true,
	}

	if config.NextProtos == nil && len(srv.proto) > 0{
		config.NextProtos = []string{srv.proto}
	}

	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	return srv.Serve(tls.NewListener(ln, config))
}

// 服务处理监听
func (srv *Server) Serve(ln net.Listener) error {
	srv.mu.Lock()
	for _, i := range srv.listeners {
		if i == ln {
			return fmt.Errorf("ln is serve status")
		}
	}
	srv.listeners = append(srv.listeners, ln)
	if srv.defaultHandle == nil {
		srv.defaultHandle = &http.HttpHandler{}
	}
	srv.ctx = context.Background()
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
	return nil
}

func (srv *Server) newConnServe(c net.Conn) {
	remoteAddr := c.RemoteAddr().String()
	ctx := context.WithValue(srv.ctx, "LocalAddrContextKey", remoteAddr)
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
			fmt.Printf("http: TLS handshake error from %s: %v\n", c.RemoteAddr(), err)
			return
		}
		
		if proto := tlsConn.ConnectionState().NegotiatedProtocol; validNPN(proto) && proto == srv.proto {
			srv.nextHandle.EudoreConn(ctx, c)
			return
		}
	}
	srv.defaultHandle.EudoreConn(ctx, c)
}

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

func (srv *Server) Shutdown(ctx context.Context) error {
	var stop = make(chan error)
	go func(){
		stop <- srv.Close()
	}()
	for {
		select {
		case <- ctx.Done():
			return ctx.Err()
		case err := <- stop:
			return err
		}
	}
	return nil
}

func (srv *Server) SetHandler(h protocol.HandlerConn) {
	srv.defaultHandle = h
}

func (srv *Server) SetNextHandler(proto string, h protocol.HandlerConn) error{
	switch proto {
	case "h2":
		srv.proto, srv.nextHandle = proto, h
		return nil
	}
	return fmt.Errorf("tls nosuppered npn proto.")
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

