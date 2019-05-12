# Eudore

本框架为个人学习研究的重框架，每周最多同步更新一次，未稳定前不欢迎issue、pr、using，可查看http及go web框架[相关文档][docs]，交流q群373278915。

## Features

- 核心全部接口化,支持重写Application、Context、Request、Response、Router、Middleware、Logger、Server、Config、Cache、Session、View、Bind、Render、Controller。
- 对象语义明确，框架源码简单易懂，无注释部分变动可能较大。

## Lsit

功能列表，在未稳定前不会编写演示

- Application
- [x] 程序启动流程自定义实现，内置两种
- [x] server异步启动
- [ ] 完全配置化启动程序
- Context
- [x] Context与server完全解耦
- [x] 根据Content-type自动序列化数据
- [x] 根据Accept反序列化数据
- [x] http2 push
- Config
- [x] 配置库集成
- [x] 支持基于路径读写对象
- [ ] 解析参数环境变量和差异化配置
- [x] map和结构体相互转换
- [x] 数据类型转换工具
- [ ] 生成对象帮助信息
- Server
- [x] 支持net/http启动server
- [x] 支持fasthttp启动server
- [x] 重新实现http协议server
- [x] http2协议两种server支持
- [x] fastcgi协议两种server支持
- [x] websocket协议支持
- [ ] 重新实现websocket协议
- [x] 支持TLS和双向TLS
- [ ] 自动自签TLS证书
- [ ] http3协议学不动了
- [x] server热重启支持
- [x] server后台启动
- Router
- [x] 零内存复制快速路由器实现
- [x] 严格路由匹配顺序
- [x] RESTful风格基于方法的路由注册匹配
- [x] 基于Host进行路由注册匹配
- [x] 路由器注册初始化时请求处理
- [x] 组路由注册支持
- [x] 全局中间件注册
- [x] 组级中间件注册
- [x] api级中间件注册
- [x] 路由默认参数
- [ ] 默认参数匹配注册中间件
- [x] 路由匹配参数捕获
- [x] 路由匹配参数校验
- [ ] 路由匹配参数正则捕捉
- [x] 路由匹配通配符捕捉
- [x] 路由匹配通配符校验
- [ ] 路由匹配通配符正则捕捉
- Logger
- [x] 初始化期间日志处理
- [x] 日志条目属性支持
- [x] 自定义模板格式化
- [ ] 日志写入到es
- View
- [ ] 多模板库接入
- Mvc
- [x] mvc支持
- [x] 控制器函数输入参数
- Tools
- [x] 程序启动命令解析
- [x] 信号响应支持
- [ ] SRI值自动设置
- [x] 自动http2 push
- [x] http代理实现
- [ ] pprof支持
- [ ] expvar支持
- [ ] 运行时对象数据显示
- Session
- [x] Session实现
- Middleware
- [x] gzip压缩
- [x] 限流
- [x] 黑名单
- [x] 异常捕捉
- [x] 访问日志

## issue

setting 基于配置初始化对象未实现

fasthttp不支持多端口和hijack

handlefile处理304缓存

handlerproxy未支持101

组件debug日志

websocket未完善

client未实现

反射接口处理

## Component

| 组件名称 | 介绍 | 定义库 |
| ------------ | ------------ | ------------ |
| router-radix | 使用基数树实现标准功能路由器 | 内置 |
| router-full | 使用基数树实现完整功能路由器 | 内置 |
| router-init | 初始时使用的路由处理 | github.com/eudore/eudore/component/router/init |
| router-host | 匹配host路由到不同路由器处理 | github.com/eudore/eudore/component/router/host |
| logger-init | 初始化日志处理，保存日志由设置的日志对象处理 | 内置 |
| logger-std | 基础日志库实现 | 内置 |
| logger-elastic | 将日志直接输出到es中 | 未更新、github.com/eudore/eudore/component/eslogger |
| server-std | 使用net/http封装标准库Server | 内置 |
| server-eudore  | 使用protocol库封装Server | github.com/eudore/eudore/component/server/eudore |
| server-fasthttp | 使用fasthttp启动服务 | github.com/eudore/eudore/component/server/fasthttp |
| cache-map | 使用Sync.Map实现的缓存 | 内置 |
| cache-group | 使用前缀匹配的多缓存组合 | 内置 |
| config-map | 使用map[string]interface{}存储配置 | 内置 |
| config-eudore | 使用反射来操作结构体和map自由嵌套的配置对象 | 内置 |
| view-std | 使用标准库html/template渲染模板 | 内置、未测试 |

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

Application默认有两种实现Core和Eudore，Core只有启动函数，Eudore多一些辅助函数封装。

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

Server是eudore用于启动http服务的顶级接口对象之一。

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

router-std路由器支持组路由、组级中间件、路径参数、通配符参数、默认参数。

router-full路由器支持组路由、组级中间件、路径参数、通配符参数、默认参数、参数校验、通配符校验，未实现多参数正则捕捉。

可实现Router接口重写路由器。

`curl -XGET http://localhost:8088/api/v1/`

`curl -XGET http://localhost:8088/api/v1/get/eudore`

`curl -XGET http://localhost:8088/api/v1/set/eudore`


```golang
package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware/logger"
	"github.com/eudore/eudore/middleware/recover"
)

// eudore core
func main() {
	// 创建App
	app := eudore.NewCore()
	app.RegisterComponent("logger-std", &eudore.LoggerStdConfig{
		Std:	true,
		Level:	eudore.LogDebug,
		Format:	"json",
	})
	// 全局级请求处理中间件
	app.AddMiddleware(
		logger.NewLogger(eudore.GetRandomString).Handle,
	)

	// 创建子路由器
	// apiv1 := eudore.NewRouterClone(app.Router)
	apiv1 := app.Group("/api/v1 version:v1")
	// 路由级请求处理中间件
	apiv1.AddMiddleware(recover.RecoverFunc)
	{
		apiv1.GetFunc("/get/:name", handleget)
		// Api级请求处理中间件
		apiv1.AnyFunc("/*", handlepre1, handleparam)
	}
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
	ctx.WriteString("\nhandlepre1\n")
}
func handleparam(ctx eudore.Context) {
	ctx.WriteString(ctx.GetParam("*"))
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
	_ "github.com/eudore/eudore/component/router/init"
)

func main() {
	app := eudore.NewCore()
	app.RegisterComponent(eudore.ComponentRouterInitName, eudore.HandlerFunc(func(ctx eudore.Context){
		conn, err := eudore.UpgradeHttp(ctx) 
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
					break
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
