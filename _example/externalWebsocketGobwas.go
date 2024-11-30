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

		// 取出请求上下文的Logger，否则Context在sync.Pool再次分配后可能数据冲突和竞态冲突。
		log := ctx.Value(eudore.ContextKeyLogger).(eudore.Logger)
		go func() {
			defer conn.Close()
			for {
				// 读取数据
				msg, op, err := wsutil.ReadClientData(conn)
				if err != nil {
					// handle error
					log.Error(err)
					break
				}
				log.Info(string(msg))

				// 写入数据
				err = wsutil.WriteServerMessage(conn, op, msg)
				if err != nil {
					// handle error
					log.Error(err)
					break
				}
			}
		}()
	}))

	app.Listen(":8088")
	app.Run()
}
