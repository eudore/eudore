package main

/*
按照一下命令执行程序。

go build -o server
./server --command=daemon
./server --command=status
./server --command=stop
./server --command=status


command包解析启动命令，支持start、daemon、status、stop四个命令，需要定义command和pidfile两个配置参数。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/command"
)

func main() {
	app := eudore.NewEudore()
	app.GetFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore")
	})

	app.RegisterInit("eudore-command", 0x007, command.InitCommand)
	app.RegisterInit("init-listen", 0x016, func(app *eudore.Eudore) error {
		app.Listen(":8088")
		return nil
	})
	app.Run()
}
