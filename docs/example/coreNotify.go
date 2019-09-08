package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/notify"
	"os"
)

func main() {
	app := eudore.NewCore()

	// 设置编译命令、启动命令、监听目录
	app.Config.Set("component.notify.buildcmd", "go build -o server coreNotify.go")
	app.Config.Set("component.notify.startcmd", "./server")
	app.Config.Set("component.notify.watchdir", ".")
	notify.NewNotify(app.App).Run()

	// 如果是启动的notify，则阻塞主进程等待。
	if !eudore.GetStringBool(os.Getenv(eudore.EnvEudoreIsNotify)) {
		defer app.Logger.Sync()
		if initlog, ok := app.Logger.(eudore.LoggerInitHandler); ok {
			app.Logger, _ = eudore.NewLoggerStd(nil)
			initlog.NextHandler(app.Logger)
		}
		<-app.Done()
		return
	}

	app.Listen(":8088")
	app.Run()
}
