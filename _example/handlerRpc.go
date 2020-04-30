package main

/*
rpc扩展函数要求函数类型为：func(eudore.Context, Request) (Response, error)
其中Request类型要求为map、struct、ptr之一，会调用ctx.Bind接续数据，Response会调用ctx.Render返回。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

type (
	// Request 定义一个请求结构
	Request struct {
		Name string `json:"name"`
		Num  int    `json:"num"`
	}
	// Response 定义一个响应结构
	Response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
)

func main() {
	app := eudore.NewApp()
	app.AnyFunc("/*", func(ctx eudore.Context, req Request) (Response, error) {
		ctx.Debugf("%#v", req)
		return Response{200, "Success"}, nil
	})

	// 请求测试
	client := httptest.NewClient(app)
	client.NewRequest("PUT", "/").WithBodyJSON(map[string]interface{}{
		"name": "eudore",
		"num":  44,
	}).WithHeaderValue("Accept", " application/json").Do().Out()

	app.CancelFunc()
	app.Run()
}
