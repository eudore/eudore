package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

type MyBaseController struct {
	eudore.ControllerSingleton
	visitor uint64
}

// 每次初始化访问次数加一
func (ctl *MyBaseController) Init(ctx eudore.Context) error {
	ctl.visitor++
	return nil
}

// 返回访问次数
func (ctl *MyBaseController) Any() interface{} {
	return ctl.visitor
}

// 单例控制器Context对象必须要参数传入，Init保存Context会并发不安全。
func (ctl *MyBaseController) Path(ctx eudore.Context) interface{} {
	return ctx.Path()
}

func (*MyBaseController) ControllerRoute() map[string]string {
	return map[string]string{
		// 修改Path方法的路由注册
		"Path": "/path/*",
	}
}

func main() {
	app := eudore.NewCore()
	app.AddController(new(MyBaseController))

	// 请求测试
	client := httptest.NewClient(app)
	client.NewRequest("GET", "/mybase/").Do().CheckStatus(200).CheckBodyContainString("1")
	client.NewRequest("GET", "/mybase/").Do().CheckStatus(200).CheckBodyContainString("2")
	client.NewRequest("GET", "/mybase/path/eudore").Do().CheckStatus(200).CheckBodyContainString("/path/eudore")
	client.NewRequest("GET", "/mybase/").Do().CheckStatus(200).CheckBodyContainString("4")
	client.NewRequest("GET", "/").Do().CheckStatus(200)
	for client.Next() {
		app.Error(client.Error())
	}

	app.Listen(":8088")
	app.Run()
}
