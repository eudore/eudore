package gzip

import (
	"fmt"
	"strings"
	"net/http"
	"io/ioutil"
	"path/filepath"
	"compress/gzip"
	"eudore"
)

type (
	GzipResponse struct {
		eudore.ResponseWriter
		writer	*gzip.Writer
		size 	int
		code	int
		rwcode	bool
	}
)

func GzipFunc(ctx eudore.Context) {
	// 检查是否使用Gzip
	if !shouldCompress(ctx) {
		ctx.Next()
		return
	}
	// 初始化ResponseWriter
	w, err := NewGzipResponse(ctx.Response()) 
	if err != nil {
		// 初始化失败，正常写入
		ctx.Error(err)
		ctx.Next()
		return
	}
	w.Header().Set(eudore.HeaderContentEncoding, "gzip")
	w.Header().Set(eudore.HeaderVary, eudore.HeaderAcceptEncoding)
	ctx.SetResponse(w)
	ctx.Next()
	w.Header().Set(eudore.HeaderContentLength, fmt.Sprint(w.size))
	w.writer.Close()
}


func NewGzipResponse(w eudore.ResponseWriter) (*GzipResponse, error) {
	gz, err := gzip.NewWriterLevel(ioutil.Discard, 5)
	if err != nil {
		return nil, err
	}
	gz.Reset(w)
	return &GzipResponse{w, gz, 0, 200, true}, nil
}

func (w *GzipResponse) Write(data []byte) (int, error) {
	if w.rwcode {
		w.ResponseWriter.WriteHeader(w.code)
		w.rwcode = false
	}

	n, err := w.writer.Write(data)
	w.size += n
	return n, err
}


func (w *GzipResponse) WriteHeader(code int) {
	w.Header().Del(eudore.HeaderContentLength)
	w.code = code
}

func (w *GzipResponse) Flush() {
	w.writer.Flush()
	w.ResponseWriter.Flush()
}

func (w *GzipResponse) Size() int {
	return w.size
}

func (w *GzipResponse) Status() int {
	return w.code
}


func shouldCompress(ctx eudore.Context) bool {
	h := ctx.Request().Header()
	if !strings.Contains(h.Get(eudore.HeaderAcceptEncoding), "gzip") ||
		strings.Contains(h.Get(eudore.HeaderConnection), "Upgrade") ||
	        strings.Contains(h.Get(eudore.HeaderContentType), "text/event-stream") {

		return false
	}

	extension := filepath.Ext(ctx.Path())
	if len(extension) < 4 { // fast path
		return true
	}

	switch extension {
	case ".png", ".gif", ".jpeg", ".jpg":
		return false
	default:
		return true
	}
}


// Push initiates an HTTP/2 server push.
// Push returns ErrNotSupported if the client has disabled push or if push
// is not supported on the underlying connection.
func (w *GzipResponse) Push(target string, opts *http.PushOptions) error {
	pusher, ok := w.ResponseWriter.(http.Pusher)
	if ok && pusher != nil {
		return pusher.Push(target, setAcceptEncodingForPushOptions(opts))
	}
	return http.ErrNotSupported
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
