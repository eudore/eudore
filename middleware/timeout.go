package middleware

import (
	"bytes"
	"errors"
	"github.com/eudore/eudore"
	"sync"
	"time"
)

// NewTimeoutFunc 函数创建一个处理超时中间件。
func NewTimeoutFunc(t time.Duration) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		tw := &timeoutWriter{
			ResponseWriter: ctx.Response(),
		}
		ctx.SetResponse(tw)

		done := make(chan struct{})
		go func() {
			ctx.Next()
			close(done)
		}()
		select {
		case <-done:
			tw.WriteSuccess()
		case <-time.After(t):
			ctx.End()
			tw.WriteFatal()
		case <-ctx.Done():
			ctx.End()
			tw.WriteFatal()
		}

	}
}

// ErrHandlerTimeout is returned on ResponseWriter Write calls
// in handlers which have timed out.
var ErrHandlerTimeout = errors.New("http: Handler timeout")

// TimeoutBody 是超时返回的响应body
var TimeoutBody = []byte("<html><head><title>Timeout</title></head><body><h1>Timeout</h1></body></html>")

type timeoutWriter struct {
	eudore.ResponseWriter
	wbuf bytes.Buffer

	mu          sync.Mutex
	timedOut    bool
	wroteHeader bool
	code        int
}

func (tw *timeoutWriter) Write(p []byte) (int, error) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut {
		return 0, ErrHandlerTimeout
	}
	if !tw.wroteHeader {
		tw.wroteHeader = true
		tw.code = 200
	}
	return tw.wbuf.Write(p)
}

func (tw *timeoutWriter) WriteHeader(code int) {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	if tw.timedOut || tw.wroteHeader {
		return
	}
	tw.wroteHeader = true
	tw.code = code
}

func (tw *timeoutWriter) WriteSuccess() {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	tw.ResponseWriter.WriteHeader(tw.code)
	tw.ResponseWriter.Write(tw.wbuf.Bytes())
}
func (tw *timeoutWriter) WriteFatal() {
	tw.mu.Lock()
	defer tw.mu.Unlock()
	tw.ResponseWriter.WriteHeader(eudore.StatusServiceUnavailable)
	tw.ResponseWriter.Write(TimeoutBody)
}
