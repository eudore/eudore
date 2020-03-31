package main

/*
componse/show是一个测试性组件，可能发生变化，其目标是使用反射获取的app全部允许数据，方便调试查看实际数据。

使用show.RegisterObject方法注册一个需要显示的数据对象，如果访问注册的对象其中一个可导出的属性，可以使用url路径访问。

例如注册A对象的名称为aaa，访问A对象的B属性，访问url路径aaa/B即可，使用eudore.Get方法获取的子属性。

/eudore/debug/show/ 显示全部注册的对象，返回值是一个数组
/eudore/debug/show/app 显示app对象的数据
/eudore/debug/show/app/router 显示app.router的数据
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/component/show"
)

func main() {
	app := eudore.NewCore()
	show.RoutesInject(app.Group("/eudore/debug"))
	show.RegisterObject("app", app.App)

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/eudore/debug/show/").Do().OutBody()
	client.NewRequest("GET", "/eudore/debug/show/app/config").Do().OutBody()
	for client.Next() {
		app.Error(client.Error())
	}
	app.Run()
}
