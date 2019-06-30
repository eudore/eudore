package websocket

import (
	"fmt"
	// "bytes"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/protocol/websocket"
	"testing"
	// "github.com/gobwas/ws/wsutil"
	_ "github.com/eudore/eudore/component/router/init"
)

func TestWebsocket(*testing.T) {
	app := eudore.NewCore()
	app.RegisterComponent(eudore.ComponentRouterInitName, eudore.HandlerFunc(func(ctx eudore.Context) {
		conn, err := eudore.UpgradeHttp(ctx)
		fmt.Println("----------", err)
		// go func() {
		wsconn := websocket.NewConn(conn)
		var body = make([]byte, 2048)
		for {
			// wsconn.ReadFrame()
			fmt.Println("start")
			n, err := wsconn.Read(body)
			if err != nil {
				fmt.Println(err)
				break
			}
			// fmt.Println(ws.ReadFrame(bytes.NewReader(body)))
			// fmt.Println(websocket.ReadFrame(bytes.NewReader(body)))
			fmt.Println(n, err, body[:n], string(body[:n]))
			// frame, _ := websocket.ReadFrame(bytes.NewReader(body))
			// frame.Read(body[:n])
			// fmt.Println(string(frame.Body()))
			wsconn.Write([]byte("kkkkk"))
			// wsutil.WriteServerMessage(conn, 0x01, []byte("kkkkk"))
		}
		conn.Close()
		// }()
	}))

	app.Listen(":8088")
	app.Run()
}
