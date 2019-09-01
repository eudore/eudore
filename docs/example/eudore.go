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
*/

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewEudore()
	app.RegisterInit("init-router", 0x015, func(app *eudore.Eudore) error {
		app.GetFunc("/*", func(ctx eudore.Context) {
			ctx.WriteString("hello eudore")
		})
		return nil
	})
	app.RegisterInit("init-listen", 0x016, func(app *eudore.Eudore) error {
		app.Listen(":8088")
		return nil
	})
	app.Run()
}
