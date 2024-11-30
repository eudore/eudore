package eudore

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var eventReplacer = strings.NewReplacer("\n", "\ndata: ")

var clientStreamHeaders = []string{
	HeaderAccept,
	MimeTextEventStream,
	HeaderConnection,
	HeaderValueKeepAlive,
	HeaderCacheControl,
	HeaderValueNoCache,
}

// Event defines [Server sent events] message.
type Event[T any] struct {
	ID      int
	Event   string
	Data    T
	Retry   int
	Comment string
}

// The Bytes method implements [Event] message encoding.
//
// If T is not of type string, []byte, or [fmt.Stringer],
// [json.NewEncoder] is used to encode T.
func (e Event[T]) Bytes() []byte {
	buf := bytes.NewBuffer(make([]byte, 0, 512))
	e.format(buf)
	return buf.Bytes()
}

func (e Event[T]) String() string {
	buf := bytes.NewBuffer(make([]byte, 0, 512))
	e.format(buf)
	return buf.String()
}

func (e Event[T]) format(buf *bytes.Buffer) {
	if e.Event != "" {
		buf.WriteString("event: ")
		buf.WriteString(e.Event)
		buf.WriteByte('\n')
	}
	if e.ID != 0 {
		buf.WriteString("id: ")
		buf.WriteString(strconv.Itoa(e.ID))
		buf.WriteByte('\n')
	}
	if e.Retry != 0 {
		buf.WriteString("retry: ")
		buf.WriteString(strconv.Itoa(e.Retry))
		buf.WriteByte('\n')
	}

	switch v := any(e.Data).(type) {
	case string:
		buf.WriteString("data: ")
		_, _ = eventReplacer.WriteString(buf, v)
		buf.WriteByte('\n')
	case []byte:
		buf.WriteString("data: ")
		_, _ = eventReplacer.WriteString(buf, string(v))
		buf.WriteByte('\n')
	case fmt.Stringer:
		buf.WriteString("data: ")
		_, _ = eventReplacer.WriteString(buf, v.String())
		buf.WriteByte('\n')
	default:
		buf.WriteString("data: ")
		err := json.NewEncoder(buf).Encode(v)
		if err != nil {
			buf.WriteString(err.Error())
			buf.WriteByte('\n')
		}
	}

	if e.Comment != "" {
		buf.WriteString(": ")
		buf.WriteString(e.Comment)
		buf.WriteByte('\n')
	}
	buf.WriteByte('\n')
}

// The HandlerEvent method implements the [Event] handshake.
//
// When the request [HeaderAccept] is [MimeTextEventStream],
// return the [Event] response Header and cancel the response Write Deadline.
func HandlerEvent(ctx Context) {
	if ctx.GetHeader(HeaderAccept) != MimeTextEventStream {
		return
	}
	w := ctx.Response()
	h := w.Header()
	h.Set(HeaderContentType, MimeTextEventStream)
	h.Set(HeaderTransferEncoding, HeaderValueChunked)
	h.Set(HeaderConnection, HeaderValueKeepAlive)
	h.Set(HeaderCacheControl, HeaderValueNoCache)
	h.Set(HeaderVary, strings.Join(append(h[HeaderVary], HeaderAccept), ", "))
	w.WriteHeader(StatusOK)
	w.Flush()

	var next http.ResponseWriter = w
	for {
		switch i := next.(type) {
		case interface{ SetWriteDeadline(t time.Time) error }:
			_ = i.SetWriteDeadline(time.Time{})
			return
		case interface{ Unwrap() http.ResponseWriter }:
			next = i.Unwrap()
		default:
			return
		}
	}
}

// The NewClientOptionEventID function creates [ClientOption] when
// id is greater than zero and sets [HeaderLastEventID].
func NewClientOptionEventID(id int) http.Header {
	if id <= 0 {
		return nil
	}
	return http.Header{HeaderLastEventID: []string{strconv.Itoa(id)}}
}

// NewClientEventCancel function creates [ClientOption] and executes the
// cancel operation when [StatusNoContent] is returned.
func NewClientEventCancel(cancel func()) *ClientOption {
	return &ClientOption{
		ResponseHooks: []func(*http.Response) error{func(resp *http.Response) error {
			if resp.StatusCode == StatusNoContent {
				cancel()
			}
			return nil
		}},
	}
}

// The NewClientEventHandler function creates [ClientOption] to process [Event].
//
// Allows return status code [StatusOK] or [StatusNoContent],
// and does not implement automatic retry.
//
// Type T can be []byte, string, JSON,
// and the function is called for each [Event] parsed.
//
// Type T can be []byte [Event].Data will reuse memory.
func NewClientEventHandler[T any](fn func(e *Event[T]) error) *ClientOption {
	handler := func(e *Event[[]byte]) error {
		return fn(&Event[T]{
			ID:      e.ID,
			Event:   e.Event,
			Retry:   e.Retry,
			Comment: e.Comment,
			Data:    unmarshalEvent[T](e.Data),
		})
	}

	var zero T
	if _, ok := any(zero).([]byte); ok {
		handler = any(fn).(func(*Event[[]byte]) error)
	}

	return &ClientOption{
		Headers: clientStreamHeaders,
		ResponseHooks: []func(*http.Response) error{func(resp *http.Response) error {
			switch resp.StatusCode {
			case StatusOK:
				return parseEvent(handler, resp.Body)
			case StatusNoContent:
				return nil
			default:
				return fmt.Errorf("not connet to stream: %d", resp.StatusCode)
			}
		}},
	}
}

func handlerEventChan[T any](events chan *Event[T]) func(*Event[[]byte]) error {
	return func(e *Event[[]byte]) error {
		events <- &Event[T]{
			ID:      e.ID,
			Event:   e.Event,
			Retry:   e.Retry,
			Comment: e.Comment,
			Data:    unmarshalEvent[T](e.Data),
		}
		return nil
	}
}

// NewClientEventChan function creates [ClientOption] and
// uses chan to receive [Event].
//
// If the server returns [StatusNoContent], chan will be closed.
//
// refer: [NewClientEventHandler].
func NewClientEventChan[T any](events chan *Event[T]) *ClientOption {
	handler := handlerEventChan(events)
	return &ClientOption{
		Headers: clientStreamHeaders,
		ResponseHooks: []func(*http.Response) error{func(resp *http.Response) error {
			switch resp.StatusCode {
			case StatusOK:
				return parseEvent(handler, resp.Body)
			case StatusNoContent:
				close(events)
				return nil
			default:
				return fmt.Errorf("not connet to stream: %d", resp.StatusCode)
			}
		}},
	}
}

func unmarshalEvent[T any](data []byte) T {
	var zero T
	var val any
	switch any(zero).(type) {
	case string:
		val = string(data)
	case []byte:
		dst := make([]byte, len(data))
		copy(dst, data)
		val = dst
	default:
		val2 := new(T)
		_ = json.Unmarshal(data, val2)
		return *val2
	}
	return val.(T)
}

func parseEvent(fn func(*Event[[]byte]) error, reader io.ReadCloser) error {
	scanner := bufio.NewScanner(reader)
	// split data from \n
	scanner.Split(func(data []byte, atEOF bool) (int, []byte, error) {
		if pos := bytes.IndexByte(data, '\n'); pos >= 0 {
			return pos + 1, data[0:pos], nil
		}
		if atEOF && len(data) > 0 {
			return len(data), data, nil
		}
		return 0, nil, nil
	})

	e := &Event[[]byte]{}
	for scanner.Scan() {
		data := scanner.Bytes()
		switch {
		case bytes.HasPrefix(data, []byte("data: ")):
			e.Data = append(e.Data, data[6:]...)
		case bytes.HasPrefix(data, []byte("id: ")):
			e.ID = GetAnyByString[int](string(data[4:]))
		case bytes.HasPrefix(data, []byte("event: ")):
			e.Event = string(data[7:])
		case bytes.HasPrefix(data, []byte("retry: ")):
			e.Retry = GetAnyByString[int](string(data[7:]))
		case bytes.HasPrefix(data, []byte(": ")):
			e.Comment = string(data[1:])
		case len(data) == 0:
			err := fn(e)
			if err != nil {
				return err
			}
			e = &Event[[]byte]{Data: e.Data[0:0]}
		default:
			return fmt.Errorf(ErrClientParseEventInvalid, data)
		}
	}

	return scanner.Err()
}

// EventHub defines the Websocket or [Event] message center.
type EventHub[T any] interface {
	// The Topics method returns all currently registered topics.
	Topics() []string
	// The Subscribe method uses chan to subscribe,
	// and the callback can obtain the current last message and
	// implement processing [HeaderLastEventID].
	//
	// The callback function will block the topic corresponding to [EventHub].
	Subscribe(name string, ch chan<- T, callback func(T))
	// The Unsubscribe method closes and cancels the chan subscription.
	Unsubscribe(name string, ch chan<- T, callback func(T))
	Publish(name string, data T)
	Broadcast(data T)
}

type MetadataEventHub struct {
	Health bool     `json:"health" protobuf:"1,name=health" yaml:"health"`
	Name   string   `json:"name" protobuf:"2,name=name" yaml:"name"`
	Topics []string `json:"topics" protobuf:"3,name=topics" yaml:"topics"`
}

type eventHub[T any] struct {
	topics  sync.Map
	timeout time.Duration
	ticker  []*time.Ticker
}

// NewEventHub creates the default [EventHub].
//
// If the Publish message is timed out, the message will be discarded;
// if the topic sends message to the chan messages,
// the chan will be removed if the timeout occurs.
func NewEventHub[T any](timeout time.Duration) EventHub[T] {
	return &eventHub[T]{
		timeout: timeout,
	}
}

// NewEventHubWithOptions creates [EventHub] with additional support for
// heartbeat and cleanup.
//
// Each cleanup cycle will check all topics.
// If there is no chan connection,
// the topic will be closed. Do not set the interval too low.
//
// Each heartbeat cycle will send heartdata to all topics.
func NewEventHubWithOptions[T any](timeout, cleanup,
	heartbeat time.Duration, heartdata T,
) EventHub[T] {
	hub := &eventHub[T]{
		timeout: timeout,
	}
	if cleanup != 0 {
		ticker := time.NewTicker(cleanup)
		hub.ticker = append(hub.ticker, ticker)
		go hub.runRange(ticker, func(key, val any) bool {
			t := val.(*eventTopic[T])
			if atomic.LoadInt64(&t.state) == 1 {
				t.operate <- topicOp[T]{4, nil, nil}
			} else {
				hub.topics.Delete(key)
			}
			return true
		})
	}
	if heartbeat != 0 {
		ticker := time.NewTicker(heartbeat)
		hub.ticker = append(hub.ticker, ticker)
		go hub.runRange(ticker, func(_, val any) bool {
			val.(*eventTopic[T]).send(heartdata)
			return true
		})
	}
	return hub
}

func (hub *eventHub[T]) runRange(ticker *time.Ticker, fn func(_, val any) bool) {
	for range ticker.C {
		hub.topics.Range(fn)
	}
}

func (hub *eventHub[T]) Unmount(context.Context) {
	hub.topics.Range(func(key, value any) bool {
		hub.topics.Delete(key)
		t := value.(*eventTopic[T])
		close(t.quit)
		return true
	})
	for _, t := range hub.ticker {
		t.Stop()
	}
}

func (hub *eventHub[T]) Metadata() any {
	return MetadataEventHub{
		Health: true,
		Name:   "eudore.eventHub",
		Topics: hub.Topics(),
	}
}

func (hub *eventHub[T]) Topics() []string {
	var names []string
	hub.topics.Range(func(key, _ any) bool {
		names = append(names, key.(string))
		return true
	})
	return names
}

func (hub *eventHub[T]) Subscribe(name string, ch chan<- T, call func(T)) {
	t := hub.getTopic(name)
	for atomic.LoadInt64(&t.state) != 2 {
		select {
		case t.operate <- topicOp[T]{1, ch, call}:
			return
		case <-time.After(time.Millisecond * 10):
			if atomic.CompareAndSwapInt64(&t.state, 0, 1) {
				go t.Run()
			}
		}
	}
}

func (hub *eventHub[T]) Unsubscribe(name string, ch chan<- T, call func(T)) {
	t := hub.getTopic(name)
	if atomic.LoadInt64(&t.state) == 1 {
		t.operate <- topicOp[T]{2, ch, call}
	}
}

func (hub *eventHub[T]) Publish(name string, data T) {
	hub.getTopic(name).send(data)
}

func (hub *eventHub[T]) Broadcast(data T) {
	hub.topics.Range(func(_, val any) bool {
		val.(*eventTopic[T]).send(data)
		return true
	})
}

func (hub *eventHub[T]) getTopic(name string) *eventTopic[T] {
	v, ok := hub.topics.Load(name)
	if ok {
		return v.(*eventTopic[T])
	}

	t := &eventTopic[T]{
		name:    name,
		quit:    make(chan struct{}),
		data:    make(chan T),
		operate: make(chan topicOp[T]),
		timeout: hub.timeout,
	}
	v, ok = hub.topics.LoadOrStore(name, t)
	if ok {
		return v.(*eventTopic[T])
	}
	return t
}

type eventTopic[T any] struct {
	name     string
	state    int64
	quit     chan struct{}
	data     chan T
	connects []chan<- T
	operate  chan topicOp[T]
	last     T
	timeout  time.Duration
}

type topicOp[T any] struct {
	kind int
	ch   chan<- T
	call func(T)
}

func (t *eventTopic[T]) Run() {
	var zero T
	t.last = zero
	timeout := time.NewTimer(time.Hour)
	timeout.Stop()
	for {
		select {
		case <-t.quit:
			atomic.StoreInt64(&t.state, 2)
			return
		case op := <-t.operate:
			if t.handleOperate(op) {
				atomic.StoreInt64(&t.state, 0)
				return
			}
		case data := <-t.data:
			t.last = data
			for i := len(t.connects) - 1; i >= 0; i-- {
				conn := t.connects[i]
				select {
				case conn <- data:
					continue
				default:
					timeout.Reset(t.timeout)
				}

				select {
				case conn <- data:
					if !timeout.Stop() {
						<-timeout.C
					}
				case <-timeout.C:
					t.connects = append(t.connects[:i], t.connects[i+1:]...)
					close(conn)
				}
			}
		}
	}
}

func (t *eventTopic[T]) send(data T) {
	if atomic.LoadInt64(&t.state) == 1 {
		select {
		case t.data <- data:
			return
		default:
		}
		select {
		case t.data <- data:
		case <-time.After(t.timeout):
		}
	}
}

func (t *eventTopic[T]) handleOperate(op topicOp[T]) bool {
	if op.call != nil {
		op.call(t.last)
	}
	switch op.kind {
	case 1:
		t.connects = append(t.connects, op.ch)
	case 2:
		index := sliceIndex(t.connects, op.ch)
		if index != -1 {
			copy(t.connects[index:], t.connects[index+1:])
			t.connects = t.connects[:len(t.connects)-1]
			close(op.ch)
		}
	case 4:
		if len(t.connects) == 0 {
			return true
		}
	}
	return false
}
