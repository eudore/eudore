package eudore

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"
)

// Server 定义启动http服务的对象。
type Server interface {
	SetHandler(http.Handler)
	Serve(net.Listener) error
	Shutdown(context.Context) error
}

// ServerStdConfig 定义ServerStd使用的配置
type ServerStdConfig struct {
	// ReadTimeout is the maximum duration for reading the entire request, including the body.
	//
	// Because ReadTimeout does not let Handlers make per-request decisions on each request body's acceptable deadline or upload rate,
	// most users will prefer to use ReadHeaderTimeout. It is valid to use them both.
	ReadTimeout time.Duration `alias:"readtimeout" description:"Http server read timeout."`

	// ReadHeaderTimeout is the amount of time allowed to read request headers.
	// The connection's read deadline is reset after reading the headers and the Handler can decide what is considered too slow for the body.
	ReadHeaderTimeout time.Duration `alias:"readheaderTimeout"` // Go 1.8

	// WriteTimeout is the maximum duration before timing out writes of the response.
	// It is reset whenever a new request's header is read.
	// Like ReadTimeout, it does not let Handlers make decisions on a per-request basis.
	WriteTimeout time.Duration `alias:"writetimeout" description:"Http server write timeout."`

	// IdleTimeout is the maximum amount of time to wait for the next request when keep-alives are enabled.
	// If IdleTimeout is zero, the value of ReadTimeout is used. If both are zero, ReadHeaderTimeout is used.
	IdleTimeout time.Duration `alias:"idleTimeout"` // Go 1.8

	// MaxHeaderBytes controls the maximum number of bytes the server will read parsing the request header's keys and values, including the request line.
	// It does not limit the size of the request body. If zero, DefaultMaxHeaderBytes is used.
	MaxHeaderBytes int `alias:"maxheaderbytes"`

	// BaseContext optionally specifies a function that returns the base context for incoming requests on this server.
	// The provided Listener is the specific Listener that's about to start accepting requests.
	// If BaseContext is nil, the default is context.Background(). If non-nil, it must return a non-nil context.
	BaseContext func(net.Listener) context.Context `alias:"basecontext"` // Go 1.13

	// ConnContext optionally specifies a function that modifies the context used for a new connection c.
	// The provided ctx is derived from the base context and has a ServerContextKey value.
	ConnContext func(context.Context, net.Conn) context.Context `alias:"conncontext"` // Go 1.13
}

// ServerStd 定义使用net/http启动http server。
type ServerStd struct {
	*http.Server
	Print func(...interface{}) `alias:"print"`
}

// netHTTPLog 实现一个函数处理log.Logger的内容，用于捕捉net/http.Server输出的error内容。
type netHTTPLog struct {
	print func(...interface{})
}

// ServerFcgi 定义fastcgi server
type ServerFcgi struct {
	http.Handler
	listeners []net.Listener
}

// ServerGrace 定义热重启服务。
type ServerGrace struct {
	Server
	listeners []net.Listener
}

// ServerListenConfig 定义一个通用的端口监听配置。
type ServerListenConfig struct {
	Addr      string `alias:"addr" description:"Listen addr."`
	HTTPS     bool   `alias:"https" description:"Is https."`
	HTTP2     bool   `alias:"http2" description:"Is http2."`
	Mutual    bool   `alias:"mutual" description:"Is mutual tls."`
	Certfile  string `alias:"certfile" description:"Http server cert file."`
	Keyfile   string `alias:"keyfile" description:"Http server key file."`
	TrustFile string `alias:"trustfile" description:"Http client ca file."`
}

// NewServerStd 创建一个标准server。
func NewServerStd(arg interface{}) Server {
	srv := &ServerStd{
		Server: &http.Server{
			ReadTimeout:  12 * time.Second,
			WriteTimeout: 4 * time.Second,
			IdleTimeout:  60 * time.Second,
			TLSNextProto: nil,
		},
	}
	ConvertTo(arg, srv.Server)
	return srv
}

// SetHandler 方法设置server的http处理者。
func (srv *ServerStd) SetHandler(h http.Handler) {
	srv.Server.Handler = h
}

// Set 方法允许Server设置输出函数和配置
func (srv *ServerStd) Set(key string, value interface{}) error {
	switch val := value.(type) {
	case func(...interface{}):
		srv.Print = val
		srv.Server.ErrorLog = newNetHTTPLogger(srv.Print)
	case ServerStdConfig, *ServerStdConfig:
		ConvertTo(value, srv.Server)
	default:
		cf := new(ServerStdConfig)
		Set(cf, key, value)
		ConvertTo(cf, srv.Server)
	}
	return nil
}

// newNetHTTPLog 实现将一个日志处理函数适配成log.Logger对象。
func newNetHTTPLogger(fn func(...interface{})) *log.Logger {
	e := &netHTTPLog{
		print: fn,
	}
	return log.New(e, "", 0)
}

func (e *netHTTPLog) Write(p []byte) (n int, err error) {
	e.print(string(p))
	return 0, nil
}

// NewServerFcgi 函数创建一个fcgi server。
func NewServerFcgi() Server {
	return &ServerFcgi{Handler: http.NotFoundHandler()}
}

// SetHandler 方法设置fcgi处理对象。
func (srv *ServerFcgi) SetHandler(h http.Handler) {
	srv.Handler = h
}

// Serve 方法启动一个新的fcgi监听。
func (srv *ServerFcgi) Serve(ln net.Listener) error {
	srv.listeners = append(srv.listeners, ln)
	return fcgi.Serve(ln, srv.Handler)
}

// Shutdown 方法关闭fcgi关闭监听。
func (srv *ServerFcgi) Shutdown(ctx context.Context) error {
	var errs muliterror
	for _, ln := range srv.listeners {
		err := ln.Close()
		errs.HandleError(err)
	}
	return errs.GetError()
}

// NewServerGrace 函数包装一个热重启服务。
func NewServerGrace(srv Server) Server {
	return &ServerGrace{Server: srv}
}

// Serve 方法记录Serve使用的net.Listener。
func (srv *ServerGrace) Serve(ln net.Listener) error {
	srv.listeners = append(srv.listeners, ln)
	return srv.Server.Serve(ln)
}

// Shutdown 方法关闭服务，如果context.Context包含AppServerGrace则使用热重启。
func (srv *ServerGrace) Shutdown(ctx context.Context) error {
	val := ctx.Value(ServerGraceContextKey)
	if val != nil {
		err := startNewProcess(srv.listeners)
		if err != nil {
			return err
		}
	}
	return srv.Server.Shutdown(ctx)
}

// Set 方法传递Set数据。
func (srv *ServerGrace) Set(key string, value interface{}) error {
	return Set(srv.Server, key, value)
}

// startNewProcess 函数启动一个新的服务。
func startNewProcess(lns []net.Listener) error {
	addrs, files := getAllListener(lns)
	envs := []string{}
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, EnvEudoreGracefulAddrs) {
			envs = append(envs, value)
		}
	}
	envs = append(envs, fmt.Sprintf("%s=%s", EnvEudoreGracefulAddrs, strings.Join(addrs, ",")))

	// fork new process
	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Env = envs
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = files
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf(ErrFormatStartNewProcessError, err)
	}
	return nil
}

// getAllListener 函数获取多个net.Listener的监听地址和fd。
func getAllListener(lns []net.Listener) ([]string, []*os.File) {
	var addrs = make([]string, 0, len(lns))
	var files = make([]*os.File, 0, len(lns))
	for _, ln := range lns {
		fd, err := getListenerFile(ln)
		if fd != nil && err == nil {
			addrs = append(addrs, fmt.Sprintf("%s://%s", ln.Addr().Network(), ln.Addr().String()))
			files = append(files, fd)
		}
	}
	return addrs, files
}

func getListenerFile(ln net.Listener) (*os.File, error) {
	if ln == nil {
		return nil, nil
	}
	lnf, ok := ln.(interface{ File() (*os.File, error) })
	if ok {
		return lnf.File()
	}

	iValue := reflect.ValueOf(ln)
	if iValue.Kind() == reflect.Ptr {
		iValue = iValue.Elem()
	}
	iType := iValue.Type()
	for i := 0; i < iType.NumField(); i++ {
		if iType.Field(i).Type == typeNetListener {
			file, err := getListenerFile(iValue.Field(i).Interface().(net.Listener))
			if file != nil && err == nil {
				return file, err
			}
		}
	}
	return nil, nil
}

// ListenWithFD 创建一个地址监听，同时会从fd里面创建监听。
func ListenWithFD(network, address string) (net.Listener, error) {
	var port string
	pos := strings.IndexByte(address, ':')
	if pos != -1 {
		address, port = address[:pos], address[pos:]
	}
	switch address {
	case "", "[::]", "0.0.0.0":
		address = "[::]" + port
	case "127.0.0.1", "localhost":
		address = "127.0.0.1" + port
	}
	if network == "" {
		network = "tcp"
	}
	return listenWithFD(network, address)
}

func listenWithFD(network, address string) (net.Listener, error) {
	proaddr := fmt.Sprintf("%s://%s", network, address)
	for i, str := range strings.Split(os.Getenv(EnvEudoreGracefulAddrs), ",") {
		if str != proaddr {
			continue
		}
		file := os.NewFile(uintptr(i+3), "")
		return net.FileListener(file)
	}
	return net.Listen(network, address)
}

// Listen 方法使ServerListenConfig实现serverListener接口，用于使用对象创建监听。
func (slc *ServerListenConfig) Listen() (net.Listener, error) {
	// set default port
	if len(slc.Addr) == 0 {
		if slc.HTTPS {
			slc.Addr = ":80"
		} else {
			slc.Addr = ":443"
		}
	}
	if !slc.HTTPS {
		return ListenWithFD("", slc.Addr)
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
	config.Certificates[0], err = loadCertificate(slc.Certfile, slc.Keyfile)
	if err != nil {
		return nil, err
	}

	// set mutual tls
	if slc.Mutual {
		data, err := ioutil.ReadFile(slc.TrustFile)
		if err != nil {
			return nil, err
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(data)
		config.ClientCAs = pool
		config.ClientAuth = tls.RequireAndVerifyClientCert
	}

	ln, err := ListenWithFD("", slc.Addr)
	if err != nil {
		return nil, err
	}
	return tls.NewListener(ln, config), nil
}

// loadCertificate 实现加载证书，如果证书配置文件为空，则自动创建一个私有证书。
func loadCertificate(cret, key string) (tls.Certificate, error) {
	if cret != "" && key != "" {
		return tls.LoadX509KeyPair(cret, key)
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
	}
	pool := x509.NewCertPool()
	pool.AddCert(ca)

	priv, _ := rsa.GenerateMultiPrimeKey(rand.Reader, 2, 2048)
	caByte, err := x509.CreateCertificate(rand.Reader, ca, ca, &priv.PublicKey, priv)

	return tls.Certificate{
		Certificate: [][]byte{caByte},
		PrivateKey:  priv,
	}, err
}
