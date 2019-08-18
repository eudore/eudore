package eudore

// 实现端口监听和热重启习惯内容。

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"time"
)

type (
	// ServerListenConfig 定义一个通用的端口监听配置。
	ServerListenConfig struct {
		Addr      string `set:"addr" description:"Listen addr."`
		Https     bool   `set:"https" description:"Is https."`
		Http2     bool   `set:"http2" description:"Is http2."`
		Mutual    bool   `set:"mutual" description:"Is mutual tls."`
		Certfile  string `set:"certfile" description:"Http server cert file."`
		Keyfile   string `set:"keyfile" description:"Http server key file."`
		TrustFile string `set:"trustfile" description:"Http client ca file."`
	}
	// serverListener 定义获得net.Listener的接口。
	serverListener interface {
		Listen() (net.Listener, error)
	}
	// serverListenFiler 定义用于获得net.Listener的fd的接口。
	serverListenFiler interface {
		File() (*os.File, error)
	}
	// serverListenFile 定义实现serverListenFiler接口，用于封装tls Listener。
	serverListenFile struct {
		net.Listener
		source net.Listener
	}
)

var (
	typeListener reflect.Type = reflect.TypeOf((*serverListener)(nil)).Elem()
)

// ListenWithFD 创建一个地址监听，同时会从fd里面创建监听。
func ListenWithFD(addr string) (net.Listener, error) {
	addr = translateAddr(addr)
	for i, str := range strings.Split(os.Getenv(EUDORE_GRACEFUL_ADDRS), ",") {
		if str != addr {
			continue
		}
		file := os.NewFile(uintptr(i+3), "")
		return net.FileListener(file)
	}
	return newListener(addr)
}

// translateAddr 实现转换tcp地址，统一地址。
func translateAddr(addr string) string {
	if strings.HasPrefix(addr, "unix://") {
		return addr
	}
	if strings.HasPrefix(addr, "tcp://") {
		addr = addr[6:]
	}
	pos := strings.IndexByte(addr, ':')
	if pos == -1 {
		return addr
	}
	switch addr[:pos] {
	case "", "[::]", "0.0.0.0":
		return "tcp://[::]" + addr[pos:]
	case "127.0.0.1", "localhost":
		return "tcp://127.0.0.1" + addr[pos:]
	default:
		return addr
	}
}

// newListener 使用net.Listen创建一个监听。
func newListener(addr string) (net.Listener, error) {
	if strings.HasPrefix(addr, "unix://") {
		return net.Listen("unix", addr[7:])
	}
	if strings.HasPrefix(addr, "tcp://") {
		return net.Listen("tcp", addr[6:])
	}
	return net.Listen("tcp", addr)
}

// newServerListens 使用配置获得多个serverListener对象。
func newServerListens(i interface{}) ([]serverListener, error) {
	if i == nil {
		return nil, nil
	}
	iType := reflect.TypeOf(i)
	iValue := reflect.ValueOf(i)
	if iType.Kind() == reflect.Slice {
		var err error
		data := make([]serverListener, iValue.Len())
		for i := 0; i < iValue.Len(); i++ {
			data[i], err = getServerListen(iValue.Index(i))
		}
		return data, err
	}
	ln, err := getServerListen(iValue)
	return []serverListener{ln}, err
}

// getServerListen 实现转换一个对象成serverListener。
func getServerListen(i reflect.Value) (serverListener, error) {
	if i.Type().Implements(typeListener) {
		return i.Interface().(serverListener), nil
	}
	sl := &ServerListenConfig{}
	ConvertTo(i.Interface(), sl)
	if sl.Addr != "" {
		return sl, nil
	}
	return nil, errors.New("not convet to serverListener")
}

// Listen 方法使ServerListenConfig实现serverListener接口，用于使用对象创建监听。
func (slc *ServerListenConfig) Listen() (net.Listener, error) {
	// set default port
	if len(slc.Addr) == 0 {
		if slc.Https {
			slc.Addr = ":80"
		} else {
			slc.Addr = ":443"
		}
	}
	// get listen
	// ln, err := GlobalListener.Listen(slc.Addr)
	ln, err := ListenWithFD(slc.Addr)

	if err != nil {
		return nil, err
	}
	if !slc.Https {
		return ln, nil
	}
	// set tls
	config := &tls.Config{
		NextProtos:   []string{"http/1.1"},
		Certificates: make([]tls.Certificate, 1),
	}
	if slc.Http2 {
		config.NextProtos = []string{"h2"}
	}

	config.Certificates[0], err = loadCertificate(slc.Certfile, slc.Keyfile)
	if err != nil {
		return nil, err
	}

	// set mutual tls
	if slc.Mutual {
		config.ClientAuth = tls.RequireAndVerifyClientCert
		pool := x509.NewCertPool()
		data, err := ioutil.ReadFile(slc.TrustFile)
		if err != nil {
			return nil, err
		}
		pool.AppendCertsFromPEM(data)
		config.ClientCAs = pool
	}
	return &serverListenFile{Listener: tls.NewListener(ln, config), source: ln}, nil
}

// loadCertificate 实现加载证书，如果证书配置文件为空，则自动创建一个本地证书。
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

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}
	caByte, err := x509.CreateCertificate(rand.Reader, ca, ca, &priv.PublicKey, priv)

	return tls.Certificate{
		Certificate: [][]byte{caByte},
		PrivateKey:  priv,
	}, err
}

// File 方法使serverListenFile实现serverListenFiler接口，可以获得net.Listener的fd对象。
func (l *serverListenFile) File() (*os.File, error) {
	f, ok := l.source.(serverListenFiler)
	if ok {
		return f.File()
	}
	return nil, errors.New("Listener is not implements ServerListenFiler.")
}

// GetAllListener 函数获取多个net.Listener的监听地址和fd。
func GetAllListener(lns []net.Listener) ([]string, []*os.File) {
	var addrs = make([]string, 0, len(lns))
	var files = make([]*os.File, 0, len(lns))
	for _, ln := range lns {
		if ln == nil {
			continue
		}
		fd, err := ln.(serverListenFiler).File()
		if err == nil {
			addrs = append(addrs, fmt.Sprintf("%s://%s", ln.Addr().Network(), ln.Addr().String()))
			files = append(files, fd)
		}
	}
	return addrs, files
}

// StartNewProcess start new process to handle HTTP Connection。
func StartNewProcess(lns []net.Listener) error {
	// get addrs and socket listen fds
	addrs, files := GetAllListener(lns)

	// set graceful restart env flag
	envs := []string{}
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, EUDORE_GRACEFUL_ADDRS) {
			envs = append(envs, value)
		}
	}
	envs = append(envs, fmt.Sprintf("%s=%s", EUDORE_GRACEFUL_ADDRS, strings.Join(addrs, ",")))

	// fork new process
	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Env = envs
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = files
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to forkexec: %v", err)
	}
	return nil
}
