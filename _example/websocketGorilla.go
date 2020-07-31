package main

import (
	"github.com/eudore/eudore"
	"github.com/gorilla/websocket"
)

func main() {
	app := eudore.NewApp()
	app.AnyFunc("/*", eudore.HandlerFunc(func(ctx eudore.Context) {
		conn, err := websocket.Upgrade(ctx.Response(), ctx.Request(), nil, 1024, 1024)
		if err != nil {
			ctx.Fatal(err)
			return
		}

		go func() {
			defer conn.Close()
			for {
				messageType, p, err := conn.ReadMessage()
				if err != nil {
					return
				}
				if err := conn.WriteMessage(messageType, p); err != nil {
					return
				}
			}
		}()
	}))

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
