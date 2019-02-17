#

[文档][docs]

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

### issue

Config setdata反射设置

ReloadSignal 清除旧规则

binder from和url实现

setting 基于配置初始化对象

signal 未防止重复注册

热重启失效




## Example

- [Application](#application)
- [Server](#Server)
- [Logger](#logger)
- [Router and Middleware](#router-and-middleware)
- [Middleware](#middleware)
	- [Jwt and Session](#jwt-and-session)
	- [Ram]
	- [Gzip]
- [Context]
	- [Bind]
	- [Param]
	- [Header]
	- [Cookie]

	- [Render]
	- [View]
	- [Push]
	- [Redirect]

	- [Logger]



- [Application]


```golang
func main() {
	// 运行core
	core := eudore.NewCore()
	go core.Run()

	// 运行eudore
	e := eudore.NewEudore()
	e.Run()
}
```


- [Server]

Server是eudore顶级接口对象之一。

```golang
func main() {
	e := eudore.NewEudore()
	// 直接设置Server对象
	e.Server, _ = eudore.NewServer("server-std", nil)

	// 加载Server组件，设置一个启动端口，设置超时时间
	e.RegisterComponent("server", &eudore.ServerConfigGeneral{
		Addr:	"8088",
		ReadTimeout:	12 * time.Second,
		WriteTimeout:	4 * time.Second,
	})

	// 启动多个端口
	// 端口8085：设置证书、开启https http2、关闭双向htttps
	e.RegisterComponent("server-multi", &eudore.ServerMultiConfig{
		Configs:	[]interface{}{
			&eudore.ServerConfigGeneral{
				Name:	"server",
				Addr:	"8085",
				Https:	true,
				Http2:	true,
				Mutual:	false,
				Certfile:	"/etc/...",
				Keyfile:	"/etc/...",
			},
			&eudore.ServerConfigGeneral{
				Name:	"server",
				Addr:	"8086",
			},
		},
	})
	e.Run()
}
```

- [Logger] 

NewEudore创建App时，会创建logger-init日志组件，实现LoggerInitHandler接口，改组件会将日志条目存储起来，直到加载下一个日志组件时，使用新日志组件处理所有存储的日子条目。

logger-std的Std用于输出标准输出；若Path为空会强制Std为true；Format会使用存储的LoggerFormatFunc函数，若无法匹配会使用text/template模板格式化日志。

```golang
func main() {
	e := eudore.NewEudore()
	e.Debug("init 1")
	e.Get("/get", get)
	e.Post("/post", post)
	e.RegisterComponent("logger-std", &eudore.LoggerStdConfig{
		Std:	true,
		Path:	"access.log",
		Level:	"debug",
//		Format:	"json",
		Format:	"{{.time}} {{.level}} {{.message}}",
	})
	e.Debug("init 2")
	e.Run()
}
```

- [Router and Middleware]

router-std路由器支持默认参数、路径参数、通配符参数，目前不支持正则参数和参数效验，可重新实现一个Router来实现这些功能。

`http://localhost:8088/api/v1/`
`http://localhost:8088/api/v1/get/eudore`


```golang
package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware/logger"
	"github.com/eudore/eudore/middleware/gzip"
	"github.com/eudore/eudore/middleware/recover"
)

func main() {
	// 创建App
	app := eudore.NewCore()
	// 修改日志配置
	app.RegisterComponent("logger-std", &eudore.LoggerStdConfig{
		Std:	true,
		Level:	eudore.LogDebug,
		Format:	"json",
	})
	// 全局级请求处理中间件
	app.AddHandler(
		logger.NewLogger(eudore.GetRandomString),
		gzip.NewGzip(5),
	)

	// 创建子路由器
	apiv1 := eudore.NewRouterClone(app.Router)
	// 路由级请求处理中间件
	apiv1.AddHandler(eudore.HandlerFunc(recover.RecoverFunc))
	{
		apiv1.GetFunc("/get/:name", handleget)
		// Api级请求处理中间件
		apiv1.Any("/*", eudore.NewMiddlewareLink(
			eudore.HandlerFunc(handlepre1),
			eudore.HandlerFunc(handleparam),
		))
	}
	// app注册api子路由
	app.SubRoute("/api/v1 version:v1", apiv1)
	// 默认路由
	app.AnyFunc("/*path", func(ctx eudore.Context){
		ctx.WriteString(ctx.Method() + " " + ctx.Path())
		ctx.WriteString("\nstar param: " + " " + ctx.GetParam("path"))
	})
	// 启动server
	app.Listen(":8088")
	app.Run()
}

func handleget(ctx eudore.Context) {
	ctx.Debug("Get: " + ctx.GetParam("name"))
	ctx.WriteString("Get: " + ctx.GetParam("name"))
}
func handlepre1(ctx eudore.Context) {
	// 添加参数
	ctx.AddParam("pre1", "1")
	ctx.AddParam("pre1", "2")
}
func handleparam(ctx eudore.Context) {
	// 将ctx的参数以Json格式返回
	ctx.WriteJson(ctx.Params())
	// 将ctx的参数感觉请求格式返回
	ctx.WriteRender(ctx.Params())
}
```


- [Jwt and Session]

```golang
func main() {
	core := eudore.NewCore()
	core.AddHandler(
		jwt.NewJwt(jwt.NewVerifyHS256("1234")),
		eudore.HandlerFunc(session.SessionFunc),
	)
	core.Any("/", any)
	core.Run()
}

func any(ctx *eudore.Context) {
	ctx.Info(ctx.Value(eudore.ValueJwt))
	ctx.Info(ctx.Value(eudore.ValueSession))

	// inti
	sess := cast.NweMap(ctx.Value(eudore.ValueJwt))
	// get
	ctx.Info(sess.GetInt("uid"))
	// set
	sess.Set("name", "ky2")
	// release，未实现检查map是否变化然后自动释放。
	ctx.SetValue(eudore.ValueSession, sess)

	// inti
	jwt := cast.NweMap(ctx.Value(eudore.ValueSession))
	// get
	ctx.Info(jwt.Get("exp"))
}

```
- []
- []
- []







[docs]: tree/master/docs
