# x/net/websocket

golang websocke常见实现有两个库`x/net/websocket`和`github.com/gorilla/websocket`，`x/net/websocket`是扩展库，`github.com/gorilla/websocket`是功能比较齐全的三方库实现，本文档解析`x/net/websocket`库的实现以及websocket的实现原理。


## prototcl

WebSocket 握手

GET /chat HTTP/1.1
Host: server.example.com
Upgrade: websocket
Connection: Upgrade
Sec-WebSocket-Key: x3JJHMbDL1EzLkh9GBhXDw==
Sec-WebSocket-Protocol: chat, superchat
Sec-WebSocket-Version: 13
Origin: http://example.com

重点如下两个 header：


Upgrade: websocket
Connection: Upgrade

Upgrade 表示升级到 WebSocket 协议，Connection 表示这个 HTTP 请求是一次协议升级，Origin 表示发请求的来源。

## Hijacker

Websocket依靠http.Hijacker接口,其意思是劫持，会劫持net.Conn连接，然后http.Server就不会处理，由自己去处理劫持的连接。

定义：

```golang
// https://golang.org/src/net/http/server.go?s=6796:7615#L168
type Hijacker interface {
        // Hijack lets the caller take over the connection.
        // After a call to Hijack the HTTP server library
        // will not do anything else with the connection.
        //
        // It becomes the caller's responsibility to manage
        // and close the connection.
        //
        // The returned net.Conn may have read or write deadlines
        // already set, depending on the configuration of the
        // Server. It is the caller's responsibility to set
        // or clear those deadlines as needed.
        //
        // The returned bufio.Reader may contain unprocessed buffered
        // data from the client.
        //
        // After a call to Hijack, the original Request.Body must not
        // be used. The original Request's Context remains valid and
        // is not canceled until the Request's ServeHTTP method
        // returns.
        Hijack() (net.Conn, *bufio.ReadWriter, error)
}
```

Hijacker只是一个Response可选的实现接口，标准库的response实现了这个接口，需要使用获得的http.ResponseWriter类型转换一下。

```golang
// https://golang.org/src/net/http/server.go#L1900
// Hijack implements the Hijacker.Hijack method. Our response is both a ResponseWriter
// and a Hijacker.
func (w *response) Hijack() (rwc net.Conn, buf *bufio.ReadWriter, err error) {
	if w.handlerDone.isSet() {
		panic("net/http: Hijack called after ServeHTTP finished")
	}
	if w.wroteHeader {
		w.cw.flush()
	}

	c := w.conn
	c.mu.Lock()
	defer c.mu.Unlock()

	// Release the bufioWriter that writes to the chunk writer, it is not
	// used after a connection has been hijacked.
	rwc, buf, err = c.hijackLocked()
	if err == nil {
		putBufioWriter(w.w)
		w.w = nil
	}
	return rwc, buf, err
}
```

## example

```golang
// https://golang.org/pkg/net/http/#example_Hijacker
package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/hijack", func(w http.ResponseWriter, r *http.Request) {
		hj, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
			return
		}
		conn, bufrw, err := hj.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Don't forget to close the connection:
		defer conn.Close()
		bufrw.WriteString("Now we're speaking raw TCP. Say hi: ")
		bufrw.Flush()
		s, err := bufrw.ReadString('\n')
		if err != nil {
			log.Printf("error reading string: %v", err)
			return
		}
		fmt.Fprintf(bufrw, "You said: %q\nBye.\n", s)
		bufrw.Flush()
	})
}
```
`hj, ok := w.(http.Hijacker)`首先进行类型转换，转换成`http.Hijacker`类型；再使用`Hijack()`方法获取`net.Conn`和`*bufio.ReadWriter`，写入数据，最后关闭掉这个连接。


## golang Websocket

```golang
package main

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

func Echo(ws *websocket.Conn) {
	var err error

	for {
		var reply string

		if err = websocket.Message.Receive(ws, &reply); err != nil {
			fmt.Println("Can't receive")
			break
		}

		fmt.Println("Received back from client: " + reply)

		msg := "Received:  " + reply
		fmt.Println("Sending to client: " + msg)

		if err = websocket.Message.Send(ws, msg); err != nil {
			fmt.Println("Can't send")
			break
		}
	}
}

func main() {
	http.Handle("/", websocket.Handler(Echo))

	if err := http.ListenAndServe(":1234", nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
```

[]: https://github.com/gorilla/websocket
