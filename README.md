

## Object

Application、Context、Request、Response

Router、Middleware、Logger、Server

Bind、Render、View

Config、Cache

## Features

- 核心对象接口化 支持重写
- 标准库解耦 可自定义http协议解析
- 多端口启动 热重启 热加载
- 路由匹配参数 路由额外附加参数 子路由
- 全局配置 自定义配置解析过程 远程读取 自动生成帮助信息
- 自定义日志处理方式 全链路日志
- 信号响应 systemctl支持
- pprof

### issue

Config setdata反射设置

Bind url编码序列化

ReloadSignal 清除旧规则

### issue2

binder from和url实现

eudore server解析Context

handler MutilHandler不支持Context.Next

router any重复注册

setting ...

signal 未防止重复注册

config reflect set/get

热重启失效
