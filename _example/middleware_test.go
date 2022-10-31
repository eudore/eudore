package eudore_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
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

	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderAuthorization, "Basic ZXVkb3JlOmhlbGxv"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderAuthorization, "eudore"))

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareBodyLimit(*testing.T) {
	app := eudore.NewApp()
	app.AddMiddleware("global", middleware.NewBodyLimitFunc(32))
	app.AnyFunc("/", func(ctx eudore.Context) {
		ctx.Body()
	})

	app.NewRequest(nil, "GET", "/", eudore.NewClientBodyString("123456"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientBodyString("1234567890abcdefghijklmnopqrstuvwxyz"))
	// limit chunck
	app.NewRequest(nil, "GET", "/", eudore.NewClientBodyFormValues(map[string]string{
		"name": "eudore", "value": "1234567890abcdefghijklmnopqrstuvwxyz",
	}))

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
	app.AddMiddleware("global", middleware.NewHeaderWithSecureFunc(http.Header{"Server": []string{"eudore"}}))
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

	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderXForwardedFor, "172.17.0.1"))
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

func TestMiddlewareBlack(*testing.T) {
	middleware.NewBlackFunc(map[string]bool{
		"192.168.0.0/16": true,
		"0.0.0.0/0":      false,
	}, nil)

	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewBlackFunc(map[string]bool{
		"192.168.100.0/24": true,
		"192.168.75.0/30":  true,
		"192.168.1.100/30": true,
		"127.0.0.1/32":     true,
		"10.168.0.0/16":    true,
		"0.0.0.0/0":        false,
	}, app.Group("/eudore/debug")))
	app.AnyFunc("/*", eudore.HandlerEmpty)

	app.NewRequest(nil, "GET", "/eudore/debug/black/ui")
	app.NewRequest(nil, "GET", "/eudore/debug/black/ui")
	app.NewRequest(nil, "PUT", "/eudore/debug/black/black/10.127.87.0?mask=24")
	app.NewRequest(nil, "PUT", "/eudore/debug/black/white/10.127.87.0?mask=24")
	app.NewRequest(nil, "GET", "/eudore/debug/black/data")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/black/10.127.87.0?mask=24")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/white/10.127.87.0?mask=24")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/black/10.127.87.0?mask=24")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/white/10.127.87.0?mask=24")

	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "127.0.0.1:29398"), eudore.NewClientCheckStatus(403))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "127.0.0.1:29398"), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "192.168.75.1:8298"), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "192.168.100.3/28"), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "192.168.100.0"), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "192.168.100.1"), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "192.168.100.77"), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "192.168.100.148"), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "192.168.100.222"), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "192.168.75.4"), eudore.NewClientCheckStatus(403))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "192.168.75.5"), eudore.NewClientCheckStatus(403))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "192.168.75.6"), eudore.NewClientCheckStatus(403))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "192.168.1.99"), eudore.NewClientCheckStatus(403))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "192.168.1.100"), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "192.168.1.101"), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "192.168.1.102"), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "192.168.1.103"), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "192.168.1.104"), eudore.NewClientCheckStatus(403))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "192.168.1.105"), eudore.NewClientCheckStatus(403))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "127.0.0.1"))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientCheckStatus(403))

	app.NewRequest(nil, "DELETE", "/eudore/debug/black/white/0.0.0.0?mask=0")
	app.NewRequest(nil, "PUT", "/eudore/debug/black/white/192.168.75.4?mask=30")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/white/192.168.75.1")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/white/192.168.75.5")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/white/192.168.75.7")
	app.NewRequest(nil, "PUT", "/eudore/debug/black/white/10.16.0.0?mask=16")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/white/192.168.75.4?mask=30")

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareBreaker(*testing.T) {
	middleware.NewBreakerFunc(nil)

	app := eudore.NewApp()
	// 创建熔断器并注入管理路由
	breaker := middleware.NewBreaker()
	breaker.MaxConsecutiveSuccesses = 3
	breaker.MaxConsecutiveFailures = 3
	breaker.OpenWait = 0
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.AddMiddleware(breaker.NewBreakerFunc(app.Group("/eudore/debug")))
	app.AnyFunc("/*", func(ctx eudore.Context) {
		if len(ctx.Querys()) > 0 {
			ctx.Fatal("test err")
			return
		}
		ctx.WriteString("route: " + ctx.GetParam("route"))
	})

	// 错误请求
	for i := 0; i < 10; i++ {
		app.NewRequest(nil, "GET", "/1?a=1")
	}
	for i := 0; i < 5; i++ {
		time.Sleep(time.Millisecond * 500)
		app.NewRequest(nil, "GET", "/1?a=1")
	}
	// 除非熔断后访问
	for i := 0; i < 5; i++ {
		time.Sleep(time.Millisecond * 500)
		app.NewRequest(nil, "GET", "/1")
	}

	app.NewRequest(nil, "GET", "/eudore/debug/breaker/ui")
	app.NewRequest(nil, "GET", "/eudore/debug/breaker/ui")
	app.NewRequest(nil, "GET", "/eudore/debug/breaker/data", eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeApplicationJSON))
	app.NewRequest(nil, "GET", "/eudore/debug/breaker/1")
	app.NewRequest(nil, "GET", "/eudore/debug/breaker/100")
	app.NewRequest(nil, "PUT", "/eudore/debug/breaker/1/state/0")
	app.NewRequest(nil, "PUT", "/eudore/debug/breaker/1/state/3")
	app.NewRequest(nil, "PUT", "/eudore/debug/breaker/3/state/3")

	time.Sleep(time.Microsecond * 100)
	app.CancelFunc()
	app.Run()
}

func TestMiddlewareCache(*testing.T) {
	app := eudore.NewApp()
	app.AddMiddleware("global", middleware.NewLoggerFunc(app, "route"))
	app.AddMiddleware(middleware.NewCacheFunc(time.Second/10, app.Context, func(ctx eudore.Context) string {
		// 自定义缓存key函数，默认实现方法
		if ctx.Method() != eudore.MethodGet || ctx.GetHeader(eudore.HeaderUpgrade) != "" {
			return ""
		}
		return ctx.Request().URL.RequestURI()
	}))
	app.AnyFunc("/sf", func(ctx eudore.Context) {
		ctx.Redirect(301, "/")
		ctx.Debug(ctx.Response().Status(), ctx.Response().Size())
	})
	app.AnyFunc("/*", func(ctx eudore.Context) {
		time.Sleep(time.Second / 3)
		ctx.WriteString("hello eudore")
	})

	app.NewRequest(nil, "GET", "/sf")
	wg := sync.WaitGroup{}
	wg.Add(5)
	for n := 0; n < 5; n++ {
		go func() {
			for i := 0; i < 3; i++ {
				app.NewRequest(nil, "GET", "/?c="+fmt.Sprint(i))
				app.NewRequest(nil, "GET", "/?c="+fmt.Sprint(i))
				time.Sleep(time.Millisecond * 200)
				app.NewRequest(nil, "GET", "/?c="+fmt.Sprint(i))
			}
			wg.Done()
		}()
	}
	wg.Wait()

	app.NewRequest(nil, "GET", "/sf")
	app.NewRequest(nil, "POST", "/sf")
	app.NewRequest(nil, "GET", "/s")

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareCacheStore(*testing.T) {
	app := eudore.NewApp()
	app.AddMiddleware("global", middleware.NewLoggerFunc(app, "route"))
	app.AddMiddleware(middleware.NewCacheFunc(time.Second/100, app.Context, new(cacheMap)))
	app.AnyFunc("/sf", func(ctx eudore.Context) {
		ctx.Redirect(301, "/")
		ctx.Debug(ctx.Response().Status(), ctx.Response().Size())
	})
	app.AnyFunc("/*", func(ctx eudore.Context) {
		time.Sleep(time.Second / 3)
		ctx.WriteString("hello eudore")
	})

	app.NewRequest(nil, "GET", "/sf")
	wg := sync.WaitGroup{}
	wg.Add(5)
	for n := 0; n < 5; n++ {
		go func() {
			for i := 0; i < 3; i++ {
				app.NewRequest(nil, "GET", "/?c="+fmt.Sprint(i))
				app.NewRequest(nil, "GET", "/?c="+fmt.Sprint(i))
				time.Sleep(time.Millisecond * 20)
				app.NewRequest(nil, "GET", "/?c="+fmt.Sprint(i))
			}
			wg.Done()
		}()
	}
	wg.Wait()

	app.NewRequest(nil, "GET", "/sf")
	app.NewRequest(nil, "POST", "/sf")
	app.NewRequest(nil, "GET", "/s")

	app.CancelFunc()
	app.Run()
}

type cacheMap struct {
	sync.Map
}

func (m *cacheMap) Load(key string) *middleware.CacheData {
	data, ok := m.Map.Load(key)
	if !ok {
		return nil
	}
	item := data.(*middleware.CacheData)
	if time.Now().After(item.Expired) {
		m.Map.Delete(key)
		return nil
	}
	fmt.Println("cache", key)
	return item
}

func (m *cacheMap) Store(key string, val *middleware.CacheData) {
	fmt.Println("new", key)
	m.Map.Store(key, val)
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
	app.NewRequest(nil, "OPTIONS", "/2", eudore.NewClientHeader("Origin", eudore.DefaultClientInternalHost))
	app.NewRequest(nil, "OPTIONS", "/3", eudore.NewClientHeader("Origin", "http://localhost"))
	app.NewRequest(nil, "OPTIONS", "/4", eudore.NewClientHeader("Origin", "http://127.0.0.1:8088"))
	app.NewRequest(nil, "OPTIONS", "/5", eudore.NewClientHeader("Origin", "http://127.0.0.1:8089"))
	app.NewRequest(nil, "OPTIONS", "/6", eudore.NewClientHeader("Origin", "http://example.com"))
	app.NewRequest(nil, "OPTIONS", "/6", eudore.NewClientHeader("Origin", "http://www.eudore.cn"))
	app.NewRequest(nil, "GET", "/1")
	app.NewRequest(nil, "GET", "/2", eudore.NewClientHeader("Origin", eudore.DefaultClientHost))
	app.NewRequest(nil, "GET", "/3", eudore.NewClientHeader("Origin", "http://localhost"))
	app.NewRequest(nil, "GET", "/4", eudore.NewClientHeader("Origin", "http://127.0.0.1:8088"))
	app.NewRequest(nil, "GET", "/5", eudore.NewClientHeader("Origin", "http://127.0.0.1:8089"))
	app.NewRequest(nil, "GET", "/6", eudore.NewClientHeader("Origin", "http://example.com"))
	app.NewRequest(nil, "GET", "/6", eudore.NewClientHeader("Origin", "http://www.eudore.cn"))

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareCsrf(*testing.T) {
	app := eudore.NewApp()
	app.AnyFunc("/query", middleware.NewCsrfFunc("query: csrf", "_csrf"), eudore.HandlerEmpty)
	app.AnyFunc("/header", middleware.NewCsrfFunc("header: "+eudore.HeaderXCSRFToken, eudore.SetCookie{Name: "_csrf", MaxAge: 86400}), eudore.HandlerEmpty)
	app.AnyFunc("/form", middleware.NewCsrfFunc("form: csrf", &eudore.SetCookie{Name: "_csrf", MaxAge: 86400}), eudore.HandlerEmpty)
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
	app.NewRequest(nil, "POST", "/1", eudore.NewClientQuery("csrf", csrfval), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "POST", "/query", eudore.NewClientQuery("csrf", csrfval), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "POST", "/header", eudore.NewClientHeader(eudore.HeaderXCSRFToken, csrfval), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "POST", "/form", eudore.NewClientBodyFormValue("csrf", csrfval), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "POST", "/form", eudore.NewClientBodyJSONValue("csrf", csrfval), eudore.NewClientCheckStatus(400))
	app.NewRequest(nil, "POST", "/fn", eudore.NewClientQuery("csrf", csrfval), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "POST", "/nil", eudore.NewClientQuery("csrf", csrfval), eudore.NewClientCheckStatus(200))

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
	app.AnyFunc("/gzip", middleware.NewCompressGzipFunc(5), func(ctx eudore.Context) {
		ctx.WriteString("gzip body")
	})
	app.AnyFunc("/gziperr", func(ctx eudore.Context) {
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
	app.NewRequest(nil, "GET", "/gzip", eudore.NewClientHeader(eudore.HeaderAcceptEncoding, "gzip"))
	app.NewRequest(nil, "GET", "/gziperr")
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

func TestMiddlewareGzip(*testing.T) {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewCompressDeflateFunc(100))
	app.AddMiddleware(middleware.NewCompressGzipFunc(10))
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Debugf("%#v", ctx.Request().Header)
		ctx.Push("/stat", nil)
		ctx.Response().Push("/stat", nil)
		ctx.Response().Push("/stat", &http.PushOptions{})
		ctx.Response().Push("/stat", &http.PushOptions{Header: make(http.Header)})
		ctx.WriteString("gzip")
		ctx.Response().Flush()
	})

	app.NewRequest(nil, "GET", "/1")
	app.NewRequest(nil, "GET", "/1", eudore.NewClientHeader(eudore.HeaderAcceptEncoding, "deflate"))
	app.NewRequest(nil, "GET", "/1", eudore.NewClientHeader(eudore.HeaderAcceptEncoding, "none"))

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
	app.NewRequest(nil, "GET", "/eudore/debug/look/?d=3", eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeApplicationJSON))
	app.NewRequest(nil, "GET", "/eudore/debug/look/?d=3", eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeTextHTML))
	app.NewRequest(nil, "GET", "/eudore/debug/look/?d=3", eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeText))

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
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderAccept, middleware.MimeValueJSON))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderAccept, middleware.MimeValueJSON+","+eudore.MimeApplicationJSON))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderAccept, middleware.MimeValueHTML))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderAccept, middleware.MimeValueText))

	// time.Sleep(100* time.Microsecond)
	app.CancelFunc()
	app.Run()
}

func TestMiddlewarePprof(*testing.T) {
	app := eudore.NewApp()
	app.Group("/eudore/debug").AddController(middleware.NewPprofController())

	app.NewRequest(nil, "GET", "/eudore/debug/pprof/expvar", eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeApplicationJSON))
	app.NewRequest(nil, "GET", "/eudore/debug/pprof/?format=json")
	app.NewRequest(nil, "GET", "/eudore/debug/pprof/?format=text")
	app.NewRequest(nil, "GET", "/eudore/debug/pprof/?format=html")
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

func TestMiddlewareRateRequest(*testing.T) {
	app := eudore.NewApp()
	app.AnyFunc("/*", middleware.NewRateRequestFunc(1, 3, app.Context), eudore.HandlerEmpty)

	for i := 0; i < 8; i++ {
		app.NewRequest(nil, "GET", "/")
	}

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareRateSpeed1(*testing.T) {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewRateSpeedFunc(16*1024, 64*1024, app.Context))
	app.PostFunc("/post", func(ctx eudore.Context) {
		ctx.Debug(string(ctx.Body()))
	})
	app.AnyFunc("/srv", func(ctx eudore.Context) {
		ctx.WriteString("rate speed 16kB")
	})
	app.AnyFunc("/*", eudore.HandlerEmpty)

	app.NewRequest(nil, "POST", "/post", eudore.NewClientBodyString("return body"))
	app.NewRequest(nil, "PUT", "/srv")

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareRateSpeed2(*testing.T) {
	app := eudore.NewApp()
	app.AnyFunc("/*", middleware.NewRateRequestFunc(1, 3, app.Context, time.Millisecond*100), eudore.HandlerEmpty)

	app.NewRequest(nil, "PUT", "/")
	app.NewRequest(nil, "PUT", "/")
	app.NewRequest(nil, "PUT", "/")
	time.Sleep(time.Second / 2)
	app.NewRequest(nil, "PUT", "/")
	app.NewRequest(nil, "PUT", "/")
	app.NewRequest(nil, "PUT", "/")

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareRateSpeed3(*testing.T) {
	app := eudore.NewApp()
	app.AnyFunc("/*", middleware.NewRateRequestFunc(1, 2, app.Context, time.Microsecond*49), eudore.HandlerEmpty)

	app.NewRequest(nil, "PUT", "/")
	app.NewRequest(nil, "PUT", "/")
	app.NewRequest(nil, "PUT", "/")
	app.NewRequest(nil, "PUT", "/")
	app.NewRequest(nil, "PUT", "/")
	time.Sleep(time.Second)

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareRateSpeedCannel1(*testing.T) {
	app := eudore.NewApp()
	app.AddMiddleware("/out", func(ctx eudore.Context) {
		c1 := ctx.GetContext()
		c2, cannel := context.WithTimeout(context.Background(), time.Millisecond*20)
		go func() {
			cannel()
		}()
		ctx.SetContext(c2)
		ctx.Next()
		ctx.SetContext(c1)
	})
	app.AddMiddleware(middleware.NewRateRequestFunc(1, 3, app.Context, time.Millisecond*10, func(ctx eudore.Context) string {
		return ctx.RealIP()
	}))
	app.AnyFunc("/out", eudore.HandlerEmpty)
	app.AnyFunc("/*", eudore.HandlerEmpty)

	app.NewRequest(nil, "PUT", "/")
	app.NewRequest(nil, "PUT", "/")
	time.Sleep(50 * time.Millisecond)
	app.NewRequest(nil, "PUT", "/out")
	app.NewRequest(nil, "PUT", "/")
	app.NewRequest(nil, "PUT", "/")
	app.NewRequest(nil, "PUT", "/")
	app.NewRequest(nil, "PUT", "/")
	app.NewRequest(nil, "PUT", "/")
	app.NewRequest(nil, "PUT", "/out")
	app.NewRequest(nil, "PUT", "/")

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareRateSpeedCannel2(*testing.T) {
	app := eudore.NewApp()
	app.AddMiddleware("/out", func(ctx eudore.Context) {
		c, cannel := context.WithTimeout(ctx.GetContext(), time.Millisecond*2)
		cannel()
		ctx.SetContext(c)
	})
	app.AddMiddleware(middleware.NewRateRequestFunc(1, 3, app.Context, time.Millisecond*10, func(ctx eudore.Context) string {
		return ctx.RealIP()
	}))
	app.AnyFunc("/out", func(ctx eudore.Context) {
		time.Sleep(time.Millisecond * 5)
	})
	app.AnyFunc("/*", eudore.HandlerEmpty)

	app.NewRequest(nil, "PUT", "/")
	app.NewRequest(nil, "PUT", "/")
	app.NewRequest(nil, "PUT", "/")
	app.NewRequest(nil, "PUT", "/")
	app.NewRequest(nil, "PUT", "/out")
	app.NewRequest(nil, "PUT", "/out")
	app.NewRequest(nil, "PUT", "/out")
	app.NewRequest(nil, "PUT", "/out")

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareRateSpeedTimeout(*testing.T) {
	app := eudore.NewApp()
	app.SetHandler(http.TimeoutHandler(app, 2*time.Second, ""))

	// /done限速512B
	app.PostFunc("/done", func(ctx eudore.Context) {
		c, cannel := context.WithCancel(ctx.GetContext())
		ctx.SetContext(c)
		cannel()
	}, middleware.NewRateSpeedFunc(512, 1024, app.Context), func(ctx eudore.Context) {
		ctx.Debug(string(ctx.Body()))
	})

	// 测试数据限速16B
	app.AddMiddleware(middleware.NewRateSpeedFunc(16, 128, app.Context))
	app.AnyFunc("/get", func(ctx eudore.Context) {
		for i := 0; i < 10; i++ {
			ctx.WriteString("rate speed =16B\n")
		}
	})
	app.PostFunc("/post", func(ctx eudore.Context) {
		ctx.Debug(string(ctx.Body()))
	})
	app.AnyFunc("/*", eudore.HandlerEmpty)

	app.NewRequest(nil, "GET", "/get")
	app.NewRequest(nil, "POST", "/post", eudore.NewClientBodyString("read body is to long,body太大，会中间件超时无法完全读取。"))
	app.NewRequest(nil, "POST", "/done", eudore.NewClientBodyString("hello"))

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareReferer(*testing.T) {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewRefererFunc(map[string]bool{
		"":                         true,
		"origin":                   false,
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

	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderReferer, ""))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHost("www.eudore.cn"), eudore.NewClientHeader(eudore.HeaderReferer, "http://www.eudore.cn/"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderReferer, "http://www.eudore.cn/"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderReferer, "http://www.example.com"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderReferer, "http://www.example.com/"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderReferer, "http://www.example.com/1"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderReferer, "http://www.example.com/1/1"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderReferer, "http://www.example.com/1/2"))
	app.NewRequest(nil, "GET", "/", eudore.NewClientHeader(eudore.HeaderReferer, "http://127.0.0.1/1"))

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

func TestMiddlewareNethttpBasicAuth(*testing.T) {
	data := map[string]string{"user": "pw"}

	app := eudore.NewApp()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {})
	app.SetHandler(middleware.NewNetHTTPBasicAuthFunc(mux, data))

	app.NewRequest(nil, "GET", "/1")
	app.NewRequest(nil, "GET", "/2", eudore.NewClientHeader("Authorization", "Basic dXNlcjpwdw=="))

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareNethttpBlack(*testing.T) {
	app := eudore.NewApp()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {})
	app.SetHandler(middleware.NewNetHTTPBlackFunc(mux, map[string]bool{
		"127.0.0.1/8":    true,
		"192.168.0.0/16": true,
		"10.0.0.0/8":     false,
	}))

	app.NewRequest(nil, "GET", "/eudore/debug/black/ui")
	app.NewRequest(nil, "GET", "/eudore/debug/black/ui")
	app.NewRequest(nil, "PUT", "/eudore/debug/black/black/10.127.87.0?mask=24")
	app.NewRequest(nil, "PUT", "/eudore/debug/black/white/10.127.87.0?mask=24")
	app.NewRequest(nil, "GET", "/eudore/debug/black/data")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/black/10.127.87.0?mask=24")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/white/10.127.87.0?mask=24")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/black/10.127.87.0?mask=24")
	app.NewRequest(nil, "DELETE", "/eudore/debug/black/white/10.127.87.0?mask=24")

	app.NewRequest(nil, "GET", "/eudore")
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXForwardedFor, "192.168.1.4 192.168.1.1"))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "127.0.0.1:29398"))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "192.168.75.1:8298"))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "10.1.1.1:2334"))
	app.NewRequest(nil, "GET", "/eudore", eudore.NewClientHeader(eudore.HeaderXRealIP, "172.17.1.1:2334"))

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareNethttpRateRequest(*testing.T) {
	app := eudore.NewApp()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {})
	app.SetHandler(middleware.NewNetHTTPRateRequestFunc(mux, 1, 3, func(req *http.Request) string {
		// 自定义限流key
		return req.UserAgent()
	}))

	app.NewRequest(nil, "GET", "/")
	app.NewRequest(nil, "GET", "/")
	app.NewRequest(nil, "GET", "/")
	app.NewRequest(nil, "GET", "/")
	app.NewRequest(nil, "GET", "/")

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareNethttpRewrite(*testing.T) {
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
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {})
	app.SetHandler(middleware.NewNetHTTPRewriteFunc(mux, rewritedata))

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

/*
goos: linux
goarch: amd64
BenchmarkMiddlewareBlackTree-2        	 1000000	      1212 ns/op	       0 B/op	       0 allocs/op
BenchmarkMiddlewareBlackArray-2       	 1000000	      1956 ns/op	       0 B/op	       0 allocs/op
BenchmarkMiddlewareBlackIp2intbit-2   	 1000000	      1654 ns/op	     320 B/op	       5 allocs/op
BenchmarkMiddlewareBlackNetParse-2    	 1000000	      1989 ns/op	     360 B/op	      20 allocs/op
PASS
ok  	command-line-arguments	6.919s
*/

var ips []string = []string{
	"10.0.0.0/4", "127.0.0.1/8", "192.168.1.0/24", "192.168.75.0/24", "192.168.100.0/24",
}

var requests []uint64 = []uint64{
	725415979, 2727437335, 889276411, 4005535794, 3864288534, 3906172701, 282878927, 1284469666, 730935782, 3371086418,
	1506312450, 1351422527, 1427742110, 1787801507, 2252116061, 229145224, 2463885032, 977944943, 3785363053, 3752670878,
	1109101831, 523139815, 2692892509, 822628332, 1521829731, 1137604504, 3946127316, 3492727158, 3701842868, 1345785201,
	2479587981, 1525387624, 2335875430, 2742578379, 842531784, 4164034788, 4067025409, 3579565778, 1135250289, 2272239320,
	2221887036, 47163049, 756685807, 3064055796, 2298095091, 3099116819, 4070972416, 1014033, 3023215026, 555430525,
	3702021454, 2340802113, 2507760403, 510831888, 3073321492, 4221140315, 1198583294, 1495418697, 827583711, 813333453,
	2746343126, 3755199452, 1697814659, 365059279, 3478405321, 2147566177, 281339662, 2742376600, 2293307920, 2061663865,
	913999062, 542572186, 4225265321, 633066366, 2063795404, 522841846, 195572401, 124532676, 2456662794, 3902204181,
	2491401143, 4233234751, 69766498, 388520887, 1017105985, 62871287, 3328355052, 1705168586, 2260082173, 3340006743,
	2211140888, 1906467873, 1247205260, 1492905294, 1014862918, 2587182986, 1040587870, 3570772999, 3084952258, 2425691705,
}

var requeststrs []string = []string{
	"43.60.248.43", "162.145.100.23", "53.1.71.251", "238.191.160.50", "230.84.93.22", "232.211.119.29", "16.220.99.207", "76.143.115.162", "43.145.49.230", "200.238.178.82",
	"89.200.129.2", "80.141.18.63", "85.25.157.158", "106.143.175.163", "134.60.144.93", "13.168.122.136", "146.219.230.232", "58.74.65.111", "225.160.14.109", "223.173.54.158",
	"66.27.141.7", "31.46.122.231", "160.130.71.93", "49.8.79.236", "90.181.71.99", "67.206.119.152", "235.53.31.212", "208.46.201.118", "220.165.163.180", "80.55.13.113",
	"147.203.130.141", "90.235.145.104", "139.58.161.102", "163.120.108.203", "50.56.3.200", "248.50.32.228", "242.105.226.1", "213.91.214.210", "67.170.139.113", "135.111.158.216",
	"132.111.78.60", "2.207.166.169", "45.26.27.239", "182.161.199.244", "136.250.37.243", "184.184.197.19", "242.166.28.0", "0.15.121.17", "180.50.153.178", "33.27.50.125",
	"220.168.93.78", "139.133.206.65", "149.121.99.19", "30.114.173.16", "183.47.42.20", "251.153.125.91", "71.112.237.254", "89.34.71.73", "49.83.236.223", "48.122.123.205",
	"163.177.222.214", "223.211.203.220", "101.50.152.131", "21.194.92.207", "207.84.64.201", "128.1.66.97", "16.196.231.14", "163.117.88.152", "136.177.26.16", "122.226.126.121",
	"54.122.132.214", "32.86.254.154", "251.216.110.169", "37.187.211.126", "123.3.4.204", "31.41.238.246", "11.168.50.177", "7.108.55.196", "146.109.179.10", "232.150.233.21",
	"148.127.195.183", "252.82.9.63", "4.40.141.98", "23.40.91.183", "60.159.206.65", "3.191.86.247", "198.98.170.236", "101.162.206.202", "134.182.29.253", "199.20.117.87",
	"131.203.85.24", "113.162.100.33", "74.86.215.140", "88.251.237.78", "60.125.148.70", "154.53.71.138", "62.6.28.94", "212.213.172.7", "183.224.162.194", "144.149.30.57",
}

/*
func TestMiddlewareBlackResult(t *testing.T) {
	tree := new(middleware.BlackNode)
	array := new(BlackNodeArray)
	for _, ip := range ips {
		tree.Insert(ip)
		array.Insert(ip)
	}
	for _, ip := range requests {
		if tree.Look(ip) != array.Look(ip) {
			t.Logf("tree: %t array: %t result not equal %d %s", tree.Look(ip), array.Look(ip), ip, int2ip(ip))
		}
	}
}

func BenchmarkMiddlewareBlackTree(b *testing.B) {
	node := new(middleware.BlackNode)
	for _, ip := range ips {
		node.Insert(ip)
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, ip := range requests {
			node.Look(ip)
		}
	}
}
*/

func BenchmarkMiddlewareBlackArray(b *testing.B) {
	node := new(BlackNodeArray)
	b.ReportAllocs()
	for _, ip := range ips {
		node.Insert(ip)
	}
	for i := 0; i < b.N; i++ {
		for _, ip := range requests {
			node.Look(ip)
		}
	}
}

func TestMiddlewareBlackParseip(t *testing.T) {
	for _, ip := range ips {
		ip1, bit1 := ip2intbit(ip)
		ip2, bit2 := ip2netintbit(ip)
		if ip1 != ip2 || bit1 != bit2 {
			t.Log("ip parse error", ip, ip1, ip2, bit1, bit2)
		}
	}
}

func BenchmarkMiddlewareBlackIp2intbit(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, ip := range ips {
			ip2intbit(ip)
		}
	}
}

func BenchmarkMiddlewareBlackNetParse(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, ip := range ips {
			ip2netintbit(ip)
		}
	}
}

// BlackNodeArray 定义数组遍历实现ip解析
type BlackNodeArray struct {
	Data  []uint64
	Mask  []uint
	Count []uint64
}

// Insert 方法给黑名单节点新增一个ip或ip段。
func (node *BlackNodeArray) Insert(ip string) {
	iip, bit := ip2intbit(ip)
	node.Data = append(node.Data, iip>>(32-bit))
	node.Mask = append(node.Mask, 32-bit)
	node.Count = append(node.Count, 0)
}

// Look 方法匹配ip是否在黑名单节点，命中则节点计数加一。
func (node *BlackNodeArray) Look(ip uint64) bool {
	for i := range node.Data {
		if node.Data[i] == (ip >> node.Mask[i]) {
			node.Count[i]++
			return true
		}
	}
	return false
}

// BlackNodeArrayNet 定义基于net库实现ip遍历匹配，支持ipv6.
type BlackNodeArrayNet struct {
	Data  []net.IP
	Mask  []net.IPMask
	Count []uint64
}

// Insert 方法给黑名单节点新增一个ip或ip段。
func (node *BlackNodeArrayNet) Insert(ip string) {
	_, ipnet, _ := net.ParseCIDR(ip)
	node.Data = append(node.Data, ipnet.IP)
	node.Mask = append(node.Mask, ipnet.Mask)
	node.Count = append(node.Count, 0)
}

// Look 方法匹配ip是否在黑名单节点，命中则节点计数加一。
func (node *BlackNodeArrayNet) Look(ip string) bool {
	netip := net.ParseIP(ip)
	for i := range node.Data {
		if node.Data[i].Equal(netip.Mask(node.Mask[i])) {
			node.Count[i]++
			return true
		}
	}
	return false
}

func ip2netintbit(ip string) (uint64, uint) {
	ipaddr, ipnet, _ := net.ParseCIDR(ip)
	length := len(ipaddr)
	bit, _ := ipnet.Mask.Size()
	var sum uint64
	sum += uint64(ipaddr[length-4]) << 24
	sum += uint64(ipaddr[length-3]) << 16
	sum += uint64(ipaddr[length-2]) << 8
	sum += uint64(ipaddr[length-1])
	return sum, uint(bit)
}

func ip2intbit(ip string) (uint64, uint) {
	bit := 32
	pos := strings.Index(ip, "/")
	if pos != -1 {
		bit, _ = strconv.Atoi(ip[pos+1:])
		ip = ip[:pos]
	}
	return ip2int(ip), uint(bit)
}

func ip2int(ip string) uint64 {
	bits := strings.Split(ip, ".")
	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum uint64
	sum += uint64(b0) << 24
	sum += uint64(b1) << 16
	sum += uint64(b2) << 8
	sum += uint64(b3)
	return sum
}

func int2ip(ip uint64) string {
	var bytes [4]uint64
	bytes[0] = ip & 0xFF
	bytes[1] = (ip >> 8) & 0xFF
	bytes[2] = (ip >> 16) & 0xFF
	bytes[3] = (ip >> 24) & 0xFF
	return fmt.Sprintf("%d.%d.%d.%d", bytes[3], bytes[2], bytes[1], bytes[0])
}

func BenchmarkMiddlewareRewrite(b *testing.B) {
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
	app.SetValue(eudore.ContextKeyLogger, eudore.NewLoggerInit())
	app.AddMiddleware("global", middleware.NewRewriteFunc(rewritedata))
	app.AnyFunc("/*", eudore.HandlerEmpty)
	paths := []string{"/", "/js/", "/js/index.js", "/api/v1/user", "/api/v1/user/new", "/api/v1/users/v3/orders/8920", "/api/v1/users/orders", "/api/v2", "/api/v2/user", "/d/3", "/help/history", "/help/historyv2"}
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			r.URL.Path = path
			app.ServeHTTP(w, r)
		}
	}
}
func BenchmarkMiddlewareRewriteWithZero(b *testing.B) {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyLogger, eudore.NewLoggerInit())
	app.AnyFunc("/*", eudore.HandlerEmpty)
	paths := []string{"/", "/js/", "/js/index.js", "/api/v1/user", "/api/v1/user/new", "/api/v1/users/v3/orders/8920", "/api/v1/users/orders", "/api/v2", "/api/v2/user", "/d/3", "/help/history", "/help/historyv2"}
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			r.URL.Path = path
			app.ServeHTTP(w, r)
		}
	}
}

func BenchmarkMiddlewareRewriteWithRouter(b *testing.B) {
	routerdata := map[string]interface{}{
		"/js/*0":                     newRewriteFunc("/public/js/$0"),
		"/api/v1/users/:0/orders/*1": newRewriteFunc("/api/v3/user/$0/order/$1"),
		"/d/*0":                      newRewriteFunc("/d/$0-$0"),
		"/api/v1/*0":                 newRewriteFunc("/api/v3/$0"),
		"/api/v2/*0":                 newRewriteFunc("/api/v3/$0"),
		"/help/history*0":            newRewriteFunc("/api/v3/history"),
		"/help/history":              newRewriteFunc("/api/v3/history"),
		"/help/*0":                   newRewriteFunc("$0"),
	}
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyLogger, eudore.NewLoggerInit())
	app.AddMiddleware("global", middleware.NewRouterFunc(routerdata))
	app.AnyFunc("/*", eudore.HandlerEmpty)
	paths := []string{"/", "/js/", "/js/index.js", "/api/v1/user", "/api/v1/user/new", "/api/v1/users/v3/orders/8920", "/api/v1/users/orders", "/api/v2", "/api/v2/user", "/d/3", "/help/history", "/help/historyv2"}
	w, r := httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, path := range paths {
			r.URL.Path = path
			app.ServeHTTP(w, r)
		}
	}
}

func newRewriteFunc(path string) eudore.HandlerFunc {
	paths := strings.Split(path, "$")
	Index := make([]string, 1, len(paths)*2-1)
	Data := make([]string, 1, len(paths)*2-1)
	Index[0] = ""
	Data[0] = paths[0]
	for _, path := range paths[1:] {
		Index = append(Index, path[0:1])
		Data = append(Data, "")
		if path[1:] != "" {
			Index = append(Index, "")
			Data = append(Data, path[1:])
		}
	}
	return func(ctx eudore.Context) {
		buffer := bytes.NewBuffer(nil)
		for i := range Index {
			if Index[i] == "" {
				buffer.WriteString(Data[i])
			} else {
				buffer.WriteString(ctx.GetParam(Index[i]))
			}
		}
		ctx.Request().URL.Path = buffer.String()
	}
}
