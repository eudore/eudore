package middleware

import (
	"bytes"
	"compress/gzip"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/eudore/eudore"
)

// NewDumpFunc function creates middleware to implement dump request.
//
// Create a Websocket connection to receive dump data through
// the [eudore.HandlerFunc] registered by the [eudore.Router].
//
// dump will load the body into memory,
// Large body requests may affect memory usage.
//
// If registered in the global middleware,
// the returned Handlers will all have the same value.
//
// This middleware does not support cluster mode.
func NewDumpFunc(router eudore.Router) Middleware {
	var d dump
	router.GetFunc("/dump/connect", d.handler)
	release := func(ctx eudore.Context, w *responseWriterDump) {
		req := ctx.Request()
		body, _ := ctx.Body()
		msg := &dumpMessage{
			Time:           time.Now().Format(eudore.DefaultContextFormatTime),
			Path:           ctx.Path(),
			Host:           ctx.Host(),
			RemoteAddr:     req.RemoteAddr,
			Proto:          req.Proto,
			Method:         req.Method,
			RequestURI:     req.RequestURI,
			RequestHeader:  req.Header,
			RequestBody:    body,
			Status:         w.Status(),
			ResponseHeader: w.Header(),
			ResponseBody:   w.GetBodyData(),
			Params:         *ctx.Params(),
			Handlers:       getContextHandlerName(ctx),
		}

		r := recover()
		if r != nil {
			msg.Status = eudore.StatusInternalServerError
			msg.ResponseBody = fmt.Appendf(nil, "panic: %v", r)
			d.writeMessage(msg)
			panic(r)
		}
		d.writeMessage(msg)
	}
	return func(ctx eudore.Context) {
		if d.number() == 0 {
			return
		}

		w := &responseWriterDump{ResponseWriter: ctx.Response()}
		defer release(ctx, w)
		_, err := ctx.Body()
		if err != nil {
			ctx.Fatal(err)
			return
		}

		ctx.SetResponse(w)
		ctx.Next()
	}
}

var (
	wsKey  = []byte("258EAFA5-E914-47DA-95CA-C5AB0DC85B11")
	wsResp = []byte(strings.Join([]string{
		"HTTP/1.1 101 Switching Protocols",
		"Connection: Upgrade",
		"Upgrade: websocket",
		"Sec-WebSocket-Accept: ",
	}, "\r\n"))
	wsline = []byte("\r\n\r\n")
)

type dump struct {
	sync.RWMutex
	conns []net.Conn
}

func (d *dump) number() int {
	d.RLock()
	defer d.RUnlock()
	return len(d.conns)
}

func (d *dump) handler(ctx eudore.Context) {
	conn, buf, err := ctx.Response().Hijack()
	if err != nil {
		ctx.Fatal(err)
		return
	}
	h := sha1.New()
	h.Write([]byte(ctx.GetHeader("Sec-WebSocket-Key")))
	h.Write(wsKey)
	_, _ = buf.Write(wsResp)
	_, _ = buf.Write([]byte(base64.StdEncoding.EncodeToString(h.Sum(nil))))
	_, _ = buf.Write(wsline)
	buf.Flush()

	d.Lock()
	d.conns = append(d.conns, conn)
	d.Unlock()
	ctx.End()
}

func (d *dump) writeMessage(msg *dumpMessage) {
	body, err := json.Marshal(msg)
	if err == nil {
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
		for i := 0; i < len(d.conns); i++ {
			_, _ = d.conns[i].Write(head)
			_, err := d.conns[i].Write(body)
			if err != nil {
				d.conns[i] = d.conns[len(d.conns)-1]
				d.conns = d.conns[:len(d.conns)-1]
				i--
			}
		}
		d.Unlock()
	}
}

type dumpMessage struct {
	Time           string      `json:"time"`
	Path           string      `json:"path"`
	Host           string      `json:"host"`
	RemoteAddr     string      `json:"remoteAddr"`
	Proto          string      `json:"proto"`
	Method         string      `json:"method"`
	RequestURI     string      `json:"requestURI"`
	RequestHeader  http.Header `json:"requestHeader"`
	RequestBody    []byte      `json:"requestBody"`
	Status         int         `json:"status"`
	ResponseHeader http.Header `json:"responseHeader"`
	ResponseBody   []byte      `json:"responseBody"`
	Params         []string    `json:"params"`
	Handlers       []string    `json:"handlers"`
}

func getContextHandlerName(ctx eudore.Context) []string {
	_, handlers := ctx.GetHandlers()
	names := make([]string, len(handlers))
	for i := range handlers {
		names[i] = handlers[i].String()
	}
	return names
}

type responseWriterDump struct {
	eudore.ResponseWriter
	w bytes.Buffer
}

// The Unwrap method is not used yet.
func (w *responseWriterDump) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *responseWriterDump) Write(data []byte) (int, error) {
	w.w.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriterDump) WriteString(data string) (int, error) {
	w.w.WriteString(data)
	return w.ResponseWriter.WriteString(data)
}

// refer: [responseWriterTimeout.Body].
func (w *responseWriterDump) Body() []byte {
	return w.w.Bytes()
}

// The GetBodyData method gets the written body content
// and decodes it if it is gzip encoded.
func (w *responseWriterDump) GetBodyData() []byte {
	if w.ResponseWriter.Header().Get(eudore.HeaderContentEncoding) == "gzip" {
		reader, err := gzip.NewReader(&w.w)
		if err != nil {
			return fmt.Appendf(nil, "Reader gzip body size %d error: %s",
				w.w.Len(), err.Error(),
			)
		}
		body, _ := io.ReadAll(reader)
		reader.Close()
		return body
	}
	return w.w.Bytes()
}
