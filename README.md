#  Eudore

本框架为个人学习研究的重框架，每周最多同步更新一次，未稳定前不欢迎issue、pr、using，可查看http及go web框架[相关文档][docs]，联系方式q群373278915。

## Features

- 核心全部接口化,支持重写Application、Context、Request、Response、Router、Middleware、Logger、Server、Config、Cache、Bind、Render、View。
- 对象语义明确。
- net/http库解耦,可实现或接入其他http协议解析库。
- Middleware和Router自己重写实现，与其他框架相比性能不可能会是短板。

## issue

Config setdata反射设置

setting 基于配置初始化对象未实现

signal 可重复注册

热重启失效

没有i18n

缺少完整websocket实现，仅有upgrade部分

重写SubRouter未完成，并导致Group Middleare无法实现

Gzip异常

Logger基于GC优化

## Example

**Example部分未更新**

- [Application](#application)
- [Server](#Server)
- [Logger](#logger)
- [Router and Middleware](#router-and-middleware)
- [Middleware](#middleware)
	- [Jwt and Session](#jwt-and-session)
	- [Ram](#ram)
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
- [Websocket](#websocket)


## Application


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


## Server

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

## Logger

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
		Format:	`[{{.Timestamp.Format "Jan 02, 2006 15:04:05 UTC"}}] {{.Level}}: {{.Message}}`,
	})
	e.Debug("init 2")
	e.Run()
}
```

## Router and Middleware

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
	// 将ctx的参数根据请求格式返回
	ctx.WriteRender(ctx.Params())
}
```


### Jwt and Session

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
	sess := cast.NewMap(ctx.Value(eudore.ValueJwt))
	// get
	ctx.Info(sess.GetInt("uid"))
	// set
	sess.Set("name", "ky2")
	// release，未实现检查map是否变化然后自动释放。
	ctx.SetValue(eudore.ValueSession, sess)

	// inti
	jwt := cast.NewMap(ctx.Value(eudore.ValueSession))
	// get
	ctx.Info(jwt.Get("exp"))
}

```

### Ram

资源访问管理

`github.com/eudore/eudore/middleware/ram`包共有acl、pbac、rbac、shell四种鉴权方式。

ram.RamHttp对象定义：

```golang
type RamHttp struct {
    RamHandler
    GetId     GetIdFunc
    GetAction GetActionFunc
    Forbidden ForbiddenFunc
}
type RamHttp
    func NewRamHttp(rams ...RamHandler) *RamHttp
    func (r *RamHttp) Handle(ctx eudore.Context)
    func (r *RamHttp) Set(f1 GetIdFunc, f2 GetActionFunc, f3 ForbiddenFunc) *RamHttp
```

RamHttp需要Set设置获取id、获取行为、403执行，三个行为参数。

```golang
// bug
package main

import (
	"strconv"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware/ram"
	"github.com/eudore/eudore/middleware/ram/acl"
)

func main() {
	app := eudore.NewCore()	
	eudore.SetComponent(app.Logger, eudore.LoggerHandleFunc(eudore.LoggerHandleJson))
	// add
	ramAcl := acl.NewAcl()
	ramAcl.AddAllowPermission(1, []string{"Show", "Get"})
	app.AddHandler(
		ram.NewRamHttp(
			// 执行acl鉴权
			ramAcl,
			// rbac.NewRbac(),
			// 默认执行拒绝
			ram.DenyHander,
		).Set(getid, nil, nil),
	)

	app.AnyFunc("/:id/:action ss:00", func(ctx eudore.Context) {
		ctx.Info(ctx.GetParam("action"))
		ctx.WriteString("Allow " + ctx.GetParam("action"))
	})
	app.AnyFunc("/", func(ctx eudore.Context) {
		ctx.Info(ctx.Path())
	})
	// start
	app.Listen(":8088")
	app.Run()
}

func getid(ctx eudore.Context) int {
	i, err := strconv.Atoi(ctx.GetParam("id"))
	if err != nil {
		return 0
	}
	return i
}
```


# Websocket

目前没有独立的websocket库，且不与net/http兼容,推荐使用`github.com/gobwas/ws`库。

eudore.UpgradeHttp获取net.Conn链接并写入建立请求响应，然后wsutil库读写数据。

`ctx.Response().Hijack()`可以直接获得原始tcp连接，然后读取header判断请求，写入101数据，再操作websocket连接。

```golang
package main

import (
	"github.com/eudore/eudore"
	"github.com/gobwas/ws/wsutil"
)

func main() {
	app := eudore.NewCore()
	eudore.SetComponent(app.Logger, eudore.LoggerHandleFunc(eudore.LoggerHandleJson))
	app.RegisterComponent(eudore.ComponentRouterEmptyName, eudore.HandlerFunc(func(ctx eudore.Context){
		conn, _, err := eudore.UpgradeHttp(ctx) 
		if err != nil {
			// handle error
			ctx.Error(err)
		}
		go func() {
			defer conn.Close()

			for {
				msg, op, err := wsutil.ReadClientData(conn)
				if err != nil {
					ctx.Error(err)
					// handle error
				}
				ctx.Info(string(msg))
				err = wsutil.WriteServerMessage(conn, op, msg)
				if err != nil {
					// handle error
				}
			}
		}()
	}))

	app.Listen(":8088")
	app.Run()
}
```

[docs]: docs
