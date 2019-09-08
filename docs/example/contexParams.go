package main

/*
Params相关方法定义。
type Context interface {
	Params() Params
	GetParam(string) string
	SetParam(string, string)
	AddParam(string, string)
	...
}

type Params interface {
	GetParam(string) string
	AddParam(string, string)
	SetParam(string, string)
}
*/

import (
	"fmt"
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewCore()
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		fmt.Println(ctx.GetParam("route"))
	})
	app.Listen(":8088")
	app.Run()
}
