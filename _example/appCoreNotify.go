package main

/*
先app.Config设置notify配置，然后启动notify。
如果是notify的程序可以通过环境变量eudore.EnvEudoreIsNotify检测。
当程序启动时会如果eudore.EnvEudoreIsNotify不存在，则使用notify开始监听阻塞app后续初始化，否在就忽略notify然后进行正常app启动。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/component/notify"
	"os"
)

type loggerInitHandler3 interface {
	NextHandler(eudore.Logger)
}

func main() {
	app := eudore.NewCore()
	httptest.NewClient(app).Stop(0)

	// 设置编译命令、启动命令、监听目录
	app.Config.Set("component.notify.buildcmd", "go build -o server appCoreNotify.go")
	app.Config.Set("component.notify.startcmd", "./server")
	app.Config.Set("component.notify.watchdir", ".")
	notify.NewNotify(app.App).Run()

	// 如果是启动的notify，则阻塞主进程等待。
	if !eudore.GetStringBool(os.Getenv(eudore.EnvEudoreIsNotify)) {
		defer app.Logger.Sync()
		if initlog, ok := app.Logger.(loggerInitHandler3); ok {
			app.Logger = eudore.NewLoggerStd(nil)
			initlog.NextHandler(app.Logger)
		}
		<-app.Done()
		return
	}

	app.Listen(":8088")
	app.Run()
}
