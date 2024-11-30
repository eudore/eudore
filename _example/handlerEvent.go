package main

import (
	"encoding/json"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Event struct {
	ID      int
	Type    string
	Message string
}

func main() {
	hub := eudore.NewEventHub[*Event](time.Second)
	app := eudore.NewApp()
	app.AddMiddleware(
		middleware.NewRequestIDFunc(nil),
		middleware.NewLoggerWithEventFunc(app),
		middleware.NewRecoveryFunc(),
		middleware.NewCompressionMixinsFunc(nil),
	)
	// app.GetFunc("/pprof/*", middleware.NewPProfFunc())
	app.GetFunc("/events/:name", func(ctx eudore.Context) {
		switch {
		case ctx.GetHeader(eudore.HeaderAccept) == eudore.MimeTextEventStream:
			eudore.HandlerEvent(ctx)
			topic := make(chan *Event, 10)
			name := ctx.GetParam("name")
			ctx.Info("start event", name)
			defer app.Warning("stop event", name)

			hub.Subscribe(name, topic, func(last *Event) {
				lastid := eudore.GetAny[int](ctx.GetHeader(eudore.HeaderLastEventID))
				if last == nil || lastid == 0 {
					return
				}

			})

			w := ctx.Response()
			for {
				c := ctx.Context()
				select {
				case data := <-topic:
					if data == nil {
						return
					}
					e := &eudore.Event[*Event]{
						ID:    data.ID,
						Event: data.Type,
						Data:  data,
					}
					w.Write(e.Bytes())
					w.Flush()
				case <-c.Done():
					hub.Unsubscribe(name, topic, nil)
					return
				}
			}
		case ctx.GetHeader(eudore.HeaderConnection) == eudore.HeaderValueUpgrade:
			conn, _, _, err := ws.UpgradeHTTP(ctx.Request(), ctx.Response())
			if err != nil {
				ctx.Fatal(err)
				return
			}

			name := ctx.GetParam("name")
			topic := make(chan *Event, 10)
			hub.Subscribe(name, topic, nil)
			log := ctx.Value(eudore.ContextKeyLogger).(eudore.Logger)
			log.Info("start ws", name)
			go func() {
				defer hub.Unsubscribe(name, topic, nil)
				defer log.Warning("stop ws", name)
				defer conn.Close()
				for e := range topic {
					if e == nil {
						return
					}
					body, _ := json.Marshal(e)
					err = wsutil.WriteServerMessage(conn, ws.OpText, body)
					if err != nil {
						break
					}
				}
			}()
		default:
			ctx.WriteStatus(415)
			ctx.Fatal("invalid type")
		}
	})
	id := 0
	app.GetFunc("/events/push", func(ctx eudore.Context) {
		id++
		hub.Publish("message", &Event{
			ID:      id,
			Type:    "message",
			Message: "event",
		})
	})
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString(`
<!DOCTYPE html>
<html>
<body/>
<script>
let m = (e)=>{
	const dom = document.createElement("li");
	dom.textContent = (e.lastEventId==""?"websocket":"sse last: "+e.lastEventId) + " data: " + e.data;
	document.body.appendChild(dom);
	console.log(e)
}
let sse = new EventSource("/events/message");
sse.onmessage = m;
sse.onerror = function (e) {console.log(e)}
function connect() {
	let ws = new WebSocket("/events/message");
	ws.onmessage = m;
	ws.onerror = function (e) {console.log(e)}
	ws.onclose = function(e) {setTimeout(connect, 3000);};
}
connect();
</script>
</html>`)
	})
	app.NewRequest("GET", "/events/push")

	app.Listen(":8088")
	app.ListenTLS(":8089", "tls.crt", "tls.key")
	app.Run()
}
