package middleware

import (
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/eudore/eudore"
)

// NewDumpFunc 函数创建一个截取请求信息的中间件，将匹配请求使用webscoket输出给客户端。
//
// router参数是eudore.Router类型，然后注入拦截路由处理。
//
// 注意：dump在集群模式下只能连接到一个server。
func NewDumpFunc(router eudore.Router) eudore.HandlerFunc {
	var d dump
	router.AnyFunc("/dump/ui", HandlerAdmin)
	router.AnyFunc("/dump/connect", d.dumphandler)
	return func(ctx eudore.Context) {
		// not handler panic
		ctx.Body()
		dumpresp := &dumpResponset{ResponseWriter: ctx.Response()}
		ctx.SetResponse(dumpresp)
		ctx.Next()
		req := ctx.Request()
		msg := &dumpMessage{
			Time:           time.Now(),
			Path:           ctx.Path(),
			Host:           ctx.Host(),
			RemoteAddr:     req.RemoteAddr,
			Proto:          req.Proto,
			Method:         req.Method,
			RequestURI:     req.RequestURI,
			RequestHeader:  req.Header,
			RequestBody:    ctx.Body(),
			Status:         ctx.Response().Status(),
			ResponseHeader: ctx.Response().Header(),
			ResponseBody:   dumpresp.GetBodyData(),
			Params:         ctx.Params(),
			Handlers:       getContextHandlerName(ctx),
		}
		d.WriteMessage(msg)
	}
}

type dump struct {
	sync.RWMutex
	dumpconn []net.Conn
}

func (d *dump) dumphandler(ctx eudore.Context) {
	err := d.newDumpConn(ctx)
	if err != nil {
		ctx.Fatal(err)
	}
	ctx.End()
}

func (d *dump) newDumpConn(ctx eudore.Context) error {
	conn, buf, err := ctx.Response().Hijack()
	if err != nil {
		return err
	}
	h := sha1.New()
	h.Write([]byte(ctx.GetHeader("Sec-WebSocket-Key") + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	buf.Write([]byte("HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: websocket\r\nSec-WebSocket-Accept: "))
	buf.Write([]byte(base64.StdEncoding.EncodeToString(h.Sum(nil))))
	buf.Write([]byte("\r\nX-Eudore-Admin: dump\r\n\r\n"))
	buf.Flush()

	d.Lock()
	d.dumpconn = append(d.dumpconn, conn)
	d.Unlock()
	return nil
}

func (d *dump) WriteMessage(msg *dumpMessage) {
	body, _ := json.Marshal(msg)
	var head []byte
	length := len(body)
	if length <= 0xffff {
		head = []byte{129, 126, uint8(length >> 8), uint8(length & 0xff)}
	} else {
		head = []byte{129, 127, 0, 0, 0, 0, 0, 0, 0, 0}
		for i := uint(0); i < 7; i++ {
			head[9-i] = uint8(length >> (8 * i) & 0xff)
		}
	}

	d.Lock()
	for i := 0; i < len(d.dumpconn); i++ {
		d.dumpconn[i].Write(head)
		_, err := d.dumpconn[i].Write(body)
		if err != nil {
			d.dumpconn[i] = d.dumpconn[len(d.dumpconn)-1]
			d.dumpconn = d.dumpconn[:len(d.dumpconn)-1]
		}
	}
	d.Unlock()
}

type dumpMessage struct {
	Time           time.Time
	Path           string
	Host           string
	RemoteAddr     string
	Proto          string
	Method         string
	RequestURI     string
	RequestHeader  http.Header
	RequestBody    []byte
	Status         int
	ResponseHeader http.Header
	ResponseBody   []byte
	Params         *eudore.Params
	Handlers       []string
}

func getContextHandlerName(ctx eudore.Context) []string {
	_, handlers := ctx.GetHandler()
	names := make([]string, len(handlers))
	for i := range handlers {
		names[i] = fmt.Sprint(handlers[i])
	}
	return names
}

type dumpResponset struct {
	eudore.ResponseWriter
	Buffer bytes.Buffer
}

// Write 方法实现ResponseWriter中的Write方法。
func (w *dumpResponset) Write(data []byte) (int, error) {
	w.Buffer.Write(data)
	return w.ResponseWriter.Write(data)
}

// GetBodyData 方法获取写入的body内容，如果是gzip编码则解压。
func (w *dumpResponset) GetBodyData() []byte {
	if w.ResponseWriter.Header().Get(eudore.HeaderContentEncoding) == "gzip" {
		gread := new(gzip.Reader)
		gread.Reset(&w.Buffer)
		body, err := ioutil.ReadAll(gread)
		if err != nil {
			return w.Buffer.Bytes()
		}
		gread.Close()
		return body
	}
	return w.Buffer.Bytes()
}
