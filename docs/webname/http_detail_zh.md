# http细节

记录学习过程中遇到的http实现细节问题。

## http长连接实现

如果http发生请求时，如果带有`Connection: keep-alive`这个header，服务端才会使用长连接，http在处理完一个请求就会等待下一个请求然继续处理，如果请求没有发送keepalive这个值，服务端就会关闭这个http连接

http客户端如何知道http服务端已经处理完了这个请求呢？就是从tcp读完全部响应内容，可以复用发送下一个请求，实现方法有两种。1、Content-Length header记录body长度；2、传输编码使用分块

### Content-Length



### Transfer-Encoding


## http反向代理实现

## http跨域请求

## http cookie实现原理

## http重定向过程

## http gzip压缩

## http 304响应

## http range header分段实现

## websocket握手过程

## http2握手过程