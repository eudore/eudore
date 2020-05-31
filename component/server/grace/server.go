package grace

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"strings"
)

type contextKey struct {
	name string
}

var (
	// ServerGraceContextKey 定义从context.Value中获取是否进行热重启的key。
	ServerGraceContextKey = &contextKey{"server-grace"}
	// EnvEudoreGracefulAddrs 按顺序记录fork多端口fd对应的地址。
	EnvEudoreGracefulAddrs = "EnvEudoreGracefulAddrs"
	// ErrFormatStartNewProcessError 在StartNewProcess函数fork启动新进程错误。
	ErrFormatStartNewProcessError = "StartNewProcess failed to forkexec error: %v"
	typeNetListener               = reflect.TypeOf((*net.Listener)(nil)).Elem()
)

// Server 定义Server接口。
type Server interface {
	SetHandler(http.Handler)
	Serve(net.Listener) error
	Shutdown(context.Context) error
}

// ServerGrace 定义热重启服务。
type ServerGrace struct {
	Server
	listeners []net.Listener
}

// ServerStd 定义http.Server实现Server接口。
type ServerStd struct {
	*http.Server
}

// NewServerGrace 函数包装一个热重启服务，支持*http.Server和eudore.Server。
func NewServerGrace(i interface{}) Server {
	srv, ok := i.(Server)
	if ok {
		return &ServerGrace{Server: srv}
	}
	srv2, ok := i.(*http.Server)
	if ok {
		return &ServerGrace{Server: ServerStd{srv2}}
	}
	return nil
}

// SetHandler 方法设置http.Server的http.Handler对象。
func (srv ServerStd) SetHandler(h http.Handler) {
	srv.Handler = h
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

// Listen 创建一个地址监听，同时会从fd里面创建监听。
func Listen(network, address string) (net.Listener, error) {
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
