package main

/*
/eudore/debug/show/ 先注册的对象
/eudore/debug/show/app 显示app对象的数据
/eudore/debug/show/app/router 显示app的router的数据
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/show"
)

func main() {
	app := eudore.NewCore()
	show.RoutesInject(app.Group("/eudore/debug"))
	show.RegisterObject("app", app.App)

	app.Listen(":8088")
	app.Run()
}
