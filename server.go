package eudore

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/fcgi"
	"sync"
	"sync/atomic"
	"time"
)

// Server 定义启动http服务的对象。
type Server interface {
	SetHandler(http.Handler)
	Serve(net.Listener) error
	Shutdown(context.Context) error
}

// ServerStdConfig 定义serverStd使用的配置
type ServerStdConfig struct {
	// set default ServerHandler
	Handler http.Handler

	// ReadTimeout is the maximum duration for reading the entire request, including the body.
	//
	// Because ReadTimeout does not let Handlers make per-request decisions on each request body's acceptable deadline or upload rate,
	// most users will prefer to use ReadHeaderTimeout. It is valid to use them both.
	ReadTimeout TimeDuration `alias:"readtimeout" json:"readtimeout" description:"Http server read timeout."`

	// ReadHeaderTimeout is the amount of time allowed to read request headers.
	// The connection's read deadline is reset after reading the headers and the Handler can decide what is considered too slow for the body.
	ReadHeaderTimeout TimeDuration `alias:"readheadertimeout" json:"readheadertimeout"  description:"Http server read header timeout."` // Go 1.8

	// WriteTimeout is the maximum duration before timing out writes of the response.
	// It is reset whenever a new request's header is read.
	// Like ReadTimeout, it does not let Handlers make decisions on a per-request basis.
	WriteTimeout TimeDuration `alias:"writetimeout" json:"writetimeout" description:"Http server write timeout."`

	// IdleTimeout is the maximum amount of time to wait for the next request when keep-alives are enabled.
	// If IdleTimeout is zero, the value of ReadTimeout is used. If both are zero, ReadHeaderTimeout is used.
	IdleTimeout TimeDuration `alias:"idletimeout" json:"idletimeout"  description:"Http server idle timeout."` // Go 1.8

	// MaxHeaderBytes controls the maximum number of bytes the server will read parsing the request header's keys and values, including the request line.
	// It does not limit the size of the request body. If zero, DefaultMaxHeaderBytes is used.
	MaxHeaderBytes int `alias:"maxheaderbytes" json:"maxheaderbytes"  description:"Http server max header size."`

	// ErrorLog specifies an optional logger for errors accepting
	// connections, unexpected behavior from handlers, and
	// underlying FileSystem errors.
	// If nil, logging is done via the log package's standard logger.
	ErrorLog *log.Logger // Go 1.3

	// BaseContext optionally specifies a function that returns the base context for incoming requests on this server.
	// The provided Listener is the specific Listener that's about to start accepting requests.
	// If BaseContext is nil, the default is context.Background(). If non-nil, it must return a non-nil context.
	BaseContext func(net.Listener) context.Context `alias:"basecontext" json:"-"` // Go 1.13

	// ConnContext optionally specifies a function that modifies the context used for a new connection c.
	// The provided ctx is derived from the base context and has a ServerContextKey value.
	ConnContext func(context.Context, net.Conn) context.Context `alias:"conncontext" json:"-"` // Go 1.13
}

// serverStd 定义使用net/http启动http server。
type serverStd struct {
	*http.Server
	Print         func(...interface{}) `alias:"print"`
	Mutex         sync.Mutex
	localListener localListener
	Ports         []string
	Counter       int64
}

type serverStdMetadata struct {
	Health     bool     `json:"health" xml:"health"`
	Name       string   `json:"name" xml:"name"`
	Ports      []string `json:"ports" xml:"ports"`
	ErrorCount int64    `json:"error_count" xml:"error-count"`
}

// serverFcgi 定义fastcgi server
type serverFcgi struct {
	http.Handler
	sync.Mutex
	listeners []net.Listener
}

// ServerListenConfig 定义一个通用的端口监听配置,监听https仅支持单证书。
type ServerListenConfig struct {
	NewListen   func(string, string) (net.Listener, error) `alias:"newlisten" json:"-" description:"create listener func, default: net.Listen"`
	Addr        string                                     `alias:"addr" json:"addr" description:"Listen addr."`
	HTTPS       bool                                       `alias:"https" json:"https" description:"Is https."`
	HTTP2       bool                                       `alias:"http2" json:"http2" description:"Is http2."`
	Mutual      bool                                       `alias:"mutual" json:"mutual" description:"Is mutual tls."`
	Certfile    string                                     `alias:"certfile" json:"certfile" description:"Http server cert file."`
	Keyfile     string                                     `alias:"keyfile" json:"keyfile" description:"Http server key file."`
	Trustfile   string                                     `alias:"trustfile" json:"trustfile" description:"Http client ca file."`
	Certificate *x509.Certificate                          `alias:"certificate" json:"certificate" description:"https use tls certificate."`
}

// NewServerStd 创建一个标准server。
func NewServerStd(arg interface{}) Server {
	srv := &serverStd{
		Server: &http.Server{
			ReadTimeout:       60 * time.Second,
			ReadHeaderTimeout: 60 * time.Second,
			WriteTimeout:      60 * time.Second,
			IdleTimeout:       60 * time.Second,
			TLSNextProto:      nil,
		},
	}
	// 捕捉net/http.Server输出的error内容。
	srv.Server.ErrorLog = log.New(srv, "", 0)
	ConvertTo(arg, srv.Server)
	return srv
}

// Mount 方法获取ContextKeyApp.(Logger)用于输出http.Server错误日志。
// 获取ContextKeyApp.(http.Handler)作为http.Server的处理对象。
func (srv *serverStd) Mount(ctx context.Context) {
	srv.Print = NewPrintFunc(ctx.Value(ContextKeyApp).(Logger))
	srv.SetHandler(ctx.Value(ContextKeyApp).(http.Handler))
	// if go1.13+ set http.Server.BaseContext
	if Get(srv, "BaseContext") == nil {
		Set(srv.Server, "BaseContext", func(net.Listener) context.Context {
			return ctx
		})
	}
}

// Unmount 方法等待DefaulerServerShutdownWait(默认60s)优雅停机。
func (srv *serverStd) Unmount(ctx context.Context) {
	ctx, _ = context.WithTimeout(context.Background(), DefaulerServerShutdownWait)
	srv.Shutdown(ctx)
}

// SetHandler 方法设置server的http处理者。
func (srv *serverStd) SetHandler(h http.Handler) {
	srv.Mutex.Lock()
	defer srv.Mutex.Unlock()
	srv.Server.Handler = h
}

// Serve 方法阻塞监听请求。
func (srv *serverStd) Serve(ln net.Listener) error {
	srv.Mutex.Lock()
	srv.Ports = append(srv.Ports, ln.Addr().String())
	srv.Mutex.Unlock()
	err := srv.Server.Serve(ln)
	if err == http.ErrServerClosed {
		err = nil
	}
	return err
}

// ServeConn 方法出来一个连接，第一次请求初始化localListener。
func (srv *serverStd) ServeConn(conn net.Conn) {
	srv.Mutex.Lock()
	if srv.localListener.Ch == nil {
		srv.localListener.Ch = make(chan net.Conn)
		srv.Ports = append(srv.Ports, srv.localListener.Addr().String())
		go srv.Server.Serve(&srv.localListener)
	}
	srv.Mutex.Unlock()
	srv.localListener.Ch <- conn
}

// Metadata 方法返回serverStd元数据。
func (srv *serverStd) Metadata() interface{} {
	srv.Mutex.Lock()
	defer srv.Mutex.Unlock()
	return serverStdMetadata{
		Health:     true,
		Name:       "eudore.ServerStd",
		Ports:      srv.Ports,
		ErrorCount: atomic.LoadInt64(&srv.Counter),
	}
}

func (srv *serverStd) Write(p []byte) (n int, err error) {
	srv.Print(string(p))
	atomic.AddInt64(&srv.Counter, 1)
	return 0, nil
}

type localListener struct {
	Ch    chan net.Conn
	close bool
}

func (ln *localListener) Accept() (net.Conn, error) {
	for conn := range ln.Ch {
		if conn != nil {
			// panic(conn)
			return conn, nil
		}
	}
	return nil, errors.New("server close")
}

func (ln *localListener) Close() error {
	if !ln.close {
		close(ln.Ch)
		ln.close = true
	}
	return nil
}

func (ln *localListener) Addr() net.Addr {
	return &net.IPAddr{
		IP: net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 255, 255, 127, 0, 0, 1},
	}
}

// NewServerFcgi 函数创建一个fcgi server。
func NewServerFcgi() Server {
	return &serverFcgi{Handler: http.NotFoundHandler()}
}

// Mount 获取ContextKeyApp.(http.Handler)作为http.Server的处理对象。
func (srv *serverFcgi) Mount(ctx context.Context) {
	srv.SetHandler(ctx.Value(ContextKeyApp).(http.Handler))
}

// Unmount 方法等待DefaulerServerShutdownWait(默认60s)优雅停机。
func (srv *serverFcgi) Unmount(ctx context.Context) {
	ctx, _ = context.WithTimeout(context.Background(), DefaulerServerShutdownWait)
	srv.Shutdown(ctx)
}

// SetHandler 方法设置fcgi处理对象。
func (srv *serverFcgi) SetHandler(h http.Handler) {
	srv.Handler = h
}

// Serve 方法启动一个新的fcgi监听。
func (srv *serverFcgi) Serve(ln net.Listener) error {
	srv.Lock()
	srv.listeners = append(srv.listeners, ln)
	srv.Unlock()
	return fcgi.Serve(ln, srv.Handler)
}

// Shutdown 方法关闭fcgi关闭监听。
func (srv *serverFcgi) Shutdown(ctx context.Context) error {
	srv.Lock()
	defer srv.Unlock()
	var errs errormulit
	for _, ln := range srv.listeners {
		errs.HandleError(ln.Close())
	}
	return errs.Unwrap()
}

// Listen 方法使ServerListenConfig实现serverListener接口，用于使用对象创建监听。
func (slc *ServerListenConfig) Listen() (net.Listener, error) {
	if slc.NewListen == nil {
		slc.NewListen = net.Listen
	}
	// set default port
	if len(slc.Addr) == 0 {
		if slc.HTTPS {
			slc.Addr = ":80"
		} else {
			slc.Addr = ":443"
		}
	}
	if !slc.HTTPS {
		return slc.NewListen("tcp", slc.Addr)
	}

	// set tls
	config := &tls.Config{
		NextProtos:   []string{"http/1.1"},
		Certificates: make([]tls.Certificate, 1),
	}
	if slc.HTTP2 {
		config.NextProtos = []string{"h2"}
	}

	var err error
	config.Certificates[0], slc.Certificate, err = loadCertificate(slc.Certfile, slc.Keyfile)
	if err != nil {
		return nil, err
	}

	// set mutual tls
	if slc.Mutual {
		data, err := ioutil.ReadFile(slc.Trustfile)
		if err != nil {
			return nil, err
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(data)
		config.ClientCAs = pool
		config.ClientAuth = tls.RequireAndVerifyClientCert
	}

	ln, err := slc.NewListen("tcp", slc.Addr)
	if err != nil {
		return nil, err
	}
	return tls.NewListener(ln, config), nil
}

// loadCertificate 实现加载证书，如果证书配置文件为空，则自动创建一个私有证书。
func loadCertificate(cret, key string) (tls.Certificate, *x509.Certificate, error) {
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

		IsCA:        true,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		DNSNames:    []string{"localhost"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
	}
	pool := x509.NewCertPool()
	pool.AddCert(ca)

	priv, _ := rsa.GenerateMultiPrimeKey(rand.Reader, 2, 2048)
	caByte, err := x509.CreateCertificate(rand.Reader, ca, ca, &priv.PublicKey, priv)

	return tls.Certificate{
		Certificate: [][]byte{caByte},
		PrivateKey:  priv,
	}, ca, err
}
