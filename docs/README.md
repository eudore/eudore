# 目录

[eudore例子](example)

[框架设计文档](frame)

[标准库和框架总结文档](ideas)

[解决方案](program)

[http协议相关文档](webname)



没介绍的部分未完善，其他不完善部分请反馈。

|  编号 | 标题  | 介绍  |
| ------------ | ------------ | ------------ |
| 1  | http协议和技术  |  |
| 1.1 | [http协议](docs/webname/proto_http_zh.md) |  |
| 1.2 | [https协议](docs/webname/proto_https_zh.md) |  |
| 1.3 | [http2协议](docs/webname/proto_http2_zh.md) |  |
| 1.4 | [webscoket协议](docs//webname/proto_websocket_zh.md) |  |
| 1.5 | [CORS跨域资源共享](docs/webname/http_cors_zh.md) |  |
| 1.6 | [http协议简单实现](component/server/simple/) | 简单实现的一个http服务端和客户端 |
| 1.7 | [http协议实现细节](docs/webname/http_detail_zh.md) | 记录自己实现http服务端遇到的一些细节问题 |
| 1.8 | [cookie实现原理](docs/webname/http_cookie_zh.md) | cookie原理简述和操作 |
| 1.10 | [post请求](docs/webname/http_postdata_zh.md) | post发送的数据解析 |
|    |  |  |
| 2  |  net/http库 |   |
| 2.1  |  [net/http Server主流程分析](docs/ideas/readNetHttpServer_zh.md) | net/http启动Server处理一个请求  |
| 2.2  |  [x/net/http2 Server主流程分析] |
| 2.3  |  [github.com/gobwas/ws 解析websocket协议] | |
| 3  | golang http框架内容  |   |
| 3.1  |  [golang基于context的web范式](docs/ideas/baseContextWeb_zh.md)  |  golang中http框架主要内容的分析和总结 |
| 3.2  |  [Web框架简化实现](docs/ideas/microWeb.go) | 一个非常简单的http框架雏形  |
| 3.3  |  [golf分析](docs/ideas/readDineverGolf_zh.md) |  golf框架源码符合前面的总结 |
|   |   |   |
| 4  | [eudore相关内容](docs/frame/README.md)  | eudore设计文档目录页 |
| 4.1 | [Application](docs/frame/application_zh.md) | 
| 4.2 | [Context](docs/frame/context_zh.md) | |
| 4.5 | [Router](docs/frame/router_zh.md) | |eudore运行对象主体  |
| 4.6 | [Middleware](docs/frame/middleware_zh.md) | |
| 4.8 | [Server] | |
| 4.14 | [Controller](docs/frame/controller_zh.md) | |
|   |   |   |
| 5 | eudore问题和场景  |   |
|   | jwt使用  |   |
| | 实现鉴权 | | 
| | api熔断器及后台| |
| | 渲染状态资源sri值 | |
| | 分析静态文件自动push | |
| | 类似Rpc编写处理请求 |  |
| | [后台启动程序](component/command) |  |
| | [代码更新自动编译重启](component/notify) | |