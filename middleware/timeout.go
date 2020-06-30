package middleware

/*
实现难点：写入中超时状态码异常、panic栈无法捕捉信息异常、http.Header并发读写、sync.Pool回收了Context
*/
import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/eudore/eudore"
)

// NewTimeoutFunc 函数创建一个处理超时中间件。
func NewTimeoutFunc(t time.Duration) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		ctxt, cancel := context.WithTimeout(ctx.GetContext(), t)
		defer cancel()
		ctx.WithContext(ctxt)

		w := &timeoutWriter{
			ResponseWriter: ctx.Response(),
			header:         make(http.Header),
			code:           200,
		}
		copyHeader(ctx.Response().Header(), w.header)
		ctx.SetResponse(w)

		done := make(chan struct{})
		panicChan := make(chan interface{}, 1)
		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChan <- &timeoutError{
						error: fmt.Sprint(p),
						stack: eudore.GetPanicStack(6),
					}
				}
			}()
			ctx.Next()
			close(done)
		}()
		select {
		case p := <-panicChan:
			panic(p)
		case <-done:
			w.WriteSuccess()
		case <-ctx.Done():
			ctx.End()
			w.WriteFatal()
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
	code   int
	header http.Header
	buffer bytes.Buffer

	mu          sync.Mutex
	timeout     bool
	wroteHeader bool
}

func (w *timeoutWriter) Header() http.Header {
	return w.header
}

func (w *timeoutWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timeout {
		return 0, ErrHandlerTimeout
	}
	if !w.wroteHeader {
		w.wroteHeader = true
	}
	return w.buffer.Write(p)
}

func (w *timeoutWriter) WriteHeader(code int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timeout || w.wroteHeader {
		return
	}
	w.code = code
}

func (w *timeoutWriter) Size() int {
	return w.buffer.Len()
}
func (w *timeoutWriter) Status() int {
	return w.code
}

func (w *timeoutWriter) WriteSuccess() {
	w.mu.Lock()
	defer w.mu.Unlock()
	copyHeader(w.header, w.ResponseWriter.Header())
	w.ResponseWriter.WriteHeader(w.code)
	w.ResponseWriter.Write(w.buffer.Bytes())
}

func (w *timeoutWriter) WriteFatal() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.timeout = true
	w.ResponseWriter.WriteHeader(eudore.StatusServiceUnavailable)
	w.ResponseWriter.Write(TimeoutBody)
}

func copyHeader(src, dst http.Header) {
	for k, vv := range src {
		dst[k] = vv
	}
}

type timeoutError struct {
	error string
	stack []string
}

func (err *timeoutError) Error() string {
	return err.error
}

func (err *timeoutError) GetStack() []string {
	return err.stack
}
