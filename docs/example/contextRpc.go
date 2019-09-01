package main

import (
	"github.com/eudore/eudore"
)

/*
如果返回响应为： map[a:1 b:2]

需要指定请求header Accept，返回json需要添加Accept: application/json

测试命令：

curl 127.0.0.1:8088

curl -H 'Accept: application/json' 127.0.0.1:8088

*/

func main() {
	app := eudore.NewCore()
	app.AnyFunc("/*", func(ctx eudore.Context, request map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{
			"a": 1,
			"b": 2,
		}, nil
	})
	app.Listen(":8088")
	app.Run()
}
