# http protocol

详细信息请查看[rfc2616][rfc2616][中文pdf][rfc2616cn]。

## Request

http请求报文体分为三部分，请求请求行、请求头、请求正文。

例如

```
GET / HTTP/1.1
User-Agent: curl/7.29.0
Host: 127.0.0.1
Accept: */*

```

第一行请求行是说明的请求的方法、uri和http协议版本。

## Response

状态行、响应头(Response Header)、响应正文。

```
HTTP/1.1 302 Moved Temporarily
Server: wejass
Date: Sun, 27 Jan 2019 08:50:13 GMT
Content-Type: text/html
Content-Length: 155
Connection: keep-alive
Location: https://www.wejass.com/
Strict-Transport-Security: max-age=15768000

<html>
<head><title>302 Found</title></head>
<body bgcolor="white">
<center><h1>302 Found</h1></center>
<hr><center>wejass</center>
</body>
</html>
```


## Method

HTTP/1.1的常用方法有OPTIONS、GET 、HEAD 、POST 、PUT 、DELETE 、TRACE 、CONNECT，定义见rfc2616

## Status 

## Header

## Simple

利用tcp协议简单实现一个http server和http client。

[链接][server-simple]

## tools

### curl

curl使用参数-v可以查看请求和响应的内容。

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

### tcpdump

### wireshark


[rfc2616]: https://tools.ietf.org/html/rfc2616
[rfc2616cn]: ../resource/rfc-2616-hypertext-transfer-protocol-chinese.pdf
[server-simple]: ../../component/server/simple