package eudore_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/xml"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/eudore/eudore"
	. "github.com/eudore/eudore/middleware"
	"golang.org/x/net/http2"
)

func TestClientOptions(t *testing.T) {
	app := NewApp()
	app.SetValue(ContextKeyClient, NewClient(
		NewClientOptionURL(""),
		NewClientOptionURL("/\n"),
		NewClientOptionURL("/api"),
		NewClientOptionURL("/user"),
		NewClientQuery("name", ""),
		NewClientQuery("name", "eudore"),
		url.Values{"debug": {"1"}},
		NewClientOptionHost("eudore.cn"),
		NewClientOptionUserAgent("Client-Eudore"),
		&Cookie{Name: "name", Value: "\"eudore ,\007"},
		&CookieSet{Name: "name", Value: "eudore"},
		http.Header{"X-Client": {"eudore"}},
		NewClientOption(nil),
		&ClientOption{
			Trace: &ClientTrace{},
		},
		&ClientTrace{},
		[]any{},
	))
	app.NewClient(&http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
			return tls.Dial(network, addr, cfg)
		},
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	})
	app.GetRequest("/?a=1")
	app.GetRequest("/\007")
	app.GetRequest("/", context.Background())
	app.GetRequest("/", NewClientHookRetry(1, nil, nil))
	app.GetRequest("/", NewClientHookTimeout(time.Millisecond*100))
	app.DeleteRequest("/")
	app.HeadRequest("/")
	app.PatchRequest("/")

	app.ListenTLS(":8088", "", "")
	client := app.NewClient(func(rt http.RoundTripper) {
		tp, ok := rt.(*http.Transport)
		if ok {
			tp.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
	})
	app.Debug(app.GetRequest("https://localhost:8088/app"))

	trace := &ClientTrace{}
	app.WithField("trace", trace).Debug(client.GetRequest("https://localhost:8088/client", trace))

	func() {
		defer func() { recover() }()
		NewClientOption([]any{0})
	}()

	app.CancelFunc()
	app.Run()
}

func TestClientHooks(t *testing.T) {
	retryInterval := []time.Duration{
		time.Millisecond, 0,
		2 * time.Millisecond, time.Millisecond,
	}
	loggernull := NewLoggerNull()
	loggernull.SetLevel(LoggerDebug)
	jar, _ := cookiejar.New(nil)

	app := NewApp()
	app.SetValue(ContextKeyClient, NewClientCustom(
		NewClientHookCookie(jar),
		NewClientHookTimeout(time.Millisecond*100),
		NewClientHookRedirect(nil),
		NewClientHookRetry(3, nil, nil),
		NewClientHookLogger(LoggerInfo, time.Millisecond*20),
		NewClientHookDigest("user", "Guest"),
	))
	app.Client.(interface{ Metadata() interface{} }).Metadata()

	app.AddMiddleware(NewRequestIDFunc(nil))
	app.AddHandler("404", "", HandlerRouter404)
	app.GetFunc("/time/slow", func(ctx Context) {
		time.Sleep(time.Millisecond * 40)
		ctx.WriteString("eudore")
	})
	app.GetFunc("/time/timeout", func() {
		time.Sleep(time.Second)
	})
	app.AnyFunc("/status/308", func(ctx Context) {
		ctx.WriteHeader(308)
	})
	app.AnyFunc("/status/503", func(ctx Context) {
		ctx.WriteHeader(503)
		ctx.WriteString("503")
	})
	app.GetFunc("/body", func(ctx Context) {
		ctx.WriteHeader(200)
		for i := 0; i < 1000; i++ {
			_, err := ctx.WriteString("eudore")
			if err != nil {
				return
			}
		}
	})
	app.GetFunc("/cookie", func(ctx Context) {
		ctx.SetCookieValue("value", ctx.GetCookie("value")+"0", 0)
		ctx.WriteString("eudore")
	})

	for _, i := range []int{301, 302, 307, 308} {
		code := i
		app.AnyFunc("/redirect/"+strconv.Itoa(code), func(ctx Context) {
			ctx.Redirect(code, "/")
		})
	}
	app.AnyFunc("/loc/err", func(ctx Context) {
		ctx.SetHeader(HeaderLocation, ":/")
		ctx.WriteHeader(308)
	})

	app.SetValue(ContextKeyClient, NewClientCustom(
		NewClientHookLogger(LoggerInfo, time.Millisecond*20),
		&ClientTrace{},
		context.WithValue(app, ContextKeyLogger, loggernull),
	))
	app.GetRequest("/log")
	app.GetRequest("/time/slow")
	app.GetRequest("/log", bytes.NewBufferString("eudore"),
		NewClientHookLogger(
			LoggerDebug, time.Millisecond*10,
			"scheme", "query", "byte-in", "byte-out", "x-response-id",
			"request-header", "response-header", "trace",
		),
	)
	app.GetRequest("/log", context.WithValue(app,
		ContextKeyLogger, DefaultLoggerNull,
	))
	app.GetRequest("/log", func(req *http.Request) {
		req.URL.Scheme = ""
	})

	app.SetValue(ContextKeyClient, NewClientCustom(
		NewClientHookCookie(nil),
	))
	app.GetRequest("/cookie")
	app.GetRequest("/cookie", NewClientCheckBody("eudore"))

	app.SetValue(ContextKeyClient, NewClientCustom(
		NewClientHookTimeout(time.Millisecond*100),
	))
	app.GetRequest("/slow")
	app.GetRequest("/timeout")
	app.GetRequest("/body", NewClientHookTimeout(-time.Millisecond*10))
	NewClientCustom(app,
		&clientHookBody{},
		NewClientHookTimeout(time.Millisecond*10),
	).GetRequest("/body")
	app.GetRequest("/body", NewClientHookTimeout(time.Millisecond*10),
		func(resp *http.Response) error {
			io.ReadAll(resp.Body)
			return nil
		},
	)
	app.GetRequest("/body", NewClientHookTimeout(time.Millisecond*10),
		func(resp *http.Response) error {
			time.Sleep(time.Millisecond * 20)
			io.ReadAll(resp.Body)
			return nil
		},
	)

	app.SetValue(ContextKeyClient, NewClientCustom(
		NewClientHookRedirect(func(req *http.Request, via []*http.Request) error {
			switch req.Response.StatusCode {
			case 301:
				return http.ErrNoLocation
			case 307:
				return http.ErrUseLastResponse
			}
			return nil
		}),
	))
	app.PostRequest("/redirect/301")
	app.PostRequest("/redirect/301", func(req *http.Request) {
		req.URL.User = url.UserPassword("eudore", "pass")
	})
	app.PostRequest("/redirect/302", func(req *http.Request) {
		req.URL.User = url.UserPassword("eudore", "pass")
	})
	app.PostRequest("/redirect/302", NewClientHookRedirect(nil))
	app.GetRequest("/redirect/307")
	app.GetRequest("/redirect/308")
	app.PostRequest("/redirect/308", strings.NewReader("string"), NewClientOptionHost("eudore.cn"))
	app.PostRequest("/redirect/308", strings.NewReader("string"), func(req *http.Request) {
		req.GetBody = nil
	})
	app.PostRequest("/redirect/308", strings.NewReader("string"), func(req *http.Request) {
		req.GetBody = func() (io.ReadCloser, error) {
			return nil, io.EOF
		}
	})
	app.GetRequest("/loc/err", func(resp *http.Response) error {
		fmt.Println(resp.Header.Get(HeaderLocation))
		return nil
	})
	app.GetRequest("/status/308")
	app.GetRequest("httpss:///")

	app.SetValue(ContextKeyClient, NewClientCustom(
		NewClientHookRetry(3, retryInterval, nil),
	))
	app.GetRequest("/status/308")
	app.GetRequest("/status/503")
	app.GetRequest("/status/503", strings.NewReader("body"))
	app.GetRequest("/status/503", strings.NewReader("body"),
		func(req *http.Request) {
			req.GetBody = func() (io.ReadCloser, error) {
				return nil, io.EOF
			}
		},
	)
	app.GetRequest("/status/503", strings.NewReader("body"),
		func(req *http.Request) { req.GetBody = nil },
	)

	app.CancelFunc()
	app.Run()
}

type clientHookBody struct {
	next http.RoundTripper
}

func (*clientHookBody) Name() string {
	return "empty"
}

func (*clientHookBody) Wrap(rt http.RoundTripper) http.RoundTripper {
	return &clientHookBody{rt}
}

func (hook *clientHookBody) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := hook.next.RoundTrip(req)
	if resp != nil {
		resp.Body = nil
	}
	return resp, err
}

func TestClientBody(t *testing.T) {
	type Body struct {
		Name string
	}
	app := NewApp()
	app.SetValue(ContextKeyClient, app.NewClient(NewClientHookLogger(LoggerError, time.Millisecond*20)))
	app.AnyFunc("/redirect", func(ctx Context) {
		ctx.Body()
		ctx.Redirect(308, "/")
	})
	app.AnyFunc("/*", func(ctx Context) {
		io.Copy(ctx, ctx)
	})

	app.GetRequest("/body/string", strings.NewReader("eudore body string"))
	app.GetRequest("/body/json", NewClientBodyJSON(nil))
	app.GetRequest("/body/jsonstruct", NewClientBodyJSON(Body{"eudor"}))
	app.GetRequest("/body/jsonmap", NewClientBodyJSON(map[string]interface{}{"name": "eudore"}))
	app.GetRequest("/body/xml", NewClientBodyXML(Body{"eudor"}))
	app.GetRequest("/body/form", NewClientBodyForm(url.Values{
		"name": {"eudore"},
	}))
	bodyForm := NewClientBodyForm(nil)
	bodyForm.AddValue("name", "eudore")
	bodyForm.AddFile("file", "bytes.txt", []byte("file bytes"))
	bodyForm.AddFile("file", "buffer.txt", bytes.NewBufferString("file buffer"))
	bodyForm.AddFile("file", "rc.txt", io.NopCloser(bytes.NewBufferString("file rc")))
	bodyForm.AddFile("file", "none.txt", nil)
	bodyForm.AddFile("file", "", "appNew.go")
	bodyForm.Close()
	app.GetRequest("/body/formfile", bodyForm)

	bodyForm = NewClientBodyForm(nil)
	bodyForm.AddValue("name", "eudore")
	bodyJSON := NewClientBodyJSON(nil)
	bodyJSON.AddValue("name", "eudore")
	bodyJSON.AddFile("file", "", "appNew.go")
	bodyJSON = NewClientBodyJSON(&Body{})
	bodyJSON.AddValue("name", "eudore")
	bodyJSON.Close()

	file, err := os.Open("README.md")
	if err == nil {
		bodyfile := NewClientBodyFile("", file)
		bodyfile.AddValue("", "")
		bodyfile.AddFile("", "", nil)
		app.PutRequest("/redirect", bodyfile)
	}

	app.PutRequest("/redirect", strings.NewReader("eudore body string"))
	app.PutRequest("/redirect", bodyForm)
	app.PutRequest("/redirect", bodyJSON)

	app.CancelFunc()
	app.Run()
}

func TestClientAuthorization(t *testing.T) {
	digest := []string{
		`Digest realm="digest@eudore.cn", algorithm=MD5, nonce="H4GiTo0v", qop="auth, auth-int", opaque="CUYo5tdS"`,
		`Digest realm="digest@eudore.cn", algorithm=MD5, nonce="H4GiTo0v", opaque="CUYo5tdS",qop="auth-int"`,
		`Digest realm="digest@eudore.cn", algorithm=MD5, nonce="H4GiTo0v"`,
		`Digest realm="digest@eudore.cn", algorithm=MD5-SESS, nonce="H4GiTo0v"`,
		`Digest realm="digest@eudore.cn", algorithm=SHA-256, nonce="H4GiTo0v"`,
		`Basic  realm="digest@eudore.cn", algorithm=MD5, nonce="H4GiTo0v", opaque="CUYo5tdS"`,
		`Digest realm="digest@eudore.cn", algorithm, nonce="H4GiTo0v", opaque="CUYo5tdS"`,
		`Digest realm="digest@eudore.cn", algorithm=MD5, nonce="H4GiTo0v", cnonce="H4GiTo0v", opaque="CUYo5tdS"`,
		`Digest realm="digest@eudore.cn", algorithm=RCR32, nonce="H4GiTo0v", opaque="CUYo5tdS"`,
		`Digest realm="digest@eudore.cn", algorithm=MD5, nonce="H4GiTo0v", opaque="CUYo5tdS", qop=int`,
	}

	app := NewApp()
	app.GetFunc("/500", func(ctx Context) {
		ctx.WriteHeader(StatusInternalServerError)
	})
	app.GetFunc("/auth", func(ctx Context) {
		ctx.Debug(ctx.GetHeader(HeaderAuthorization))
	})
	app.GetFunc("/digest", func(ctx Context) {
		if ctx.GetHeader(HeaderAuthorization) == "" {
			ctx.SetHeader(HeaderWWWAuthenticate, digest[GetAnyByString(ctx.GetQuery("d"), 0)])
			ctx.WriteHeader(StatusUnauthorized)
		} else {
			ctx.Debug(ctx.GetHeader(HeaderAuthorization))
		}
	})

	app.GetRequest("/auth", NewClientOptionBearer("Bearer .eyJ1c2VyX25hbWUiOiJHdWVzdCIsImV4cGlyYXRpb24iOjEwNDEzNzkyMDAwfQ.vNTXrJNVqRLLY01w6weQWMRo_HDeBeVpX4HZtVfYUBY"))
	app.GetRequest("/auth", NewClientOptionBasicauth("Guest", ""))

	client := app.NewClient(NewClientHookDigest("Guest", "Guest"))
	for i := range digest {
		client.GetRequest("/digest", url.Values{"d": {fmt.Sprint(i)}})
	}
	client.GetRequest("/digest?d=1", strings.NewReader("digest body"))
	client.GetRequest("/digest?d=1", strings.NewReader("digest body"),
		func(req *http.Request) { req.GetBody = nil },
	)
	client.GetRequest("/500")
	client.GetRequest("/digest?d=1", io.NopCloser(strings.NewReader("digest body")))

	form := NewClientBodyForm(nil)
	form.AddFile("file", "name", strings.NewReader("file bodt"))
	client.GetRequest("/digest?d=1", form)

	app.CancelFunc()
	app.Run()
}

func TestClientResponse(t *testing.T) {
	app := NewApp()
	app.SetValue(ContextKeyClient, app.NewClient(NewClientHookLogger(LoggerError, time.Millisecond*20)))
	app.SetValue(ContextKeyRender, NewHandlerDataRenders(map[string]HandlerDataFunc{
		MimeAll:             HandlerDataRenderJSON,
		MimeText:            HandlerDataRenderText,
		MimeTextPlain:       HandlerDataRenderText,
		MimeTextHTML:        NewHandlerDataRenderTemplates(nil, nil),
		MimeApplicationJSON: HandlerDataRenderJSON,
		MimeApplicationXML: func(ctx Context, data any) error {
			ctx.SetHeader(HeaderContentType, MimeApplicationXML)
			return xml.NewEncoder(ctx).Encode(data)
		},
	}))
	app.SetValue(ContextKeyContextPool, NewContextBasePool(app))
	app.GetFunc("/body/*", func(Context) interface{} {
		return MetadataConfig{Name: "config"}
	})
	app.GetFunc("/err", func(ctx Context) error {
		return fmt.Errorf("test err")
	})
	app.GetFunc("/errcode", func(Context) error {
		return NewErrorWithCode(fmt.Errorf("test code err"), 10005)
	})
	app.GetFunc("/longbody", func(ctx Context) {
		for i := 0; i < 50; i++ {
			ctx.WriteString("0123456789")
		}
	})

	// app.GetRequest( "https://goproxy.cn", time.Second, &eudore.ClientTrace{})
	// app.GetRequest( "https://golang.org", time.Second, &eudore.ClientTrace{})
	app.GetRequest("/check/status", NewClientCheckStatus(200))
	app.GetRequest("/check/status", NewClientCheckStatus(404))

	for _, accept := range []string{MimeApplicationJSON, MimeApplicationXML, MimeApplicationProtobuf, MimeTextHTML} {
		conf := &MetadataConfig{}
		app.GetRequest("/body/proxy",
			NewClientHeader(HeaderAccept, accept),
			NewClientHeader(HeaderAccept, ""),
			NewClientParse(conf),
		)
	}

	clientBodyError := func(w *http.Response) error {
		w.Body = &responseBody{}
		return nil
	}
	var str string
	app.GetRequest("/body/parse", NewClientParse(&str))
	app.GetRequest("/body/parse", NewClientParse(&responseBody{}))
	app.GetRequest("/body/parse", clientBodyError, NewClientParse(&str))
	app.GetRequest("/body/parseif", NewClientParseIf(201, nil))
	app.GetRequest("/body/parsein", NewClientParseIn(300, 308, nil))
	app.GetRequest("/body/parsestr", NewClientParseErr(), NewClientParse(&str))
	app.GetRequest("/body/check", NewClientCheckBody("config"))
	app.GetRequest("/body/check", NewClientCheckBody("123456"))
	app.GetRequest("/body/check", clientBodyError, NewClientCheckBody(""))

	app.GetRequest("/err", NewClientParseErr(), NewClientParse(&str))
	app.GetRequest("/err", clientBodyError, NewClientParseErr())
	app.GetRequest("/errcode", NewClientParseErr())
	app.GetRequest("/longbody", NewClientParseIf(200, nil))
	app.GetRequest("/longbody", NewClientCheckBody("eudore"))

	app.CancelFunc()
	app.Run()
}

type responseBody struct{}

func (r *responseBody) Write(p []byte) (int, error) {
	return 0, nil
}

func (r *responseBody) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("test error")
}

func (r *responseBody) Close() error {
	return nil
}
