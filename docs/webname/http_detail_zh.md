# http细节

记录学习过程中遇到的http实现细节问题。

## http长连接实现

如果http发生请求时，如果带有`Connection: keep-alive`这个header，服务端才会使用长连接，http在处理完一个请求就会等待下一个请求然继续处理，如果请求没有发送keepalive这个值，服务端就会关闭这个http连接。

http客户端如何知道http服务端已经处理完了这个请求呢？就是从tcp读多少才是全部响应内容，然后才可以复用连接发送下一个请求，响应内容长度获得实现方法有两种。1、Content-Length header记录body长度；2、传输编码使用分块`Transfer-Encoding: chunked`。

如果不使用`Connection: keep-alive`长连接呢？在请求结束时服务器断开连接，客户端就读取的全部body就是一个请求的响应。

### Content-Length

在第一次写入数据时，判断响应是否设置了Content-Length header，如果设置了内容长度，就会在header里面写入内容长度，然后继续写入全部数据。

在go标准库和eudore实现中，在两种情况下会存在Content-Length header

一是第一次写入数据前，处理函数设置这个header，表示了这个返回数据有多少。

二是请求处理结束时，在写入响应body数据时，会先放到缓冲中，缓冲满了才会写入到连接，求处理结束时会调用finalFlush，将剩余全部数据返回。

这个时候判断是否写入了状态行和header，没有设置状态行就设置Content-Length的值为缓冲长度，虽然body的长度没有设置，但是是已知的就是缓冲长度。

### Transfer-Encoding

如果设置了Content-Length，那么长度是已知的，就不会使用分块传输。

但是没有设置Content-Length呢？缓冲的长度是一个固定值，如果内容写入数据超过缓冲区，就必须返回给客户端，就要使用分块传输。

在响应写入状态行和header时，无法知道Content-Length，就设置传输编码为“chunked（分块）”，body就是使用分块编码。

分块传输的一个响应体：

```
HTTP/1.1 200 OK 
Content-Type: text/plain 
Transfer-Encoding: chunked

7\r\n
Mozilla\r\n 
9\r\n
Developer\r\n
7\r\n
Network\r\n
0\r\n 
\r\n
```

body先写入一个块的长度例如7然后\r\n换行，再写入长度为7的内容然后\r\n换行，这样就是一个块，最后写入`0\r\n\r\n` 告诉客户端分块传输完毕。

客户端读取分块传输，就这样一块块body读取完毕，组合起来的数据就完整的body，就知道读取一个响应完毕了，可以复用这个连接发送下一个请求了。

### Trailer

Trailer Header是一个响应header，必须在分块传输下使用，允许发送方在分块发送的消息后面添加额外的元信息，这些元信息可能是随着消息主体的发送动态生成的，比如消息的完整性校验，消息的数字签名，或者消息经过处理之后的最终状态等。

Trailer可以理解成“预告”，先告诉客户端最后会有什么header，然后在body后再发送这些header。

Trailer的响应例子：

```
HTTP/1.1 200 OK 
Content-Type: text/plain 
Transfer-Encoding: chunked
Trailer: Expires

7\r\n 
Mozilla\r\n 
9\r\n 
Developer\r\n 
7\r\n 
Network\r\n 
0\r\n 
Expires: Wed, 21 Oct 2015 07:28:00 GMT\r\n
\r\n
```

可以发现在分块传输body的基础上有所变化，`0\r\n\r\n`的分块结束标识中间写入了header，而尾部header的名称就是开始写入header的Trailer的值，多个header名称使用逗号分隔。

Trailer不能发送Content-Encoding、Content-Type、Content-Range、Trailer等header。

### Buffer

关于缓存实现，标准库使用bufio.Writer作为缓冲，每写满一个块2kb缓冲，bufio.Writer就会flush数据，将缓冲返回，第一次缓冲满了就必须要写入状态行和header，才能继续写入body。

eudore http使用一个[]byte作为缓冲，会记录未发送的缓冲数据，如果新写入数据会使缓冲区溢出，就将缓冲数据和新数据一起发送
，否则将新数据追加到缓冲中。

标准库在第一次Write后就不能设置状态行和Header，eudore http会在第一次写满缓冲后才不能设置状态行和Header。

## http反向代理实现

http反向代理的原理就是请求转发，处理者接收请求然后发送目标。

转发时需要移除跳对跳Header，并写入正确的header,如果遇到Upgrade返回101之后就进行双向io.Cpoy进行tcp隧道代理。

在rfc里面定义了端对端 and 跳对跳header，端对端主要是指客户端和服务端之间传输的header，跳对跳是针对有http proxy的情况，从一个代理到达下一个代理就一跳。

如果一个请求是这样的，client -> proxy -> proxy -> server，端就是client到server，跳就三跳。

在跳对跳header里面，保存者两个http服务间的，连接信息、传输信息，在每一个跳之间都有自己的连接信息和传输信息，所以转发时需要移除跳对跳Header。

端对端Header的协议header和body就正常的转发就好了。

简单实现：

没有处理err，没有处理跳对跳Header，没有处理101。

```golang
func(w http.ResponseWriter, r *http.Request) {
	r2, _ := http.NewRequest(r.Method, r.URL.Path, r.Body)
	w2, _ := http.DefaultClient.Do(r2)
	defer w2.Body.Close()
	io.Copy(w, w2.Body)
}
```

## http跨域请求

[Cors](http_cors_zh.md)

## http cookie实现原理

[Cookie](http_cookie_zh.md)

## http重定向过程

返回30x状态和Location header，状态码表示重定向类型，而Location里面记录了重定向地址。

## http gzip压缩

gzip是内容编码，重写Write方法封装一层gzip.Writer，将原始数据写入gzip.Writer，最后将gzip编码后的数据输出到http.ResponseWriter。

## http 304响应

## http range header分段实现

## websocket握手过程

[Websocket Upgrade](../webname/proto_websocket_zh.md#Upgrade)

## http2握手过程

h2握手利用了tls的ALPN或NPN机制，如果想ws那样的握手就是h2c，基于http实现的http2，不是基于https。

tls配置有一项NextProtos属性，用于tls的ALPN扩展，如果值是h2，在tls握手时就可以完成h2握手，客户端在tls握手时，发现APLPN扩展的值是h2，就自动服务端支持h2，然后客户端发送固定的请求行`PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n`，服务端解析到就将http Connn转换成http2 Conn，然后安装h2协议开始操作链接。