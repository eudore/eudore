package main

/*
如果控制器嵌入一个名称后缀为Controller的控制器，会获得该对象的方法注册成路由。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

// Controllerwebsite 是基础方法的控制器
type Controllerwebsite struct {
	eudore.ControllerAutoRoute
}
type methodController struct {
	// Controllerwebsite 因名称后缀不为Controller所以不会注册Hello方法为路由。
	Controllerwebsite
}
type tableController struct {
	eudore.ControllerAutoRoute
}

// routeController 从tableController嵌入两个方法注册成路由。
type routeController struct {
	tableController
}

func main() {
	app := eudore.NewApp()
	app.AddController(new(methodController))
	app.AddController(new(routeController))

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/method/").Do().Out()
	client.NewRequest("GET", "/route/hello").Do().Out()
	client.NewRequest("PUT", "/route/").Do().Out()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

// Hello 方法返回heelo
func (ctl *Controllerwebsite) Hello() string {
	return "hello eudore"
}

// Get 方法不会被组合
func (ctl *Controllerwebsite) Get(ctx eudore.Context) {}

// Any 方法处理控制器全部请求。
func (ctl *methodController) Any(ctx eudore.Context) {
	ctx.Debug("methodController Any", ctl.Hello())
}

func (ctl *tableController) Hello() interface{} {
	return "hello eudore"
}

func (ctl *tableController) Any(ctx eudore.Context) {
	ctx.Debug("tableController Any", ctl.Hello())
}
