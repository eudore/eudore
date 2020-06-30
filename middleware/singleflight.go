package middleware

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"net/http"
	"sync"

	"github.com/eudore/eudore"
)

// 定义SingleFlight响应错误
var (
	ErrsingleFlightResponseHijack = errors.New("SingleFlight not support eudore.ReadWriter.Hijack method")
	ErrsingleFlightResponsePush   = errors.New("SingleFlight not support eudore.ReadWriter.Push method")
)

// NewSingleFlightFunc 函数创建一个SingleFlight处理函数。
func NewSingleFlightFunc() eudore.HandlerFunc {
	mu := sync.Mutex{}
	calls := make(map[string]*singleFlightResponse)
	return func(ctx eudore.Context) {
		// 非幂等方法不允许启用SingleFlight
		switch ctx.Method() {
		case eudore.MethodPost, eudore.MethodPatch:
			return
		}

		key := ctx.Path()
		mu.Lock()
		if call, ok := calls[key]; ok {
			mu.Unlock()
			call.Wait()
			call.WriteData(ctx.Response())
			ctx.End()
			return
		}

		call := &singleFlightResponse{
			header: make(http.Header),
			code:   200,
		}
		call.Add(1)
		calls[key] = call
		mu.Unlock()

		w := ctx.Response()
		ctx.SetResponse(call)
		ctx.Next()
		ctx.SetResponse(w)
		call.Done()
		call.WriteData(w)

		mu.Lock()
		delete(calls, key)
		mu.Unlock()
	}
}

// singleFlightResponse 定义SingleFlight请求写入的响应。
type singleFlightResponse struct {
	sync.WaitGroup
	code   int
	header http.Header
	buffer bytes.Buffer
}

// WriteData 方法将SingleFlight响应数据写入到请求响应。
func (w *singleFlightResponse) WriteData(resp eudore.ResponseWriter) {
	resp.WriteHeader(w.code)
	resp.Write(w.buffer.Bytes())
}

// Header 方法返回响应header。
func (w *singleFlightResponse) Header() http.Header {
	return w.header
}

// Write 方法写入数据。
func (w *singleFlightResponse) Write(p []byte) (int, error) {
	return w.buffer.Write(p)
}

// WriteHeader 方法设置响应状态码。
func (w *singleFlightResponse) WriteHeader(code int) {
	w.code = code
}

// Flush 方法立刻写入请求数据，SingleFlight方法未实现改方法。
func (w *singleFlightResponse) Flush() {
	// Do nothing because singleFlightResponse not support flush.
}

// Hijack 方法劫持请求net.Conn连接，SingleFlight不支持Hijack。
func (w *singleFlightResponse) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, ErrsingleFlightResponseHijack
}

// Push 方法返回一条push请求，SingleFlight不支持Push。
func (w *singleFlightResponse) Push(string, *http.PushOptions) error {
	return ErrsingleFlightResponsePush
}

// Size 方法返回写入数据长度。
func (w *singleFlightResponse) Size() int {
	return w.buffer.Len()
}

// Status 方法返回响应状态码。
func (w *singleFlightResponse) Status() int {
	return w.code
}
