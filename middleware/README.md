# Middleware

Middleware包实现部分基础eudore请求中间件。

```golang
func InitMidd(app *eudore.Eudore) error {
	app.AddMiddleware(
		// add logger middleware
		middleware.NewLoggerFunc(app, "route"),
		// 熔断器
		middleware.NewCircuitBreaker(app.Group("/eudore/debug/breaker")).NewBreakFunc(),
		// 处理超时
		middleware.NewTimeoutFunc(10 * time.Second),
		// cors
		middleware.NewCorsFunc(nil, map[string]string{
			"Access-Control-Allow-Credentials": "true",
			"Access-Control-Allow-Headers": "Authorization,DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,X-Parent-Id",
			"Access-Control-Expose-Headers": "X-Request-Id",
			"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, HEAD",
			"Access-Control-Max-Age": "1000",
		}),
		// black list
		// 黑名单
		middleware.NewDenialsFunc(app.Cache, 72*time.Hour),
		// 流控
		middleware.NewRateFunc(app, 10, 30),
		middleware.NewBasicAuthFunc("", map[string]string{
			"root": "111",
		}),
		// gzip压缩
		middleware.NewGzipFunc(5),
		// 捕捉panic
		middleware.NewRecoverFunc(),
	)
}
```