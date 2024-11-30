package main

import (
	"context"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(
		"global",
		middleware.NewLoggerWithEventFunc(app),
		middleware.NewRequestIDFunc(nil),
	)
	id := 0
	app.GetFunc("/events", func(ctx eudore.Context) {
		if ctx.GetHeader(eudore.HeaderAccept) != eudore.MimeTextEventStream {
			ctx.Fatal("invalid Header Accept")
			return
		}
		eudore.HandlerEvent(ctx)

		c := ctx.Context()
		w := ctx.Response()
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			id++
			select {
			case <-ticker.C:
				if rand.Intn(100) == 0 {
					id = id + 10 + rand.Intn(10)
					return
				}
				w.Write(eudore.Event[string]{
					ID:    id,
					Event: "message",
					Data:  "eudore",
				}.Bytes())
				w.Flush()
			case <-c.Done():
				return
			}
		}
	})

	go func() {
		last, retry := 0, 3000
		client := app.Client.NewClient(eudore.NewClientHookTimeout(-1))
		ctx, cancel := context.WithCancel(app)
		defer cancel()
		for {
			client.GetRequest("/events", ctx,
				eudore.NewClientOptionEventID(last),
				eudore.NewClientEventCancel(cancel),
				eudore.NewClientEventHandler(func(e *eudore.Event[string]) error {
					if e.ID > 0 {
						last = e.ID
					}
					if e.Retry > 0 {
						retry = e.Retry
					}
					if e.Event != "ping" {
						app.Info(strings.TrimSpace(strings.ReplaceAll(e.String(), "\n", " ")))
					}
					return nil
				}),
			)
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Duration(retry) * time.Millisecond)
			}
		}
	}()

	app.Listen(":8088")
	app.Run()
}
