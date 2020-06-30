package httptest

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"runtime"
	"strings"
)

var (
	// HTTPTestHost 定义默认使用的测试Host header。
	HTTPTestHost = "eudore-httptest"
	// ErrResponseWriterTestNotSupportHijack ResponseWriterTest对象的Hijack不支持。
	ErrResponseWriterTestNotSupportHijack = errors.New("ResponseWriterTest no support hijack")
)

type (
	// Client 定义httptest客户端。
	Client struct {
		context.Context
		http.Handler
		*http.Client
		http.CookieJar
		RemoteAddr string
		Args       url.Values
		Headers    http.Header
		Index      int
		Errs       []error
		Out        io.Writer
	}
)

// NewClient 方法创建一个httptest客户端。
func NewClient(handler http.Handler) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		Context:    context.Background(),
		Handler:    handler,
		Client:     http.DefaultClient,
		CookieJar:  jar,
		RemoteAddr: "192.0.2.1:1234",
		Args:       make(url.Values),
		Headers:    make(http.Header),
		Out:        os.Stdout,
	}
}

// NewRequest 方法创建一个新请求。
func (clt *Client) NewRequest(method, path string) *RequestReaderTest {
	return NewRequestReaderTest(clt, method, path)
}

// Next 方法检查是否存在下一个错误。
func (clt *Client) Next() bool {
	return clt.Index < len(clt.Errs)
}

// Error 方法返回当前错误。
func (clt *Client) Error() string {
	if clt.Next() {
		clt.Index++
		return clt.Errs[clt.Index-1].Error()
	}
	return ""
}

// Println 方法客户端输出字符串。
func (clt *Client) Println(args ...interface{}) (int, error) {
	return fmt.Fprintln(clt.Out, args...)
}

// Printf 方法客户端可视化输出字符串。
func (clt *Client) Printf(format string, args ...interface{}) (int, error) {
	return fmt.Fprintf(clt.Out, format, args...)
}

// GetCookie 获取客户端存储的请求路由对应的cookie值。
func (clt *Client) GetCookie(path, key string) string {
	u, err := url.Parse(path)
	if err != nil {
		//
		return ""
	}
	if u.Host == "" {
		u.Host = HTTPTestHost
	}
	if u.Scheme == "" {
		u.Scheme = "http"
	}
	for _, cookie := range clt.CookieJar.Cookies(u) {
		if cookie.Name == key {
			return cookie.Value
		}
	}
	return ""
}

// WithAddParam 方法添加客户端全局参数。
func (clt *Client) WithAddParam(key, val string) *Client {
	clt.Args.Add(key, val)
	return clt
}

// WithHeaders 方法添加客户端多个header。
func (clt *Client) WithHeaders(headers http.Header) *Client {
	for key, vals := range headers {
		for _, val := range vals {
			clt.Headers.Add(key, val)
		}
	}
	return clt
}

// WithHeaderValue 方法给客户端添加一个header值。
func (clt *Client) WithHeaderValue(key, val string) *Client {
	clt.Headers.Add(key, val)
	return clt
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
