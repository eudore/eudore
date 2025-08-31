package eudore_test

import (
	"bufio"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	. "github.com/eudore/eudore"
	. "github.com/eudore/eudore/middleware"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type signingMethod struct{}

func (fn signingMethod) Verify(_ string, _ []byte, _ any) error {
	return nil
}

func (fn signingMethod) Alg() string {
	return "HS256"
}

func brarerSigned(claims any, salt ...string) string {
	salt = append(salt, "", "", "")
	payload, err := json.Marshal(claims)
	if err != nil {
		return err.Error()
	}
	payload = append(payload, []byte(salt[0])...)
	unsigned := `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.` +
		base64.RawURLEncoding.EncodeToString(payload) + salt[1]

	h := hmac.New(sha256.New, []byte("secret"))
	h.Write([]byte(unsigned))
	return unsigned + "." + base64.RawURLEncoding.EncodeToString(h.Sum(nil)) + salt[2]
}

func TestMiddlewareBearerAuth(*testing.T) {
	type user struct {
		Userid     int    `json:"userid"`
		Username   string `json:"username,omitempty"`
		NotBefore  int64  `json:"nbf,omitempty"`
		Expiration int64  `json:"exp,omitempty"`
	}
	type user2 struct {
		Userid func()
	}
	app := NewApp()
	app.AddMiddleware(
		NewBearerAuthFunc("secret",
			NewOptionBearerSignaturer(&signingMethod{}),
			NewOptionBearerPayload[user2](NewContextKey("user")),
		),
		NewBearerAuthFunc(
			[]byte("secret"),
			NewOptionKeyFunc(func(ctx Context) string {
				auth := ctx.GetHeader(HeaderAuthorization)
				if auth != "" {
					return auth
				}

				auth = ctx.GetQuery("access_token")
				if auth != "" {
					return "Bearer " + auth
				}
				return ""
			}),
			NewOptionBearerSignaturer(nil),
			NewOptionBearerPayload[user](NewContextKey("user")),
		),
	)
	app.GetFunc("/*", HandlerEmpty)

	bearers := []string{
		"",
		"x." + brarerSigned(&user{Userid: 10000}),
		"x" + brarerSigned(&user{Userid: 10000}),
		brarerSigned(&user{Userid: 10000}) + "x",
		brarerSigned(&user{Userid: 10000}, "00"),
		brarerSigned(&user{Userid: 10000}, "", "=="),
		brarerSigned(&user{Userid: 10000}, "", "", "=="),
		brarerSigned(&user{Userid: 10000, NotBefore: time.Now().Add(10 * time.Hour).Unix()}),
		brarerSigned(&user{Userid: 10000, Expiration: time.Now().Add(-10 * time.Hour).Unix()}),
		brarerSigned(&user{Userid: 10000, Username: "eudoore"}),
	}

	for _, bearer := range bearers {
		app.GetRequest("/", NewClientOptionBearer(bearer))
	}

	app.CancelFunc()
	app.Run()
}

type metadatable func() any

func (fn metadatable) Metadata() any {
	return fn()
}

func TestMiddlewareHealth(*testing.T) {
	app := NewApp()
	app.GetFunc("/health", NewHealthCheckFunc(app))
	app.GetFunc("/meta/*name", NewMetadataFunc(app))
	app.GetRequest("/health", NewClientCheckStatus(200))
	app.GetRequest("/meta/", NewClientCheckStatus(200))
	app.GetRequest("/meta/router", NewClientCheckStatus(200))
	app.GetRequest("/meta/app", NewClientCheckStatus(404))

	app.SetValue(NewContextKey("m1"), metadatable(func() any {
		return 0
	}))
	app.SetValue(NewContextKey("m2"), metadatable(func() any {
		return &struct{}{}
	}))
	app.SetValue(NewContextKey("m3"), metadatable(func() any {
		return struct{ Health bool }{false}
	}))
	app.SetValue(NewContextKey("h1"), NewHandlerExtenderTree())
	app.SetValue(NewContextKey("h2"), NewHandlerExtenderWrap(
		NewHandlerExtenderTree(), DefaultHandlerExtender),
	)
	app.GetRequest("/health", NewClientCheckStatus(503))
	app.GetRequest("/meta/", NewClientCheckStatus(503))

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareLoggerAccess(*testing.T) {
	log := NewLogger(&LoggerConfig{
		Stdout: false,
		Handlers: []LoggerHandler{
			NewLoggerFormatterText(DefaultLoggerFormatterFormatTime),
		},
	})
	app := NewApp()
	app.AddMiddleware("global",
		NewLoggerFunc(log,
			"remote_addr",
			"scheme",
			"querys",
			"byte_in",
			"request:"+HeaderXForwardedFor,
			"response:"+HeaderXRequestID,
			"response:"+HeaderXTraceID,
			"response:"+HeaderLocation,
			"param:path",
			"cookie:sessionid",
		),
		NewLoggerWithEventFunc(log),
		NewLoggerLevelFunc(func(Context) int { return 4 }),
	)
	app.AddMiddleware("global", NewRequestIDFunc(func(Context) string {
		return GetStringRandom(16)
	}))
	app.AnyFunc("/long", func(ctx Context) {
		time.Sleep(time.Millisecond * 10)
	})
	app.AnyFunc("/500", func(ctx Context) {
		ctx.Fatal("test error")
	})
	app.AnyFunc("/s-v6", func(ctx Context) {
		ctx.Request().TLS = &tls.ConnectionState{}
		ctx.Request().RemoteAddr = "[::1]:35002"
	})
	app.AnyFunc("/sse", HandlerEvent)

	app.GetRequest("/", http.Header{HeaderXForwardedFor: {"172.17.0.1"}})
	app.PostRequest("/long")
	app.PostRequest("/500")
	app.PostRequest("/s-v6")
	app.GetRequest("/sse", http.Header{HeaderAccept: {MimeTextEventStream}})

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareLoggerLevel(*testing.T) {
	app := NewApp()
	app.SetLevel(LoggerInfo)
	app.AddMiddleware(NewLoggerLevelFunc(nil))
	app.AnyFunc("/*", HandlerEmpty)

	app.GetRequest("/")
	app.GetRequest("/?eudore_debug=0")
	app.GetRequest("/?eudore_debug=1")
	app.GetRequest("/?eudore_debug=5")

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareCSRF(*testing.T) {
	app := NewApp()
	app.AddMiddleware(NewCSRFFunc("_csrf",
		NewOptionCSRFCookie(http.Cookie{Name: "_csrf"}),
	))
	app.AnyFunc("/*", HandlerEmpty)

	var csrfval string
	app.GetRequest("/",
		NewClientCheckStatus(200),
		func(w *http.Response) error {
			csrfval = w.Header.Get(HeaderSetCookie)
			app.Info("csrf token:", csrfval)
			return nil
		},
	)

	app.PostRequest("/2", NewClientCheckStatus(400))
	app.PostRequest("/1",
		NewClientQuery("_csrf", strings.TrimPrefix(csrfval, "_csrf=")),
		NewClientHeader("Cookie", csrfval),
		NewClientCheckStatus(200),
	)

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareDump(*testing.T) {
	type dumpMessage struct {
		Time          string      `json:"time"`
		Path          string      `json:"path"`
		Host          string      `json:"host"`
		RemoteAddr    string      `json:"remoteAddr"`
		Proto         string      `json:"proto"`
		Method        string      `json:"method"`
		RequestURI    string      `json:"requestURI"`
		RequestHeader http.Header `json:"requestHeader"`
		// RequestBody    []byte        `json:"requestBody"`
		Status         int         `json:"status"`
		ResponseHeader http.Header `json:"responseHeader"`
		// ResponseBody   []byte        `json:"responseBody"`
		Params   []string `json:"params"`
		Handlers []string `json:"handlers"`
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
		}
	}

	app := NewApp()
	app.AddMiddleware(
		"global",
		NewLoggerLevelFunc(func(Context) int { return 4 }),
		func(ctx Context) {
			if ctx.GetQuery("nodump") != "" {
				ctx.SetResponse(&responseWriterNoHijack{ctx.Response()})
			}
			if ctx.GetQuery("nobody") != "" {
				ctx.Request().Body = io.NopCloser(bodyNoreader{})
			}
		})
	app.AddMiddleware(
		NewRecoveryFunc(),
		NewDumpFunc(app.Group("/eudore/debug")),
	)
	app.AnyFunc("/echo", func(ctx Context) {
		io.Copy(ctx, ctx)
		w := ctx.Response()
		u, ok := w.(interface{ Unwrap() http.ResponseWriter })
		if ok {
			u.Unwrap()
		}
		b, ok := w.(interface{ Body() []byte })
		if ok {
			b.Body()
		}
	})
	app.AnyFunc("/panic", func(Context) {
		panic("recover")
	})
	app.AnyFunc("/bigbody", func(ctx Context) {
		ctx.Write([]byte("0123456789abcdef0123456789abcdef0123456789abcdefx"))
		ctx.Write(make([]byte, 0xffff))
	})
	app.AnyFunc("/gzip", NewCompressionFunc(CompressionNameGzip, nil), func(ctx Context) {
		ctx.WriteString("gzip body")
		for i := 0; i < 40; i++ {
			ctx.Write([]byte("0123456789abcdef0123456789abcdef0123456789abcdefxx"))
		}
	})
	app.AnyFunc("/gziperr1", func(ctx Context) {
		ctx.SetHeader(HeaderContentEncoding, "gzip")
		ctx.Write([]byte("gzip body"))
		ctx.WriteString("gzip body")
	})
	app.AnyFunc("/gziperr2", func(ctx Context) {
		ctx.SetHeader(HeaderContentEncoding, "gzip")
		ctx.WriteString("gzip body")
	})
	app.AnyFunc("/*", HandlerEmpty)
	app.Listen(":8088")
	time.Sleep(200 * time.Millisecond)

	app.GetRequest("/echo")
	go ReadDumpMessage("ws://localhost:8088/eudore/debug/dump/connect", 10)
	go ReadDumpMessage("ws://localhost:8088/eudore/debug/dump/connect?nodump=1", 1)
	time.Sleep(200 * time.Millisecond)

	app.GetRequest("http://localhost:8088/eudore/debug/dump/connect")
	app.GetRequest("/gzip", http.Header{HeaderAcceptEncoding: {"gzip"}})
	app.GetRequest("/gzip", http.Header{HeaderAcceptEncoding: {"identity"}})
	app.GetRequest("/gziperr1")
	app.GetRequest("/gziperr2")
	app.GetRequest("/echo")
	app.GetRequest("/echo?nobody=1", strings.NewReader("123456"))
	app.GetRequest("/panic")
	app.GetRequest("/bigbody", func(resp *http.Response) error {
		io.Copy(io.Discard, resp.Body)
		return nil
	})

	time.Sleep(200 * time.Millisecond)
	app.CancelFunc()
	app.Run()
}

type bodyNoreader struct{}

func (bodyNoreader) Read([]byte) (int, error) {
	return 0, fmt.Errorf("no read")
}

type responseWriterNoHijack struct {
	ResponseWriter
}

func (responseWriterNoHijack) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, fmt.Errorf("no Hijack")
}

func TestMiddlewareCompressGzip(*testing.T) {
	longtext := "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"
	longdata := []byte(longtext)
	app := NewApp()
	app.AddMiddleware(
		NewHeaderAddFunc(http.Header{HeaderVary: []string{"User-Agent"}}),
		NewCompressionFunc(CompressionNameGzip, nil),
	)
	app.GetFunc("/*", func(ctx Context) {
		w := ctx.Response()
		w.WriteStatus(StatusOK)
		w.WriteHeader(StatusOK)
		w.Size()
		w.Status()
		w.Push("/stat", nil)
		w.Push("/stat", &http.PushOptions{})
		w.Push("/stat", &http.PushOptions{Header: make(http.Header)})
		w.WriteString("compress")
		w.Flush()
	})
	app.GetFunc("/empty", HandlerEmpty)
	app.GetFunc("/code", func(ctx Context) {
		ctx.SetHeader(HeaderContentLength, "6")
		ctx.WriteHeader(200)
		ctx.WriteString("eudore")
	})
	app.GetFunc("/gzip", func(ctx Context) {
		ctx.SetHeader(HeaderContentType, "application/gzip;encoding=gzip")
		w := ctx.Response()
		for i := 0; i < 20; i++ {
			w.WriteString(longtext)
		}
		w.Flush()
	})
	app.GetFunc("/long", func(ctx Context) {
		ctx.SetHeader(HeaderVary, HeaderOrigin)
		ctx.SetHeader(HeaderContentLength, "610")
		w := ctx.Response()
		for i := 0; i < 5; i++ {
			w.Write(longdata)
			w.WriteString(longtext)
		}
		w.Flush()
	})
	app.GetFunc("/longs", func(ctx Context) {
		w := ctx.Response()
		ctx.SetHeader(HeaderContentLength, "1220")
		for i := 0; i < 10; i++ {
			w.Write(longdata)
			w.WriteString(longtext)
		}
		w.Flush()
	})

	h := http.Header{HeaderAcceptEncoding: {CompressionNameGzip}}
	app.GetRequest("/", http.Header{HeaderAcceptEncoding: {CompressionNameDeflate}})
	app.GetRequest("/empty", h)
	app.GetRequest("/code", h, NewClientCheckBody("eudore"))
	app.GetRequest("/gzip", h)
	app.GetRequest("/", h)
	app.GetRequest("/long", h)
	app.GetRequest("/longs", h)
	app.GetRequest("/long", NewClientCheckBody(longtext))
	app.GetRequest("/longs", NewClientCheckBody(longtext))
	// app.GetRequest("/gzip", NewClientCheckBody(longtext))
	// app.GetRequest("/long", NewClientCheckBody(longtext))

	func() {
		defer func() {
			recover()
		}()
		NewCompressionFunc("miss", nil)
	}()
	func() {
		defer func() {
			recover()
		}()
		NewCompressionFunc("miss", func() any {
			return nil
		})
	}()

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareCompressMixins(*testing.T) {
	DefaultCompressionEncoder["undefined"] = func() interface{} {
		return gzip.NewWriter(io.Discard)
	}
	defer func() {
		delete(DefaultCompressionEncoder, "undefined")
	}()

	app := NewApp()
	app.AddMiddleware(NewCompressionMixinsFunc(nil))
	app.AnyFunc("/*", func(ctx Context) {
		w := ctx.Response()
		w.WriteString("mixins")
		w.Size()
		w.Flush()
	})

	app.GetRequest("/")
	app.GetRequest("/", http.Header{HeaderAcceptEncoding: {CompressionNameGzip}})
	app.GetRequest("/", http.Header{HeaderAcceptEncoding: {CompressionNameDeflate}})
	app.GetRequest("/", http.Header{HeaderAcceptEncoding: {CompressionNameIdentity}})
	app.GetRequest("/", http.Header{HeaderAcceptEncoding: {"gzip;q=0"}})
	app.GetRequest("/", http.Header{HeaderAcceptEncoding: {"none"}})

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
		"bytes":       []byte(`AddHeader(HeaderAcceptEncoding, "none")`),
	}
	app2 := NewApp()
	app2.SetValue(ContextKeyLogger, NewLoggerInit())
	app2.Set("conf", config)
	app2.Set("logger", app2.Logger)
	app2.Set("router", app2.Router)

	app := NewApp()
	app.AddMiddleware(NewLoggerLevelFunc(func(Context) int { return 4 }))
	app.AnyFunc("/eudore/debug/look/*", NewLookFunc(app2))
	app.AnyFunc("/eudore/debug/data", NewLookFunc(func(Context) interface{} {
		return nil
	}))

	app.GetRequest("/eudore/debug/data")
	app.GetRequest("/eudore/debug/look/?d=3")
	app.GetRequest("/eudore/debug/look/?all=1")
	app.GetRequest("/eudore/debug/look/?format=text")
	app.GetRequest("/eudore/debug/look/?format=json")
	app.GetRequest("/eudore/debug/look/?format=t2")
	app.GetRequest("/eudore/debug/look/Config/Keys/2")
	app.GetRequest("/eudore/debug/look/?d=3", http.Header{HeaderAccept: {MimeApplicationJSON}})
	app.GetRequest("/eudore/debug/look/?d=3", http.Header{HeaderAccept: {MimeTextHTML}})
	app.GetRequest("/eudore/debug/look/?d=3", http.Header{HeaderAccept: {MimeText}})

	app.CancelFunc()
	app.Run()
}

func TestMiddlewarePprof(*testing.T) {
	app := NewApp()
	app.AddMiddleware(
		NewHeaderDeleteFunc(nil, nil),
		NewRecoveryFunc(),
		NewCompressionMixinsFunc(nil),
	)
	app.AnyFunc("/eudore/debug/pprof/*", NewPProfFunc())
	app.AnyFunc("/wait", func(ctx Context) {
		time.Sleep(time.Second)
	})

	for i := 0; i < 4; i++ {
		go app.GetRequest("/wait")
	}

	app.GetRequest("/eudore/debug/pprof/expvar", http.Header{HeaderAccept: {MimeApplicationJSON}})
	app.GetRequest("/eudore/debug/pprof/?format=json")
	app.GetRequest("/eudore/debug/pprof/?format=text")
	app.GetRequest("/eudore/debug/pprof/?format=html")
	app.GetRequest("/eudore/debug/pprof/allocs")
	app.GetRequest("/eudore/debug/pprof/block")
	app.GetRequest("/eudore/debug/pprof/heap")
	app.GetRequest("/eudore/debug/pprof/mutex")
	app.GetRequest("/eudore/debug/pprof/goroutine?debug=0")
	app.GetRequest("/eudore/debug/pprof/goroutine?debug=1")
	app.GetRequest("/eudore/debug/pprof/goroutine?debug=1&format=json")
	app.GetRequest("/eudore/debug/pprof/goroutine?debug=1&format=text")
	app.GetRequest("/eudore/debug/pprof/goroutine?debug=1&format=html")
	app.GetRequest("/eudore/debug/pprof/goroutine?debug=2&format=json")
	app.GetRequest("/eudore/debug/pprof/goroutine?debug=2&format=text")
	app.GetRequest("/eudore/debug/pprof/goroutine?debug=2&format=html")

	app.CancelFunc()
	app.Run()
}

func TestMiddlewareTimeout(*testing.T) {
	app := NewApp()
	app.AddMiddleware(func(ctx Context) {
		if ctx.Path() == "/cancel" {
			c, cancel := context.WithCancel(ctx.Context())
			cancel()
			ctx.SetContext(c)
		}
		ctx.SetValue("Name", "eudore")
	})
	app.AddMiddleware(
		NewRecoveryFunc(),
		NewLoggerLevelFunc(func(Context) int { return 4 }),
		NewTimeoutSkipFunc(app.ContextPool, time.Millisecond*20, nil),
		NewBodyLimitFunc(4<<20),
	)
	app.AnyFunc("/panic", func() {
		panic(0)
	})
	app.AnyFunc("/hello", func(ctx Context) {
		ctx.SetHeader("Name", "eudore")
		w := ctx.Response()
		ctx.Info(
			w.Status(), w.Size(), ctx.GetParam("route"),
			ctx.Value("Name"), ctx.FormValue("name"),
		)
		w.WriteStatus(200)
		w.WriteHeader(200)
		w.Flush()
		w.Hijack()
		b, ok := w.(interface{ Body() []byte })
		if ok {
			b.Body()
		}
	})
	app.AnyFunc("/timeout", func(ctx Context) {
		time.Sleep(time.Millisecond * 30)
		ctx.Write(nil)
		ctx.WriteString("timeout")
		ctx.Response().Push("/", nil)
	})
	app.AnyFunc("/*", func(ctx Context) {
		ctx.WriteString("hello")
		ctx.Write(nil)
		ctx.Response().Push("/", nil)
	})

	form := NewClientBodyForm(nil)
	form.AddValue("name", "eudore")
	form.AddFile("file", "bytes.txt", []byte("file bytes"))

	app.GetRequest("/", NewClientBodyJSON(map[string]any{}))
	app.GetRequest("/", NewClientCheckStatus(200))
	app.GetRequest("/hello", NewClientCheckStatus(200))
	app.GetRequest("/hello", NewClientCheckStatus(200), form)
	app.GetRequest("/panic", NewClientCheckStatus(500), NewClientCheckBody("runtime.gopanic"))
	app.GetRequest("/cancel", NewClientCheckStatus(503))
	app.GetRequest("/timeout", NewClientCheckStatus(503))
	app.GetRequest("/timeout", NewClientCheckStatus(200), NewClientHeader(HeaderAccept, MimeTextEventStream))
	time.Sleep(time.Millisecond * 30)

	app.CancelFunc()
	app.Run()
}
