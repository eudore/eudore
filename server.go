package eudore

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/http/fcgi"
	"strconv"
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
	ReadTimeout TimeDuration `alias:"readtimeout" description:"Http server read timeout."`

	// ReadHeaderTimeout is the amount of time allowed to read request headers.
	// The connection's read deadline is reset after reading the headers and the Handler can decide what is considered too slow for the body.
	ReadHeaderTimeout TimeDuration `alias:"readheadertimeout"` // Go 1.8

	// WriteTimeout is the maximum duration before timing out writes of the response.
	// It is reset whenever a new request's header is read.
	// Like ReadTimeout, it does not let Handlers make decisions on a per-request basis.
	WriteTimeout TimeDuration `alias:"writetimeout" description:"Http server write timeout."`

	// IdleTimeout is the maximum amount of time to wait for the next request when keep-alives are enabled.
	// If IdleTimeout is zero, the value of ReadTimeout is used. If both are zero, ReadHeaderTimeout is used.
	IdleTimeout TimeDuration `alias:"idletimeout"` // Go 1.8

	// MaxHeaderBytes controls the maximum number of bytes the server will read parsing the request header's keys and values, including the request line.
	// It does not limit the size of the request body. If zero, DefaultMaxHeaderBytes is used.
	MaxHeaderBytes int `alias:"maxheaderbytes"`

	// BaseContext optionally specifies a function that returns the base context for incoming requests on this server.
	// The provided Listener is the specific Listener that's about to start accepting requests.
	// If BaseContext is nil, the default is context.Background(). If non-nil, it must return a non-nil context.
	BaseContext func(net.Listener) context.Context `alias:"basecontext" json:"-"` // Go 1.13

	// ConnContext optionally specifies a function that modifies the context used for a new connection c.
	// The provided ctx is derived from the base context and has a ServerContextKey value.
	ConnContext func(context.Context, net.Conn) context.Context `alias:"conncontext" json:"-"` // Go 1.13
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
		return net.Listen("tcp", slc.Addr)
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

	ln, err := net.Listen("tcp", slc.Addr)
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

// TimeDuration 定义time.Duration类型处理json
type TimeDuration time.Duration

// MarshalJSON 方法实现json序列化输出。
func (d TimeDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// UnmarshalJSON 方法实现解析json格式时间。
func (d *TimeDuration) UnmarshalJSON(b []byte) error {
	str := string(b)
	if str != "" && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}
	// parse int64
	val, err := strconv.ParseInt(str, 10, 64)
	if err == nil {
		*d = TimeDuration(val)
		return nil
	}
	// parse string
	t, err := time.ParseDuration(str)
	if err == nil {
		*d = TimeDuration(t)
		return nil
	}
	return fmt.Errorf("invalid duration type %T, value: '%s'", b, b)

}
