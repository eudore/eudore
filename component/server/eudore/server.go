package eudore

import (
	"net"
	"sync"
	"time"
	"context"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/protocol"
	"github.com/eudore/eudore/protocol/server"
	"github.com/eudore/eudore/protocol/http"
	"github.com/eudore/eudore/protocol/http2"
	"github.com/eudore/eudore/protocol/fastcgi"
)

type (
	HttpConfig = eudore.ServerListenConfig
	FastcgiConfig struct {
		Addr	string		`set:"addr"`
	}
	ServerConfig struct {
		ReadTimeout		time.Duration	`set:"readtimeout" description:"Http server read timeout."`
		WriteTimeout	time.Duration	`set:"writetimeout" description:"Http server write timeout."`
		Http		[]*HttpConfig `set:"http"`
		Fastcgi		[]*FastcgiConfig `set:"fastcgi"`
		Handler		interface{}			`set:"handler" json:"-"`
	}
	Server struct {
		Config			*ServerConfig 	`set:"config"`
		mu				sync.Mutex
		wg				sync.WaitGroup
		handler			protocol.Handler
		http			*server.Server
		fastcgi			*server.Server
		oncehttp		sync.Once
		oncefastcgi		sync.Once
	}
)

func init() {
	eudore.RegisterComponent(eudore.ComponentServerEudoreName, func(arg interface{}) (eudore.Component, error) {
		srv := NewServer()		
		return srv, srv.Set("", arg)
	})
}

func NewServer() (*Server) {
	return &Server{
		Config:	&ServerConfig{},
		handler:	protocol.HandlerFunc(func(_ context.Context, w protocol.ResponseWriter, _ protocol.RequestReader) {
			w.Write([]byte("start eudore server, this default page."))
		}),
	}
}

func (srv *Server) Start() error {
	srv.mu.Lock()
	// 设置handler
	if h, ok := srv.Config.Handler.(protocol.Handler); ok {
		srv.handler = h
	}
	if len(srv.Config.Fastcgi) + len(srv.Config.Http) == 0 {
		return eudore.ErrServerNotSetRuntimeInfo
	}
	errs := eudore.NewErrors()
	// 启动fastcgi
	for _, fastcgi := range srv.Config.Fastcgi {
		ln, err := eudore.GlobalListener.Listen(fastcgi.Addr)
		if err != nil {
			errs.HandleError(err)
			continue
		}
		srv.EnableFastcgi()
		srv.wg.Add(1)
		go func(ln net.Listener){
			errs.HandleError(srv.fastcgi.Serve(ln))
			srv.wg.Done()
		}(ln)
	}
	// 启动http
	for _, http := range srv.Config.Http {
		ln, err := http.Listen()		
		if err != nil {
			errs.HandleError(err)
			continue
		}
		srv.EnableHttp()
		srv.wg.Add(1)
		go func(ln net.Listener){
			errs.HandleError(srv.http.Serve(ln))
			srv.wg.Done()
		}(ln)
	}

	// 等待结束
	srv.mu.Unlock()
	srv.wg.Wait()
	return errs.GetError()
}

func (srv *Server) Restart() error{
	srv.mu.Lock()
	defer srv.mu.Unlock()
	err := eudore.StartNewProcess()
	if err == nil {
		srv.Shutdown(context.Background())
	}
	return err
}

func (srv *Server) Close() error {
	return srv.Shutdown(context.Background())
}

func (srv *Server) Shutdown(ctx context.Context) (err error) {
	var stop = make(chan error, 2)
	// 关闭http server
	if srv.http != nil {
		go func() {
			stop <- srv.http.Shutdown(ctx)
		}()
	}else {
		stop <- nil
	}
	// 关闭fastcgi server
	if srv.fastcgi != nil {
		go func() {
			stop <- srv.fastcgi.Shutdown(ctx)
		}()
	}else {
		stop <- nil
	}
	// 获取关闭结果
	if e := <- stop; e != nil {
		err = e
	}
	if e := <- stop; e != nil {
		err = e
	}
	return err
}

func (srv *Server) Set(key string, val interface{}) error {
	srv.mu.Lock()
	defer srv.mu.Unlock()
	switch v := val.(type) {
	case *ServerConfig:
		srv.Config = v
	case protocol.Handler:
		srv.handler = v
	case *HttpConfig:
		srv.Config.Http = append(srv.Config.Http, v)
	case *FastcgiConfig:
		srv.Config.Fastcgi = append(srv.Config.Fastcgi, v)
	case map[string]interface{}:
		_, err := eudore.ConvertStruct(srv.Config, val)
		return err
	}
	return nil
}

func (*Server) GetName() string {
	return eudore.ComponentServerEudoreName
}

func (*Server) Version() string {
	return eudore.ComponentServerEudoreVersion
}

func (srv *Server) EnableHttp() {
	srv.oncehttp.Do(func(){
		// 创建http使用的服务
		srv.http = &server.Server{
			Handler:	srv.handler,
		}
		// 设服务连接处理为http
		srv.http.SetHandler(http.NewHttpHandler(srv.handler))
		srv.http.SetNextHandler("h2", http2.NewServer(srv.handler))
	})
}


func (srv *Server) EnableFastcgi() {
	srv.oncefastcgi.Do(func(){
		// 创建fastcgi使用的服务
		srv.fastcgi = &server.Server{
			Handler:	srv.handler,
		}
		// 设服务连接处理为fastcgi
		srv.fastcgi.SetHandler(fastcgi.NewServer(srv.handler))
	})
}


func (*ServerConfig) GetName() string {
	return eudore.ComponentServerEudoreName
}
