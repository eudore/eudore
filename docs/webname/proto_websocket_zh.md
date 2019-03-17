# WebSocket

详细信息请查看[rfc6455][rfc6455][中文pdf][rfc6455cn]。

基于http协议Upgrade机制。

Websocket依靠http.Hijacker接口，获得net.Conn的tcp连接，然后利用tcp协议连接封装生成WebSocket协议，重新定义协议行为。

## Upgrade



## Read & Write

websocket协议是分帧传输，目前标准有继续帧、二进制帧、文本帧、ping帧、pong帧、close帧六种。


[rfc6455]: https://tools.ietf.org/html/rfc6455
[rfc6455cn]: ../resource/pdf/rfc-6455-websocket-protocol-in-chinese.pdf