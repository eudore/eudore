package main

/*
eudore.ControllerSingleton控制器所有请求公用一个控制器，注意并发安全。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.AddController(new(mySingletonController))

	// 请求测试
	client := httptest.NewClient(app)
	var mybasepath = "/mysingleton/"
	client.NewRequest("GET", mybasepath).Do().CheckStatus(200).CheckBodyContainString("1")
	client.NewRequest("GET", mybasepath).Do().CheckStatus(200).CheckBodyContainString("2")
	client.NewRequest("GET", "/mysingleton/path/eudore").Do().CheckStatus(200).CheckBodyContainString("/path/eudore")
	client.NewRequest("GET", mybasepath).Do().CheckStatus(200).CheckBodyContainString("4")

	client.NewRequest("GET", "/2/my/singleton").Do().CheckStatus(200)
	client.NewRequest("GET", "/88/name/my/singleton").Do().CheckStatus(200)
	client.NewRequest("GET", "/").Do().CheckStatus(404)

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

type mySingletonController struct {
	eudore.ControllerSingleton
	visitor int64
}

// 每次初始化访问次数加一
func (ctl *mySingletonController) Init(ctx eudore.Context) error {
	ctl.visitor++
	return ctl.ControllerSingleton.Init(ctx)
}

// 返回访问次数
func (ctl *mySingletonController) Any() interface{} {
	return ctl.visitor
}

// 单例控制器Context对象必须要参数传入，Init保存Context会并发不安全。
func (ctl *mySingletonController) Path(ctx eudore.Context) interface{} {
	return ctx.Path()
}

func (ctl *mySingletonController) Help(ctx eudore.Context) {}

func (*mySingletonController) ControllerRoute() map[string]string {
	return map[string]string{
		// 修改Path方法的路由注册
		"Any":  " method=any",
		"Path": "/path/*",
		"Help": "",
	}
}
