package main

/*
RpcMap是使用map作为rpc的请求和响应的存储结构实现rpc一样的方法。

要求函数类型为：func(eudore.Context, map[string]interface{}) (map[string]interface{}, error)

如果返回响应为：map[a:1 b:2]，需要指定请求header Accept，返回json需要添加Accept: application/json
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewApp()
	app.AnyFunc("/*", func(eudore.Context, map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{
			"a": 1,
			"b": 2,
		}, nil
	})

	// 请求测试
	client := httptest.NewClient(app)
	client.NewRequest("GET", "/").Do().CheckBodyString("map[a:1 b:2]").Out()
	client.NewRequest("GET", "/").WithHeaderValue("Accept", "application/json").Do().CheckHeader("Content-Type", "application/json; charset=utf-8").CheckBodyString(`{"a":1,"b":2}`).Out()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
