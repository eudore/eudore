package middleware

import (
	"compress/flate"
	"compress/gzip"
	"fmt"
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

// The NewCompressionFunc function creates middleware to implement specify
// method response body compress.
// It is necessary to specify the compress name and [compressor] constructor.
// The actual type of parameter fn is [func() compressor].
// If compressor is not passed the value in [DefaultCompressionEncoder] is used.
//
//	type compressor interface {
//		Reset(w io.Writer)
//		Write(b []byte) (int, error)
//		Flush() error
//		Close() error
//	}
//
// Use [eudore.HeaderAcceptEncoding] to negotiate whether to compress the
// response body.
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
	if fn == nil {
		fn = DefaultCompressionEncoder[name]
	}
	if fn == nil {
		panic(fmt.Errorf(ErrCompressMissingEncoder, name))
	}
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
// refer: [NewCompressionFunc].
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
	w.buffer = w.buffer[0:0]
	w.state = compressionStateUnknown
	w.code = eudore.StatusOK
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
	_, ok := fn().(compressor)
	if !ok {
		panic(fmt.Errorf(ErrCompressInvalidEncoder, name))
	}
	return &sync.Pool{
		New: func() any {
			return &responseWriterCompress{
				name:   name,
				Writer: fn().(compressor),
				buffer: make([]byte, CompressionBufferLength),
			}
		},
	}
}

// responseWriterCompress defines compressed response and
// implements the [eudore.ResponseWriter] interface.
type responseWriterCompress struct {
	eudore.ResponseWriter
	Writer compressor
	state  int
	buffer []byte
	code   int
	name   string
}

type compressor interface {
	Reset(w io.Writer)
	Write(b []byte) (int, error)
	Flush() error
	Close() error
}

func (w *responseWriterCompress) Close(pool *sync.Pool) {
	switch w.state {
	case compressionStateEnable:
		w.Writer.Close()
		w.Writer.Reset(nil)
	case compressionStateDisable:
	case compressionStateBuffer:
		w.writeHeader()
		_, _ = w.ResponseWriter.Write(w.buffer)
	case compressionStateUnknown:
		w.writeHeader()
	}
	pool.Put(w)
}

func (w *responseWriterCompress) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *responseWriterCompress) Write(data []byte) (int, error) {
	switch w.state {
	case compressionStateEnable:
		return w.Writer.Write(data)
	case compressionStateDisable:
		return w.ResponseWriter.Write(data)
	default:
		if len(data)+len(w.buffer) <= CompressionBufferLength {
			w.state = compressionStateBuffer
			w.buffer = append(w.buffer, data...)
			return len(data), nil
		}

		w.init()
		return w.Write(data)
	}
}

func (w *responseWriterCompress) WriteString(data string) (int, error) {
	switch w.state {
	case compressionStateEnable:
		// WriteString method only gzip
		return io.WriteString(w.Writer, data)
	case compressionStateDisable:
		return w.ResponseWriter.WriteString(data)
	default:
		if len(data)+len(w.buffer) <= CompressionBufferLength {
			w.state = compressionStateBuffer
			w.buffer = append(w.buffer, data...)
			return len(data), nil
		}

		w.init()
		return w.WriteString(data)
	}
}

func (w *responseWriterCompress) writeHeader() {
	if w.code > 0 {
		w.ResponseWriter.WriteHeader(w.code)
		w.code = -w.code
	}
}

func (w *responseWriterCompress) WriteStatus(code int) {
	if w.code > 0 && code > 0 {
		w.code = code
	}
}

func (w *responseWriterCompress) WriteHeader(code int) {
	if w.code > 0 && code > 0 {
		w.code = code
	}
}

func (w *responseWriterCompress) Flush() {
	switch w.state {
	case compressionStateEnable:
		w.Writer.Flush()
	case compressionStateDisable:
	default:
		w.init()
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
		w.state = compressionStateDisable
	} else {
		w.state = compressionStateEnable
		w.Writer.Reset(w.ResponseWriter)
		h.Set(eudore.HeaderContentEncoding, w.name)
		headerVary(h, eudore.HeaderAcceptEncoding)
	}
	w.writeHeader()
	if len(w.buffer) > 0 {
		_, _ = w.Write(w.buffer)
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
			Header: http.Header{eudore.HeaderAcceptEncoding: {w.name}},
		}
	case opts.Header == nil:
		opts.Header = http.Header{eudore.HeaderAcceptEncoding: {w.name}}
	case opts.Header.Get(eudore.HeaderAcceptEncoding) == "":
		opts.Header.Add(eudore.HeaderAcceptEncoding, w.name)
	}
	return w.ResponseWriter.Push(p, opts)
}

// The Size method returns the response size,
// otherwise the StateBuffer size is 0.
func (w *responseWriterCompress) Size() int {
	if w.state == compressionStateBuffer {
		return len(w.buffer)
	}
	return w.ResponseWriter.Size()
}

// The Status method returns the set http status code.
func (w *responseWriterCompress) Status() int {
	// abs
	m := w.code >> 31
	return (w.code + m) ^ m
}
