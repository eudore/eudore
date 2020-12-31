package main

/*
先app.Config设置notify配置，然后启动notify。
如果是notify的程序可以通过环境变量eudore.EnvEudoreIsNotify检测。
当程序启动时会如果eudore.EnvEudoreIsNotify不存在，则使用notify开始监听阻塞app后续初始化，否在就忽略notify然后进行正常app启动。

实现原理基于fsnotify检测目录内go文件变化，然后执行编译命令，如果编译成功就kill原进程并执行启动命令。

其他类似工具：air
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/notify"
)

func main() {
	app := eudore.NewApp()

	// 设置编译命令、启动命令、监听目录, 如果是启动的notify，则阻塞主进程等待退出。
	app.Config.Set("component.notify.buildcmd", "go build -o server appNotify.go")
	app.Config.Set("component.notify.startcmd", "./server")
	app.Config.Set("component.notify.watchdir", ".")
	n := notify.NewNotify(app)
	if n.IsRun() {
		// 启动日志输出 跳过后续初始化
		go app.Run()
		n.Run()
		return
	}

	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore")
	})
	app.Listen(":8088")
	app.Run()
}
