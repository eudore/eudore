package main

/*
eudore.UpgradeHttp获取net.Conn链接并写入建立请求响应，然后wsutil库读写数据。

`ctx.Response().Hijack()`可以直接获得原始tcp连接，然后读取header判断请求，写入101数据，再操作websocket连接。
*/

import (
	"github.com/eudore/eudore"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

func main() {
	app := eudore.NewApp()
	app.AnyFunc("/*", eudore.HandlerFunc(func(ctx eudore.Context) {
		conn, _, _, err := ws.UpgradeHTTP(ctx.Request(), ctx.Response())
		if err != nil {
			ctx.Fatal(err)
			return
		}
		go func() {
			defer conn.Close()
			for {
				// 读取数据
				msg, op, err := wsutil.ReadClientData(conn)
				if err != nil {
					// handle error
					ctx.Error(err)
					break
				}
				ctx.Info(string(msg))

				// 写入数据
				err = wsutil.WriteServerMessage(conn, op, msg)
				if err != nil {
					// handle error
					ctx.Error(err)
					break
				}
			}
		}()
	}))
	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
