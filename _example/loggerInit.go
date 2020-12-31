package main

/*
eudore.App可以设置Logger为LoggerInit对象，会保存全部日志信息，在调用NextHandler方法时，将全部日志输出给新日志器。

LoggerInit意义是将配置解析之前，未设置Logger的日子全部保存起来，来初始化Logger后处理之前的日志，在调用SetLevel方法后，在NextHandler方法会传递日志级别。

如果修改Logger、Server、Router后需要调用Set方法重写，设置目标的输出函数。
*/

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewApp(eudore.NewLoggerInit())
	app.Debug("debug")
	app.Info("info")
	app.Warning("warning")
	app.Error("error")
	app.Sync()

	logout := app.WithField("caller", "mylogout").WithFields(nil)
	logout.WithField("level", "debug").Debug("debug")
	logout.WithField("level", "info").Info("info")
	logout.WithField("level", "warning").Warning("warning")
	logout.WithField("level", "error").Error("error")

	app.AnyFunc("/*path", eudore.HandlerEmpty)
	app.Options(eudore.NewLoggerStd(nil))
	app.WithField("depth", "enable").Info("info")
	app.CancelFunc()
	app.Run()
}
