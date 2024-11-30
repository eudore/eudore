package main

import (
	"github.com/eudore/eudore"
)

// Request 定义一个请求结构
type Request struct {
	Name string `json:"name"`
	Num  int    `json:"num"`
}

// Response 定义一个响应结构
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func main() {
	app := eudore.NewApp()
	app.AddHandlerExtend(func(fn func(ContextCustom)) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			fn(ContextCustom{ctx})
		}
	})
	app.GetFunc("/ctx", func(ctx ContextCustom) {
		ctx.WriteString("ctx")
	})

	// 非固定转换函数 func NewHandlerAnyContextTypeAnyError(fn any) HandlerFunc
	// 在创建时，通过any类型拦截到全部对象，再判断函数出入参数，匹配时返回HandlerFunc。
	app.AnyFunc("/rpc/*", func(ctx eudore.Context, req Request) (Response, error) {
		ctx.Debugf("%#v", req)
		return Response{200, "Success"}, nil
	})

	app.Listen(":8088")
	app.Run()
}

type contextBase = eudore.Context
type ContextCustom struct {
	contextBase
}
