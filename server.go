package eudore

import (
	"os"
	"os/exec"
	"fmt"
	"net"
	"sync"
	"time"
	"errors"
	"io/ioutil"
	"net/http"
	"context"
	"strings"
	"crypto/tls"
	"crypto/x509"
	"golang.org/x/net/http2"
)

const (
	ServerStateInit 	ServerState		=	iota
	ServerStateRun	
	ServerStateClose
	ServerStateUnknown
	// 按顺序记录fork多端口fd对应的地址。
	GRACEFUL_ENVIRON_ADDRS	= "EUDORE_GRACEFUL_ADDRS"
)

var (
	graceServerAddrs	[]string
	graceOutput = fmt.Println
	ErrArgsNotArray		=	errors.New("args is noy array.")
	ErrArgsNotServerConfig	=	errors.New("args is noy server config.")
)
type (
	ServerState = int
	Server interface {
		Component
		Start() error
		Restart() error
		Close() error
		Shutdown(ctx context.Context) error
		GetState() ServerState
		SetErrorFunc(ErrorFunc)
		SetHandler(interface{}) error
	}

	// 通用路由配置信息。
	//
	// 记录一个sever的组件名称、监听地址、是否https、是否http2、https证书、双向https信任证书、超时时间、请求处理对象。
	ServerConfigGeneral struct {
		Name		string
		Addr		string		`description:"Listen addr."`
		Https		bool		`description:"Is https, default use http2."`
		Http2		bool		`description:"Is http2.`
		Mutual		bool		`description:"Is mutual tls.`
		Certfile	string		`description:"Http server cert file."`
		Keyfile		string		`description:"Http server key file."`
		TrustFile	string		`description:"Http client ca file."`
		ReadTimeout		time.Duration	`description:"Http server read timeout."`
		WriteTimeout	time.Duration	`description:"Http server write timeout."`
		// ServerType	string		`description:"server instance method."`
		// Handler		http.Handler`json:"-" description:"-"`	
		Handler		interface{}	`json:"-" description:"-"`
	}
	ServerStd struct {
		Servers 	[]*stdServerPort
		port		*stdServerPort
		errfunc		ErrorFunc
		state		ServerState
	}
	stdServerPort struct {
		http.Server
		*ServerConfigGeneral
		Addr		string
		Listener	net.Listener
		State		ServerState
	}
	// multi
	// ServerMulti配置，记录多server信息。
	ServerMultiConfig struct {
		Configs		[]interface{}
	}
	// 用于启动多个server组合。
	ServerMulti struct {
		*ServerMultiConfig
		Servers		[]Server
		mu			sync.Mutex
		wg			sync.WaitGroup
	}
)

func init() {
	addrs := os.Getenv(GRACEFUL_ENVIRON_ADDRS)
	if addrs != "" {
		graceServerAddrs = strings.Split(addrs,",")
		fmt.Println("Addrs", addrs)
	}
}

func NewServer(name string, arg interface{}) (Server, error) {
	name = AddComponetPre(name, "server")
	c, err := NewComponent(name, arg)
	if err != nil {
		return nil, err
	}
	s, ok := c.(Server)
	if ok {
		return s, nil
	}
	return nil, fmt.Errorf("Component %s cannot be converted to Server type", name)
}

func (sc *ServerConfigGeneral) GetName() string {
	return sc.Name
}


func NewServerStd(arg interface{}) (Server, error) {
	c, ok := arg.(*ServerConfigGeneral)
	if !ok {
		c = &ServerConfigGeneral{}
		err := MapToStruct(arg, c)
		if err != nil {
			return nil, fmt.Errorf("----: %v", err)
		}
	}
	port, err := newStdServer(c)
	if err != nil {
		return nil, err
	}
	return &ServerStd{
		port:		port,
		errfunc:	ErrorDefaultHandleFunc,
		state:		ServerStateInit,
	}, nil
}

func (s *ServerStd) Start() error {
	s.port.Server.ErrorLog = NewHttpError(s.errfunc).Logger()
	s.state = ServerStateRun
	return s.port.Run()


	/*if len(s.args) == 0 {
		return fmt.Errorf("No corresponding server information is registered.")
	}
	errs := NewErrors()
	// add group wait goroutine num
	s.wg.Add(len(s.args))
	for _, c := range s.args {
		// create a server instace
		var conf = c
		server, err := newStdServer(conf)
		if err != nil {
			errs.HandleError(err, s.Shutdown(context.Background()))
			break
		}
		go func() {
			// set state
			l := NewHttpError(s.errfunc)
			server.State = ServerStateRun
			server.Server.ErrorLog = l.Logger()
			err := server.Run()
			server.State = ServerStateClose
			if err != http.ErrServerClosed && err != nil {
				errs.HandleError(err, s.Shutdown(context.Background()))
			}
			s.wg.Done()			
		}()
		s.Servers = append(s.Servers, server)
	}
	s.wg.Wait()
	return errs.GetError()*/
}

func (s *ServerStd) Restart() error {
	pid, err := s.startNewProcess();
	if  err != nil {
		graceOutput("start new process failed: ", err,", continue serving.", err)
	} else {
		graceOutput("start new process successed, the new pid is ", pid)
		s.Shutdown(context.Background())
	}
	return nil
}

func (s *ServerStd) Close() (err error) {	
/*	s.mu.Lock()
	defer s.mu.Unlock()
	for _, s := range s.Servers {
		if s.State == ServerStateRun {
			s.State = ServerStateClose
			err = s.Close()	
			if err != nil {
				return
			}
		}
	}*/
	if s.state == ServerStateRun {
		s.state = ServerStateClose
		return s.port.Close()
	}
	return nil
}

func (s *ServerStd) Shutdown(ctx context.Context) error {
	// s.mu.Lock()
	// defer s.mu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 5 * time.Second)
	defer cancel()
	// for _, server := range s.Servers {
	// 	if server.State == ServerStateRun {
	// 		server.State = ServerStateClose
	// 		server.Shutdown(ctx)
	// 	}
	// }
	s.state = ServerStateClose
	return  s.port.Shutdown(ctx)
}




func (s *ServerStd) GetState() ServerState {
	return s.state
}
func (s *ServerStd) SetErrorFunc(fn ErrorFunc) {
	s.errfunc = fn
}


func (s *ServerStd) SetHandler(i interface{}) error {
	if h, ok := i.(http.Handler);ok {
		s.port.Server.Handler = h
		return nil
	}
	return fmt.Errorf("Server config Handler object, not convert net.http.Handler type.")
}

func (s *ServerStd) GetName() string {
	return ComponentServerStdName
}

func (s *ServerStd) Version() string {
	return ComponentServerStdVersion
}



// start new process to handle HTTP Connection
func (s *ServerStd) startNewProcess() (uintptr, error) {
	// get addrs and socket listen fds
	var addrs = make([]string, 0, len(s.Servers))
	var files = make([]*os.File, 0, len(s.Servers))
	for _, s := range s.Servers {
		fd, err := s.Listener.(*net.TCPListener).File()
		if err != nil {
			return 0, fmt.Errorf("failed to get socket file descriptor: %v", err)
		}
		addrs = append(addrs, s.Addr)
		files = append(files, fd)
	}

	// set graceful restart env flag
	envs := []string{}
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, GRACEFUL_ENVIRON_ADDRS) {
			envs = append(envs, value)
		}
	}
	envs = append(envs, fmt.Sprintf("%s=%s", GRACEFUL_ENVIRON_ADDRS, strings.Join(addrs, ",")) ) 

	// fork new process
	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Env = envs
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.ExtraFiles = files
	err := cmd.Start()
	if err != nil {
		return 0, fmt.Errorf("failed to forkexec: %v", err)
	}
	return uintptr(cmd.Process.Pid), nil
}



func newStdServer(arg interface{}) (*stdServerPort, error) {
	s := &stdServerPort{
		Server: http.Server{},
	}
	if arg != nil {
		c, ok := arg.(*ServerConfigGeneral)
		if !ok {
			return nil, ErrArgsNotServerConfig
		}
		s.ServerConfigGeneral = c
	}
	if h, ok := s.ServerConfigGeneral.Handler.(http.Handler);ok {
		s.Server.Handler = h
	}else {
		// return nil, fmt.Errorf("Server config Handler object, not convert net.http.Handler type.")
	}
	l, err :=  GetNetListener(s.Addr)
	if err != nil {
		s.Listener = l
	}
	s.Addr = s.ServerConfigGeneral.Addr
	s.Server.Addr = s.ServerConfigGeneral.Addr
	s.Server.ReadTimeout = s.ServerConfigGeneral.ReadTimeout
	s.Server.WriteTimeout = s.ServerConfigGeneral.WriteTimeout
	return s, nil
}

func GetNetListener(addr string) (ln net.Listener, err error) {
	// find addr fd
	var fd uint 
	for i, v := range graceServerAddrs {
		if addr == v {
			fd = uint(i + 3)
			break
		}
	}
	if fd == 0 {
		return nil, fmt.Errorf("The service address %s did not find the corresponding fd", addr)
	}
	// use old net socket
	file := os.NewFile(uintptr(fd), "")
	ln, err = net.FileListener(file)
	if err != nil {
		err = fmt.Errorf("net.FileListener error: %v", err)
	}
	return	
}

func (s *stdServerPort) Run() error {
	conf := s.ServerConfigGeneral
	if !conf.Https {
		return s.ListenAndServe()
	} else if conf.Mutual {
		return s.ListenAndServeMutualTLS(conf.Certfile, conf.Keyfile, conf.TrustFile)
	} else {
		return s.ListenAndServeTLS(conf.Certfile, conf.Keyfile)
	}
	return nil
}

func (s *stdServerPort) ListenAndServe() error {
	if s.Listener == nil {
		if s.Addr == "" {
			s.Addr = ":http"
		}
		ln, err := net.Listen("tcp", s.Addr)
		if err != nil {
			return err
		}
		s.Listener = ln
	}
	return s.Server.Serve(s.Listener)
}

func (s *stdServerPort) ListenAndServeTLS(certFile, keyFile string) error {
	// 初始化连接
	if s.Listener == nil {
		if s.Addr == "" {
			s.Addr = ":https"
		}
		ln, err := net.Listen("tcp", s.Addr)
		if err != nil {
			return err
		}
		s.Listener = ln
	}
	// 配置http2，可选
	if s.Http2 || strings.Contains(os.Getenv("GODEBUG"), "http2server=0") {
		http2.ConfigureServer(&s.Server, &http2.Server{})
	}
	 // 配置https
	config := &tls.Config{}
	if s.TLSConfig != nil {
		*config = *s.TLSConfig
	}
	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}
	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	// 启动tls连接
	return s.Server.Serve(tls.NewListener(s.Listener, config))
}

func (s *stdServerPort) ListenAndServeMutualTLS(certFile, keyFile, trustFile string) error {// 初始化连接
	if s.Listener == nil {
		if s.Addr == "" {
			s.Addr = ":https"
		}
		ln, err := net.Listen("tcp", s.Addr)
		if err != nil {
			return err
		}
		s.Listener = ln
	}
	// 配置http2，可选
	if s.Http2 || strings.Contains(os.Getenv("GODEBUG"), "http2server=0") {
		http2.ConfigureServer(&s.Server, &http2.Server{})
	}
	 // 配置https
	config := &tls.Config{}
	if s.TLSConfig != nil {
		*config = *s.TLSConfig
	}
	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}
	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return err
	}
	// 配置双向https
	s.TLSConfig.ClientAuth = tls.RequireAndVerifyClientCert
	pool := x509.NewCertPool()
	data, err := ioutil.ReadFile(trustFile)
	if err != nil {
		return err
	}
	pool.AppendCertsFromPEM(data)
	s.TLSConfig.ClientCAs = pool
	// 启动tls连接
	return s.Server.Serve(tls.NewListener(s.Listener, config))
}

func NewServerMulti(i interface{}) (Server, error) {
	// check args
	sc, ok := i.(*ServerMultiConfig)
	if !ok {
		sc = &ServerMultiConfig{}
		err := MapToStruct(i, sc)
		if err != nil {
			return nil, fmt.Errorf("------- error: %v", err)
		}
	}
	s := &ServerMulti{ServerMultiConfig: sc,}
	s.Servers = make([]Server, len(sc.Configs))
	var err error
	// creation servers
	for i, c := range sc.Configs {
		name := GetComponetName(c)
		if len(name) == 0 {
			return nil, fmt.Errorf("ServerMulti %dth creation parameter could not get the corresponding component name", i)
		}
		s.Servers[i], err = NewServer(name, c)
		if err != nil {
			return nil, fmt.Errorf("ServerMulti %dth creation Error: %v", i, err)
		}
	}
	return s, nil
}

func (s *ServerMulti) Start() (err error) {
	// startup all server
	errs := NewErrors()
	s.wg.Add(len(s.Servers))
	for _, server := range s.Servers {
		go func(server Server) {
			err := server.Start()
			if err != http.ErrServerClosed && err != nil {
				errs.HandleError(err)
			}
			s.wg.Done()
		}(server)
	}
	s.wg.Wait()
	return errs.GetError()
}
func (s *ServerMulti) Restart() error {
	return nil
}
func (s *ServerMulti) Close() error {
	return nil
}
func (s *ServerMulti) Shutdown(ctx context.Context) error {
	return nil
}

func (s *ServerMulti) GetState() ServerState {
	return ServerStateUnknown
}

func (s *ServerMulti) SetErrorFunc(fn ErrorFunc) {
	for _, server := range s.Servers {
		server.SetErrorFunc(fn)
	}
}

func (s *ServerMulti) SetHandler(i interface{}) error {
	errs := NewErrors()
	for _, server := range s.Servers {
		errs.HandleError(server.SetHandler(i))
	}
	return errs.GetError()
}

func (s *ServerMultiConfig) GetName() string {
	return ComponentServerMultiName	
}

func (s *ServerMultiConfig) Version() string {
	return ComponentServerMultiVersion
}
