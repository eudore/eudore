package middleware

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/eudore/eudore"
)

// NewCompressGzipFunc 函数创建一个gzip压缩处理函数。
func NewCompressGzipFunc() eudore.HandlerFunc {
	return NewCompressFunc(CompressNameGzip, func() any {
		return gzip.NewWriter(io.Discard)
	})
}

// NewCompressDeflateFunc 函数创建一个deflate压缩处理函数。
func NewCompressDeflateFunc() eudore.HandlerFunc {
	return NewCompressFunc(CompressNameDeflate, func() any {
		w, _ := flate.NewWriter(io.Discard, flate.DefaultCompression)
		return w
	})
}

func newResponseWriterCompressPool(name string, fn func() any) *sync.Pool {
	return &sync.Pool{
		New: func() any {
			return &responseWriterCompress{
				Name:   name,
				Writer: fn().(compressor),
				Buffer: make([]byte, CompressBufferLength),
			}
		},
	}
}

// NewCompressFunc 函数创建一个压缩处理函数，需要指定压缩算法和压缩对象构造函数。
func NewCompressFunc(name string, fn func() any) eudore.HandlerFunc {
	pool := newResponseWriterCompressPool(name, fn)
	return func(ctx eudore.Context) {
		// 检查是否使用压缩
		if !strings.Contains(ctx.GetHeader(eudore.HeaderAcceptEncoding), name) ||
			ctx.Response().Header().Get(eudore.HeaderContentEncoding) != "" ||
			strings.Contains(ctx.GetHeader(eudore.HeaderConnection), "Upgrade") ||
			strings.Contains(ctx.GetHeader(eudore.HeaderContentType), "text/event-stream") {
			return
		}

		handlerCompress(ctx, pool)
	}
}

// NewCompressMixinsFunc 函数创建一个混合压缩处理函数，默认具有gzip、defalte。
//
// 如果压缩ResponseWriter.Size()值为压缩后size。
//
// 如果设置middleware.DefaultComoressBrotliFunc指定brotli压缩函数，追加br压缩。
//
// HeaderAcceptEncoding值忽略非零权重值，顺序优先。
func NewCompressMixinsFunc(compresss map[string]func() any) eudore.HandlerFunc {
	if compresss == nil {
		compresss = make(map[string]func() any)
		compresss[CompressNameGzip] = func() any {
			return gzip.NewWriter(io.Discard)
		}
		compresss[CompressNameDeflate] = func() any {
			w, _ := flate.NewWriter(io.Discard, flate.DefaultCompression)
			return w
		}
		if DefaultComoressBrotliFunc != nil {
			compresss[CompressNameBrotli] = DefaultComoressBrotliFunc
		}
	}
	names := make([]string, 0, len(compresss))
	pools := make([]*sync.Pool, 0, len(compresss))
	for name := range compresss {
		names = append(names, name)
		pools = append(pools, newResponseWriterCompressPool(name, compresss[name]))
	}

	return func(ctx eudore.Context) {
		encoding := ctx.GetHeader(eudore.HeaderAcceptEncoding)
		if encoding == "" || encoding == CompressNameIdentity ||
			ctx.Response().Header().Get(eudore.HeaderContentEncoding) != "" ||
			strings.Contains(ctx.GetHeader(eudore.HeaderConnection), "Upgrade") ||
			strings.Contains(ctx.GetHeader(eudore.HeaderContentType), "text/event-stream") {
			return
		}

		for _, encoding := range strings.Split(encoding, ",") {
			name, quality, ok := strings.Cut(strings.TrimSpace(encoding), ";")
			if ok && quality == "q=0" {
				continue
			}
			for i := range names {
				if names[i] == name {
					handlerCompress(ctx, pools[i])
				}
			}
		}
	}
}

func handlerCompress(ctx eudore.Context, pool *sync.Pool) {
	// 初始化ResponseWriter
	w := pool.Get().(*responseWriterCompress)
	w.ResponseWriter = ctx.Response()
	w.Buffer = w.Buffer[0:0]
	w.State = CompressStateUnknown
	defer w.Close(pool)

	ctx.SetResponse(w)
	ctx.Next()
}

// responseWriterCompress 定义压缩响应，实现ResponseWriter接口。
type responseWriterCompress struct {
	eudore.ResponseWriter
	Writer compressor
	State  int
	Buffer []byte
	Name   string
}

type compressor interface {
	Reset(io.Writer)
	Write([]byte) (int, error)
	Flush() error
	Close() error
}

// Unwrap 方法返回原始http.ResponseWrite对象。
func (w *responseWriterCompress) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

// Write 实现ResponseWriter中的Write方法。
func (w *responseWriterCompress) Write(data []byte) (int, error) {
	switch w.State {
	case CompressStateEnable:
		return w.Writer.Write(data)
	case CompressStateDisable:
		return w.ResponseWriter.Write(data)
	default:
		if len(data)+len(w.Buffer) <= CompressBufferLength {
			w.State = CompressStateBuffer
			w.Buffer = append(w.Buffer, data...)
			return len(data), nil
		}

		w.init()
		return w.Write(data)
	}
}

func (w *responseWriterCompress) WriteString(data string) (int, error) {
	switch w.State {
	case CompressStateEnable:
		// WriteString only gzip
		return io.WriteString(w.Writer, data)
	case CompressStateDisable:
		return w.ResponseWriter.WriteString(data)
	default:
		if len(data)+len(w.Buffer) <= CompressBufferLength {
			w.State = CompressStateBuffer
			w.Buffer = append(w.Buffer, data...)
			return len(data), nil
		}

		w.init()
		return w.WriteString(data)
	}
}

func (w *responseWriterCompress) WriteHeader(code int) {
	if code == 200 {
		return
	}
	if w.State < CompressStateEnable {
		contentlength := w.ResponseWriter.Header().Get(eudore.HeaderContentLength)
		if contentlength == "" || len(contentlength) > 3 {
			w.init()
		}
		w.ResponseWriter.WriteHeader(code)
	}
}

// Flush 实现ResponseWriter中的Flush方法。
func (w *responseWriterCompress) Flush() {
	switch w.State {
	case CompressStateEnable:
		w.Writer.Flush()
	case CompressStateBuffer:
		w.init()
		w.Writer.Flush()
	}
	w.ResponseWriter.Flush()
}

func (w *responseWriterCompress) Close(pool *sync.Pool) {
	switch w.State {
	case CompressStateEnable:
		w.Writer.Close()
		w.Writer.Reset(io.Discard)
	case CompressStateBuffer:
		w.ResponseWriter.Write(w.Buffer)
	}
	pool.Put(w)
}

func (w *responseWriterCompress) init() {
	h := w.ResponseWriter.Header()
	contenttype := h.Get(eudore.HeaderContentType)
	pos := strings.IndexByte(contenttype, ';')
	if pos != -1 {
		contenttype = contenttype[:pos]
	}

	if DefaultComoressDisableMime[contenttype] || h.Get(eudore.HeaderContentEncoding) != "" {
		w.State = CompressStateDisable
	} else {
		w.State = CompressStateEnable
		w.Writer.Reset(w.ResponseWriter)
		h.Del(eudore.HeaderContentLength)
		h.Set(eudore.HeaderContentEncoding, w.Name)
		h.Set(eudore.HeaderVary, strings.Join(append(h.Values(eudore.HeaderVary), eudore.HeaderAcceptEncoding), ", "))
	}
	if len(w.Buffer) > 0 {
		w.Write(w.Buffer)
	}
}

// Push 方法给Push Header设置HeaderAcceptEncoding。
func (w *responseWriterCompress) Push(target string, opts *http.PushOptions) error {
	switch {
	case opts == nil:
		opts = &http.PushOptions{Header: http.Header{eudore.HeaderAcceptEncoding: {w.Name}}}
	case opts.Header == nil:
		opts.Header = http.Header{eudore.HeaderAcceptEncoding: {w.Name}}
	case opts.Header.Get(eudore.HeaderAcceptEncoding) == "":
		opts.Header.Add(eudore.HeaderAcceptEncoding, w.Name)
	}
	return w.ResponseWriter.Push(target, opts)
}
