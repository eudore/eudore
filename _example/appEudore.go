package main

/*
eudore相对Core实现细节更多。

eudore使用以下格式创建程序,启动逻辑定义在初始化函数中。

func main() {
	app := eudore.NewEudore()
	app.RegisterInit("init-listen", 0x016, func(app *eudore.Eudore) error {	return nil	})
	app.Run()
}

RegisterInit第一参数是字符串类型定义唯一标识名称，第二参数是优先级，数值小的先执行，第三参数是执行函数，如果为空会删除以及注册的同名称函数。
在app.Run()时，先app.Config.Parse()解析配置，然后按优先级启动全部InitFunc并阻塞app。
当app.HandleError()方法处理一个error后，才会调用解除app阻塞并返回error。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewEudore()
	app.Set("workdir", ".")
	app.RegisterInit("init-router", 0x015, func(app *eudore.Eudore) error {
		app.GetFunc("/*", func(ctx eudore.Context) {
			ctx.WriteString("hello eudore")
		})
		return nil
	})
	app.RegisterInit("init-listen", 0x016, func(app *eudore.Eudore) error {
		httptest.NewClient(app).Stop(0)
		return app.Listen(":8088")
	})
	// 最后启动Server
	app.RegisterInit("eudore-server", 0xf0f, eudore.InitServer)
	app.Run()
}
