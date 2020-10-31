package main

/*
ControllerAutoRoute将控制器方法转换成处理函数注册。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewApp()
	app.AddController(new(myAutoController))

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/index").Do()
	client.NewRequest("GET", "/info/22").Do()
	client.NewRequest("POST", "/").Do()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

type myAutoController struct {
	eudore.ControllerAutoRoute
}

// Any 方法注册 Any /*路径。
func (*myAutoController) Any(ctx eudore.Context) {
	ctx.Info("myAutoController Any")
}

// Get 方法注册 Get /*路径。
func (*myAutoController) Get(ctx eudore.Context) interface{} {
	ctx.Debug("get")
	return "get myAutoController"
}

// GetInfoById 方法注册GET /info/:id 路由路径。
func (*myAutoController) GetInfoById(ctx eudore.Context) interface{} {
	return ctx.GetParam("id")
}

// Error 方法触发路由注册报错
func (*myAutoController) Error(myAutoController) error {
	return nil
}

// String 方法返回控制器名称，响应Router.AddController输出的名称。
func (*myAutoController) String() string {
	return "hello.myAutoController"
}

// Help 方法定义一个控制器本身的方法。
func (*myAutoController) Help(ctx eudore.Context) {}

// ControllerRoute 方法返回控制器路由推导修改信息。
func (*myAutoController) ControllerRoute() map[string]string {
	return map[string]string{
		// 修改Path方法的路由注册,路径为空就是忽略改方法
		"Help":   "",
		"String": "",
		// 给Any方法自动生成的路径添加参数iaany
		"Any": " isany=1",
	}
}
