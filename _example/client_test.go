package eudore_test

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
	"golang.org/x/net/http2"
)

func TestClientOptions(t *testing.T) {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyClient, app.WithClient(
		&eudore.ClientOption{
			Values: url.Values{"debug": {"1"}},
			Header: http.Header{"X-Client": {"eudore"}},
			Trace:  &eudore.ClientTrace{},
		},
		eudore.NewClientOptionUserAgent("Client-Eudore"),
		eudore.NewClientOptionHost("eudore.cn"),
		eudore.Cookie{Name: "name", Value: "eudore"},
	))

	{
		app.GetFunc("/jar", middleware.NewLoggerFunc(app), func(ctx eudore.Context) {
			ctx.SetCookieValue("count", ctx.GetCookie("count")+"1", 0)
			ctx.Debug(ctx.GetCookie("count"))
		})
		jar, _ := cookiejar.New(nil)
		client := app.WithClient(jar)
		for i := 0; i < 5; i++ {
			client.NewRequest(nil, "", "/jar")
		}
	}
	{
		app.GetFunc("/timeout", func(ctx eudore.Context) {
			time.Sleep(time.Microsecond * 2)
		})
		client := app.WithClient(time.Microsecond)
		app.Info(client.NewRequest(nil, "", "/timeout"))
	}
	{
		app.ListenTLS(":8089", "", "")
		client := app.WithClient(&http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		})
		app.Debug(app.NewRequest(nil, "", "https://localhost:8089/app"))

		trace := &eudore.ClientTrace{}
		app.WithField("trace", trace).Debug(client.NewRequest(nil, "", "https://localhost:8089/client", trace))
	}

	{
		app.GetClient().Get("/")
	}

	{
		eudore.NewClient().WithClient(
			&http2.Transport{
				AllowHTTP: true,
				DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
					return tls.Dial(network, addr, cfg)
				},
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			&http.Transport{},
			http.DefaultClient,
		)
	}

	app.CancelFunc()
	app.Run()
}

func TestClientRequest(t *testing.T) {
	eudore.DefaultClinetLoggerLevel = eudore.LoggerDebug
	defer func() {
		eudore.DefaultClinetLoggerLevel = eudore.LoggerError
	}()

	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewRequestIDFunc(nil))
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Info("server body:", string(ctx.Body()))
		ctx.Write(ctx.Body())
	})
	app.AnyFunc("/ctx", func(ctx eudore.Context) {
		client := ctx.Value(eudore.ContextKeyClient).(eudore.Client).WithClient(
			http.Header{
				eudore.HeaderAuthorization: {ctx.GetHeader(eudore.HeaderAuthorization)},
			},
		)
		ctx.SetValue(eudore.ContextKeyClient, client)
		ctx.NewRequest("GET", "/")
	})

	client := app.WithClient(&http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	})

	client.NewRequest(nil, "GET", "/",
		strings.NewReader("eudore body"),
		context.Background(),
		time.Second,
		url.Values{"name": {"eudore"}},
		http.Header{"Cookie": {"name=eudore-client"}},
		&http.Cookie{Name: "name", Value: "eudore"},
		eudore.Cookie{Name: "name", Value: "eudore"},
		eudore.Cookie{Name: "key1", Value: "key ,space"},
		eudore.Cookie{Name: "key2", Value: "key\x03invalid"},
		&eudore.ClientTrace{},
		&eudore.ClientOption{},
	)
	app.Debug(app.NewRequest(nil, "GET", "/ctx"))
	app.Debug(app.NewRequest(nil, "GET", "/ctx\x00"))
	app.Debug(app.NewRequest(nil, "LOCK", "/ctx"))

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

	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app))
	app.GetFunc("/500", func(ctx eudore.Context) {
		ctx.WriteHeader(eudore.StatusInternalServerError)
	})
	app.GetFunc("/auth", func(ctx eudore.Context) {
		ctx.Debug(ctx.GetHeader(eudore.HeaderAuthorization))
	})
	app.GetFunc("/digest", func(ctx eudore.Context) {
		if ctx.GetHeader(eudore.HeaderAuthorization) == "" {
			ctx.WriteHeader(eudore.StatusUnauthorized)
			ctx.SetHeader(eudore.HeaderWWWAuthenticate, digest[eudore.GetAnyByString(ctx.GetQuery("d"), 0)])
		} else {
			ctx.Debug(ctx.GetHeader(eudore.HeaderAuthorization))
		}
	})

	app.NewRequest(nil, "", "/auth", eudore.NewClientOptionBearer("Bearer .eyJ1c2VyX25hbWUiOiJHdWVzdCIsImV4cGlyYXRpb24iOjEwNDEzNzkyMDAwfQ.vNTXrJNVqRLLY01w6weQWMRo_HDeBeVpX4HZtVfYUBY"))
	app.NewRequest(nil, "", "/auth", eudore.NewClientOptionBasicauth("Guest", ""))

	client := app.WithClient(eudore.NewClientRetryDigest("Guest", "Guest"))
	for i := range digest {
		client.NewRequest(nil, "", "/digest", url.Values{"d": {fmt.Sprint(i)}})
	}
	client.NewRequest(nil, "", "/digest?d=1", strings.NewReader("digest body"))
	client.NewRequest(nil, "", "/500")
	client.NewRequest(nil, "", "/digest?d=1", io.NopCloser(strings.NewReader("digest body")))

	form := eudore.NewClientBodyForm(nil)
	form.AddFile("file", "name", strings.NewReader("file bodt"))
	client.NewRequest(nil, "", "/digest?d=1", form)

	app.CancelFunc()
	app.Run()
}

func TestClientRetry(t *testing.T) {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app))
	app.GetFunc("/502", func(ctx eudore.Context) {
		ctx.WriteHeader(eudore.StatusBadGateway)
	})
	client := app.WithClient(eudore.NewClientRetryNetwork(1))

	client.NewRequest(nil, "", "/502", time.Second/10)
	app.CancelFunc()
	app.Run()
}

func TestClientRequestBody(t *testing.T) {
	type Body struct {
		Name string
	}
	app := eudore.NewApp()
	app.AnyFunc("/redirect", func(ctx eudore.Context) {
		ctx.Body()
		ctx.Redirect(308, "/")
	})
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Info("server body:", string(ctx.Body()))
		ctx.Write(ctx.Body())
	})

	app.Debug(app.NewRequest(nil, "GET", "/body/string", strings.NewReader("eudore body string")))
	app.Debug(app.NewRequest(nil, "GET", "/body/json", eudore.NewClientBodyJSON(nil)))
	app.Debug(app.NewRequest(nil, "GET", "/body/jsonstruct", eudore.NewClientBodyJSON(Body{"eudor"})))
	app.Debug(app.NewRequest(nil, "GET", "/body/jsonmap", eudore.NewClientBodyJSON(map[string]interface{}{"name": "eudore"})))
	app.Debug(app.NewRequest(nil, "GET", "/body/xml", eudore.NewClientBodyXML(Body{"eudor"})))
	app.Debug(app.NewRequest(nil, "GET", "/body/protobuf", eudore.NewClientBodyProtobuf(Body{"eudor"})))
	app.Debug(app.NewRequest(nil, "GET", "/body/form", eudore.NewClientBodyForm(url.Values{
		"name": {"eudore"},
	})))
	bodyForm := eudore.NewClientBodyForm(nil)
	bodyForm.AddValue("name", "eudore")
	bodyForm.AddFile("file", "bytes.txt", []byte("file bytes"))
	bodyForm.AddFile("file", "buffer.txt", bytes.NewBufferString("file buffer"))
	bodyForm.AddFile("file", "rc.txt", io.NopCloser(bytes.NewBufferString("file rc")))
	bodyForm.AddFile("file", "none.txt", nil)
	bodyForm.AddFile("file", "", "appNew.go")
	bodyForm.Close()
	app.Debug(app.NewRequest(nil, "GET", "/body/formfile", bodyForm))

	bodyForm = eudore.NewClientBodyForm(nil)
	bodyForm.AddValue("name", "eudore")
	bodyJSON := eudore.NewClientBodyJSON(nil)
	bodyJSON.AddValue("name", "eudore")
	bodyJSON.AddFile("file", "", "appNew.go")
	bodyJSON = eudore.NewClientBodyJSON(&Body{})
	bodyJSON.AddValue("name", "eudore")
	bodyJSON.Close()

	app.Debug(app.NewRequest(nil, "PUT", "/redirect", strings.NewReader("eudore body string")))
	app.Debug(app.NewRequest(nil, "PUT", "/redirect", bodyForm))
	app.Debug(app.NewRequest(nil, "PUT", "/redirect", bodyJSON))

	app.CancelFunc()
	app.Run()
}

func TestClientResponse(t *testing.T) {
	app := eudore.NewApp()
	app.GetFunc("/body/*", func(eudore.Context) interface{} {
		return eudore.MetadataConfig{Name: "config"}
	})
	app.GetFunc("/proxy", func(ctx eudore.Context) {
		ctx.NewRequest(ctx.Method(), "/body",
			eudore.NewClientOptionHeader(eudore.HeaderAccept, ctx.GetHeader(eudore.HeaderAccept)),
			eudore.NewClientOptionHeader(eudore.HeaderAcceptEncoding, ctx.GetHeader(eudore.HeaderAcceptEncoding)),
			eudore.NewClienProxyWriter(ctx.Response()),
		)
	})
	app.GetFunc("/err", func(eudore.Context) error {
		return fmt.Errorf("test err")
	})

	// app.NewRequest(nil, "GET", "https://goproxy.cn", time.Second, &eudore.ClientTrace{})
	// app.NewRequest(nil, "GET", "https://golang.org", time.Second, &eudore.ClientTrace{})
	// app.NewRequest(context.Background(), "GET", "/", time.Second)

	app.NewRequest(nil, "GET", "/check/status", eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "GET", "/check/status", eudore.NewClientCheckStatus(404))

	for _, accept := range []string{eudore.MimeApplicationJSON, eudore.MimeApplicationXML, eudore.MimeApplicationProtobuf, eudore.MimeTextHTML} {
		conf := &eudore.MetadataConfig{}
		app.NewRequest(nil, "GET", "/body/proxy",
			eudore.NewClientOptionHeader(eudore.HeaderAccept, accept),
			eudore.NewClientParse(conf),
		)
		app.Debugf("%#v", conf)
	}

	clientBodyError := func(w *http.Response) error {
		w.Body = &responseBody{}
		return nil
	}
	var str string
	app.NewRequest(nil, "GET", "/body/parse", eudore.NewClientParse(&str))
	app.NewRequest(nil, "GET", "/body/parse", clientBodyError, eudore.NewClientParse(&str))
	app.NewRequest(nil, "GET", "/body/parseif", eudore.NewClientParseIf(201, nil))
	app.NewRequest(nil, "GET", "/body/parsein", eudore.NewClientParseIn(300, 308, nil))

	app.NewRequest(nil, "GET", "/body/parsestr", eudore.NewClientParseErr(), eudore.NewClientParse(&str))
	app.NewRequest(nil, "GET", "/err", eudore.NewClientParseErr(), eudore.NewClientParse(&str))
	app.NewRequest(nil, "GET", "/err", clientBodyError, eudore.NewClientParseErr())

	app.NewRequest(nil, "GET", "/body/check", eudore.NewClientCheckBody("config"))
	app.NewRequest(nil, "GET", "/body/check", eudore.NewClientCheckBody("123456"))
	app.NewRequest(nil, "GET", "/body/check", clientBodyError, eudore.NewClientCheckBody(""))

	app.NewRequest(nil, "GET", "/proxy")

	app.CancelFunc()
	app.Run()
}

type responseBody struct{}

func (r *responseBody) Read(p []byte) (int, error) {
	return 0, fmt.Errorf("test error")
}

func (r *responseBody) Close() error {
	return nil
}
