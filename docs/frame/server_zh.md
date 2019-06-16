# Server

Server的功能负责http server启动，目前有std server、eudore server、 fasthttp server三种。

## Std Server

std sever是对标准库的封装，handler需要实现http.Handler，暂时没有修改成protocol.Handler接口。

## Eudore Server

eudore server是自行实现的一个简单http server，使用protocol.Handler作为请求处理接口。