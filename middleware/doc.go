/*
Package middleware 包实现eudore基础请求中间件和处理函数。

BasicAuth

实现请求BasicAuth访问认证

参数:
	map[string]string    允许的用户名和密码的键值对map。
example:
	app.AddMiddleware(middleware.NewBasicAuthFunc(map[string]string{"user": "pw"}))


Black

实现黑白名单管理及管理后台

参数:
	map[string]bool    指明初始化使用的黑白名单，true为白白名单/false为黑名单
	eudore.Router      为注入黑名单管理路由的路由器。
example:
	app.AddMiddleware(middleware.NewBlackFunc(map[string]bool{
		"192.168.100.0/24": true,
		"192.168.75.0/30":  true,
		"192.168.1.100/30": true,
		"127.0.0.1/32":     true,
		"10.168.0.0/16":    true,
		"0.0.0.0/0":        false,
	}, app.Group("/eudore/debug")))

Breaker

实现路由规则熔断

参数:
	eudore.Router
属性:
	MaxConsecutiveSuccesses uint32                   最大连续成功次数
	MaxConsecutiveFailures  uint32                   最大连续失败次数
	OpenWait                time.Duration            打开状态恢复到半开状态下等待时间
	NewHalfOpen             func(string) func() bool 创建一个路由规则半开状态下的限流函数

example:

	app.AddMiddleware(middleware.NewBreakerFunc(app.Group("/eudore/debug")))

	breaker := middleware.NewBreaker()
	breaker.OpenWait = 0
	app.AddMiddleware(breaker.NewBreakerFunc(app.Group("/eudore/debug")))

在关闭状态下连续错误一定次数后熔断器进入半开状态；在半开状态下请求将进入限流状态，半开连续错误一定次数后进入打开状态，半开连续成功一定次数后回到关闭状态；在进入关闭状态后等待一定时间后恢复到半开状态。

Cache

创建一个缓存中间件，对Get请求具有缓存和SingleFlight双重效果。

参数：
	context.Context	控制默认cacheMap清理过期数据的生命周期
	time.Duration	请求数据缓存时间，默认秒
	cacheStore	缓存存储对象
example:
	app.AddMiddleware(middleware.NewCacheFunc(time.Second*10, app.Context))

ContextWarp

使中间件之后的处理函数使用的eudore.Context对象为新的Context

参数:
	func(eudore.Context) eudore.Context    指定ContextWarp使用的eudore.Context封装函数
example:
	app.AddMiddleware(middleware.NewContextWarpFunc(newContextParams))
	func newContextParams(ctx eudore.Context) eudore.Context {
		return contextParams{ctx}
	}

Cors

跨域请求

参数:
	[]string             允许使用的origin，默认值为:[]string{"*"}
	map[string]string    CORS验证通过后给请求添加的协议headers，用来设置CORS控制信息
example:
	app.AddMiddleware("global", middleware.NewCorsFunc([]string{"www.*.com", "example.com", "127.0.0.1:*"}, map[string]string{
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Headers":     "Authorization,DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,X-Parent-Id",
		"Access-Control-Expose-Headers":    "X-Request-Id",
		"access-control-allow-methods":     "GET, POST, PUT, DELETE, HEAD",
		"access-control-max-age":           "1000",
	}))

Cors中间件注册不是全局中间件时，需要最后注册一次Options /*或404方法，否则Options请求匹配了默认404没有经过Cors中间件处理。

Csrf

校验设置CSRF token

参数:
	interface{}    指明获取csrf token的方法，下列是允许使用的值
		- "csrf"
		- "query: csrf"
		- "header: X-CSRF-Token"
		- "form: csrf"
		- func(ctx eudore.Context) string {return ctx.Query("csrf")}
		- nil
	interface{}    指明设置Cookie的基础信息，下列是允许使用的值
		- "csrf"
		- http.Cookie{Name: "csrf"}
		- nil
example:
	app.AddMiddleware(middleware.NewCsrfFunc("csrf", nil))

Dump

截取请求信息的中间件，将匹配请求使用webscoket输出给客户端。

参数:
	router参数是eudore.Router类型，然后注入拦截路由处理。
example:
	app.AddMiddleware(middleware.NewDumpFunc(app.Group("/eudore/debug")))

Gzip

对请求响应body使用gzip压缩

参数:
	int    gzip压缩等级，非法值设置为5
example:
	app.AddMiddleware(middleware.NewGzipFunc(5))

Logger

输出请求access logger并记录相关fields

参数:
	eudore.App    指定App对象，需要使用App.Logger输出日志。
	...string     指定额外添加的Params值，如果值非空则会加入到access logger fields中
example:
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))

Rate

实现请求令牌桶限流

参数:
	int               每周期(默认秒)增加speed个令牌
	int               最多拥有的令牌数量
	...interface{}    额外使用的Options,根据类型来断言设置选项
		context.Context               =>    控制cleanupVisitors退出的生命周期
		time.Duration                 =>    基础时间周期单位，默认秒
		func(eudore.Context) string   =>    限流获取key的函数，默认Context.ReadIP
example:
	// 限流 每秒一个请求，最多保存3个请求
	app.AddMiddleware(middleware.NewRateRequestFunc(1, 3, app.Context))
	// 限速 每秒32Kb流量，最多保存128Kb流量
	app.AddMiddleware(middleware.NewRateSpeedFunc(32*1024, 128*1024, app.Context))

Recover

恢复panic抛出的错误，并输出日志、返回异常响应

example:
	app.AddMiddleware(middleware.NewRecoverFunc())

Referer

检查请求Referer Header值是否有效

参数:
	map[string]bool    设置referer值是否有效
		""                         =>    其他值未匹配时使用的默认值。
		"origin"                   =>    请求Referer和Host同源情况下，检查host为referer前缀，origin检查在其他值检查之前。
		"*"                        =>    任意域名端口
		"www.eudore.cn/*"          =>    www.eudore.cn域名全部请求，不指明http或https时为同时包含http和https
		"www.eudore.cn/api/*"      =>    www.eudore.cn域名全部/api/前缀的请求
		"https://www.eudore.cn/*"  =>    www.eudore.cn仅匹配https。
example:
	app.AddMiddleware(middleware.NewRefererFunc(map[string]bool{
		"":                         true,
		"origin":                   false,
		"www.eudore.cn/*":          true,
		"www.eudore.cn/api/*":      false,
		"www.example.com/*":        true,
	}))

RequestID

给请求、响应、日志设置一个请求ID

参数:
	func() string     用于创建一个请求ID，默认使用时间戳随机数
example:
	app.AddMiddleware(middleware.NewRequestIDFunc(nil))

Rewrite

重写请求路径，需要注册全局中间件

参数:
	map[string]string    请求匹配模式对应的目标模式
example:
	app.AddMiddleware("global", middleware.NewRewriteFunc(map[string]string{
		"/js/*":          "/public/js/$0",
		"/d/*":           "/d/$0-$0",
		"/api/v1/*":      "/api/v3/$0",
		"/api/v2/*":      "/api/v3/$0",
		"/help/history*": "/api/v3/history",
		"/help/history":  "/api/v3/history",
		"/help/*":        "$0",
	}))

Router

用于执行额外的路由匹配行为

参数:
	map[string]interface{}    请求路径对应的执行函数，路径前缀不指定方法则为Any方法
example:
	app.AddMiddleware(middleware.NewRouterFunc(map[string]interface{}{
		"/api/:v/*": func(ctx eudore.Context) {
			ctx.Request().URL.Path = "/api/v3/" + ctx.GetParam("*")
		},
		"GET /api/:v/*": func(ctx eudore.Context) {
			ctx.WriteHeader(403)
			ctx.End()
		},
	}))

RouterRewrite

基于Router中间件实现路由重写，参考Rewrite

example:
	app.AddMiddleware("global", middleware.NewRouterRewriteFunc(map[string]string{
		"/js/*":          "/public/js/$0",
		"/d/*":           "/d/$0-$0",
		"/api/v1/*":      "/api/v3/$0",
		"/api/v2/*":      "/api/v3/$0",
		"/help/history*": "/api/v3/history",
		"/help/history":  "/api/v3/history",
		"/help/*":        "$0",
	}))

Timeout

设置请求处理超时时间，如果超时返回503状态码并取消context，

实现难点：写入中超时状态码异常、panic栈无法捕捉信息异常、http.Header并发读写、sync.Pool回收了Context、Context数据竟态检测

*/
package middleware // import "github.com/eudore/eudore/middleware"

import (
	"github.com/eudore/eudore"
	"runtime"
)

var adminStaticFile string

// 获取文件定义位置，静态ui文件在同目录。
func init() {
	_, file, _, ok := runtime.Caller(0)
	if ok {
		adminStaticFile = file[:len(file)-6] + "admin.html"
	}
}

// HandlerAdmin 函数返回Admin UI界面。
func HandlerAdmin(ctx eudore.Context) {
	ctx.SetHeader("X-Eudore-Admin", "ui")
	ctx.WriteFile(adminStaticFile)
}
