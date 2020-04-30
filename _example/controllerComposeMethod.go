package main

/*
如果控制器嵌入一个名称为Controller为前缀的属性，该对象的全部方法不会自动注册路由，否在可以嵌入获得该对象的方法注册成路由。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

type (
	// Controllerwebsite 是基础方法的控制器
	Controllerwebsite struct {
		eudore.ControllerData
	}
	myMethodController struct {
		// Controllerwebsite 因名称前缀为Controller所以不会注册Hello方法为路由。
		Controllerwebsite
	}
)

func main() {
	app := eudore.NewApp()
	app.AddController(new(myMethodController))

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/mymethod/").Do().Out()
	for client.Next() {
		app.Error(client.Error())
	}

	app.CancelFunc()
	app.Run()
}

// Hello 方法返回heelo
func (ctl *Controllerwebsite) Hello() string {
	return "hello eudore"
}

// Any 方法处理控制器全部请求。
func (ctl *myMethodController) Any() {
	ctl.Debug("myMethodController Any", ctl.Hello())
}
