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

// Server 定义启动http服务的对象。
type Server interface {
	SetHandler(http.Handler)
	Serve(net.Listener) error
	Shutdown(context.Context) error
}

// ServerConfig 定义serverStd使用的配置。
type ServerConfig struct {
	// set default ServerHandler
	Handler http.Handler `alias:"handler" json:"-" xml:"-" yaml:"-"`

	// ReadTimeout is the maximum duration for reading the entire request, including the body.
	//
	// Because ReadTimeout does not let Handlers make per-request decisions on each request body's acceptable deadline or upload rate,
	// most users will prefer to use ReadHeaderTimeout. It is valid to use them both.
	ReadTimeout TimeDuration `alias:"readtimeout" json:"readtimeout" xml:"readtimeout" yaml:"readtimeout" description:"Http server read timeout."`

	// ReadHeaderTimeout is the amount of time allowed to read request headers.
	// The connection's read deadline is reset after reading the headers and the Handler can decide what is considered too slow for the body.
	ReadHeaderTimeout TimeDuration `alias:"readheadertimeout" json:"readheadertimeout" xml:"readheadertimeout" yaml:"readheadertimeout" description:"Http server read header timeout."`
	// WriteTimeout is the maximum duration before timing out writes of the response.
	// It is reset whenever a new request's header is read.
	// Like ReadTimeout, it does not let Handlers make decisions on a per-request basis.
	WriteTimeout TimeDuration `alias:"writetimeout" json:"writetimeout" xml:"writetimeout" yaml:"writetimeout" description:"Http server write timeout."`

	// IdleTimeout is the maximum amount of time to wait for the next request when keep-alives are enabled.
	// If IdleTimeout is zero, the value of ReadTimeout is used. If both are zero, ReadHeaderTimeout is used.
	IdleTimeout TimeDuration `alias:"idletimeout" json:"idletimeout" xml:"idletimeout" yaml:"idletimeout" description:"Http server idle timeout."`

	// MaxHeaderBytes controls the maximum number of bytes the server will read parsing the request header's keys and values, including the request line.
	// It does not limit the size of the request body. If zero, DefaultMaxHeaderBytes is used.
	MaxHeaderBytes int `alias:"maxheaderbytes" json:"maxheaderbytes" xml:"maxheaderbytes" yaml:"maxheaderbytes" description:"Http server max header size."`

	// ErrorLog specifies an optional logger for errors accepting
	// connections, unexpected behavior from handlers, and
	// underlying FileSystem errors.
	// If nil, logging is done via the log package's standard logger.
	ErrorLog *log.Logger `alias:"errorlog" json:"-" xml:"-" yaml:"-"`

	// BaseContext optionally specifies a function that returns the base context for incoming requests on this server.
	// The provided Listener is the specific Listener that's about to start accepting requests.
	// If BaseContext is nil, the default is context.Background(). If non-nil, it must return a non-nil context.
	BaseContext func(net.Listener) context.Context `alias:"basecontext" json:"-" xml:"-" yaml:"-"`

	// ConnContext optionally specifies a function that modifies the context used for a new connection c.
	// The provided ctx is derived from the base context and has a ServerContextKey value.
	ConnContext func(context.Context, net.Conn) context.Context `alias:"conncontext" json:"-" xml:"-" yaml:"-"`
}

// serverStd 定义使用net/http启动http server。
type serverStd struct {
	*http.Server
	Mutex         sync.Mutex
	Logger        Logger
	localListener localListener
	Ports         []string
	Counter       int64
}

type MetadataServer struct {
	Health     bool     `alias:"health" json:"health" xml:"health" yaml:"health"`
	Name       string   `alias:"name" json:"name" xml:"name" yaml:"name"`
	Ports      []string `alias:"ports" json:"ports" xml:"ports" yaml:"ports"`
	ErrorCount int64    `alias:"errorcount" json:"errorcount" xml:"errorcount" yaml:"errorcount"`
}

// serverFcgi 定义fastcgi server。
type serverFcgi struct {
	http.Handler
	sync.Mutex
	listeners []net.Listener
}

// ServerListenConfig 定义一个通用的端口监听配置,监听https仅支持单证书。
type ServerListenConfig struct {
	Addr        string            `alias:"addr" json:"addr" xml:"addr" yaml:"addr" description:"Listen addr."`
	HTTPS       bool              `alias:"https" json:"https" xml:"https" yaml:"https" description:"Is https."`
	HTTP2       bool              `alias:"http2" json:"http2" xml:"http2" yaml:"http2" description:"Is http2."`
	Mutual      bool              `alias:"mutual" json:"mutual" xml:"mutual" yaml:"mutual" description:"Is mutual tls."`
	Certfile    string            `alias:"certfile" json:"certfile" xml:"certfile" yaml:"certfile" description:"Http server cert file."`
	Keyfile     string            `alias:"keyfile" json:"keyfile" xml:"keyfile" yaml:"keyfile" description:"Http server key file."`
	Trustfile   string            `alias:"trustfile" json:"trustfile" xml:"trustfile" yaml:"trustfile" description:"Http client ca file."`
	Certificate *x509.Certificate `alias:"certificate" json:"certificate" xml:"certificate" yaml:"certificate" description:"https use tls certificate."`
}

// NewServer 创建一个标准server。
func NewServer(config *ServerConfig) Server {
	if config == nil {
		config = &ServerConfig{}
	}
	srv := &serverStd{
		Server: &http.Server{
			Handler:           config.Handler,
			ReadTimeout:       GetAnyDefault(time.Duration(config.ReadTimeout), DefaultServerReadTimeout),
			ReadHeaderTimeout: GetAnyDefault(time.Duration(config.ReadHeaderTimeout), DefaultServerReadHeaderTimeout),
			WriteTimeout:      GetAnyDefault(time.Duration(config.WriteTimeout), DefaultServerWriteTimeout),
			IdleTimeout:       GetAnyDefault(time.Duration(config.IdleTimeout), DefaultServerIdleTimeout),
			MaxHeaderBytes:    config.MaxHeaderBytes,
			ErrorLog:          config.ErrorLog,
			BaseContext:       config.BaseContext,
			ConnContext:       config.ConnContext,
		},
		Logger: DefaultLoggerNull,
	}
	// 捕捉net/http.Server输出的error内容。
	if srv.ErrorLog == nil {
		srv.ErrorLog = log.New(srv, "", 0)
	}
	return srv
}

// Mount 方法获取ContextKeyApp.(Logger)用于输出http.Server错误日志。
// 获取ContextKeyApp.(http.Handler)作为http.Server的处理对象。
func (srv *serverStd) Mount(ctx context.Context) {
	if srv.Handler == nil {
		srv.SetHandler(ctx.Value(ContextKeyApp).(http.Handler))
	}
	if srv.BaseContext == nil {
		srv.BaseContext = func(net.Listener) context.Context {
			return ctx
		}
	}
	log, ok := ctx.Value(ContextKeyApp).(Logger)
	if ok {
		srv.Logger = log
	}
}

// Unmount 方法等待DefaulerServerShutdownWait(默认60s)优雅停机。
func (srv *serverStd) Unmount(context.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultServerShutdownWait)
	defer cancel()
	_ = srv.Shutdown(ctx)
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
	if errors.Is(err, http.ErrServerClosed) {
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
		go func() {
			_ = srv.Server.Serve(&srv.localListener)
		}()
	}
	srv.Mutex.Unlock()
	srv.localListener.Ch <- conn
}

// Metadata 方法返回serverStd元数据。
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

func (srv *serverStd) Write(p []byte) (n int, err error) {
	atomic.AddInt64(&srv.Counter, 1)
	log := srv.Logger.WithField(ParamDepth, "disable").WithField(ParamCaller, "*serverStd.ErrorLog.Write")
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
			lines = append(lines, strings.TrimPrefix(strs[i+1], "\t")+" "+strs[i])
		}
		log.WithField("stack", lines).Errorf("%s %s", strs[0], strs[1][:len(strs[1])-1])
	} else {
		log.Errorf(strs[0])
	}
	return 0, nil
}

type localListener struct {
	Ch    chan net.Conn
	close bool
}

func (ln *localListener) Accept() (net.Conn, error) {
	for conn := range ln.Ch {
		if conn != nil {
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
func (srv *serverFcgi) Unmount(context.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultServerShutdownWait)
	defer cancel()
	_ = srv.Shutdown(ctx)
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
func (srv *serverFcgi) Shutdown(context.Context) error {
	srv.Lock()
	defer srv.Unlock()
	var errs mulitError
	for _, ln := range srv.listeners {
		errs.HandleError(ln.Close())
	}
	return errs.Unwrap()
}

// Listen 方法使ServerListenConfig实现serverListener接口，用于使用对象创建监听。
func (slc *ServerListenConfig) Listen() (net.Listener, error) {
	// set default port
	if slc.Addr == "" {
		if slc.HTTPS {
			slc.Addr = ":80"
		} else {
			slc.Addr = ":443"
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

	var err error
	config.Certificates[0], slc.Certificate, err = loadCertificate(slc.Certfile, slc.Keyfile)
	if err != nil {
		return nil, err
	}

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
