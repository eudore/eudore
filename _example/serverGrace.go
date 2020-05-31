package main

/*
ListenTLS方法一般均默认开启了h2，如果需要仅开启https，需要手动listen监听然后使用app.Serve启动服务。
*/

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/pprof"
	"github.com/eudore/eudore/component/server/grace"
	"github.com/eudore/eudore/component/websocket/gobwas"
)

func main() {
	app := eudore.NewApp(grace.NewServerGrace(eudore.NewServerStd(nil)))
	pprof.Init(app.Group("/eudore/debug"))
	app.AddHandlerExtend(gobwas.NewExtendFuncStream)
	app.AnyFunc("/ws", func(stream eudore.Stream) {
		body := make([]byte, 2048)
		defer stream.Close()
		for {
			n, err := stream.Read(body)
			if err != nil {
				app.Error("wsio error:", err)
				return
			}
			fmt.Println("wsio read: ", string(body[:n]), stream.GetType())
			stream.SetType(stream.GetType())
			stream.Write(body[:n])
		}
	})
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

	app.CancelFunc()
	app.Run()
}
