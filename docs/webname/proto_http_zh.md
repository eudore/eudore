# http protocol

**学习协议最好的方法就是抓包，参考[tools部分](#tools)的抓包方法。**

http请求是一问一答这样的，发送一个请求返回一个对应的响应，本编主要讲解基础概念，一些实现细节见[http细节](http_detail_zh.md).

详细信息请查看[rfc2616][rfc2616][中文pdf][rfc2616cn]。

1999年定义了HTTP/1.1协议RFC2616，总计176页， 经过2007-2014年多年的努力，RFC2616和RFC2617已经废弃，被总计10份新的RFCs代替：

RFC 7230: Message Syntax and Routing

RFC 7231: Semantics and Content

RFC 7232: Conditional Requests

RFC 7233: Range Request

RFC 7234: Caching

RFC 7235: Authentication

RFC 7236: Authentication Scheme Registrations

RFC 7237: Method Registrations

RFC 7238: the 308 status code

RFC 7239: Forwarded HTTP extension


## Request

http请求报文体分为三部分，请求请求行、请求头、请求正文。

例如

```
GET / HTTP/1.1
User-Agent: curl/7.29.0
Host: www.example.com
Accept: */*

```

第一行请求行是说明的请求的方法、uri和http协议版本。

后面就是header，一行一个header的键值对，每行使用`\r\n`作为分割符。

header最后是一空行，然后是就body内容，一般GET和HEAD方法的body都是空，所以结束了请求读取。

```
POST / HTTP/1.1
User-Agent: curl/7.29.0
Host: www.example.com
Accept: */*
Content-Length: 9
Content-Type: application/x-www-form-urlencoded

Post body
```

这是一个POST最后就多了body的数据，body后面是`\r\n`的分割符，body前是一个空行的分割。

在header中增加了Context-Length这个header，记录了body的长度，这个服务端就可以正确的读取body的数据，如果没有这个header就是没有body，另外还存在分段传输这种情况，通常请求使用分段传输很少见。

## Response

状态行、响应头(Response Header)、响应正文。

```
HTTP/1.1 302 Moved Temporarily
Server: example
Date: Sun, 27 Jan 2019 08:50:13 GMT
Content-Type: text/html
Content-Length: 155
Connection: keep-alive
Location: https://www.example.com/
Strict-Transport-Security: max-age=15768000

<html>
<head><title>302 Found</title></head>
<body bgcolor="white">
<center><h1>302 Found</h1></center>
<hr><center>example</center>
</body>
</html>
```


## Method

HTTP/1.1的常用方法有OPTIONS、GET 、HEAD 、POST 、PUT 、DELETE 、TRACE 、CONNECT，定义见rfc.

CONNECT通常用于服务端隧道代理，通常不使用可忽略。

OPTIONS用于跨越时，具体看跨越。

Head方法的特点是响应数据没有body。

HEAD和GET在rfc定义中没有说明不允许存在body，但是在浏览器里面是不允许body存在的。

GET与POST主要区分在是否存在body和阅览器是否会缓存。

在RESTful api风格中，主要区分是安全性和幂等性。

安全方法是指不修改资源的 HTTP 方法。譬如，当使用 GET 或者 HEAD 作为资源 URL，都必须不去改变资源。

HTTP 幂等方法是指无论调用多少次都不会有不同结果的 HTTP 方法。它无论是调用一次，还是十次都无关紧要。

| HTTP Method	| Idempotent|	Safe |
| ------------ | ------------ | ------------ |
| OPTIONS |	yes |	yes |
| GET |	yes |	yes |
| HEAD |	yes |	yes |
| PUT	 | yes	| no |
| POST | 	no	| no |
| DELETE	 | yes	| no |
| PATCH | 	no	| no |

## Status 

HTTP 响应状态代码指示特定 HTTP 请求是否已成功完成。响应分为五类：信息响应，成功响应，重定向，客户端错误和服务器错误。

[mozilla说明](https://developer.mozilla.org/zh-CN/docs/Web/HTTP/Status)

### 信息响应

10x只有两个状态码100（继续）和101（协议转换）

100在请求发送时带有Expect: 100-continue这个header，那个客户端不会发送body，服务端在读取body时先返回100，才会继续读取body，然后返回响应。

101通常是在websocket协议握手使用，利用http的Upgrade机制，在服务端完成ws握手后，然后101告诉客户端转换协议。

### 成功响应

20x是表示响应成功，一般是200直接表示成功，其他状态码有些小的区别。

### 重定向

30x表示重定向，301 302 307 308一般使用这些重定向，区别在于是否永久重定向，表示是否会缓存这个重定向结果301 307就是永久重定向，是否保持请求方法，301 302重定向后的请求方法都是GET，而307 308重定向后的请求方法保持不变。

304是缓存重定向，表示资源没有变动，客户端使用缓存，详细实现见缓存机制。

### 客户端响应

40x是客户端错误，表示客户端请求数据不对，通常就401(未认证) 403(权限不足) 404(资源不存在) 405(请求方法不允许),较为常见，其他就是各种异常，其中有一些错误会在http客户端出现。

### 服务端响应

50x是服务端错误，一般就500 502 504，502是后端代理错误(nginx 502就表示到反向代理的后端方法异常)，504是请求超时，其他请求一般就直接500了

## Header

## Simple

利用tcp协议简单实现一个http server和http client。

[链接][server-simple]

## tools

### curl

curl使用参数-v可以查看请求和响应的内容，或者使用-I来参考请求header。

```
[root@izj6cffbpd9lzl3tcm2csxz ~]# curl -v 127.0.0.1
* About to connect() to 127.0.0.1 port 80 (#0)
*   Trying 127.0.0.1...
* Connected to 127.0.0.1 (127.0.0.1) port 80 (#0)
> GET / HTTP/1.1
> User-Agent: curl/7.29.0
> Host: 127.0.0.1
> Accept: */*
> 
< HTTP/1.1 302 Moved Temporarily
< Server: wejass
< Date: Sun, 27 Jan 2019 08:50:13 GMT
< Content-Type: text/html
< Content-Length: 155
< Connection: keep-alive
< Location: https://www.wejass.com/
< Strict-Transport-Security: max-age=15768000
< 
<html>
<head><title>302 Found</title></head>
<body bgcolor="white">
<center><h1>302 Found</h1></center>
<hr><center>wejass</center>
</body>
</html>
* Connection #0 to host 127.0.0.1 left intact
```

### telnet

telnet相当于一个tcp拨号的客户端。

`telnet 127.0.0.1 80`建立tcp连接，写入请求行和Host header，然后换行就可以发送一个http请求。

### tcpdump

### wireshark


[rfc2616]: https://tools.ietf.org/html/rfc2616
[rfc2616cn]: ../resource/pdf/rfc-2616-hypertext-transfer-protocol-chinese.pdf
[server-simple]: ../../component/server/simple