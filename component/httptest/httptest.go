package httptest

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
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
		Host       string
		RemoteAddr string
		Querys     url.Values
		Headers    http.Header
		Print      func(...interface{})
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
		Host:       HTTPTestHost,
		RemoteAddr: "192.0.2.1:1234",
		Querys:     make(url.Values),
		Headers:    make(http.Header),
		Print: func(args ...interface{}) {
			fmt.Println(args...)
		},
	}
}

// NewRequest 方法创建一个新请求。
func (clt *Client) NewRequest(method, path string) *RequestReaderTest {
	return NewRequestReaderTest(clt, method, path)
}

// Printf 方法格式化输出信息。
func (clt *Client) Printf(format string, args ...interface{}) {
	clt.Print(fmt.Sprintf(format, args...))
}

// AddQuerys 方法给客户端添加全局请求参数。
func (clt *Client) AddQuerys(querys url.Values) *Client {
	for key, vals := range querys {
		for _, val := range vals {
			clt.Querys.Add(key, val)
		}
	}
	return clt
}

// AddQuery 方法添加客户端全局参数。
func (clt *Client) AddQuery(key, val string) *Client {
	clt.Querys.Add(key, val)
	return clt
}

// AddHeaders 方法添加客户端多个header。
func (clt *Client) AddHeaders(headers http.Header) *Client {
	for key, vals := range headers {
		for _, val := range vals {
			clt.Headers.Add(key, val)
		}
	}
	return clt
}

// AddHeaderValue 方法给客户端添加一个header值。
func (clt *Client) AddHeaderValue(key, val string) *Client {
	clt.Headers.Add(key, val)
	return clt
}

// AddBasicAuth 方法给客户端设置basicauth信息。
func (clt *Client) AddBasicAuth(name, pass string) *Client {
	clt.Headers.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(name+":"+pass)))
	return clt
}

// AddCookie 方法指定url添加cookie
func (clt *Client) AddCookie(path, key, val string) *Client {
	u, err := url.Parse(path)
	if err != nil {
		clt.Printf("GetCookie parse url %s error: %s", path, err.Error())
		return clt
	}
	clt.CookieJar.SetCookies(u, []*http.Cookie{{Name: key, Value: val}})
	return clt
}

// GetCookie 获取客户端存储的请求路由对应的cookie值。
func (clt *Client) GetCookie(path, key string) string {
	u, err := url.Parse(path)
	if err != nil {
		clt.Printf("GetCookie parse url %s error: %s", path, err.Error())
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

// logFormatFileLine 函数获得调用的文件位置，默认层数加三。
//
// 文件位置会从第一个src后开始截取，处理gopath下文件位置。
func logFormatFileLine(depth int) (string, int) {
	_, file, line, _ := runtime.Caller(depth)
	// slash := strings.LastIndex(file, "/")
	slash := strings.Index(file, "src")
	if slash >= 0 {
		file = file[slash+4:]
	}
	return file, line
}
