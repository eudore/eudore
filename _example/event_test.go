package eudore_test

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func TestEventCover(t *testing.T) {
	for i := 0; i < 6; i++ {
		hub := eudore.NewEventHub[int](time.Microsecond * time.Duration(10+i))
		topic := make(chan int)
		hub.Subscribe("msg", topic, nil)
		go func() {
			for i := 1; i < 200; i++ {
				hub.Publish("msg", i)
			}
			hub.Unsubscribe("msg", topic, nil)
		}()
		for v := range topic {
			if v == 0 {
				break
			}
		}
	}
	hub := eudore.NewEventHub[int](time.Microsecond * 10)
	for i := 0; i < 200; i++ {
		topic := "get" + strconv.Itoa(i)
		for n := 0; n < 10; n++ {
			go func() {
				hub.Publish(topic, 1)
			}()
		}
	}
	time.Sleep(time.Millisecond * 20)
}

func TestEventHub(t *testing.T) {
	type Event struct {
		Message string
	}
	eudore.NewEventHub[any](0)
	hub := eudore.NewEventHubWithOptions[*eudore.Event[Event]](
		time.Millisecond*10,
		time.Millisecond*100,
		time.Millisecond*200,
		&eudore.Event[Event]{Event: "ping"},
	)
	top := make(chan *eudore.Event[Event])
	hub.Subscribe("top", top, nil)
	hub.(interface{ Metadata() interface{} }).Metadata()
	hub.Unsubscribe("top", top, nil)

	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyClient, app.NewClient(
		eudore.NewClientHookTimeout(-1),
	))
	app.SetValue(eudore.NewContextKey("hub"), hub)
	app.AddMiddleware(middleware.NewRequestIDFunc(nil))
	app.GetFunc("/events/:topic", func(ctx eudore.Context) {
		if ctx.GetHeader(eudore.HeaderAccept) != eudore.MimeTextEventStream {
			return
		}
		eudore.HandlerEvent(ctx)

		log := ctx.Value(eudore.ContextKeyLogger).(eudore.Logger)
		last := eudore.GetAnyByString[int](ctx.GetHeader(eudore.HeaderLastEventID))
		topic := make(chan *eudore.Event[Event], 10)
		hub.Subscribe(ctx.GetParam("topic"), topic, func(e *eudore.Event[Event]) {
			log.Infof("last %d -> %v", last, e)
		})

		for {
			select {
			case e := <-topic:
				if e == nil {
					fmt.Println("sn")
					return
				}
				ctx.Write(e.Bytes())
				ctx.Response().Flush()
			case <-ctx.Context().Done():
				fmt.Println("don")
				return
			}
		}
	})
	app.GetFunc("/timeout/:topic", func(ctx eudore.Context) {
		topic := make(chan *eudore.Event[Event])
		hub.Subscribe(ctx.GetParam("topic"), topic, nil)
		for i := 0; i < 3; i++ {
			time.Sleep(time.Millisecond * 10)
			<-topic
		}
		<-ctx.Context().Done()
	})
	app.GetFunc("/events/204", func(ctx eudore.Context) {
		ctx.WriteHeader(204)
	})
	app.GetFunc("/events/err", func(ctx eudore.Context) {
		ctx.SetResponse(noWriteResponse{ctx.Response()})
		eudore.HandlerEvent(ctx)
	})
	id := 0
	app.GetFunc("/push/:topic", func(ctx eudore.Context) {
		id++
		hub.Publish(ctx.GetParam("topic"), &eudore.Event[Event]{
			ID:    id,
			Event: ctx.GetParam("topic"),
			Data:  Event{"event"},
		})
		hub.Broadcast(&eudore.Event[Event]{Event: "ping"})
	})

	{

		app2 := eudore.NewApp()
		app2.GetFunc("/pprof/*", middleware.NewPProfFunc())
		app2.Listen(":8088")
		defer app2.Run()
		defer app2.CancelFunc()
	}

	run := func(path string) {
		client := app.Client.NewClient(eudore.NewClientHookTimeout(-1))
		ctx, cancel := context.WithCancel(app)
		defer cancel()
		for i := 0; i < 3; i++ {
			client.GetRequest(path, ctx,
				eudore.NewClientOptionEventID(0),
				eudore.NewClientOptionEventID(1),
				eudore.NewClientEventCancel(cancel),
				eudore.NewClientEventHandler(func(e *eudore.Event[string]) error {
					if e.Event != "ping" {
						app.Info(strings.ReplaceAll(e.String(), "\n", " "))
					}
					return nil
				}),
			)
			select {
			case <-ctx.Done():
				return
			default:
				time.Sleep(time.Duration(100) * time.Millisecond)
			}
		}
	}

	app.GetRequest("/push/message")
	app.GetRequest("/events/err")
	go run("/events/message")
	go run("/events/message")
	go run("/events/message")
	go run("/events/204")
	go run("/events/err")
	time.Sleep(time.Millisecond * 20)
	for i := 0; i < 3; i++ {
		app.GetRequest("/push/message")
	}

	time.Sleep(time.Millisecond * 200)
	app.CancelFunc()
	app.Run()
	time.Sleep(time.Millisecond * 20)
}

type noWriteResponse struct {
	eudore.ResponseWriter
}

func (w noWriteResponse) Write([]byte) (int, error) {
	return 0, io.EOF
}

func TestEventMessage(t *testing.T) {
	msg := eudore.Event[[]byte]{
		ID:      13,
		Event:   "message",
		Data:    []byte(`{"message":"body"}`),
		Retry:   3000,
		Comment: "retry time",
	}

	app := eudore.NewApp()
	app.GetFunc("/stream/200", func(ctx eudore.Context) {
		app.Info("new stream")
		ctx.Write(msg.Bytes())
		ctx.Write([]byte((eudore.Event[string]{Data: "messgae"}).String()))
		ctx.Write([]byte((eudore.Event[any]{Data: &strings.Builder{}}).String()))
		ctx.Write([]byte((eudore.Event[any]{Data: TestEventMessage}).String()))
	})
	app.GetFunc("/stream/201", func(ctx eudore.Context) {
		ctx.WriteHeader(201)
	})
	app.GetFunc("/stream/204", func(ctx eudore.Context) {
		ctx.WriteHeader(204)
	})
	app.GetFunc("/stream/invalud", func(ctx eudore.Context) {
		ctx.WriteString("message")
	})
	app.GetFunc("/stream/eof", func(ctx eudore.Context) {
		ctx.WriteString("message not EOF")
	})
	app.GetFunc("/stream/timeout", func(ctx eudore.Context) {
		ctx.SetResponse(struct{ eudore.ResponseWriter }{ctx.Response()})
	})

	type Message struct {
		Message string `json:"message"`
	}
	app.GetRequest("/stream/200", eudore.NewClientEventHandler(func(event *eudore.Event[[]byte]) error {
		return nil
	}))
	app.GetRequest("/stream/200", eudore.NewClientEventHandler(func(event *eudore.Event[string]) error {
		return nil
	}))
	app.GetRequest("/stream/204", eudore.NewClientEventHandler(func(event *eudore.Event[Message]) error {
		return nil
	}))
	app.GetRequest("/stream/200", eudore.NewClientEventHandler(func(event *eudore.Event[error]) error {
		return fmt.Errorf("new errro")
	}))
	app.GetRequest("/stream/201", eudore.NewClientEventHandler[[]byte](nil))

	events := make(chan *eudore.Event[[]byte], 12)
	app.GetRequest("/stream/200", eudore.NewClientEventChan(events))
	app.GetRequest("/stream/201", eudore.NewClientEventChan(events))
	app.GetRequest("/stream/204", eudore.NewClientEventChan(events))
	app.GetRequest("/stream/invalud", eudore.NewClientEventChan(events))
	app.GetRequest("/stream/eof", eudore.NewClientEventChan(events))
	app.GetRequest("/stream/timeout", eudore.NewClientEventChan(events))

	app.CancelFunc()
	app.Run()
}
