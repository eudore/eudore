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
	router.GetFunc("/dump/connect Action=middleware:dump:GetConnect", d.GetConnect)
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

func (d *dump) GetConnect(ctx eudore.Context) {
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
	RequestURI     string      `json:"requestUri"`
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
	buffer
}

// The Unwrap method is not used yet.
func (w *responseWriterDump) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *responseWriterDump) Write(data []byte) (int, error) {
	_, _ = w.buffer.Write(data)
	return w.ResponseWriter.Write(data)
}

func (w *responseWriterDump) WriteString(data string) (int, error) {
	_, _ = w.buffer.WriteString(data)
	return w.ResponseWriter.WriteString(data)
}

// refer: [responseWriterTimeout.Body].
func (w *responseWriterDump) Body() []byte {
	return w.buf
}

// The GetBodyData method gets the written body content
// and decodes it if it is gzip encoded.
func (w *responseWriterDump) GetBodyData() []byte {
	if w.ResponseWriter.Header().Get(eudore.HeaderContentEncoding) == "gzip" {
		reader, err := gzip.NewReader(bytes.NewReader(w.buf))
		if err != nil {
			return fmt.Appendf(nil, "Reader gzip body size %d error: %s",
				len(w.buf), err.Error(),
			)
		}
		body, _ := io.ReadAll(reader)
		reader.Close()
		return body
	}
	return w.buf
}

const dumpScript = `
function NewHandlerDump() {
function b64DecodeUnicode(str) {
	try {
		return decodeURIComponent(atob(str).split("").map(function(c) {
			return "%" + ("00" + c.charCodeAt(0).toString(16)).slice(-2)
		}).join(""), )
	} catch {
		return str
	}
}
return {
	ws: null,
	Messages: [],
	Mount(ctx) {
		this.Messages = [];
		this.ws = new WebSocket("ws://" + location.host + ctx.Config.App.FetchGroup + "dump/connect",);
		try {
			this.ws.onopen = () => {
				fetch("/hello", {method: "PUT", body: "request hello body", cache: "no-cache"})
			}
			this.ws.onmessage = (e) => {
				let data = JSON.parse(e.data) || {};
				data.display = false;
				data.info = "basic";
				this.Messages.push(data)
			}
			this.ws.onclose = () => {this.Unmount() }
			this.ws.onerror = (e) => {
				ctx.Error("dump server error:", e);
				this.Unmount()
			}
		} catch (e) {ctx.Error(e.message) }
		return true
	},
	Unmount() {if (this.ws) {this.ws.close(); this.ws = null}},
	View() {
		if(!this.ws)return["eudore server not support dump"]
		return this.Messages.map((data) => {
		return {type: 'div', class: 'dump-node', child:[
			{type: 'div', class: this.getState(data["status"]), onclick: ()=>{data.display=!data.display; }, child: [
				{type: 'span', text: data["method"]},
				{type: 'span', text: data["host"]+data["path"]},
				{type: 'span', text: data["status"]}
			]},
			{type: 'div', class: 'dump-info', if: data.display, child: [
				{type: 'ul', li: [
					{text: 'Basic Info', onclick: ()=>{data.info="basic"; }},
					{text: 'Request Info', onclick: ()=>{data.info="request"; }},
					{text: 'Response Info', onclick: ()=>{data.info="response"; }},
				]},
				{type: 'div', class:"dump-info-basic", if: data.info=="basic",
					table: {tbody: {tr: [
						{td: [{text:"Time"},	{text:data["time"]}]},
						{td: [{text:"Method"},	{text:data["method"]}]},
						{td: [{text:"URI"},		{text:data["requestUri"]}]},
						{td: [{text:"Proto"},	{text:data["proto"]}]},
						{td: [{text:"Host"},	{text:data["host"]}]},
						{td: [{text:"Remote"},	{text:data["remoteAddr"]}]},
						{td: [{text:"Status"},	{text:data["status"]}]},
						{td: [{text:"Params"},	{text: this.getParams(data["params"])}]},
						{td: [{text:"Handlers"},{p: this.getHandlerDom(data["handlers"])}]},
					]}}
				},
				{type: 'div', class: 'dump-info-request', if: data.info=="request", child: [
					{type: 'table', tr: this.getHeaderDom(data["requestHeader"]) },
					{type: 'div', pre: {code: {text: b64DecodeUnicode(data['requestBody'])||""}}}
				]},
				{type: 'div', class: 'dump-info-response', if: data.info=="response", child: [
					{type: 'table', tr: this.getHeaderDom(data["responseHeader"]) },
					{type: 'div', pre: {code: {text: b64DecodeUnicode(data['responseBody'])||""}}}
				]}
			]}
		]}
	})},
	getState(status){if(status<400){return"state state-info"}if(status<500){return"state state-warning"}return"state state-error"},
	getParams(p){return p.reduce((t,c,i)=>{if (i%2===0){t.push("${0}=${1}".format(c,p[i+1])||"")}return t},[]).join(" ")},
	getHeaderDom(data){return Object.keys(data).map((k)=>({td:[{text:k},{text:data[k].toString()}]}))},
	getParamsDom(data){return Object.entries(data).map(([key,v])=>({text:"${0}=${1}".format(k,v)}))},
	getHandlerDom(data){return data.map((item)=>({text:item}))},
}}
`
