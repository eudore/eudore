package middleware

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/eudore/eudore"
)

// NewGzipFunc 创建一个gzip压缩函数,如果压缩级别超出gzip范围默认使用5。
func NewGzipFunc(level int) eudore.HandlerFunc {
	if level < gzip.HuffmanOnly || level > gzip.BestCompression {
		level = 5
	}
	pool := sync.Pool{
		New: func() interface{} {
			gz, _ := gzip.NewWriterLevel(ioutil.Discard, level)
			return gz
		},
	}
	return func(ctx eudore.Context) {
		// 检查是否使用Gzip
		if !shouldCompress(ctx) {
			ctx.Next()
			return
		}
		// 初始化ResponseWriter
		w := &gzipResponse{
			ResponseWriter: ctx.Response(),
			Writer:         pool.Get().(*gzip.Writer),
		}
		w.Writer.Reset(ctx.Response())

		ctx.SetResponse(w)
		ctx.SetHeader(eudore.HeaderContentEncoding, "gzip")
		ctx.SetHeader(eudore.HeaderVary, eudore.HeaderAcceptEncoding)
		ctx.Next()

		w.Writer.Close()
		pool.Put(w.Writer)
	}

}

// gzipResponse 定义Gzip响应，实现ResponseWriter接口
type gzipResponse struct {
	eudore.ResponseWriter
	Writer *gzip.Writer
}

// Write 实现ResponseWriter中的Write方法。
func (w gzipResponse) Write(data []byte) (int, error) {
	return w.Writer.Write(data)
}

// Flush 实现ResponseWriter中的Flush方法。
func (w gzipResponse) Flush() {
	w.Writer.Flush()
	w.ResponseWriter.Flush()
}

func shouldCompress(ctx eudore.Context) bool {
	h := ctx.Request().Header
	if !strings.Contains(h.Get(eudore.HeaderAcceptEncoding), "gzip") ||
		strings.Contains(h.Get(eudore.HeaderConnection), "Upgrade") ||
		strings.Contains(h.Get(eudore.HeaderContentType), "text/event-stream") {

		return false
	}

	h = ctx.Response().Header()
	if strings.Contains(h.Get(eudore.HeaderContentEncoding), "gzip") {
		return false
	}

	return true
}

// Push initiates an HTTP/2 server push.
// Push returns ErrNotSupported if the client has disabled push or if push
// is not supported on the underlying connection.
func (w *gzipResponse) Push(target string, opts *http.PushOptions) error {
	return w.ResponseWriter.Push(target, setAcceptEncodingForPushOptions(opts))
}

// setAcceptEncodingForPushOptions sets "Accept-Encoding" : "gzip" for PushOptions without overriding existing headers.
func setAcceptEncodingForPushOptions(opts *http.PushOptions) *http.PushOptions {
	if opts == nil {
		opts = &http.PushOptions{
			Header: http.Header{
				eudore.HeaderAcceptEncoding: []string{"gzip"},
			},
		}
		return opts
	}

	if opts.Header == nil {
		opts.Header = http.Header{
			eudore.HeaderAcceptEncoding: []string{"gzip"},
		}
		return opts
	}

	if encoding := opts.Header.Get(eudore.HeaderAcceptEncoding); encoding == "" {
		opts.Header.Add(eudore.HeaderAcceptEncoding, "gzip")
		return opts
	}

	return opts
}
