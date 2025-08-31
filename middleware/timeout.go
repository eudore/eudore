package middleware

/*
Implementation issues:
- Timeout status code exception when writing.
- Panic stack cannot capture information exceptions.
- http.Header concurrent reading and writing.
- Context data race detection.
- sync.Pool recycles Context.
*/

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/eudore/eudore"
)

// NewTimeoutFunc function creates middleware to implement the given handler
// request time limit.
//
// [eudore.ResponseWriter] body is written to memory buffer.
// handle files is not recommended.
//
// Return [eudore.StatusServiceUnavailable] when the handler timeout.
//
// [responseWriterTimeout] implements the Body method,
// which can return the written body.
//
// NewTimeoutFunc supports the [http.Pusher] interface but does not support
// the [http.Hijacker] or [http.Flusher] interfaces.
//
// refer: [http.TimeoutHandler] [NewTimeoutSkipFunc].
//
//go:noinline
//nolint:funlen
func NewTimeoutFunc(pool *sync.Pool, timeout time.Duration) Middleware {
	release := func(c2 eudore.Context, done chan stackError) {
		r := recover()
		if r != nil {
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf(DefaultRecoveryErrorFormat, r)
			}
			done <- eudore.NewErrorWithDepth(err, 4).(stackError) //nolint:errorlint
			c2.End()
		}
		if c2.Request().MultipartForm != nil {
			_ = c2.Request().MultipartForm.RemoveAll()
		}
		close(done)
		pool.Put(c2)
	}
	start := func(c2 eudore.Context, done chan stackError) {
		defer release(c2, done)
		c2.Next()
	}
	return func(c1 eudore.Context) {
		done := make(chan stackError)
		w := &responseWriterTimeout{c: 200, h: http.Header{}, p: c1.Response()}
		ctx, cancel := context.WithTimeout(c1.Context(), timeout)
		headerCopy(w.h, w.p.Header())
		defer cancel()
		defer c1.End()

		c2 := pool.Get().(eudore.Context)
		c2.Reset(nil, c1.Request().WithContext(c1.Request().Context()))
		c2.SetContext(ctx)
		c2.SetResponse(w)
		c2.SetHandlers(c1.GetHandlers())
		p2 := c2.Params()
		*p2 = append((*p2)[0:0], *c1.Params()...)
		go start(c2, done)

		select {
		case err, ok := <-done:
			p1 := c1.Params()
			*p1 = append((*p1)[0:0], *c2.Params()...)
			if ok {
				stack := eudore.GetCallerStacks(2)
				stack[0] = strings.Replace(stack[0], ":87", ":80", 1)
				panic(eudore.NewErrorWithStack(err.Unwrap(), append(err.Stack(), stack...)))
			}

			to := c1.Response()
			th := to.Header()
			size := w.Size()
			for k, vv := range w.h {
				th[k] = vv
			}
			// should HeaderContentLength be set?
			to.WriteHeader(w.c)
			if size > 0 {
				_, _ = to.Write(w.buf)
			}
		case <-ctx.Done():
			err := ctx.Err()
			writePage(c1, eudore.StatusServiceUnavailable,
				DefaultPageTimeout, timeout.String(),
			)
			w.Lock()
			defer w.Unlock()
			if errors.Is(err, context.DeadlineExceeded) {
				w.e = http.ErrHandlerTimeout
			} else {
				w.e = err
			}
		}
	}
}

// NewTimeoutSkipFunc function creates middleware to implement
// conditional skip [NewTimeoutFunc].
//
// Skip websocket and sse by default.
//
// refer: [NewTimeoutFunc].
func NewTimeoutSkipFunc(pool *sync.Pool, timeout time.Duration,
	fn func(eudore.Context) bool,
) Middleware {
	if fn == nil {
		fn = func(ctx eudore.Context) bool {
			// skip websocket and sse
			return ctx.GetHeader(eudore.HeaderConnection) ==
				eudore.HeaderValueUpgrade ||
				ctx.GetHeader(eudore.HeaderAccept) ==
					eudore.MimeTextEventStream
		}
	}
	fntimeout := NewTimeoutFunc(pool, timeout)
	return func(ctx eudore.Context) {
		if !fn(ctx) {
			fntimeout(ctx)
		}
	}
}

type responseWriterTimeout struct {
	sync.Mutex
	buffer
	e error
	c int
	h http.Header
	p eudore.ResponseWriter
}

func (w *responseWriterTimeout) Unwrap() http.ResponseWriter {
	return w.p
}

func (w *responseWriterTimeout) Write(p []byte) (int, error) {
	w.Lock()
	defer w.Unlock()
	if w.e != nil {
		return 0, w.e
	}
	return w.buffer.Write(p)
}

func (w *responseWriterTimeout) WriteString(p string) (int, error) {
	w.Lock()
	defer w.Unlock()
	if w.e != nil {
		return 0, w.e
	}
	return w.buffer.WriteString(p)
}

func (w *responseWriterTimeout) WriteStatus(code int) {
	if code > 0 {
		w.c = code
	}
}

func (w *responseWriterTimeout) WriteHeader(code int) {
	if code > 0 {
		w.c = code
	}
}

func (w *responseWriterTimeout) Header() http.Header { return w.h }

// The Body method returns the written response body.
//
// I don't know the purpose.
func (w *responseWriterTimeout) Body() []byte {
	return w.buf
}

// The Flush method is not supported.
func (w *responseWriterTimeout) Flush() {}

// The Hijack method is not supported.
func (w *responseWriterTimeout) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, eudore.ErrContextNotHijacker
}

// The Push method implements the [http.Psuher] interface.
//
// support of HTTP/2 Server Push will be disabled by default in
// Chrome 106 and other Chromium-based browsers.
func (w *responseWriterTimeout) Push(p string, opts *http.PushOptions) error {
	w.Lock()
	defer w.Unlock()
	if w.e != nil {
		return w.e
	}
	return w.p.Push(p, opts)
}

// The Size method returns the length of the data written.
func (w *responseWriterTimeout) Size() int {
	return len(w.buf)
}

// The Status method returns the set http status code.
func (w *responseWriterTimeout) Status() int {
	return w.c
}
