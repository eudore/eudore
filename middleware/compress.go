package middleware

import (
	"compress/flate"
	"compress/gzip"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/eudore/eudore"
)

const (
	compressionStateUnknown = iota
	compressionStateBuffer
	compressionStateEnable
	compressionStateDisable
)

func compressionWriterGzip() any {
	return gzip.NewWriter(nil)
}

func compressionWriterFlate() any {
	w, _ := flate.NewWriter(nil, flate.DefaultCompression)
	return w
}

// The NewGzipFunc function creates middleware to implement [gzip] compress of
// response body.
//
// refer: [NewCompressionFunc].
//
//go:noinline
func NewGzipFunc() Middleware {
	return NewCompressionFunc(CompressionNameGzip, compressionWriterGzip)
}

// The NewCompressionFunc function creates middleware to implement specify
// method response body compress.
// It is necessary to specify the compress name and [compressor] constructor.
//
// Use [eudore.HeaderAcceptEncoding] to negotiate whether to compress the
// response body.
//
// Disable compression in the following cases.
//
// 1. Websocket	The [net.Conn] used does not have a response body.
//
// 2. Small body	If the body length is less than [CompressionBufferLength],
// the compression effect will be poor.
//
// 3. compress data		If [eudore.HeaderContentType] specifies that the body
// format is self-compressed data,
// The compression effect is poor.
// The mime define: [DefaultCompressionDisableMime].
//
//go:noinline
func NewCompressionFunc(name string, fn func() any) Middleware {
	const ae = eudore.HeaderAcceptEncoding
	const conn = eudore.HeaderConnection
	pool := newCompressPool(name, fn)
	return func(ctx eudore.Context) {
		// check enable compression
		if !strings.Contains(ctx.GetHeader(ae), name) ||
			ctx.GetHeader(conn) == eudore.HeaderValueUpgrade {
			return
		}

		handlerCompress(ctx, pool)
	}
}

// NewCompressionMixinsFunc function creates middleware to implement
// mixed compress response body,
//
// The default compresss is: [DefaultCompressionEncoder].
//
// The default compress is [CompressionNameGzip] and [CompressionNameDeflate],
// and the compress [CompressionNameZstandard] and [CompressionNameBrotli] require a
// specified constructor.
//
// [eudore.HeaderAcceptEncoding] value ignores non-zero weight values,
// and the order is based on [DefaultCompressionOrder].
//
// refer: [NewCompressionFunc] [NewGzipFunc].
func NewCompressionMixinsFunc(compresss map[string]func() any) Middleware {
	names, pools := initCompress(compresss)
	return func(ctx eudore.Context) {
		encoding := ctx.GetHeader(eudore.HeaderAcceptEncoding)
		if encoding == "" ||
			encoding == CompressionNameIdentity ||
			ctx.GetHeader(eudore.HeaderConnection) ==
				eudore.HeaderValueUpgrade {
			return
		}

		for i := range names {
			if strings.Contains(encoding, names[i]) {
				handlerCompress(ctx, pools[i])
				return
			}
		}
	}
}

func handlerCompress(ctx eudore.Context, pool *sync.Pool) {
	// init ResponseWriter
	w := pool.Get().(*responseWriterCompress)
	w.ResponseWriter = ctx.Response()
	w.Buffer = w.Buffer[0:0]
	w.State = compressionStateUnknown
	defer w.Close(pool)

	ctx.SetResponse(w)
	ctx.Next()
}

func initCompress(compresss map[string]func() any) ([]string, []*sync.Pool) {
	if compresss == nil {
		compresss = DefaultCompressionEncoder
	}

	names := make([]string, 0, len(compresss))
	pools := make([]*sync.Pool, 0, len(compresss))
	for name := range compresss {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		return getCompresssOrder(names[i]) < getCompresssOrder(names[j])
	})

	for _, name := range names {
		pools = append(pools, newCompressPool(name, compresss[name]))
	}
	return names, pools
}

func getCompresssOrder(val string) int {
	for i := range DefaultCompressionOrder {
		if val == DefaultCompressionOrder[i] {
			return i
		}
	}
	return len(DefaultCompressionOrder)
}

func newCompressPool(name string, fn func() any) *sync.Pool {
	return &sync.Pool{
		New: func() any {
			return &responseWriterCompress{
				Name:   name,
				Writer: fn().(compressor),
				Buffer: make([]byte, CompressionBufferLength),
			}
		},
	}
}

// responseWriterCompress defines compressed response and
// implements the [eudore.ResponseWriter] interface.
type responseWriterCompress struct {
	eudore.ResponseWriter
	Writer compressor
	State  int
	Buffer []byte
	Name   string
}

type compressor interface {
	Reset(w io.Writer)
	Write(b []byte) (int, error)
	Flush() error
	Close() error
}

func (w *responseWriterCompress) Close(pool *sync.Pool) {
	switch w.State {
	case compressionStateEnable:
		w.Writer.Close()
		w.Writer.Reset(nil)
	case compressionStateBuffer:
		_, _ = w.ResponseWriter.Write(w.Buffer)
	}
	pool.Put(w)
}

func (w *responseWriterCompress) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *responseWriterCompress) Write(data []byte) (int, error) {
	switch w.State {
	case compressionStateEnable:
		return w.Writer.Write(data)
	case compressionStateDisable:
		return w.ResponseWriter.Write(data)
	default:
		if len(data)+len(w.Buffer) <= CompressionBufferLength {
			w.State = compressionStateBuffer
			w.Buffer = append(w.Buffer, data...)
			return len(data), nil
		}

		w.init()
		return w.Write(data)
	}
}

func (w *responseWriterCompress) WriteString(data string) (int, error) {
	switch w.State {
	case compressionStateEnable:
		// WriteString method only gzip
		return io.WriteString(w.Writer, data)
	case compressionStateDisable:
		return w.ResponseWriter.WriteString(data)
	default:
		if len(data)+len(w.Buffer) <= CompressionBufferLength {
			w.State = compressionStateBuffer
			w.Buffer = append(w.Buffer, data...)
			return len(data), nil
		}

		w.init()
		return w.WriteString(data)
	}
}

func (w *responseWriterCompress) WriteHeader(code int) {
	switch w.State {
	case compressionStateUnknown, compressionStateBuffer:
		w.init()
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriterCompress) Flush() {
	switch w.State {
	case compressionStateEnable:
		w.Writer.Flush()
	case compressionStateBuffer:
		w.init()
		w.Writer.Flush()
	}
	w.ResponseWriter.Flush()
}

func (w *responseWriterCompress) init() {
	h := w.ResponseWriter.Header()
	contenttype := h.Get(eudore.HeaderContentType)
	pos := strings.IndexByte(contenttype, ';')
	if pos != -1 {
		contenttype = contenttype[:pos]
	}

	_, ok := DefaultCompressionDisableMime[contenttype]
	if ok || skipCompress(h) || h.Get(eudore.HeaderContentEncoding) != "" {
		w.State = compressionStateDisable
	} else {
		w.State = compressionStateEnable
		w.Writer.Reset(w.ResponseWriter)
		h.Set(eudore.HeaderContentEncoding, w.Name)
		headerVary(h, eudore.HeaderAcceptEncoding)
	}
	if len(w.Buffer) > 0 {
		_, _ = w.Write(w.Buffer)
	}
}

func skipCompress(h http.Header) bool {
	length := h.Get(eudore.HeaderContentLength)
	if length == "" {
		return false
	}

	v, err := strconv.ParseInt(length, 10, 64)
	if err != nil || v > CompressionBufferLength {
		h.Del(eudore.HeaderContentLength)
		return h.Get(eudore.HeaderContentEncoding) != ""
	}

	return true
}

// The Push method sets [HeaderAcceptEncoding] for the Push Header.
func (w *responseWriterCompress) Push(p string, opts *http.PushOptions) error {
	switch {
	case opts == nil:
		opts = &http.PushOptions{
			Header: http.Header{eudore.HeaderAcceptEncoding: {w.Name}},
		}
	case opts.Header == nil:
		opts.Header = http.Header{eudore.HeaderAcceptEncoding: {w.Name}}
	case opts.Header.Get(eudore.HeaderAcceptEncoding) == "":
		opts.Header.Add(eudore.HeaderAcceptEncoding, w.Name)
	}
	return w.ResponseWriter.Push(p, opts)
}

// The Size method returns the response size,
// otherwise the StateBuffer size is 0.
func (w *responseWriterCompress) Size() int {
	if w.State == compressionStateBuffer {
		return len(w.Buffer)
	}
	return w.ResponseWriter.Size()
}
