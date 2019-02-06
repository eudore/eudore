package eudore_test

import (
	"eudore"
)

// eudore
func ExampleNew() {
	eudore.New().Run()
}

// router
func ExampleNewStdRouter() {
	r, _ := eudore.NewRouterStd(nil)
	// Or the path is /api/v1/*path
	// 或者路径是 /api/v1/*path
	r.AnyFunc("/api/v1/*", func(ctx eudore.Context) {
		ctx.WriteString(ctx.GetParam("*"))
	})
	r.GetFunc("/api/v1/info/:name action:showname version:v1", func(ctx eudore.Context){
		// Get route additional parameters and path parameters
		// 获取路由附加参数和路径参数
		ctx.WithField("version", ctx.GetParam("version")).Info("user name is: " + ctx.GetParam("name"))
	})
}