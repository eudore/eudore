# Server

Server的功能负责http server启动，目前有std server、eudore server两种。

```golang
// Server 定义启动http服务的对象。
type Server interface {
	SetHandler(http.Handler)
	Serve(net.Listener) error
	Shutdown(ctx context.Context) error
}
```

Server定义三个方法，SetHandler用于设置Server的http.Handler，Serve启动一个端口的监听，Shutdown用于关闭Server。

## Std Server

std sever是对标准库的封装，handler需要实现http.Handler接口。

## Eudore Server

eudore server是自行实现的一个简单http server，使用protocol.Handler作为请求处理接口。