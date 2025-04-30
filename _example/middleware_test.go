package eudore_test

// middleware_test.go all.go
// middleware2_test.go other
// middleware3_test.go midd options
// middleware4_test.go midd radix

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	. "github.com/eudore/eudore"
	. "github.com/eudore/eudore/middleware"
)

func TestMiddlewareBasicAuth(*testing.T) {
	app := NewApp()
	app.AddMiddleware("global", NewBasicAuthFunc(map[string]string{"eudore": "hello"}))
	app.AnyFunc("/", NewAdminFunc())

	app.GetRequest("/", http.Header{HeaderAuthorization: {"Basic ZXVkb3JlOmhlbGxv"}}, NewClientCheckStatus(200))
	app.GetRequest("/", http.Header{HeaderAuthorization: {"eudore"}}, NewClientCheckStatus(401))

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareBodyLimit(*testing.T) {
	app := NewApp()
	app.AddMiddleware("global",
		NewGzipFunc(),
		NewBodySizeFunc(),
		NewBodyLimitFunc(32),
		NewLoggerLevelFunc(func(Context) int { return 4 }),
	)
	app.AnyFunc("/", func(ctx Context) {
		_, err := ctx.Body()
		if err != nil {
			ctx.Fatal(err)
		}
	})
	app.AnyFunc("/form", func(ctx Context) {
		_, err := ctx.FormValues()
		if err != nil {
			ctx.Fatal(err)
		}
	})

	app.GetRequest("/", NewClientCheckStatus(200))
	app.GetRequest("/", strings.NewReader("123456"), NewClientCheckStatus(200))
	app.GetRequest("/", strings.NewReader("1234567890abcdefghijklmnopqrstuvwxyz"), NewClientCheckStatus(413))
	// limit chunck
	data := url.Values{
		"name":  {"eudore"},
		"value": {"1234567890abcdefghijklmnopqrstuvwxyz"},
	}
	app.GetRequest("/", NewClientBodyForm(data), NewClientCheckStatus(413))
	app.GetRequest("/form", NewClientBodyForm(data), NewClientCheckStatus(413))

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareContextWrap(*testing.T) {
	app := NewApp()
	app.AddMiddleware(NewContextWrapperFunc(newContextParams))
	app.AnyFunc("/ctx", func(ctx Context) {
		index, handler := ctx.GetHandlers()
		ctx.SetHandlers(index, handler)
		if ctx.GetParam("") != "none" {
			panic("ContextWrap")
		}
	})
	app.AnyFunc("/*", func(ctx Context) {
		ctx.Debug("hello eudore")
		ctx.Info("hello eudore")
		ctx.End()
	})

	app.GetRequest("/")
	app.GetRequest("/ctx")

	app.CancelFunc()
	app.Run()
}

func newContextParams(ctx Context) Context {
	return contextParams{ctx}
}

type contextParams struct {
	contextBase
}

type contextBase = Context

// GetParam 方法获取一个参数的值。
func (ctx contextParams) GetParam(key string) string {
	return "none"
}

func TestMiddlewareHeader(*testing.T) {
	app := NewApp()
	app.AddMiddleware("global", NewHeaderAddSecureFunc(http.Header{"Server": {"eudore"}}))
	app.AddMiddleware("global", NewHeaderAddFunc(nil))
	app.AddMiddleware(func(ctx Context) {
		addr := ctx.GetQuery("addr")
		if addr != "" {
			ctx.Request().RemoteAddr = addr
		}
	})
	app.AnyFunc("/default", NewHeaderDeleteFunc(nil, nil))
	app.AnyFunc("/", NewHeaderDeleteFunc([]string{
		"127.0.0.0/24",
		"::1",
	}, nil))

	app.GetRequest("/")
	app.GetRequest("/")
	app.GetRequest("/?addr=192.0.0.1:50424")
	app.GetRequest("/?addr=127.0.0.1:50424")
	app.GetRequest("/?addr=[::1]:50424")

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareRecover(*testing.T) {
	app := NewApp()
	app.AddMiddleware("global",
		NewServerTimingFunc(),
		NewRequestIDFunc(nil),
		NewRecoveryFunc(),
		NewLoggerLevelFunc(func(Context) int { return 4 }),
	)
	app.AnyFunc("/panic", func(ctx Context) {
		panic("test error")
	})
	app.AnyFunc("/err", func(ctx Context) {
		panic(fmt.Errorf("test error"))
	})
	app.AnyFunc("/nil", func(ctx Context) {
		panic(nil)
	})
	app.AnyFunc("/timing", func(ctx Context) {
		ctx.WriteHeader(206)
		ctx.Write(nil)
		ctx.WriteString("hello")
		ctx.Response().Flush()
	})

	app.GetRequest("/timing", NewClientCheckStatus(206))
	app.GetRequest("/panic", NewClientCheckStatus(500))
	app.GetRequest("/err", NewClientCheckStatus(500))
	app.GetRequest("/nil", NewClientCheckStatus(200))

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareRoutes(*testing.T) {
	hend := func(ctx Context) { ctx.End() }
	h500 := func(ctx Context) { ctx.WriteHeader(500) }
	app := NewApp()
	app.AddMiddleware("global",
		NewRoutesFunc(map[string]any{
			"/api/*":   HandlerFuncs{hend, h500},
			"GET /500": hend,
		}),
	)
	app.GetFunc("/500", h500)
	app.GetFunc("/sub", NewRoutesFunc(map[string]any{}))
	app.GetFunc("/", HandlerEmpty)

	app.GetRequest("/", NewClientCheckStatus(200))
	app.GetRequest("/sub", NewClientCheckStatus(200))
	app.GetRequest("/api/v1", NewClientCheckStatus(200))
	app.GetRequest("/500", NewClientCheckStatus(200))

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareSkipHandler(*testing.T) {
	NewSkipHandlerFunc("", nil)
	NewSkipHandlerFunc("", map[string]struct{}{})
	app := NewApp()
	app.GetFunc("/path/*", NewSkipHandlerFunc("path", map[string]struct{}{"/path/200": {}}), HandlerRouter403)
	app.GetFunc("/param", NewSkipHandlerFunc("param:route", map[string]struct{}{"/param": {}}), HandlerRouter403)
	app.GetFunc("/cookie", NewSkipHandlerFunc("cookie:name", map[string]struct{}{"eudore": {}}), HandlerRouter403)
	app.GetFunc("/request", NewSkipHandlerFunc("request:name", map[string]struct{}{"eudore": {}}), HandlerRouter403)

	app.GetRequest("/path/200", NewClientCheckStatus(200))
	app.GetRequest("/path/201", NewClientCheckStatus(403))
	app.GetRequest("/param", NewClientCheckStatus(200))
	app.GetRequest("/cookie", &Cookie{"name", "eudore"}, NewClientCheckStatus(200))
	app.GetRequest("/cookie", NewClientCheckStatus(403))
	app.GetRequest("/request", http.Header{"Name": []string{"eudore"}}, NewClientCheckStatus(200))
	app.GetRequest("/request", NewClientCheckStatus(403))

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareOption(*testing.T) {
	op := NewOptionKeyFunc(func(ctx Context) string { return "" })
	NewCSRFFunc("", op)
	NewCacheFunc(0, op)
	NewCircuitBreakerFunc(op)
	NewRateRequestFunc(1, 1, op)
}

func TestMiddlewareName(t *testing.T) {
	app := NewApp()
	hs := []HandlerFunc{
		NewAdminFunc(),
		NewBasicAuthFunc(map[string]string{}),
		NewBlackListFunc(map[string]bool{}),
		NewBodyLimitFunc(4 << 20),
		NewBodySizeFunc(),
		NewCORSFunc(nil, nil),
		NewCSRFFunc("_csrf"),
		NewCacheFunc(time.Second),
		NewCircuitBreakerFunc(),
		NewCompressionFunc("gz", func() any { return nil }),
		NewCompressionMixinsFunc(nil),
		NewContextWrapperFunc(nil),
		NewDumpFunc(app.Group(" loggerkind=~all")),
		NewGzipFunc(),
		NewHeaderAddFunc(http.Header{"X": []string{"x"}}),
		NewHeaderAddSecureFunc(http.Header{}),
		NewHeaderDeleteFunc(nil, nil),
		NewHealthCheckFunc(app),
		NewLoggerFunc(app),
		NewLoggerLevelFunc(nil),
		NewLoggerWithEventFunc(app),
		NewLookFunc(app),
		NewMetadataFunc(app),
		NewPProfFunc(),
		NewRateRequestFunc(4, 16),
		NewRateSpeedFunc(4, 16),
		NewRecoveryFunc(),
		NewRefererCheckFunc(map[string]bool{}),
		NewRequestIDFunc(nil),
		NewRewriteFunc(map[string]string{}),
		NewRouterFunc(app),
		NewRoutesFunc(map[string]any{}),
		NewServerTimingFunc(),
		NewSkipHandlerFunc("path", map[string]struct{}{"/": {}}),
		NewTimeoutFunc(app.ContextPool, time.Second),
		NewTimeoutSkipFunc(app.ContextPool, time.Second, nil),
	}
	for _, h := range hs {
		rh := reflect.ValueOf(h)
		name := runtime.FuncForPC(rh.Pointer()).Name()
		if !strings.Contains(name, "/eudore/middleware.New") {
			panic(h.String())
		}
	}
}
