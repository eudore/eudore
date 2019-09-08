package httptest

import (
	"errors"
	"fmt"
	"io"
	"net/textproto"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/eudore/eudore/protocol"
)

var (
	// ErrResponseWriterTestNotSupportHijack ResponseWriterTest对象的Hijack不支持。
	ErrResponseWriterTestNotSupportHijack = errors.New("ResponseWriterTest no support hijack")
)

type (
	HeaderMap map[string][]string
	Client    struct {
		protocol.HandlerHTTP
		Args    url.Values
		Headers HeaderMap
		Index   int
		Errs    []error
		Out     io.Writer
	}
)

func NewClient(handler protocol.HandlerHTTP) *Client {
	return &Client{
		HandlerHTTP: handler,
		Args:        make(url.Values),
		Headers:     make(HeaderMap),
		Out:         os.Stdout,
	}
}

func (clt *Client) NewRequest(method, path string) *RequestReaderTest {
	r := NewRequestReaderTest(clt, method, path)
	r.File, r.Line = logFormatFileLine(2)
	return r
}

func (clt *Client) Next() bool {
	return clt.Index < len(clt.Errs)
}

func (clt *Client) Error() string {
	if clt.Next() {
		clt.Index++
		return clt.Errs[clt.Index-1].Error()
	}
	return ""
}

func (clt *Client) Println(args ...interface{}) (int, error) {
	return fmt.Fprintln(clt.Out, args...)
}

func (clt *Client) Printf(format string, args ...interface{}) (int, error) {
	return fmt.Fprintf(clt.Out, format, args...)
}

func (clt *Client) WithAddParam(key, val string) *Client {
	clt.Args.Add(key, val)
	return clt
}
func (clt *Client) WithHeader(headers protocol.Header) *Client {
	headers.Range(func(key, val string) {
		clt.Headers.Add(key, val)
	})
	return clt
}

func (clt *Client) WithHeaderValue(key, val string) *Client {
	clt.Headers.Add(key, val)
	return clt
}

// Get 方法获得一个Header值。
func (h HeaderMap) Get(key string) string {
	return textproto.MIMEHeader(h).Get(key)
}

// Set 方法设置一个Header值。
func (h HeaderMap) Set(key, value string) {
	textproto.MIMEHeader(h).Set(key, value)
}

// Add 方法添加一个Header值。
func (h HeaderMap) Add(key, value string) {
	textproto.MIMEHeader(h).Add(key, value)
}

// Del 方法删除一个Header值。
func (h HeaderMap) Del(key string) {
	textproto.MIMEHeader(h).Del(key)
}

// Range 方法遍历Header全部键值。
func (h HeaderMap) Range(fn func(string, string)) {
	for k, v := range h {
		for _, vv := range v {
			fn(k, vv)
		}
	}
}

// logFormatFileLine 函数获得调用的文件位置，默认层数加三。
//
// 文件位置会从第一个src后开始截取，处理gopath下文件位置。
func logFormatFileLine(depth int) (string, int) {
	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		file = "???"
		line = 1
	} else {
		// slash := strings.LastIndex(file, "/")
		slash := strings.Index(file, "src")
		if slash >= 0 {
			file = file[slash+4:]
		}
	}
	return file, line
}
