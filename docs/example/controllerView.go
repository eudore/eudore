package main

/*
View控制器需要Renderer支持html模板渲染，默认渲染模板路径见路由注册中的template参数。

可以通过实现interface{GetViewTemplate(string, string) string}接口,修改返回的模板路径。

如果Data数据长度不等于0且未写入数据，在Release时会自动Render返回数据。

详细过程请看日志。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

type myUserController struct {
	eudore.ControllerView
}

func (*myUserController) Any() {}
func (ctl *myUserController) Get() {
	ctl.SetTemplate("index.html")
}
func (*myUserController) GetInfoByIdName() {}
func (*myUserController) GetIndex()        {}
func (*myUserController) GetContent()      {}

func (ctl *myUserController) Release() (err error) {
	ctl.Data["method"] = ctl.Method()
	return ctl.ControllerView.Release()
}

func main() {
	app := eudore.NewCore()
	// 支持渲染模板
	app.Renderer = eudore.NewHTMLWithRender(app.Renderer, nil)
	app.AddController(new(myUserController))

	// 请求测试
	client := httptest.NewClient(app)
	// 请求必须是Accept: text/html 这样才会渲染模板
	client.NewRequest("GET", "/myuser/").WithHeaderValue("Accept", "text/html").Do().CheckStatus(200)
	client.NewRequest("PUT", "/myuser/").WithHeaderValue("Accept", "text/html").Do().CheckStatus(200)
	for client.Next() {
		app.Error(client.Error())
	}

	app.Listen(":8088")
	app.Run()
}
