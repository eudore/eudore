# Middleware

Middleware包实现基础eudore请求中间件。

| Index  |  Name |  Type | 描述 |  备注 |
| ------------ | ------------ | ------------ | ------------ | ------------ |
| 01 |  [Admin](#Admin) | 调试  | 相关组件管理后台 | [example](../_example/middlewareAdmin.go)  |
| 02 |  [BasicAuth](#BasicAuth) |  拦截 | basic认证  | [example](../_example/middlewareBasicAuth.go) [nethttp](../_example/nethttpBasicAuth.go)  |
| 03 |  [BodyLimit](#BodyLimit) | 拦截 | 限制请求body大小 | [example](../_example/middlewareBodyLimit.go)  |
| 04 |  [Black](#Black) |  拦截 |  黑白名单 | [example](../_example/middlewareBlack.go) [nethttp](../_example/nethttpBalck.go) api |
| 05 |  [Breaker](#Breaker) |  拦截 |  熔断器 | [example](../_example/middlewareBreaker.go) api groups |
| 06 |  [Cache](#Cache) |  拦截 |  请求缓存 | [example](../_example/middlewareCache.go) [example2](../_example/middlewareCacheStore.go) groups |
| 07 |  [Compress](#Compress) | 辅助 | 响应压缩| [example](../_example/middlewareCompress.go) |
| 08 |  [ContextWarp](#ContextWarp) |  辅助 |  封装Context | [example](../_example/middlewareContextWarp.go)  |
| 09 |  [Cors](#Cors) |  拦截 |  跨域处理 | [example](../_example/middlewareCors.go)  |
| 10 |  [Csrf](#Csrf) |  拦截 |  CSRF token检查 | [example](../_example/middlewareCsrf.go)  |
| 11 |  [Dump](#Dump) |  调试 |  捕捉请求信息 | [example](../_example/middlewareDump.go) api  |
| 12 |  [Header](#Header) |  追加 |  添加响应header信息 |  [example](../_example/middlewareHeader.go) |
| 13 |  [HeaderFilte](#HeaderFilte) |  追加 |   过滤外部请求header |  [example](../_example/middlewareHeaderFilte.go) |
| 14 |  [Logger](#Logger) |  追加 |  输出access日志 | [example](../_example/middlewareLogger.go)  |
| 15 |  [LoggerLevel](#LoggerLevel) |  辅助 |  请求设置独立日志级别 | [example](../_example/middlewareLoggerLevel.go)  |
| 16 |  [Look](#Look)  |  调试 |  路径访问对象 | [example](../_example/middlewareLook.go)  |
| 17 |  [Pprof](#Pprof) |  调试 |  处理pprof响应 |  [example](../_example/middlewarePprof.go) |
| 18 |  [Rate](#Rate)  |  拦截 | 限速限流  | [限流](../_example/middlewareRateRequest.go) [限速](../_example/middlewareRateSpeed.go) [nethttp限流](../_example/netttpRateRequest.go) groups |
| 19 |  [Recover](#Recover) |  追加 | 恢复panic  | [example](../_example/middlewareRecover.go)  |
| 20 |  [Referer](#Referer) |  拦截 | referer校验  | [example](../_example/middlewareReferer.go)  |
| 21 |  [RequestID](#RequestID) |  追加 | 增加请求id  | [example](../_example/middlewareRequestID.go)  |
| 22 |  [Rewrite](#Rewrite) |  辅助 |  请求路径修改 | [example](../_example/middlewareRewrite.go) [nethttp](../_example/nethttpRewrite.go) |
| 23 |  [Router](#Router)  |  辅助 |  路由自定义处理 | [example](../_example/middlewareRouter.go)  |
| 24 |  [RouterRewrite](#RouterRewrite)  |  辅助 | 重写请求路径 | [example](../_example/middlewareRouterRewrite.go) |
| 25 |  [Timeout](#Timeout) |  其他 | 处理请求超时 | [example](../_example/middlewareTimeout.go)  |
| 26 |    | 其他 | 自定义中间件处理函数  | [example](../_example/middlewareHandle.go) |
| 27 |  Policy  |  其他 | Pbac | [example](../_example/policyPbac.go) |
| 28 |  Promethues  |  其他 | prometheus采集请求信息 |   |
| 29 |  OpenTelemetry |  其他 | otel记录请求信息<br>注入tracer |   |


## BasicAuth

实现请求BasicAuth访问认证

参数:
- map[string]string    允许的用户名和密码的键值对map。

example:

`app.AddMiddleware(middleware.NewBasicAuthFunc(map[string]string{"user": "pw"}))`

## BodyLimit

限制请求body大小

参数:
- int64       指定限制body的长度

examole:

`app.AddMiddleware(middleware.NewBodyLimitFunc(32 << 20))`

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

实现路由规则熔断

参数:
- eudore.Router
属性:
- MaxConsecutiveSuccesses uint32                   最大连续成功次数
- MaxConsecutiveFailures  uint32                   最大连续失败次数
- OpenWait                time.Duration            打开状态恢复到半开状态下等待时间
- NewHalfOpen             func(string) func() bool 创建一个路由规则半开状态下的限流函数

example:
```
app.AddMiddleware(middleware.NewBreakerFunc(app.Group("/eudore/debug")))

breaker := middleware.NewBreaker()
breaker.OpenWait = 0
app.AddMiddleware(breaker.NewBreakerFunc(app.Group("/eudore/debug")))
```

在关闭状态下连续错误一定次数后熔断器进入半开状态；在半开状态下请求将进入限流状态，半开连续错误一定次数后进入打开状态，半开连续成功一定次数后回到关闭状态；在进入关闭状态后等待一定时间后恢复到半开状态。

## Cache

创建一个缓存中间件，对Get请求具有缓存和SingleFlight双重效果。

参数：
- context.Context	控制默认cacheMap清理过期数据的生命周期
- time.Duration	请求数据缓存时间，默认秒
- cacheStore	缓存存储对象

example:

`app.AddMiddleware(middleware.NewCacheFunc(time.Second*10, app.Context))`

## Compress

创建响应压缩中间件，默认提供gzip和deflate压缩
参数:
- string	压缩名称
- func() any 压缩器创建函数
-	int	压缩级别

example:
```golang
import: "github.com/andybalholm/brotli"
app.AddMiddleware(middleware.NewCompressMixinsFunc(nil))
app.AddMiddleware(middleware.NewCompressFunc("br", func() any { return brotli.NewWriter(ioutil.Discard) }))
app.AddMiddleware(middleware.NewCompressGzipFunc())
app.AddMiddleware(middleware.NewCompressDeflateFunc())
```

## ContextWarp

使中间件之后的处理函数使用的eudore.Context对象为新的Context

参数:
- func(eudore.Context) eudore.Context    指定ContextWarp使用的eudore.Context封装函数

example:

```golang
app.AddMiddleware(middleware.NewContextWarpFunc(newContextParams))
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
app.AddMiddleware("global", middleware.NewCorsFunc([]string{"www.*.com", "example.com", "127.0.0.1:*"}, map[string]string{
	"Access-Control-Allow-Credentials": "true",
	"Access-Control-Allow-Headers":     "Authorization,DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,X-Parent-Id",
	"Access-Control-Expose-Headers":    "X-Request-Id",
	"access-control-allow-methods":     "GET, POST, PUT, DELETE, HEAD",
	"access-control-max-age":           "1000",
}))
```

Cors中间件注册不是全局中间件时，需要最后注册一次Options /\*或404方法，否则Options请求匹配了默认404没有经过Cors中间件处理。

## Csrf

校验设置CSRF token

参数:
- any    指明获取csrf token的方法，下列是允许使用的值
	- "csrf"
	- "query: csrf"
	- "header: X-CSRF-Token"
	- "form: csrf"
	- func(ctx eudore.Context) string {return ctx.Query("csrf")}
	- nil
- any    指明设置Cookie的基础信息，下列是允许使用的值
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


## Header

添加响应Header

参数:
- http.Header     需要添加的Header内存

examaple:
```
app.AddMiddleware(middleware.NewHeaderFunc(http.Header{
	"Cache-Control": []string{"no-cache"},
}))
app.AddMiddleware(middleware.NewHeaderWithSecureFunc(nil))
```

## HeaderFilte

对来源于外部请求，过滤指定请求header

参数:
- []string	指定内部ip，默认[]string{"10.0.0.0/8", "172.16.0.0/12", "192.0.0.0/24", "127.0.0.1"}
- []string	指定需要过滤的请求header，默认[]string{HeaderXRealIP, HeaderXForwardedFor, HeaderXForwardedHost, HeaderXForwardedProto, HeaderXRequestID, HeaderXTraceID}

examaple:
```
	app.AddMiddleware(middleware.NewHeaderFilteFunc(nil, nil))
	app.AddMiddleware(middleware.NewHeaderFilteFunc([]string{"127.0.0.1"}, nil))
```

## Logger

输出请求access logger并记录相关fields

参数:
- eudore.App    指定App对象，需要使用App.Logger输出日志。
- ...string     指定额外添加的Params值，如果值非空则会加入到access logger fields中

example:

`app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))`

## Rate

实现请求令牌桶限流/限速

参数:
- int       每周期(默认秒)增加speed个令牌
- int       最多拥有的令牌数量
- ...any    额外使用的Options,根据类型来断言设置选项
	context.Context               =>    控制cleanupVisitors退出的生命周期
	time.Duration                 =>    基础时间周期单位，默认秒
	func(eudore.Context) string   =>    限流获取key的函数，默认Context.ReadIP

example:
```
// 限流 每秒一个请求，最多保存3个请求
app.AddMiddleware(middleware.NewRateRequestFunc(1, 3, app.Context))
// 限速 每秒32Kb流量，最多保存128Kb流量
app.AddMiddleware(middleware.NewRateSpeedFunc(32*1024, 128*1024, app.Context))
```

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

## RequestID

给请求、响应、日志设置一个请求ID

参数:
- func() string		用于创建一个请求ID，默认使用时间戳随机数

example:
```
app.AddMiddleware(middleware.NewRequestIDFunc(nil))
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
- map[string]any    请求路径对应的执行函数，路径前缀不指定方法则为Any方法
example:
```
app.AddMiddleware(middleware.NewRouterFunc(map[string]any{
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

## Timeout

设置请求处理超时时间，如果超时返回503状态码并取消context，

实现难点：写入中超时状态码异常、panic栈无法捕捉信息异常、http.Header并发读写、sync.Pool回收了Context、Context数据竟态检测

## Policy 

goto [github.com/eudore/eudore/policy](../policy)

## Prometheus

goto [github.com/eudore/endpoint/prometheus](https://github.com/eudore/endpoint/tree/master/prometheus)

## Opentracing

goto [github.com/eudore/endpoint/opentracing](https://github.com/eudore/endpoint/tree/master/tracer)

# 不将实现中间件及原因：
- Casbin 接入太简单不具有技术含量，自行添加判断逻辑；不支持pbac实现。
- Jwt 无明显效果，不如Context扩展实现相关功能。
- Session 无明显效果，不如Context扩展实现相关功能。
- Timing 核心入侵大，不如Trace。
