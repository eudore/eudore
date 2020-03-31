package main

/*
eudore.App默认Logger为LoggerInit对象，会保存全部日志信息，在调用NextHandler方法时，将全部日志输出给新日志。

LoggerInit意义是将配置解析之前，未设置Logger的日子全部保存起来，来初始化Logger后处理之前的日志，在调用SetLevel方法后，在NextHandler方法会传递日志级别。

如果修改Logger、Server、Router后需要调用Set方法重写，设置目标的输出函数。
*/

import (
	"github.com/eudore/eudore"
)

type loggerInitHandler interface {
	NextHandler(eudore.Logger)
}

func main() {
	app := eudore.NewCore()
	app.Info(1)
	app.Info(2)
	app.Info(3)
	app.AnyFunc("/*path", eudore.HandlerEmpty)

	// 判断是LoggerInit
	if initlog, ok := app.Logger.(loggerInitHandler); ok {
		// 创建日志
		app.Logger = eudore.NewLoggerStd(nil)
		// 新日志处理LoggerInit保存的日志。
		initlog.NextHandler(app.Logger)
	}
	app.Logger.Sync()
}
