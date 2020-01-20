package main

/*
按照命令执行程序。

go build -o server
./server --command=daemon
./server --command=status
./server --command=stop
./server --command=status

command包解析启动命令，支持start、daemon、status、stop、restart五个命令，需要定义command和pidfile两个配置参数。
通过向进程发送对应的系统信号实现对应的命令。
该组件不支持win系统。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/command"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewEudore()
	httptest.NewClient(app).Stop(0)
	app.RegisterInit("eudore-command", 0x007, command.InitCommand)
	app.RegisterInit("init-listen", 0x016, func(app *eudore.Eudore) error {
		app.GetFunc("/*", func(ctx eudore.Context) {
			ctx.WriteString("hello eudore")
		})
		return app.Listen(":8088")
	})
	app.Run()
}
