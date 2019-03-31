/*
Server

用于启动http服务
*/
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
)

const (
	ServerStateInit 	ServerState		=	iota
	ServerStateRun	
	ServerStateClose
	ServerStateUnknown
	// 按顺序记录fork多端口fd对应的地址。
	EUDORE_GRACEFUL_ADDRS	= "EUDORE_GRACEFUL_ADDRS"
)

var (
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
	}

	ServerListenConfig struct {
		Addr		string		`set:"addr" description:"Listen addr."`
		Https		bool		`set:"https" description:"Is https, default use http2."`
		Mutual		bool		`set:"mutual" description:"Is mutual tls.`
		Certfile	string		`set:"certfile" description:"Http server cert file."`
		Keyfile		string		`set:"keyfile" description:"Http server key file."`
		TrustFile	string		`set:"trustfile" description:"Http client ca file."`
	}

	// 通用Server配置信息。
	//
	// 记录一个sever的组件名称、监听地址、是否https、是否http2、https证书、双向https信任证书、超时时间、请求处理对象。
	ServerGeneralConfig struct {
		Name			string			`set:"name"`
		ReadTimeout		time.Duration	`set:"readtimeout" description:"Http server read timeout."`
		WriteTimeout	time.Duration	`set:"writetimeout" description:"Http server write timeout."`
		Handler			interface{}		`set:"-" json:"-" description:"-"`
		Listeners		[]*ServerListenConfig `set:"listeners"`
	}
	ServerStd struct {
		// Servers 	[]*stdServerPort
		*http.Server
		Config		*ServerGeneralConfig	`set:"config"`
		Listener	net.Listener			`set:"listener"`
		Errfunc		ErrorFunc				`set:"errfunc"`
		mu			sync.Mutex				`set:"-"`
		wg			sync.WaitGroup			`set:"-"`
		state		ServerState				`set:"-"`
	}
	// stdServerPort struct {
	// 	Addr		string
	// 	State		ServerState
	// }
	// multi
	// ServerMulti配置，记录多server信息。
	ServerMultiConfig struct {
		Name			string
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


func NewServer(name string, arg interface{}) (Server, error) {
	name = ComponentPrefix(name, "server")
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

/*func (sc *ServerConfigGeneral) GetName() string {
	return sc.Name
}
*/

func NewServerStd(arg interface{}) (Server, error) {
	scg, ok := arg.(*ServerGeneralConfig)
	if !ok {
		scg = &ServerGeneralConfig{}
		_, err := ConvertStruct(scg, arg)
		if err != nil {
			return nil, fmt.Errorf("----: %v", err)
		}
	}
	// conv listen
	//
	return &ServerStd{
		Config:		scg,
		Errfunc:	DefaultErrorHandleFunc,
		state:		ServerStateInit,
	}, nil
}

func (srv *ServerStd) Start() error {
	// update server state
	srv.mu.Lock()
	if srv.state != ServerStateInit {
		return fmt.Errorf("server state exception")
	}
	srv.state = ServerStateRun
	srv.mu.Unlock()
	// set handler
	h, ok := srv.Config.Handler.(http.Handler)
	if !ok {
		return fmt.Errorf("server not set handle")
	}	
	// create server
	srv.Server =  &http.Server{
		Handler:	h,
		ErrorLog:	NewHttpError(srv.Errfunc).Logger(),
	}
	// start server
	errs := NewErrors()
	for _, listener := range srv.Config.Listeners {
		// get listen
		ln, err := listener.Listen()
		if err != nil {
			errs.HandleError(err)
			continue
		}
		srv.wg.Add(1)
		go func(ln net.Listener){
			err := srv.Server.Serve(ln)
			if err != http.ErrServerClosed && err != nil {
				errs.HandleError(err)
			}
			srv.wg.Done()
		}(ln)
	}
	// wait over
	srv.wg.Wait()
	return errs.GetError()
}

func (srv *ServerStd) Restart() error {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	err := StartNewProcess()
	if err == nil {
		srv.Shutdown(context.Background())
	}
	return err
}

func (srv *ServerStd) Close() (err error) {	
	srv.mu.Lock()
	defer srv.mu.Unlock()
	if srv.state == ServerStateRun {
		srv.state = ServerStateClose
		return srv.Server.Close()
	}
	return nil
}

func (srv *ServerStd) Shutdown(ctx context.Context) error {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, 5 * time.Second)
	defer cancel()
	srv.state = ServerStateClose
	return  srv.Server.Shutdown(ctx)
}

func (srv *ServerStd) Set(key string, val interface{}) error {
	switch v := val.(type) {
	case ErrorFunc:
		srv.Errfunc = v
	case func(http.ResponseWriter, *http.Request):
		srv.Config.Handler = http.HandlerFunc(v)
	case http.Handler:
		srv.Config.Handler = val
	case *ServerGeneralConfig:
		srv.Config = v
	case *ServerListenConfig:
		srv.Config.Listeners = append(srv.Config.Listeners, v)
	}
	return nil
}

func (srv *ServerStd) SetErrorFunc(fn ErrorFunc) {
	srv.Errfunc = fn
}


func (srv *ServerStd) SetHandler(i interface{}) error {
	if _, ok := i.(http.Handler);ok {
		srv.Config.Handler = i
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


func NewServerMulti(i interface{}) (Server, error) {
	// check args
	sc, ok := i.(*ServerMultiConfig)
	if !ok {
		sc = &ServerMultiConfig{}
		_, err := ConvertStruct(sc, i)
		if err != nil {
			return nil, fmt.Errorf("------- error: %v", err)
		}
	}
	s := &ServerMulti{ServerMultiConfig: sc,}
	s.Servers = make([]Server, len(sc.Configs))
	var err error
	// creation servers
	for i, c := range sc.Configs {
		name := ComponentGetName(c)
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
func (srv *ServerMulti) Restart() error {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	err := StartNewProcess()
	if err == nil {
		srv.Shutdown(context.Background())
	}
	return err
}
func (srv *ServerMulti) Close() error {
	for _, server := range srv.Servers {
		server.Close()
	}
	return nil
}
func (srv *ServerMulti) Shutdown(ctx context.Context) error {
	for _, server := range srv.Servers {
		server.Shutdown(ctx)
	}
	return nil
}

func (s *ServerMultiConfig) GetName() string {
	return ComponentServerMultiName	
}

func (s *ServerMultiConfig) Version() string {
	return ComponentServerMultiVersion
}



func (sgc *ServerGeneralConfig) GetName() string {
	return sgc.Name
}



func (slc *ServerListenConfig) Listen() (net.Listener, error) {
	// set default port
	if len(slc.Addr) == 0 {
		if slc.Https {
			slc.Addr = ":80"
		}else {
			slc.Addr = ":443"
		}
	}
	// get listen
	ln, err := GlobalListener.Listen(slc.Addr)
	if err != nil {
		return nil, err
	}
	if !slc.Https {
		return ln, nil
	}
	// set tls
	config := &tls.Config{
		NextProtos:		[]string{"http/1.1"},
		Certificates:	make([]tls.Certificate, 1),
	}
	config.Certificates[0], err = tls.LoadX509KeyPair(slc.Certfile, slc.Keyfile)
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
	return tls.NewListener(ln, config), nil
}


// start new process to handle HTTP Connection
func StartNewProcess() error {
	// get addrs and socket listen fds
	addrs, files := GlobalListener.AllListener()

	// set graceful restart env flag
	envs := []string{}
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, EUDORE_GRACEFUL_ADDRS) {
			envs = append(envs, value)
		}
	}
	envs = append(envs, fmt.Sprintf("%s=%s", EUDORE_GRACEFUL_ADDRS, strings.Join(addrs, ",")) ) 

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
