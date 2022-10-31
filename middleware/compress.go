package middleware

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"

	"github.com/eudore/eudore"
)

// NewCompressFunc 函数创建一个压缩响应处理函数，需要指定压缩算法和压缩对象构造函数。
//
// import: "github.com/andybalholm/brotli"
//
// br: middleware.NewCompressFunc("br", func() interface{} { return brotli.NewWriter(ioutil.Discard) }),
//
// gzip: middleware.NewCompressGzipFunc(5)
//
// deflate: middleware.NewCompressDeflateFunc(5)
func NewCompressFunc(name string, fn func() interface{}) eudore.HandlerFunc {
	pool := sync.Pool{New: fn}
	return func(ctx eudore.Context) {
		// 检查是否使用name压缩
		if shouldNotCompress(ctx, name) {
			ctx.Next()
			return
		}
		// 初始化ResponseWriter
		w := &responseCompress{
			ResponseWriter: ctx.Response(),
			Writer:         pool.Get().(compressor),
			Name:           name,
		}
		w.Writer.Reset(ctx.Response())

		ctx.SetResponse(w)
		ctx.SetHeader(eudore.HeaderContentEncoding, name)
		ctx.SetHeader(eudore.HeaderVary, eudore.HeaderAcceptEncoding)
		ctx.Next()

		w.Writer.Close()
		pool.Put(w.Writer)
	}
}

// NewCompressGzipFunc 函数创建一个gzip压缩响应处理函数,如果压缩级别超出gzip范围默认使用5。
func NewCompressGzipFunc(level int) eudore.HandlerFunc {
	if level < gzip.HuffmanOnly || level > gzip.BestCompression {
		level = 5
	}
	return NewCompressFunc("gzip", func() interface{} {
		gz, _ := gzip.NewWriterLevel(ioutil.Discard, level)
		return gz
	})
}

// NewCompressDeflateFunc 函数创建一个deflate压缩响应处理函数,如果压缩级别超出deflate范围默认使用5。
func NewCompressDeflateFunc(level int) eudore.HandlerFunc {
	if level < flate.HuffmanOnly || level > flate.BestCompression {
		level = 5
	}
	return NewCompressFunc("deflate", func() interface{} {
		gz, _ := flate.NewWriter(ioutil.Discard, level)
		return gz
	})
}

// responseCompress 定义Gzip响应，实现ResponseWriter接口
type responseCompress struct {
	eudore.ResponseWriter
	Writer compressor
	Name   string
}

type compressor interface {
	Reset(io.Writer)
	Write([]byte) (int, error)
	Flush() error
	Close() error
}

// Write 实现ResponseWriter中的Write方法。
func (w responseCompress) Write(data []byte) (int, error) {
	return w.Writer.Write(data)
}

// Flush 实现ResponseWriter中的Flush方法。
func (w responseCompress) Flush() {
	w.Writer.Flush()
	w.ResponseWriter.Flush()
}

func shouldNotCompress(ctx eudore.Context, name string) bool {
	h := ctx.Request().Header
	if !strings.Contains(h.Get(eudore.HeaderAcceptEncoding), name) ||
		strings.Contains(h.Get(eudore.HeaderConnection), "Upgrade") ||
		strings.Contains(h.Get(eudore.HeaderContentType), "text/event-stream") {

		return true
	}

	return ctx.Response().Header().Get(eudore.HeaderContentEncoding) != ""
}

// Push initiates an HTTP/2 server push.
// Push returns ErrNotSupported if the client has disabled push or if push
// is not supported on the underlying connection.
func (w *responseCompress) Push(target string, opts *http.PushOptions) error {
	return w.ResponseWriter.Push(target, w.setAcceptEncodingForPushOptions(opts))
}

// setAcceptEncodingForPushOptions sets "Accept-Encoding" : "gzip" for PushOptions without overriding existing headers.
func (w *responseCompress) setAcceptEncodingForPushOptions(opts *http.PushOptions) *http.PushOptions {
	if opts == nil {
		opts = &http.PushOptions{
			Header: http.Header{
				eudore.HeaderAcceptEncoding: []string{w.Name},
			},
		}
		return opts
	}

	if opts.Header == nil {
		opts.Header = http.Header{
			eudore.HeaderAcceptEncoding: []string{w.Name},
		}
		return opts
	}

	if encoding := opts.Header.Get(eudore.HeaderAcceptEncoding); encoding == "" {
		opts.Header.Add(eudore.HeaderAcceptEncoding, w.Name)
		return opts
	}

	return opts
}
