package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
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

		// 取出请求上下文的Logger，否则Context在sync.Pool再次分配后可能竞态冲突。
		log := ctx.Logger()
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
	app.ListenTLS(":8089", "", "")

	client := httptest.NewClient(app)

	client.Client.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	go func() {
		client.NewRequest("GET", "/example/wsio").WithWebsocket(handlerGobwasWebsocket).Do().Out()
		client.NewRequest("GET", "http://localhost:8088/example/wsio").WithWebsocket(handlerGobwasWebsocket).Do().Out()
		client.NewRequest("GET", "https://localhost:8089/example/wsio").WithWebsocket(handlerGobwasWebsocket).Do().Out()
		// app.CancelFunc()
	}()

	app.Run()
}

func handlerGobwasWebsocket(conn net.Conn) {
	go func() {
		wsutil.WriteClientBinary(conn, []byte("aaaaaa"))
		wsutil.WriteClientBinary(conn, []byte("bbbbbb"))
		wsutil.WriteClientBinary(conn, []byte("ccccc"))
	}()

	defer conn.Close()
	for {
		b, err := wsutil.ReadServerBinary(conn)
		fmt.Println("ws io client read: ", string(b), err)
		if err != nil {
			fmt.Println("gobwas client err:", err)
			return
		}
	}
}
