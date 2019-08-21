# Eudore

[![Go Report Card](https://goreportcard.com/badge/github.com/eudore/eudore)](https://goreportcard.com/report/github.com/eudore/eudore)
[![GoDoc](https://godoc.org/github.com/eudore/eudore?status.svg)](https://godoc.org/github.com/eudore/eudore)

eudore是一个高扩展、高效的http框架及[http文档库](docs)。

反馈和交流[q群373278915](//shang.qq.com/wpa/qunwpa?idkey=869ec8f1272b4757771c3e406349f1128cfa3bd9ca668937dda8dfb223261a60)。

## Features

- 易扩展：主要设计目标，核心全部解耦，接口即可逻辑。
- 简单：对象语义明确，框架代码量少复杂度低，无依赖库。
- 易用：支持各种Appcation和Context扩展添加功能。
- 高性能：各部分在同类库中没有明显性能问题。
- 两项创新：[新Radix路由实现](https://github.com/eudore/erouter)和处理函数扩展机制

## 学习指南

包含http相关、go web相关以及eudore相关文档。

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
| 4.6 | [Middleware](middleware_zh.md) | |
| 4.8 | [Server] | |
| 4.14 | [Controller](docs/controller_zh.md) | |
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

## 功能列表及演示

[eudore例子](docs/example)，暂时缺省的文档请看[godoc](https://godoc.org/github.com/eudore/eudore)和源码。

- Application
- [x] 程序启动流程自定义实现，内置两种
- [x] 信号处理
- Context
- [x] Context与server完全解耦
- [x] [根据Content-type自动序列化数据](docs/example/bind.go)
- [x] 根据Accept反序列化数据
- [x] [http2 push](docs/example/serverPush.go)
- Config
- [x] [配置库集成](docs/frame/config_zh.md)
- [x] 支持基于路径读写对象
- [x] 解析参数环境变量和[差异化配置](docs/example/configMods.go)
- [x] map和结构体相互转换
- [x] 数据类型转换工具
- [ ] 生成对象帮助信息
- Server
- [x] 支持net/http启动server
- [x] 支持fasthttp启动server
- [x] 重新实现http协议server
- [x] http2协议两种server支持
- [x] fastcgi协议两种server支持
- [x] [websocket协议支持](docs/example/websocket.go)
- [ ] 重新实现websocket协议
- [x] 支持TLS和双向TLS
- [x] 自动自签TLS证书
- [ ] http3协议学不动了
- [x] server热重启支持
- [x] server后台启动
- Router
- [x] [零内存复制快速路由器实现](docs/frame/router_zh.md#RouterRadix)
- [x] [严格路由匹配顺序](docs/example/core.go#L26-L28)
- [x] RESTful风格基于方法的路由注册匹配
- [x] 基于Host进行路由注册匹配
- [x] [组路由注册支持](docs/example/core.go#L22-L29)
- [x] 匹配前全局中间件
- [x] [全局中间件注册](docs/example/core.go#L17-L19)
- [x] [组级中间件注册](docs/example/core.go#L24)
- [x] [api级中间件注册](docs/example/core.go#L27)
- [x] [路由默认参数](docs/example/core.go#L22)
- [ ] 默认参数匹配注册中间件
- [x] [路由匹配参数捕获](docs/example/core.go#L28)
- [x] 路由匹配参数校验
- [ ] 路由匹配参数正则捕捉
- [x] [路由匹配通配符捕捉](docs/example/core.go#L27)
- [x] 路由匹配通配符校验
- [ ] 路由匹配通配符正则捕捉
- Logger
- [x] 初始化期间日志处理
- [x] 日志条目属性支持
- Mvc
- [x] [mvc支持](docs/example/mvc.go)
- [x] [控制器函数输入参数](docs/example/mvc.go#L27-L30)
- [x] [自定义控制器执行](docs/frame/controller_zh.md#路由器控制器解析函数)
- Tools
- [x] 程序启动命令解析
- [x] 信号响应支持
- [ ] SRI值自动设置
- [x] 自动http2 push
- [x] [http代理实现](docs/example/httpProxy.go)
- [x] [pprof支持](component/pprof)
- [x] expvar支持
- [x] [运行时对象数据显示](component/show)
- [x] api模拟工具
- [x] [更新代码自动重启](component/notify)
- Session
- [ ] [Session实现](docs/example/session.go)
- Middleware
- [x] [熔断器及管理后台](docs/example/breaker.go)
- [x] RAM资源访问管理
- [x] BasicAuth
- [x] CORS跨域资源共享
- [x] gzip压缩
- [x] 限流
- [x] 异常捕捉
- [ ] 请求超时
- [x] 访问日志

## 许可

MIT

框架使用无限制且不负责，文档转载需声明出处。