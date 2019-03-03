# Protocol

脱离net/http重写实现http相关协议。

- protocol 定义通用http接口

- protocol/server 实现http server

- protocol/http 实现http协议解析

- protocol/http2 实现http2协议解析，基于golang.org/x/net/http2修改。

- protocol/fastcgi 实现fastcgi协议解析，基于net/http/fcgi修改。

根据处理过程，可以将应用程序分层分为：服务层、连接层、应用层。

服务层启动服务监听连接，创建连接后将连接传递给连接层，使用接口protocol.HandlerConn传递到连接层。

连接层解析连接读写数据，生成protocol.RequestReader和protocol.ResponseWriter两个接口对象，然后调用protocol.Handler接口处理请求。

应用层将接口protocol.Handler转换成实际使用对象，例如eudore.Context。

以上server包就是服务层对象，http包和http2包是连接层对象，eudore框架为应用层实例。

# issue

http包解解析不够完善

http2协议和factcgi协议下如何实现hijack

