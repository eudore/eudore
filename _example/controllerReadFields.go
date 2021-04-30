package main

/*
eudore.ControllerBase等默认控制器，会将可导出的非空属性复制一份给新的控制器。
这些只读属性不要修改，读写属性应该在init时完成数据的初始化。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewApp()
	app.AddController(&myFieldsController{
		Name: "eudore",
		Num:  10,
		Func: func() string {
			return "hello"
		},
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/my/fields/").Do().CheckStatus(200).Out()
	client.NewRequest("GET", "/my/fields/num").Do().CheckStatus(200).Out()
	client.NewRequest("GET", "/my/fields/name").Do().CheckStatus(200).Out()
	client.NewRequest("GET", "/file/data/2").Do().CheckStatus(200).Out()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

type myFieldsController struct {
	eudore.ControllerBase
	Name string
	Num  int
	Func func() string
}

func (ctl *myFieldsController) Any() {
	ctl.Info("myFieldsController Any")
}
func (ctl *myFieldsController) Get() interface{} {
	return "get myFieldsController " + ctl.Func()
}
func (ctl *myFieldsController) GetNum() interface{} {
	return ctl.Num
}

func (ctl *myFieldsController) GetName() interface{} {
	return ctl.Name
}

func (ctl *myFieldsController) GetInfoById() interface{} {
	return ctl.GetParam("id")
}
