# Eudore

eudore具有以下对象，除Application以为均为接口，每个对象都具有明确语义，Application是最顶级对象可以通过组合方式实现重写，其他对象为接口定义直接重新实现，或组合接口实现部分重写。

| 名称 | 作用 | 定义 |
| ------------ | ------------ | ------------ |
| [Application](application_zh.md) | 运行对象主体 | app.go core.go eudore.go |
| [Context](context_zh.md) | 请求处理上下文 | context.go contextExtend.go |
| Request | Http请求数据 | request.go |
| Response | http响应写入 | response.go |
| [Router](router_zh.md) | 请求路由选择 | router.go routerRadix.go routerFull.go |
| [Middleware](middleware_zh.md) | 多Handler组合运行 | handler.go |
| Logger | App和Ctx日志输出 | logger.go |
| Server | http Server启动 | server.go |
| Config | 配置数据管理 | config.go configparse.go |
| Cache | 全局缓存对象 | cache.go |
| View | 模板渲染 | view.go |
| Client | http客户端 | client.go |
| [Session](session_zh.md) | 会话数据管理 | session.go |
| [Controller](controller_zh.md) | 解析执行控制器 | controller.go |
| Bind | 请求数据反序列化 | bind.go |
| Render | 响应数据序列化 | render.go |
| Websocket | websocket协议读写 | websocket.go |

其他文件定义内容

| 文件 | 作用 |
| ------------ | ------------ |
| command.go | 启动命令解析 |
| component.go | 组件定义 |
| const.go | 定义常量 |
| doc.go | godoc内容 |
| error.go | 定义错误 |
| listener.go | 全局监听管理 |
| reflect.go | 各类反射辅助函数 |
| setting.go | 配置化启动程序 |
| signal.go | 全局信号管理 |
| util.go | 辅助函数 |
| version.go | 版本信息常量 |


# RequestReader & ResponseWriter

```golang
type (
	// Get the method, version, uri, header, body from the RequestReader according to the http protocol request body. (There is no host in the golang net/http library header)
	//
	// Read the remote connection address and TLS information from the net.Conn connection.
	//
	// 根据http协议请求体，从RequestReader获取方法、版本、uri、header、body。(golang net/http库header中没有host)
	//
	// 从net.Conn连接读取远程连接地址和TLS信息。
	RequestReader interface {
		// http protocol data
		Method() string
		Proto() string
		RequestURI() string
		Header() Header
		Read([]byte) (int, error)
		Host() string
		// conn data
		RemoteAddr() string
		TLS() *tls.ConnectionState
	}
	// ResponseWriter接口用于写入http请求响应体status、header、body。
	//
	// net/http.response实现了flusher、hijacker、pusher接口。
	ResponseWriter interface {
		// http.ResponseWriter
		Header() http.Header
		Write([]byte) (int, error)
		WriteHeader(codeCode int)
		// http.Flusher 
		Flush()
		// http.Hijacker
		Hijack() (net.Conn, *bufio.ReadWriter, error)
		// http.Pusher
		Push(string, *PushOptions) error
		Size() int
		Status() int
	}
)
```