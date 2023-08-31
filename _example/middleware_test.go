package eudore_test

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
)

func TestMiddlewareAdmin(*testing.T) {
	app := eudore.NewApp()
	app.AnyFunc("/", middleware.HandlerAdmin)

	app.NewRequest(nil, "GET", "/")

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareBasicAuth(*testing.T) {
	app := eudore.NewApp()
	app.AddMiddleware("global", middleware.NewBasicAuthFunc(map[string]string{"eudore": "hello"}))

	app.NewRequest(nil, "GET", "/", http.Header{eudore.HeaderAuthorization: {"Basic ZXVkb3JlOmhlbGxv"}})
	app.NewRequest(nil, "GET", "/", http.Header{eudore.HeaderAuthorization: {"eudore"}})

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareBodyLimit(*testing.T) {
	app := eudore.NewApp()
	app.AddMiddleware("global",
		middleware.NewCompressGzipFunc(),
		middleware.NewBodyLimitFunc(32),
	)
	app.AnyFunc("/", func(ctx eudore.Context) {
		ctx.Body()
	})
	app.AnyFunc("/form", func(ctx eudore.Context) {
		ctx.FormValues()
	})

	app.NewRequest(nil, "GET", "/")
	app.NewRequest(nil, "GET", "/", strings.NewReader("123456"))
	app.NewRequest(nil, "GET", "/", strings.NewReader("1234567890abcdefghijklmnopqrstuvwxyz"))
	// limit chunck
	data := url.Values{
		"name":  {"eudore"},
		"value": {"1234567890abcdefghijklmnopqrstuvwxyz"},
	}
	app.NewRequest(nil, "GET", "/", eudore.NewClientBodyForm(data))
	app.NewRequest(nil, "GET", "/form", eudore.NewClientBodyForm(data))

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareContextwarp(*testing.T) {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewContextWarpFunc(newContextParams))
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.AnyFunc("/ctx", func(ctx eudore.Context) {
		index, handler := ctx.GetHandler()
		ctx.Debug(index, handler)
		ctx.SetHandler(index, handler)
	})
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Debug("hello eudore")
		ctx.Info("hello eudore")
		ctx.End()
	})

	app.NewRequest(nil, "GET", "/")
	app.NewRequest(nil, "GET", "/ctx")

	app.CancelFunc()
	app.Run()
}

func newContextParams(ctx eudore.Context) eudore.Context {
	return contextParams{ctx}
}

type contextParams struct {
	eudore.Context
}

// GetParam 方法获取一个参数的值。
func (ctx contextParams) GetParam(key string) string {
	ctx.Debug("eudore.Context GetParam", key)
	return ctx.Context.GetParam(key)
}

func TestMiddlewareHeader(*testing.T) {
	app := eudore.NewApp()
	app.AddMiddleware("global", middleware.NewHeaderWithSecureFunc(http.Header{"Server": {"eudore"}}))
	app.AddMiddleware("global", middleware.NewHeaderFunc(nil))

	app.NewRequest(nil, "GET", "/")

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareHeaderFilte(*testing.T) {
	app := eudore.NewApp()
	app.AnyFunc("/1", middleware.NewHeaderFilteFunc(nil, nil))
	app.AnyFunc("/2", middleware.NewHeaderFilteFunc([]string{"127.0.0.0/24"}, nil))
	app.Listen(":8088")

	app.NewRequest(nil, "GET", "http://localhost:8088/1")
	app.NewRequest(nil, "GET", "/1")
	app.NewRequest(nil, "GET", "/2")

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareLogger(*testing.T) {
	app := eudore.NewApp()
	app.AddMiddleware("global", middleware.NewLoggerFunc(app, "route"))
	app.AddMiddleware("global", middleware.NewRequestIDFunc(func(eudore.Context) string {
		return uuid.New().String()
	}))
	app.AnyFunc("/500", func(ctx eudore.Context) {
		ctx.Fatal("test error")
	})

	app.NewRequest(nil, "GET", "/", http.Header{eudore.HeaderXForwardedFor: {"172.17.0.1"}})
	app.NewRequest(nil, "POST", "/500")

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareLoggerLevel(*testing.T) {
	app := eudore.NewApp()
	app.SetLevel(eudore.LoggerInfo)
	app.AddMiddleware(middleware.NewLoggerLevelFunc(nil))
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.AnyFunc("/api/v1/user", func(ctx eudore.Context) {
		ctx.Debug("Get User")
	})
	app.AnyFunc("/api/v1/meta", func(ctx eudore.Context) {
		ctx.Info("Get Meta", ctx.GetQuery("eudore_debug"))
	})
	app.AnyFunc("/*", eudore.HandlerEmpty)
	app.AddHandler("404", "", eudore.HandlerRouter404)
	app.AddHandler("405", "", eudore.HandlerRouter405)

	app.NewRequest(nil, "GET", "/api/v1/user")
	app.NewRequest(nil, "GET", "/api/v1/meta?eudore_debug=0")
	app.NewRequest(nil, "GET", "/api/v1/meta?eudore_debug=1")
	app.NewRequest(nil, "GET", "/api/v1/meta?eudore_debug=5")

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareRecover(*testing.T) {
	app := eudore.NewApp()
	app.AddMiddleware("global", middleware.NewRecoverFunc())
	app.AnyFunc("/", func(ctx eudore.Context) {
		panic("test error")
	})
	app.AnyFunc("/nil", func(ctx eudore.Context) {
		panic(nil)
	})

	app.NewRequest(nil, "GET", "/")
	app.NewRequest(nil, "GET", "/nil")

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareRequestID(*testing.T) {
	app := eudore.NewApp()
	app.AddMiddleware("global", middleware.NewRequestIDFunc(nil))

	app.NewRequest(nil, "GET", "/")

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareCors(*testing.T) {
	middleware.NewCorsFunc(nil, map[string]string{
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Headers":     "Authorization,DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,X-Parent-Id",
		"Access-Control-Expose-Headers":    "X-Request-Id",
		"access-control-max-age":           "1000",
	})

	app := eudore.NewApp()
	app.AddMiddleware("global", middleware.NewCorsFunc([]string{"www.*.com", "example.com", "127.0.0.1:*"}, map[string]string{
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Headers":     "Authorization,DNT,X-CustomHeader,Keep-Alive,User-Agent,X-Requested-With,If-Modified-Since,Cache-Control,Content-Type,X-Parent-Id",
		"Access-Control-Expose-Headers":    "X-Request-Id",
		"access-control-allow-methods":     "GET, POST, PUT, DELETE, HEAD",
		"access-control-max-age":           "1000",
	}))

	app.NewRequest(nil, "OPTIONS", "/1")
	app.NewRequest(nil, "OPTIONS", "/2", http.Header{eudore.HeaderOrigin: {eudore.DefaultClientInternalHost}})
	app.NewRequest(nil, "OPTIONS", "/3", http.Header{eudore.HeaderOrigin: {"http://localhost"}})
	app.NewRequest(nil, "OPTIONS", "/4", http.Header{eudore.HeaderOrigin: {"http://127.0.0.1:8088"}})
	app.NewRequest(nil, "OPTIONS", "/5", http.Header{eudore.HeaderOrigin: {"http://127.0.0.1:8089"}})
	app.NewRequest(nil, "OPTIONS", "/6", http.Header{eudore.HeaderOrigin: {"http://example.com"}})
	app.NewRequest(nil, "OPTIONS", "/6", http.Header{eudore.HeaderOrigin: {"http://www.eudore.cn"}})
	app.NewRequest(nil, "GET", "/1")
	app.NewRequest(nil, "GET", "/2", http.Header{eudore.HeaderOrigin: {eudore.DefaultClientHost}})
	app.NewRequest(nil, "GET", "/3", http.Header{eudore.HeaderOrigin: {"http://localhost"}})
	app.NewRequest(nil, "GET", "/4", http.Header{eudore.HeaderOrigin: {"http://127.0.0.1:8088"}})
	app.NewRequest(nil, "GET", "/5", http.Header{eudore.HeaderOrigin: {"http://127.0.0.1:8089"}})
	app.NewRequest(nil, "GET", "/6", http.Header{eudore.HeaderOrigin: {"http://example.com"}})
	app.NewRequest(nil, "GET", "/6", http.Header{eudore.HeaderOrigin: {"http://www.eudore.cn"}})

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareCsrf(*testing.T) {
	app := eudore.NewApp()
	app.AnyFunc("/query", middleware.NewCsrfFunc("query: csrf", "_csrf"), eudore.HandlerEmpty)
	app.AnyFunc("/header", middleware.NewCsrfFunc("header: "+eudore.HeaderXCSRFToken, eudore.CookieSet{Name: "_csrf", MaxAge: 86400}), eudore.HandlerEmpty)
	app.AnyFunc("/form", middleware.NewCsrfFunc("form: csrf", &eudore.CookieSet{Name: "_csrf", MaxAge: 86400}), eudore.HandlerEmpty)
	app.AnyFunc("/fn", middleware.NewCsrfFunc(func(ctx eudore.Context) string { return ctx.GetQuery("csrf") }, "_csrf"), eudore.HandlerEmpty)
	app.AnyFunc("/*", middleware.NewCsrfFunc(nil, nil), eudore.HandlerEmpty)

	var csrfval string
	app.NewRequest(nil, "GET", "/1",
		eudore.NewClientCheckStatus(200),
		func(w *http.Response) error {
			csrfval = w.Header.Get(eudore.HeaderSetCookie)
			app.Info("csrf token:", csrfval)
			return nil
		},
	)
	app.NewRequest(nil, "POST", "/2", eudore.NewClientCheckStatus(400))
	app.NewRequest(nil, "POST", "/1", url.Values{"csrf": {csrfval}}, eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "POST", "/query", url.Values{"csrf": {csrfval}}, eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "POST", "/header", http.Header{eudore.HeaderXCSRFToken: {csrfval}}, eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "POST", "/form", eudore.NewClientBodyForm(url.Values{"csrf": {csrfval}}), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "POST", "/form", eudore.NewClientBodyJSON(map[string]any{"csrf": csrfval}), eudore.NewClientCheckStatus(400))
	app.NewRequest(nil, "POST", "/fn", url.Values{"csrf": {csrfval}}, eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "POST", "/nil", url.Values{"csrf": {csrfval}}, eudore.NewClientCheckStatus(200))

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareDump(*testing.T) {
	type dumpMessage struct {
		Time           time.Time
		Path           string
		Host           string
		RemoteAddr     string
		Proto          string
		Method         string
		RequestURI     string
		RequestHeader  http.Header
		Status         int
		ResponseHeader http.Header
		Params         map[string]string
		Handlers       []string
	}

	var wsdialer ws.Dialer
	wsdialer.Timeout = time.Second * 1
	ReadDumpMessage := func(urlstr string, count int) {
		conn, _, _, err := wsdialer.Dial(context.Background(), urlstr)
		if err != nil {
			return
		}
		defer conn.Close()
		for i := 0; i < count; i++ {
			b, err := wsutil.ReadServerText(conn)
			if err != nil {
				break
			}
			var msg dumpMessage
			err = json.Unmarshal(b, &msg)
			if err != nil {
				break
			}
			fmt.Printf("%# v\n", msg)
		}
	}

	app := eudore.NewApp()
	app.AddMiddleware("global", middleware.NewLoggerFunc(app, "route"))
	app.AddMiddleware(func(ctx eudore.Context) {
		if ctx.GetQuery("nodump") != "" {
			ctx.SetResponse(&nodumpResponse022{ctx.Response()})
		}
	})
	app.AddMiddleware(middleware.NewDumpFunc(app.Group("/eudore/debug")))
	app.AnyFunc("/gzip", middleware.NewCompressGzipFunc(), func(ctx eudore.Context) {
		ctx.WriteString("gzip body")
	})
	app.AnyFunc("/gziperr1", func(ctx eudore.Context) {
		ctx.SetHeader(eudore.HeaderContentEncoding, "gzip")
		ctx.Write([]byte("gzip body"))
		ctx.WriteString("gzip body")
	})
	app.AnyFunc("/gziperr2", func(ctx eudore.Context) {
		ctx.SetHeader(eudore.HeaderContentEncoding, "gzip")
		ctx.WriteString("gzip body")
	})
	app.AnyFunc("/echo", func(ctx eudore.Context) {
		ctx.Write(ctx.Body())
	})
	app.AnyFunc("/bigbody", func(ctx eudore.Context) {
		ctx.Write([]byte("0123456789abcdef0123456789abcdef0123456789abcdefx"))
		ctx.Write(make([]byte, 0xffff))
	})
	app.AnyFunc("/*", eudore.HandlerEmpty)
	app.Listen(":8088")
	time.Sleep(200 * time.Millisecond)

	go ReadDumpMessage("ws://localhost:8088/eudore/debug/dump/connect", 1)
	go ReadDumpMessage("ws://localhost:8088/eudore/debug/dump/connect?nodump=1", 1)
	time.Sleep(200 * time.Millisecond)

	app.NewRequest(nil, "GET", "http://localhost:8088/eudore/debug/dump/connect")
	app.NewRequest(nil, "GET", "/gzip", http.Header{eudore.HeaderAcceptEncoding: {"gzip"}})
	app.NewRequest(nil, "GET", "/gziperr1")
	app.NewRequest(nil, "GET", "/gziperr2")
	app.NewRequest(nil, "GET", "/echo")
	app.NewRequest(nil, "GET", "/bigbody", func(resp *http.Response) {
		io.Copy(io.Discard, resp.Body)
	})

	time.Sleep(200 * time.Millisecond)
	app.CancelFunc()
	app.Run()
}

type nodumpResponse022 struct {
	eudore.ResponseWriter
}

func (nodumpResponse022) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, fmt.Errorf("nodump")
}

func TestMiddlewareCompressGzip(*testing.T) {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewCompressDeflateFunc())
	app.AddMiddleware(middleware.NewCompressGzipFunc())
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Debugf("%#v", ctx.Request().Header)
		ctx.WriteHeader(eudore.StatusOK)
		ctx.Response().Push("/stat", nil)
		ctx.Response().Push("/stat", &http.PushOptions{})
		ctx.Response().Push("/stat", &http.PushOptions{Header: make(http.Header)})
		ctx.WriteString("compress")
	})
	app.AnyFunc("/gzip", func(ctx eudore.Context) {
		ctx.SetHeader(eudore.HeaderContentType, "application/gzip;encoding=gzip")
		for i := 0; i < 20; i++ {
			ctx.WriteString("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY")
		}
		ctx.Response().Flush()
	})
	app.AnyFunc("/long", func(ctx eudore.Context) {
		data := []byte("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY")
		for i := 0; i < 20; i++ {
			ctx.Write(data)
		}
		ctx.Response().Flush()
	})
	app.AnyFunc("/longs", func(ctx eudore.Context) {
		for i := 0; i < 20; i++ {
			ctx.WriteString("1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY")
		}
		ctx.Response().Flush()
	})

	app.NewRequest(nil, "GET", "/")
	app.NewRequest(nil, "GET", "/", http.Header{eudore.HeaderAcceptEncoding: {middleware.CompressNameDeflate}})
	app.NewRequest(nil, "GET", "/", http.Header{eudore.HeaderAcceptEncoding: {middleware.CompressNameGzip}})
	app.NewRequest(nil, "GET", "/gzip", http.Header{eudore.HeaderAcceptEncoding: {middleware.CompressNameGzip}})
	app.NewRequest(nil, "GET", "/long", http.Header{eudore.HeaderAcceptEncoding: {middleware.CompressNameGzip}})
	app.NewRequest(nil, "GET", "/longs", http.Header{eudore.HeaderAcceptEncoding: {middleware.CompressNameGzip}})

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareCompressMixins(*testing.T) {
	middleware.DefaultComoressBrotliFunc = func() interface{} {
		return gzip.NewWriter(io.Discard)
	}
	defer func() {
		middleware.DefaultComoressBrotliFunc = nil
	}()

	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewCompressMixinsFunc(nil))
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Debugf("%#v", ctx.Request().Header)
		ctx.WriteString("mixins")
		ctx.Response().Flush()
	})

	app.NewRequest(nil, "GET", "/")
	app.NewRequest(nil, "GET", "/", http.Header{eudore.HeaderAcceptEncoding: {middleware.CompressNameGzip}})
	app.NewRequest(nil, "GET", "/", http.Header{eudore.HeaderAcceptEncoding: {middleware.CompressNameDeflate}})
	app.NewRequest(nil, "GET", "/", http.Header{eudore.HeaderAcceptEncoding: {middleware.CompressNameIdentity}})
	app.NewRequest(nil, "GET", "/", http.Header{eudore.HeaderAcceptEncoding: {"gzip;q=0"}})
	app.NewRequest(nil, "GET", "/", http.Header{eudore.HeaderAcceptEncoding: {"none"}})

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareLook(*testing.T) {
	var i interface{}
	config := map[interface{}]interface{}{
		true:          1,
		1:             2,
		uint(1):       3,
		1.0:           4.0,
		complex(1, 1): complex(5, 5),
		i:             6,
		struct{}{}:    7,
		"bytes":       []byte(`    app.NewRequest(nil, "GET", "/1").AddHeader(eudore.HeaderAcceptEncoding, "none")`),
	}

	app := eudore.NewApp()
	{
		app2 := eudore.NewApp()
		app2.SetValue(eudore.ContextKeyLogger, eudore.NewLoggerInit())
		app2.Set("conf", config)
		app.AnyFunc("/eudore/debug/look/*", middleware.NewLookFunc(app2))
		app.AnyFunc("/eudore/debug/data", middleware.NewLookFunc(func(eudore.Context) interface{} {
			return nil
		}))
	}

	app.NewRequest(nil, "GET", "/eudore/debug/data")
	app.NewRequest(nil, "GET", "/eudore/debug/look/?d=3")
	app.NewRequest(nil, "GET", "/eudore/debug/look/?all=1")
	app.NewRequest(nil, "GET", "/eudore/debug/look/?format=text")
	app.NewRequest(nil, "GET", "/eudore/debug/look/?format=json")
	app.NewRequest(nil, "GET", "/eudore/debug/look/?format=t2")
	app.NewRequest(nil, "GET", "/eudore/debug/look/Config/Keys/2")
	app.NewRequest(nil, "GET", "/eudore/debug/look/?d=3", http.Header{eudore.HeaderAccept: {eudore.MimeApplicationJSON}})
	app.NewRequest(nil, "GET", "/eudore/debug/look/?d=3", http.Header{eudore.HeaderAccept: {eudore.MimeTextHTML}})
	app.NewRequest(nil, "GET", "/eudore/debug/look/?d=3", http.Header{eudore.HeaderAccept: {eudore.MimeText}})

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareLookRender(*testing.T) {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyRender, middleware.NewBindLook(map[string]eudore.HandlerDataFunc{
		eudore.MimeApplicationXML: nil,
		eudore.MimeTextXML:        eudore.RenderXML,
	}))
	app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))
	app.AnyFunc("/", func(ctx eudore.Context) interface{} {
		return map[string]interface{}{
			"name": "eudore",
			"date": time.Now(),
		}
	})
	app.NewRequest(nil, "GET", "/", http.Header{eudore.HeaderAccept: {middleware.MimeValueJSON}})
	app.NewRequest(nil, "GET", "/", http.Header{eudore.HeaderAccept: {middleware.MimeValueJSON + "," + eudore.MimeApplicationJSON}})
	app.NewRequest(nil, "GET", "/", http.Header{eudore.HeaderAccept: {middleware.MimeValueHTML}})
	app.NewRequest(nil, "GET", "/", http.Header{eudore.HeaderAccept: {middleware.MimeValueText}})

	// time.Sleep(100* time.Microsecond)
	app.CancelFunc()
	app.Run()
}

func TestMiddlewarePprof(*testing.T) {
	app := eudore.NewApp()
	app.AddMiddleware(
		middleware.NewHeaderFilteFunc(nil, nil),
		middleware.NewRecoverFunc(),
		middleware.NewLoggerFunc(app),
		middleware.NewCompressMixinsFunc(nil),
	)
	app.AnyFunc("/eudore/debug/pprof/*", middleware.HandlerPprof)
	app.AnyFunc("/wait", func(ctx eudore.Context) {
		time.Sleep(time.Second)
	})

	for i := 0; i < 4; i++ {
		go app.NewRequest(nil, "GET", "/wait")
	}

	app.NewRequest(nil, "GET", "/eudore/debug/pprof/expvar", http.Header{eudore.HeaderAccept: {eudore.MimeApplicationJSON}})
	app.NewRequest(nil, "GET", "/eudore/debug/pprof/?format=json")
	app.NewRequest(nil, "GET", "/eudore/debug/pprof/?format=text")
	app.NewRequest(nil, "GET", "/eudore/debug/pprof/?format=html")
	app.NewRequest(nil, "GET", "/eudore/debug/pprof/allocs")
	app.NewRequest(nil, "GET", "/eudore/debug/pprof/block")
	app.NewRequest(nil, "GET", "/eudore/debug/pprof/heap")
	app.NewRequest(nil, "GET", "/eudore/debug/pprof/mutex")
	app.NewRequest(nil, "GET", "/eudore/debug/pprof/goroutine?debug=0")
	app.NewRequest(nil, "GET", "/eudore/debug/pprof/goroutine?debug=1")
	app.NewRequest(nil, "GET", "/eudore/debug/pprof/goroutine?debug=1&format=json")
	app.NewRequest(nil, "GET", "/eudore/debug/pprof/goroutine?debug=1&format=text")
	app.NewRequest(nil, "GET", "/eudore/debug/pprof/goroutine?debug=1&format=html")
	app.NewRequest(nil, "GET", "/eudore/debug/pprof/goroutine?debug=2&format=json")
	app.NewRequest(nil, "GET", "/eudore/debug/pprof/goroutine?debug=2&format=text")
	app.NewRequest(nil, "GET", "/eudore/debug/pprof/goroutine?debug=2&format=html")

	app.CancelFunc()
	app.Run()
}
func TestMiddlewareReferer(*testing.T) {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewRefererFunc(map[string]bool{
		"":                         true,
		"origin":                   true,
		"www.eudore.cn/*":          true,
		"www.eudore.cn/api/*":      false,
		"www.example.com/*":        true,
		"www.example.com/*/*":      false,
		"www.example.com/*/2":      true,
		"http://127.0.0.1/*":       true,
		"http://126.0.0.1:*/*":     true,
		"http://127.0.0.1:*/*":     true,
		"http://128.0.0.1:*/*":     true,
		"http://localhost/api/*":   true,
		"http://localhost:*/api/*": true,
	}))
	app.AnyFunc("/*", eudore.HandlerEmpty)

	app.NewRequest(nil, "GET", "/", http.Header{eudore.HeaderReferer: {""}})
	app.NewRequest(nil, "GET", "/",
		eudore.NewClientOptionHost("www.eudore.cn"),
		eudore.NewClientOptionHeader(eudore.HeaderReferer, "http://www.eudore.cn/"),
	)
	app.NewRequest(nil, "GET", "/", http.Header{eudore.HeaderReferer: {"http://www.eudore.cn/"}})
	app.NewRequest(nil, "GET", "/", http.Header{eudore.HeaderReferer: {"http://www.example.com"}})
	app.NewRequest(nil, "GET", "/", http.Header{eudore.HeaderReferer: {"http://www.example.com/"}})
	app.NewRequest(nil, "GET", "/", http.Header{eudore.HeaderReferer: {"http://www.example.com/1"}})
	app.NewRequest(nil, "GET", "/", http.Header{eudore.HeaderReferer: {"http://www.example.com/1/1"}})
	app.NewRequest(nil, "GET", "/", http.Header{eudore.HeaderReferer: {"http://www.example.com/1/2"}})
	app.NewRequest(nil, "GET", "/", http.Header{eudore.HeaderReferer: {"http://127.0.0.1:80/1"}})
	app.NewRequest(nil, "GET", "/", http.Header{eudore.HeaderReferer: {"http://127.0.0.10:80"}})

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareRewrite(*testing.T) {
	rewritedata := map[string]string{
		"/js/*":                    "/public/js/$0",
		"/api/v1/users/*/orders/*": "/api/v3/user/$0/order/$1",
		"/d/*":                     "/d/$0-$0",
		"/api/v1/*":                "/api/v3/$0",
		"/api/v2/*":                "/api/v3/$0",
		"/api/v1/group*":           "/api/v3/$0",
		"/api/v1/group":            "/api/v3/$0",
		"/api/v1/name*":            "/api/v3/$0",
		"/api/v1/name":             "/api/v3/$0",
		"/help/history*":           "/api/v3/history",
		"/help/history":            "/api/v3/history",
		"/help/*":                  "$0",
	}

	app := eudore.NewApp()
	app.AddMiddleware("global", middleware.NewRewriteFunc(rewritedata))
	app.AddMiddleware(middleware.NewLoggerFunc(app))
	app.AnyFunc("/*", eudore.HandlerEmpty)

	app.NewRequest(nil, "GET", "/")
	app.NewRequest(nil, "GET", "/js/")
	app.NewRequest(nil, "GET", "/js/index.js")
	app.NewRequest(nil, "GET", "/api/v1/user")
	app.NewRequest(nil, "GET", "/api/v1/user/new")
	app.NewRequest(nil, "GET", "/api/v1/users/v3/orders/8920")
	app.NewRequest(nil, "GET", "/api/v1/users/orders")
	app.NewRequest(nil, "GET", "/api/v2")
	app.NewRequest(nil, "GET", "/api/v2/user")
	app.NewRequest(nil, "GET", "/d/3")
	app.NewRequest(nil, "GET", "/help/history")
	app.NewRequest(nil, "GET", "/help/historyv2")

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareRrouter(*testing.T) {
	routerdata := map[string]interface{}{
		"/api/:v/*": func(ctx eudore.Context) {
			ctx.Request().URL.Path = "/api/v3/" + ctx.GetParam("*")
		},
		"GET /api/:v/*": func(ctx eudore.Context) {
			ctx.WriteHeader(403)
			ctx.End()
		},
	}

	app := eudore.NewApp()
	app.AddMiddleware("global", middleware.NewLoggerFunc(app, "route", "*"))
	app.AddMiddleware(middleware.NewRouterFunc(routerdata))
	app.AnyFunc("/*", eudore.HandlerEmpty)

	app.NewRequest(nil, "GET", "/api/v1/user")
	app.NewRequest(nil, "PUT", "/api/v1/user")
	app.NewRequest(nil, "PUT", "/api/v2/user")

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareRrouterRewrite(*testing.T) {
	rewritedata := map[string]string{
		"/js/*":                    "/public/js/$0",
		"/api/v1/users/*/orders/*": "/api/v3/user/$0/order/$1",
		"/d/*":                     "/d/$0-$0",
		"/api/v1/*":                "/api/v3/$0",
		"/api/v2/*":                "/api/v3/$0",
		"/help/history*":           "/api/v3/history",
		"/help/history":            "/api/v3/history",
		"/help/*":                  "$0",
	}

	app := eudore.NewApp()
	app.AddMiddleware("global", middleware.NewRouterRewriteFunc(rewritedata))
	app.AddMiddleware(middleware.NewLoggerFunc(app))
	app.AnyFunc("/*", eudore.HandlerEmpty)

	app.NewRequest(nil, "GET", "/")
	app.NewRequest(nil, "GET", "/js/")
	app.NewRequest(nil, "GET", "/js/index.js")
	app.NewRequest(nil, "GET", "/api/v1/user")
	app.NewRequest(nil, "GET", "/api/v1/user/new")
	app.NewRequest(nil, "GET", "/api/v1/users/v3/orders/8920")
	app.NewRequest(nil, "GET", "/api/v1/users/orders")
	app.NewRequest(nil, "GET", "/api/v2")
	app.NewRequest(nil, "GET", "/api/v2/user")
	app.NewRequest(nil, "GET", "/d/3")
	app.NewRequest(nil, "GET", "/help/history")
	app.NewRequest(nil, "GET", "/help/historyv2")

	app.CancelFunc()
	app.Run()
}
