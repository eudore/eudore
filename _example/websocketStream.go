package main

import (
	"fmt"
	"net"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/component/pprof"
	"github.com/eudore/eudore/component/websocket/gobwas"
	"github.com/eudore/eudore/component/websocket/gorilla"
)

func main() {
	app := eudore.NewApp()
	pprof.Init(app.Group("/eudore/debug"))
	// gorilla和gobwas两种Stream扩展均可使用，只需要注册一个就可支持eudore.Stream处理websocket。
	app.AddHandlerExtend(gorilla.NewExtendFuncStream)
	app.AddHandlerExtend(gobwas.NewExtendFuncStream)
	app.AnyFunc("/ui", func(ctx eudore.Context) {
		ctx.WriteFile("websocket.html")
	})
	// 使用rpc一样的方式读写对象，websocket使用的json编码。
	app.AnyFunc("/example/ws", func(stream eudore.Stream) {
		defer stream.Close()
		for {
			msg := make(map[string]interface{})
			err := stream.RecvMsg(&msg)
			if err != nil {
				app.Error("ws error:", err)
				return
			}
			fmt.Println("ws map read:", msg)
			stream.SendMsg(map[string]interface{}{
				"msg":  "success",
				"data": msg,
			})
		}
	})
	// 使用io.ReadWriter对象一样使用io
	app.AnyFunc("/example/wsio", func(stream eudore.Stream) {
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

	client := httptest.NewClient(app)
	go func() {
		client.NewRequest("ws", "/example/ws").WithWebsocket(HandlerWebsocketMap).Do().Out()
		client.NewRequest("ws", "/example/wsio").WithWebsocket(HandlerWebsocket).Do().Out()
		client.NewRequest("ws", "ws://localhost:8088/example/wsio").WithWebsocket(HandlerWebsocket).Do().Out()
		for client.Next() {
			app.Error(client.Error())
		}
		app.CancelFunc()
	}()

	app.Listen(":8088")
	app.Run()
}

func HandlerWebsocket(conn net.Conn) {
	stream := gobwas.NewStreamWebsocketClient(conn, "111")
	go func() {
		stream.Write([]byte("aaaaaa"))
		stream.Write([]byte("bbbbbb"))
		stream.Write([]byte("ccccc"))
	}()

	body := make([]byte, 2048)
	defer stream.Close()
	for {
		n, err := stream.Read(body)
		fmt.Println("ws io client read: ", string(body[:n]), err)
		if err != nil || string(body[:n]) == "ccccc" {
			return
		}
	}
}

func HandlerWebsocketMap(conn net.Conn) {
	stream := gobwas.NewStreamWebsocketClient(conn, "111")
	go func() {
		for i := 8; i < 12; i++ {
			stream.SendMsg(map[string]interface{}{
				"name":  "eudore",
				"count": i,
			})
		}
	}()

	body := make([]byte, 2048)
	defer stream.Close()
	for {
		n, err := stream.Read(body)
		fmt.Println("ws map client read: ", string(body[:n]), err)
		if err != nil || n == 54 {
			return
		}
	}
}
