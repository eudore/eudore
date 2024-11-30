package main

/*
通过AppDaemon()函数后台启动程序，也可以通过命令解析启动程序。

当第一次启动时，使用os.Exec执行启动命令后台启动进程、关闭进程并附加环境变量，
第二次启动时检测到环境变量即为后台启动，会忽略后台启动逻辑。然后执行正常启动。
*/

import (
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/daemon"
)

func main() {
	daemon.AppDaemon()

	app := eudore.NewApp()
	app.GetFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("server daemon")
	})

	go func() {
		select {
		case <-app.Done():
		case <-time.After(60 * time.Second):
			app.CancelFunc()
		}
	}()
	app.Listen(":8088")
	app.Run()
}
