# WebSocket

详细信息请查看[rfc6455][rfc6455][中文pdf][rfc6455cn]。

Websocket协议握手基于http协议Upgrade机制。

Websocket依靠http.Hijacker接口，获得net.Conn的tcp连接，然后利用tcp协议连接封装生成WebSocket协议，重新定义协议行为。

## Upgrade

ws握手礼仪了http的Upgrade机制

例如一次ws握手请求header

```
GET /chat HTTP/1.1
Host: server.example.com
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Key: x3JJHMbDL1EzLkh9GBhXDw==
Sec-WebSocket-Protocol: chat, superchat
Sec-WebSocket-Version: 13
Origin: http://example.com
```

重点如下两个 header：

Upgrade: websocket
Connection: Upgrade

Upgrade 表示升级到 WebSocket 协议，Connection 表示这个 HTTP 请求是一次协议升级，Origin 表示发请求的来源。

Sec-WebSocket-Key是一个随机值，服务端需要返回Sec-Websocket-Accept这个header，值为Sec-WebSocket-Key的值加固定字符串"258EAFA5-E914-47DA-95CA-C5AB0DC85B11",然后计算sha1.

Sec-WebSocket-Protocol是支持的子协议。

Sec-WebSocket-Version是ws版本，一般都是13.

## 消息帧

websocket协议是分帧传输，目前标准有继续帧、二进制帧、文本帧、ping帧、pong帧、close帧六种。

基本来源于rfc5.2章。

桢结构

```
  0                   1                   2                   3
  0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
 +-+-+-+-+-------+-+-------------+-------------------------------+
 |F|R|R|R| opcode|M| Payload len |    Extended payload length    |
 |I|S|S|S|  (4)  |A|     (7)     |             (16/64)           |
 |N|V|V|V|       |S|             |   (if payload len==126/127)   |
 | |1|2|3|       |K|             |                               |
 +-+-+-+-+-------+-+-------------+ - - - - - - - - - - - - - - - +
 |     Extended payload length continued, if payload len == 127  |
 + - - - - - - - - - - - - - - - +-------------------------------+
 |                               |Masking-key, if MASK set to 1  |
 +-------------------------------+-------------------------------+
 | Masking-key (continued)       |          Payload Data         |
 +-------------------------------- - - - - - - - - - - - - - - - +
 :                     Payload Data continued ...                :
 + - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - +
 |                     Payload Data continued ...                |
 +---------------------------------------------------------------+
```

fin是第一个bit，表示是否是最后一个分片，如果是0就是最后一个。

rsv123，是三个标志位一般都是0，如果使用协议扩展，可能非0，需要客户端和服务端协商好。

opcode就桢的类型，目前使用了六种，其他作为保留，0x0延续桢，0x1文本桢,0x2二进制桢,0x8关闭桢,0x9 ping桢,0xa ping桢。

mask 表示是否使用掩码，客户端发送消息需要掩码mask=1，服务端返回消息不要掩码mask=1，该项要求在rfc有说明。

Payload length为数据长度，有三种含义，0-125是数据长度，126表示后续两byte是一个16位无符号整数为长度，127表示后续8 byte是一个64位无符号整数为长度（最高有效位必须是0），

往后就是payload扩展长度，如果mask=1后续就有4位是掩码的长度。

最后是Payload Data，就是桢的数据，前面都是固定的header结构和长度，Payload的长度在header中可以解析到，一个完整的桢就是header+payload，通过长度来实现桢边界分割。


[rfc6455]: https://tools.ietf.org/html/rfc6455
[rfc6455cn]: ../resource/pdf/rfc-6455-websocket-protocol-in-chinese.pdf