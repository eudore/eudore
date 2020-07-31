# Middleware

Middleware包实现部分基础eudore请求中间件。

example:

```golang
func InitMiddleware(app *App) error {
	// admin
	admin := app.Group("/eudore/debug godoc=https://golang.org")
	admin.AddMiddleware(middleware.NewBasicAuthFunc("", map[string]string{
		"root": "111",
	}))
	pprof.Init(admin)
	admin.AnyFunc("/pprof/look/*", pprof.NewLook(app))
	admin.AnyFunc("/admin/ui", middleware.HandlerAdmin)

	// 增加全局中间件
	app.AddMiddleware(
		middleware.NewLoggerFunc(app.App, "route", "action", "ram", "basicauth", "resource", "browser", "sql"),
		middleware.NewDumpFunc(admin),
		middleware.NewBlackFunc(map[string]bool{
			"0.0.0.0/0":      false,
			"127.0.0.1/32":   true,
			"192.168.0.0/16": true,
			"172.0.0.0./8":   true,
		}, admin),
		middleware.NewRateFunc(10, 100, app),
		middleware.NewBreaker().InjectRoutes(admin).NewBreakFunc(),
		NewAddHeaderFunc(),
		middleware.NewTimeoutFunc(5*time.Second),
		middleware.NewCorsFunc(nil, map[string]string{
			"Access-Control-Allow-Credentials": "true",
			"Access-Control-Allow-Headers":     "Authorization,DNT,Keep-Alive,User-Agent,Cache-Control",
			"Access-Control-Expose-Headers":    "X-Request-Id",
			"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, HEAD",
			"Access-Control-Max-Age":           "1000",
		}),
		middleware.NewGzipFunc(5),
		middleware.NewRecoverFunc(),
	)
	// /api/v1/
	app.AnyFunc("/api/v1/*", eudore.HandlerRouter404)
	app.AddMiddleware(
		"/api/v1/",
		// 需要自行实现获取用户信息(jwt session)和权限控制(ram casbin)
		// NewUserInfoFunc(app),
		// app.RAM.NewRAMFunc(),
	)
	// 404 405
	app.AddHandler("404", "", eudore.HandlerRouter404)
	app.AddHandler("405", "", eudore.HandlerRouter405)
	return nil
}
```

## BasicAuth

实现请求BasicAuth访问认证

参数:
- map[string]string    允许的用户名和密码的键值对map。

example:
`app.AddMiddleware(middleware.NewBasicAuthFunc(map[string]string{"user": "pw"}))`


## Black

实现黑白名单管理及管理后台

参数:
- map[string]bool    指明初始化使用的黑白名单，true为白白名单/false为黑名单
- eudore.Router      为注入黑名单管理路由的路由器。

example:
```
app.AddMiddleware(middleware.NewBlackFunc(map[string]bool{
  "192.168.100.0/24": true,
  "192.168.75.0/30":  true,
  "192.168.1.100/30": true,
  "127.0.0.1/32":     true,
  "10.168.0.0/16":    true,
  "0.0.0.0/0":        false,
}, app.Group("/eudore/debug")))
```

## Breaker

重构中

## ContextWarp

使中间件之后的处理函数使用的eudore.Context对象为新的Context

参数:
- func(eudore.Context) eudore.Context    指定ContextWarp使用的eudore.Context封装函数

example:
```app.AddMiddleware(middleware.NewContextWarpFunc(newContextParams))
func newContextParams(ctx eudore.Context) eudore.Context {
  return contextParams{ctx}
}
```

## Cors

跨域请求

参数:
- []string             允许使用的origin，默认值为:[]string{"*"}
- map[string]string    CORS验证通过后给请求添加的协议headers，用来设置CORS控制信息

example:
```
app.AddMiddleware(middleware.NewCorsFunc([]string{"www.*.com", "example.com", "127.0.0.1:*"}, map[string]string{
	"Access-Control-Allow-Credentials": "true",
	"Access-Control-Allow-Headers":     "Authorization,DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,X-Parent-Id",
	"Access-Control-Expose-Headers":    "X-Request-Id",
	"access-control-allow-methods":     "GET, POST, PUT, DELETE, HEAD",
	"access-control-max-age":           "1000",
}))
```

## Csrf

校验设置CSRF token

参数:
- interface{}    指明获取csrf token的方法，下列是允许使用的值
	- "csrf"
	- "query: csrf"
	- "header: X-CSRF-Token"
	- "form: csrf"
	- func(ctx eudore.Context) string {return ctx.Query("csrf")}
	- nil
- interface{}    指明设置Cookie的基础信息，下列是允许使用的值
	- "csrf"
	- http.Cookie{Name: "csrf"}
	- nil

example:

`app.AddMiddleware(middleware.NewCsrfFunc("csrf", nil))`

## Dump

截取请求信息的中间件，将匹配请求使用webscoket输出给客户端。

参数:
- router参数是eudore.Router类型，然后注入拦截路由处理。

example:
`app.AddMiddleware(middleware.NewDumpFunc(app.Group("/eudore/debug")))`

## Gzip

对请求响应body使用gzip压缩

参数:
- int    gzip压缩等级，非法值设置为5

example:
`app.AddMiddleware(middleware.NewGzipFunc(5))`

## Logger

输出请求access logger并记录相关fields

参数:
- eudore.App    指定App对象，需要使用App.Logger输出日志。
- ...string     指定额外添加的Params值，如果值非空则会加入到access logger fields中

example:
`app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))`

## Rate

实现请求令牌桶限流

参数:
- int               每周期(默认秒)增加speed个令牌
- int               最多拥有的令牌数量
- ...interface{}    额外使用的Options,根据类型来断言设置选项
	context.Context               =>    控制cleanupVisitors退出的生命周期
	time.Duration                 =>    基础时间周期单位，默认秒
	func(eudore.Context) string   =>    限流获取key的函数，默认Context.ReadIP

example:
`app.AddMiddleware(middleware.NewRateFunc(1, 3, app.Context))`

## Recover

恢复panic抛出的错误，并输出日志、返回异常响应

example:
`app.AddMiddleware(middleware.NewRecoverFunc())`

## Referer

检查请求Referer Header值是否有效

参数:
- map[string]bool    设置referer值是否有效
	- ""                         =>    其他值未匹配时使用的默认值。
	- "origin"                   =>    请求Referer和Host同源情况下，检查host为referer前缀，origin检查在其他值检查之前。
	- "\*"                        =>    任意域名端口
	- "www.eudore.cn/*"          =>    www.eudore.cn域名全部请求，不指明http或https时为同时包含http和https
	- "www.eudore.cn/api/*"      =>    www.eudore.cn域名全部/api/前缀的请求
	- "https://www.eudore.cn/*"  =>    www.eudore.cn仅匹配https。

example:
```
app.AddMiddleware(middleware.NewRefererFunc(map[string]bool{
	"":                         true,
	"origin":                   false,
	"www.eudore.cn/*":          true,
	"www.eudore.cn/api/*":      false,
	"www.example.com/*":        true,
}))
```

## Rewrite

重写请求路径，需要注册全局中间件

参数:
- map[string]string    请求匹配模式对应的目标模式

example:
```
app.AddMiddleware("global", middleware.NewRewriteFunc(map[string]string{
	"/js/*":          "/public/js/$0",
	"/d/*":           "/d/$0-$0",
	"/api/v1/*":      "/api/v3/$0",
	"/api/v2/*":      "/api/v3/$0",
	"/help/history*": "/api/v3/history",
	"/help/history":  "/api/v3/history",
	"/help/*":        "$0",
}))
```

## Router

用于执行额外的路由匹配行为

参数:
- map[string]interface{}    请求路径对应的执行函数，路径前缀不指定方法则为Any方法
example:
```
app.AddMiddleware(middleware.NewRouterFunc(map[string]interface{}{
	"/api/:v/*": func(ctx eudore.Context) {
		ctx.Request().URL.Path = "/api/v3/" + ctx.GetParam("*")
	},
	"GET /api/:v/*": func(ctx eudore.Context) {
		ctx.WriteHeader(403)
		ctx.End()
	},
}))
```

## RouterRewrite

基于Router中间件实现路由重写，参考Rewrite

example:
```
app.AddMiddleware("global", middleware.NewRouterRewriteFunc(map[string]string{
	"/js/*":          "/public/js/$0",
	"/d/*":           "/d/$0-$0",
	"/api/v1/*":      "/api/v3/$0",
	"/api/v2/*":      "/api/v3/$0",
	"/help/history*": "/api/v3/history",
	"/help/history":  "/api/v3/history",
	"/help/*":        "$0",
}))
```

## SingleFlight

同时多次请求同一资源时，缓存一份处理结果返回给全部请求

example:
`app.AddMiddleware(middleware.NewSingleFlightFunc())`

## Timeout

设置请求处理超时时间，如果超时返回503状态码并取消context，

实现难点：写入中超时状态码异常、panic栈无法捕捉信息异常、http.Header并发读写、sync.Pool回收了Context、Context数据竟态检测


不将实现中间件及原因：
- BodyLimit 实现太简单不具有技术含量，自行重定义Request.Body。
- Casbin 实现太简单不具有技术含量，自行添加判断逻辑；不支持pbac实现。
- Jaeger 简单的全局中间件初始化sp效果太差，需要依赖Context.Logger完整封装。
- Jwt 无明显效果，不如Context扩展实现相关功能。
- RequestID 实现太简单不具有技术含量，自己添加第三方库生成ID加入Header。
- Secure 实现太简单不具有技术含量，自行添加Header。
- Session 无明显效果，不如Context扩展实现相关功能。