package main

/*
Cors中间件具有两个参数
第一个参数是一个字符串数组，保存全部运行的Origin。
第二个参数是一个map，保存option请求匹配后返回的额外Header。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewCore()
	app.AddMiddleware(middleware.NewCorsFunc([]string{"example.com"}, map[string]string{
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Headers":     "Authorization,DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,X-Parent-Id",
		"Access-Control-Expose-Headers":    "X-Request-Id",
		"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, HEAD",
		"Access-Control-Max-Age":           "1000",
	}))
	app.Listen(":8088")
	app.Run()
}
