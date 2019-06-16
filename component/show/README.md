# 数据显示工具

改工具用于显示运行时程序内部数据，目前序列化方式未决定。

示例：

```golang
package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/show"
)

func main() {
	app := eudore.NewCore()
	show.Inject(app.Group("/eudore/debug"))
	show.RegisterObject("app", app.App)
	
	app.Listen(":8088")
	app.Run()
}
```

# 其他

`show.Inject`方法会给路由器注入库使用的路由，示例中访问

`http://localhost:8088/eudore/debug/show/` 显示全部注册的对象的key

`http://localhost:8088/eudore/debug/show/app` 显示注册的key为app的属性

`http://localhost:8088/eudore/debug/show/app/router` 显示app的router属性

注入方法实现：

```golang
func Inject(r eudore.RouterMethod) {
	r = r.Group("/show")
	r.GetFunc("/", List)
	r.GetFunc("/*key", Showkey)
}
```