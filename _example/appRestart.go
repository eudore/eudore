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
该组件不支持windows系统。
*/

import (
	"context"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/daemon"
)

func main() {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyLogger, eudore.NewLoggerInit())
	app.ParseOption(
		daemon.NewParseDaemon(app),
		NewParseLogger(app),
		daemon.NewParseRestart(),
	)
	app.Parse()
	defer app.Run()

	if app.Err() == nil {
		app.GetFunc("/*", func(ctx eudore.Context) {
			ctx.WriteString("server daemon")
		})
		app.Listen(":8087")

		go func() {
			select {
			case <-app.Done():
			case <-time.After(600 * time.Second):
				app.CancelFunc()
			}
		}()
	}
}

func NewParseLogger(app *eudore.App) eudore.ConfigParseFunc {
	return func(context.Context, eudore.Config) error {
		app.SetValue(eudore.ContextKeyLogger, eudore.NewLogger(&eudore.LoggerConfig{
			Stdout: true,
			Path:   "/tmp/daemon.log",
		}))
		return nil
	}
}
