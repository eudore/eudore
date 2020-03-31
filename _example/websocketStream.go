package main

import (
	"fmt"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/component/websocket/gobwas"
	"github.com/eudore/eudore/component/websocket/gorilla"
)

func main() {
	app := eudore.NewCore()
	httptest.NewClient(app).Stop(0)
	// gorilla和gobwas两种Stream扩展均可使用，只需要注册一个就可支持eudore.Stream处理websocket。
	app.AddHandlerExtend(gorilla.NewExtendFuncStream)
	app.AddHandlerExtend(gobwas.NewExtendFuncStream)
	app.AnyFunc("/ui", func(ctx eudore.Context) {
		ctx.WriteFile("websocket.html")
	})
	// 使用rpc一样的方式读写对象，websocket使用的json编码。
	app.AnyFunc("/example/ws", func(stream eudore.Stream) {
		for {
			msg := make(map[string]interface{})
			err := stream.RecvMsg(&msg)
			if err != nil {
				app.Error("ws error:", err)
				return
			}
			fmt.Println(msg)
			stream.SendMsg(map[string]interface{}{
				"msg": "success",
			})
		}
	})
	// 使用io.ReadWriter对象一样使用io
	app.AnyFunc("/example/wsio", func(stream eudore.Stream) {
		body := make([]byte, 2048)
		for {
			n, err := stream.Read(body)
			if err != nil {
				app.Error("wsio error:", err)
				return
			}
			fmt.Println(string(body[:n]))
			stream.SetType(stream.GetType())
			stream.Write(body[:n])
		}
	})
	app.Listen(":8088")
	app.Run()
}
