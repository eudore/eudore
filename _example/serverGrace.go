package main

/*
ListenTLS方法一般均默认开启了h2，如果需要仅开启https，需要手动listen监听然后使用app.Serve启动服务。
*/

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/server/grace"
)

func main() {
	app := eudore.NewApp(
		grace.NewServerGrace(eudore.NewServerStd(nil)),
	)
	app.Options(app.Parse())

	ln, err := grace.Listen("tcp", ":8088")
	app.Options(err)
	if err == nil {
		app.Info("grace listen:", ln.Addr().String())
		app.Serve(ln)
	}

	app.Info("start pid", os.Getpid())

	go func() {
		signalChan := make(chan os.Signal)
		signal.Notify(signalChan, syscall.Signal(0x0c))
		<-signalChan
		err := app.Shutdown(context.WithValue(context.Background(), grace.ServerGraceContextKey, "1"))
		app.Options(err)
		app.CancelFunc()
	}()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
