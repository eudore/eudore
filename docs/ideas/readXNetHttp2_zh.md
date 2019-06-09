# x/net/http2

实现源码分析

`golang.org/x/net/http2`包处理请求主要堆栈，标准库也是使用的这套代码。

处理一个http2请求的主要函数调用堆栈：

http2.ConfigureServer(s *http.Server, conf *Server) error

	http2.*Server.ServeConn(c net.Conn, opts *ServeConnOpts) 

		http2.*serverConn.serve()

			http2.*serverConn.processFrameFromReader(res readFrameResult) bool

				http2.*serverConn.processFrame(f Frame) error

					http2.*serverConn.processSettings

					http2.*serverConn.processHeaders

						http2.*serverConn.newWriterAndRequest(*stream, *MetaHeadersFrame) (*responseWriter, *http.Request, error)

							http2.*serverConn.newWriterAndRequestNoBody(*stream, requestParam) (*responseWriter, *http.Request, error) 

						http2.*serverConn.runHandler(*responseWriter, *http.Request, func(http.ResponseWriter, *http.Request))

					http2.*serverConn.processData

					http2.*serverConn....







## http2.ConfigureServer

http2.ConfigureServer部分主要目的是设置http.Server.TLSNextProto对象，ConfigureServer函数前面是tls检查等，主要是闭包生产一个TLSNextProto的处理函数，NextProtoTLS的值一定是`h2`，如果只是https值为`http/1.1`。

当http2开始处理时，就conf.ServeConn处理连接了，下一步看`http2.Server.ServeConn`方法

`注意：此处将http.Server的连接传递给http2来处理`

```golang
func ConfigureServer(s *http.Server, conf *Server) error {
	...
	
	if s.TLSNextProto == nil {
		s.TLSNextProto = map[string]func(*http.Server, *tls.Conn, http.Handler){}
	}
	protoHandler := func(hs *http.Server, c *tls.Conn, h http.Handler) {
		if testHookOnConn != nil {
			testHookOnConn()
		}
		conf.ServeConn(c, &ServeConnOpts{
			Handler:    h,
			BaseConfig: hs,
		})
	}
	s.TLSNextProto[NextProtoTLS] = protoHandler
	return nil
}

```

## http.Server.ServeConn

ServeConn开始连接中间部分就是各种连接初始化了，主要就是创建的http2连接使用serve方法处理。

```golang
func (s *Server) ServeConn(c net.Conn, opts *ServeConnOpts) {
	baseCtx, cancel := serverConnBaseContext(c, opts)
	defer cancel()

	sc := &serverConn{
		srv:                         s,
		...
	}

	...

	sc.serve()
}
```

## http2.serverConn.serve

