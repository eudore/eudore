package eudore

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Server defines the object that starts the http service.
type Server interface {
	SetHandler(h http.Handler)
	Serve(ln net.Listener) error
	Shutdown(ctx context.Context) error
}

// ServerConfig defines the configuration used by [NewServer].
type ServerConfig struct {
	// set default ServerHandler
	Handler http.Handler `alias:"handler" json:"-" yaml:"-"`

	// ReadTimeout is the maximum duration for reading the entire
	// request, including the body. A zero or negative value means
	// there will be no timeout.
	//
	// Because ReadTimeout does not let Handlers make per-request
	// decisions on each request body's acceptable deadline or
	// upload rate, most users will prefer to use
	// ReadHeaderTimeout. It is valid to use them both.
	ReadTimeout TimeDuration `alias:"readTimeout" json:"readTimeout" yaml:"readTimeout"`

	// WriteTimeout is the maximum duration before timing out
	// writes of the response. It is reset whenever a new
	// request's header is read. Like ReadTimeout, it does not
	// let Handlers make decisions on a per-request basis.
	// A zero or negative value means there will be no timeout.
	WriteTimeout TimeDuration `alias:"writeTimeout" json:"writeTimeout" yaml:"writeTimeout"`

	// ReadHeaderTimeout is the amount of time allowed to read
	// request headers. The connection's read deadline is reset
	// after reading the headers and the Handler can decide what
	// is considered too slow for the body. If zero, the value of
	// ReadTimeout is used. If negative, or if zero and ReadTimeout
	// is zero or negative, there is no timeout.
	ReadHeaderTimeout TimeDuration `alias:"readHeaderTimeout" json:"readHeaderTimeout" yaml:"readHeaderTimeout"`

	// IdleTimeout is the maximum amount of time to wait for the
	// next request when keep-alives are enabled. If zero, the value
	// of ReadTimeout is used. If negative, or if zero and ReadTimeout
	// is zero or negative, there is no timeout.
	IdleTimeout TimeDuration `alias:"idleTimeout" json:"idleTimeout" yaml:"idleTimeout"`

	// MaxHeaderBytes controls the maximum number of bytes the
	// server will read parsing the request header's keys and
	// values, including the request line. It does not limit the
	// size of the request body.
	// If zero, DefaultMaxHeaderBytes is used.
	MaxHeaderBytes int `alias:"maxHeaderBytes" json:"maxHeaderBytes" yaml:"maxHeaderBytes"`

	// ErrorLog specifies an optional logger for errors accepting
	// connections, unexpected behavior from handlers, and
	// underlying FileSystem errors.
	// If nil, logging is done via the log package's standard logger.
	ErrorLog *log.Logger `alias:"errorLog" json:"-" yaml:"-"`

	// BaseContext optionally specifies a function that returns
	// the base context for incoming requests on this server.
	// The provided Listener is the specific Listener that's
	// about to start accepting requests.
	// If BaseContext is nil, the default is context.Background().
	// If non-nil, it must return a non-nil context.
	BaseContext func(net.Listener) context.Context `alias:"baseContext" json:"-" yaml:"-"`

	// ConnContext optionally specifies a function that modifies
	// the context used for a new connection c. The provided ctx
	// is derived from the base context and has a ServerContextKey
	// value.
	ConnContext func(context.Context, net.Conn) context.Context `alias:"connContext" json:"-" yaml:"-"`
}

// serverStd defines using [http.Server] to start the http server.
type serverStd struct {
	*http.Server `alias:"server"`
	Mutex        sync.Mutex       `alias:"mutex"`
	listener     internalListener `alias:"listener"`
	Ports        []string         `alias:"ports"`
	Counter      int64            `alias:"counter"`
}

type MetadataServer struct {
	Health     bool     `json:"health" protobuf:"1,name=health" yaml:"health"`
	Name       string   `json:"name" protobuf:"2,name=name" yaml:"name"`
	Ports      []string `json:"ports" protobuf:"3,name=ports" yaml:"ports"`
	ErrorCount int64    `json:"errorCount" protobuf:"4,name=errorCount" yaml:"errorCount"`
}

// serverStd defines using [fcgi.Serve] to start the http server.
type serverFcgi struct {
	http.Handler
	sync.Mutex
	listeners []net.Listener
}

// ServerListenConfig defines a common port listening configuration.
type ServerListenConfig struct {
	Addr        string            `alias:"addr" json:"addr" yaml:"addr"`
	HTTPS       bool              `alias:"https" json:"https" yaml:"https"`
	HTTP2       bool              `alias:"http2" json:"http2" yaml:"http2"`
	Mutual      bool              `alias:"mutual" json:"mutual" yaml:"mutual"`
	Certfile    string            `alias:"certfile" json:"certfile" yaml:"certfile"`
	Keyfile     string            `alias:"keyfile" json:"keyfile" yaml:"keyfile"`
	Trustfile   string            `alias:"trustfile" json:"trustfile" yaml:"trustfile"`
	Certificate *x509.Certificate `alias:"certificate" json:"certificate" yaml:"certificate"`
}

// The NewServer function creates a [Server] implemented by warp [http.Server].
func NewServer(config *ServerConfig) Server {
	if config == nil {
		config = &ServerConfig{}
	}
	srv := &serverStd{
		Server: &http.Server{
			Handler: config.Handler,
			ReadTimeout: GetAnyDefault(
				time.Duration(config.ReadTimeout),
				DefaultServerReadTimeout,
			),
			ReadHeaderTimeout: GetAnyDefault(
				time.Duration(config.ReadHeaderTimeout),
				DefaultServerReadHeaderTimeout,
			),
			WriteTimeout: GetAnyDefault(
				time.Duration(config.WriteTimeout),
				DefaultServerWriteTimeout,
			),
			IdleTimeout: GetAnyDefault(
				time.Duration(config.IdleTimeout),
				DefaultServerIdleTimeout,
			),
			MaxHeaderBytes: config.MaxHeaderBytes,
			ErrorLog:       config.ErrorLog,
			BaseContext:    config.BaseContext,
			ConnContext:    config.ConnContext,
		},
	}
	// fix http2 server in golang ?-1.22
	// https://github.com/golang/go/issues/65785
	if srv.Server.ReadTimeout < 0 {
		srv.Server.ReadTimeout = 0
	}
	if srv.Server.ReadHeaderTimeout < 0 {
		srv.Server.ReadHeaderTimeout = 0
	}
	if srv.Server.WriteTimeout < 0 {
		srv.Server.WriteTimeout = 0
	}
	if srv.Server.IdleTimeout < 0 {
		srv.Server.IdleTimeout = 0
	}
	return srv
}

// The Mount method gets [ContextKeyHTTPHandler] or [ContextKeyApp] from
// [context.Context] as [http.Handler],
//
// Get [ContextKeyApp] or [ContextKeyLogger] as [Logger] to
// receive [http.Server.ErrorLog].
func (srv *serverStd) Mount(ctx context.Context) {
	if srv.Handler == nil {
		for _, key := range [...]any{ContextKeyHTTPHandler, ContextKeyApp} {
			h, ok := ctx.Value(key).(http.Handler)
			if ok {
				srv.SetHandler(h)
				break
			}
		}
	}

	if srv.ErrorLog == nil {
		// Capture the error content output by net/http.Server.
		for _, key := range [...]any{ContextKeyApp, ContextKeyLogger} {
			logger, ok := ctx.Value(key).(Logger)
			if ok {
				out := &serverLogger{
					Logger:  logger,
					Counter: &srv.Counter,
				}
				srv.ErrorLog = log.New(out, "", 0)
				break
			}
		}
	}
}

// Unmount method waits for [DefaulerServerShutdownWait] to use
// [http.Server.Shutdown] to shut down [Server] listening.
func (srv *serverStd) Unmount(context.Context) {
	ctx, cancel := context.WithTimeout(context.Background(),
		DefaultServerShutdownWait,
	)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func (srv *serverStd) SetHandler(h http.Handler) {
	srv.Mutex.Lock()
	defer srv.Mutex.Unlock()
	srv.Server.Handler = h
}

func (srv *serverStd) Serve(ln net.Listener) error {
	srv.Mutex.Lock()
	srv.Ports = append(srv.Ports, ln.Addr().String())
	srv.Mutex.Unlock()
	err := srv.Server.Serve(ln)
	if errors.Is(err, http.ErrServerClosed) {
		err = nil
	}
	return err
}

// The ServeConn method handles a [net.Conn].
//
// Implement [net.Listen] to pass [net.Conn] to [http.Servr].
func (srv *serverStd) ServeConn(conn net.Conn) {
	srv.Mutex.Lock()
	if srv.listener.Ch == nil {
		srv.listener.Ch = make(chan net.Conn)
		srv.Ports = append(srv.Ports, srv.listener.Addr().String())
		go func() {
			_ = srv.Server.Serve(&srv.listener)
		}()
	}
	srv.Mutex.Unlock()
	srv.listener.Ch <- conn
}

// The Metadata method returns [MetadataServer].
func (srv *serverStd) Metadata() any {
	srv.Mutex.Lock()
	defer srv.Mutex.Unlock()
	return MetadataServer{
		Health:     true,
		Name:       "eudore.serverStd",
		Ports:      srv.Ports,
		ErrorCount: atomic.LoadInt64(&srv.Counter),
	}
}

type serverLogger struct {
	Logger  Logger
	Counter *int64
}

func (srv *serverLogger) Write(p []byte) (n int, err error) {
	atomic.AddInt64(srv.Counter, 1)
	log := srv.Logger.WithField(ParamDepth, DefaultLoggerDepthKindDisable).
		WithField(ParamCaller, "serverStd.ErrorLog")
	strs := strings.Split(string(p), "\n")
	if strings.HasPrefix(strs[0], "http: panic serving ") {
		lines := []string{}
		for i := 2; i < len(strs)-1; i += 2 {
			if strings.HasPrefix(strs[i], "created by ") {
				strs[i] = strs[i][11:]
			} else {
				end := strings.LastIndexByte(strs[i], '(')
				if end != -1 {
					strs[i] = strs[i][:end]
				}
			}
			pos := strings.IndexByte(strs[i+1], ' ')
			if pos != -1 {
				strs[i+1] = strs[i+1][:pos]
			}
			lines = append(lines,
				strings.TrimPrefix(strs[i+1], "\t")+" "+strs[i],
			)
		}
		log.WithField("stack", lines).Errorf("%s %s",
			strs[0], strs[1][:len(strs[1])-1],
		)
	} else {
		log.Errorf(strs[0])
	}
	return 0, nil
}

type internalListener struct {
	Ch    chan net.Conn
	close bool
}

func (ln *internalListener) Accept() (net.Conn, error) {
	for conn := range ln.Ch {
		if conn != nil {
			return conn, nil
		}
	}
	return nil, http.ErrServerClosed
}

func (ln *internalListener) Close() error {
	if !ln.close {
		close(ln.Ch)
		ln.close = true
	}
	return nil
}

func (ln *internalListener) Addr() net.Addr {
	return &net.IPAddr{
		IP: net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 255, 127, 0, 0, 1},
	}
}

// The NewServerFcgi function creates a [Server] using [fcgi.Serve].
func NewServerFcgi() Server {
	return &serverFcgi{}
}

// The Mount method gets [ContextKeyHTTPHandler] or [ContextKeyApp] from
// [context.Context] as [http.Handler].
func (srv *serverFcgi) Mount(ctx context.Context) {
	if srv.Handler == nil {
		for _, key := range [...]any{ContextKeyHTTPHandler, ContextKeyApp} {
			h, ok := ctx.Value(key).(http.Handler)
			if ok {
				srv.SetHandler(h)
				break
			}
		}
	}
}

// Unmount method waits for [DefaulerServerShutdownWait] shuts down all
// fcgi listeners.
func (srv *serverFcgi) Unmount(context.Context) {
	ctx, cancel := context.WithTimeout(context.Background(),
		DefaultServerShutdownWait,
	)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func (srv *serverFcgi) SetHandler(h http.Handler) {
	srv.Handler = h
}

func (srv *serverFcgi) Serve(ln net.Listener) error {
	srv.Lock()
	srv.listeners = append(srv.listeners, ln)
	srv.Unlock()
	return fcgi.Serve(ln, srv.Handler)
}

// The Shutdown method shuts down all fcgi listeners.
func (srv *serverFcgi) Shutdown(context.Context) error {
	srv.Lock()
	defer srv.Unlock()
	var err error
	for _, ln := range srv.listeners {
		if cerr := ln.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}
	return err
}

// The Listen method uses the port configuration to create a listener,
// and uses Certificate to save the parsed TLS certificate.
//
// If https is enabled but there is no certificate, a private certificate will
// be created.
func (slc *ServerListenConfig) Listen() (net.Listener, error) {
	// set default port
	if slc.Addr == "" {
		if slc.HTTPS {
			slc.Addr = ":443"
		} else {
			slc.Addr = ":80"
		}
	}
	if !slc.HTTPS {
		return DefaultServerListen("tcp", slc.Addr)
	}

	// set tls
	config := &tls.Config{
		NextProtos:   []string{"http/1.1"},
		Certificates: make([]tls.Certificate, 1),
	}
	if slc.HTTP2 {
		config.NextProtos = []string{"h2"}
	}

	cert, ca, err := loadCertificate(slc.Certfile, slc.Keyfile)
	if err != nil {
		return nil, err
	}
	config.Certificates[0], slc.Certificate = cert, ca

	// set mutual tls
	if slc.Mutual {
		data, err := os.ReadFile(slc.Trustfile)
		if err != nil {
			return nil, err
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(data)
		config.ClientCAs = pool
		config.ClientAuth = tls.RequireAndVerifyClientCert
	}

	ln, err := DefaultServerListen("tcp", slc.Addr)
	if err != nil {
		return nil, err
	}
	return tls.NewListener(ln, config), nil
}

func loadCertificate(cret, key string) (tls.Certificate, *x509.Certificate,
	error,
) {
	if cret != "" && key != "" {
		cret509, err := tls.LoadX509KeyPair(cret, key)
		if err != nil {
			return cret509, nil, err
		}
		ca, _ := x509.ParseCertificate(cret509.Certificate[0])
		return cret509, ca, err
	}

	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1653),
		Subject: pkix.Name{
			Country:            []string{"China"},
			Organization:       []string{"eudore"},
			OrganizationalUnit: []string{"eudore"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		SubjectKeyId:          []byte{1, 2, 3, 4, 5},
		BasicConstraintsValid: true,

		IsCA: true,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}
	pool := x509.NewCertPool()
	pool.AddCert(ca)

	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	caByte, err := x509.CreateCertificate(rand.Reader, ca, ca,
		&priv.PublicKey, priv,
	)

	return tls.Certificate{
		Certificate: [][]byte{caByte},
		PrivateKey:  priv,
	}, ca, err
}
