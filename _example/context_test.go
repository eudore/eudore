package eudore_test

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"

	. "github.com/eudore/eudore"
)

func TestContextRequest(*testing.T) {
	app := NewApp()
	app.SetValue(ContextKeyContextPool, NewContextBasePool(app))

	app.AnyFunc("/*", func(ctx Context) {
		ctx.Info(ctx.Method(), ctx.Path())
	})
	app.GetFunc("/rectx", func(ctx Context) {
		ctx.SetRequest(ctx.Request().WithContext(context.Background()))
	})
	app.GetFunc("/info", func(ctx Context) {
		ctx.WriteString("host: " + ctx.Host())
		ctx.WriteString("\nmethod: " + ctx.Method())
		ctx.WriteString("\npath: " + ctx.Path())
		ctx.WriteString("\nparams: " + ctx.Params().String())
		ctx.WriteString("\nreal ip: " + ctx.RealIP())
		body, _ := ctx.Body()
		if len(body) > 0 {
			ctx.WriteString("\nbody: " + string(body))
		}
	})
	app.GetFunc("/realip", func(ctx Context) {
		ctx.WriteString("real ip: " + ctx.RealIP())
	})
	type bindData struct {
		Name string
	}
	app.AnyFunc("/bind", func(ctx Context) {
		var data bindData
		ctx.Bind(&data)
	})
	app.Listen(":8088")

	app.NewRequest("GET", "/info")
	app.NewRequest("GET", "/rectx")
	app.NewRequest("GET", "/realip")
	app.NewRequest("GET", "http://localhost:8088/realip")
	app.NewRequest("GET", "/realip", http.Header{HeaderXRealIP: {"47.11.11.11"}})
	app.NewRequest("GET", "/realip", http.Header{HeaderXForwardedFor: {"47.11.11.11"}})
	app.NewRequest("GET", "/bind", NewClientBodyJSON(bindData{"eudore"}))
	app.NewRequest("GET", "/bind",
		http.Header{HeaderContentType: {MimeApplicationJSON}},
		strings.NewReader("eudore"),
	)
	app.NewRequest("POST", "/bind",
		http.Header{HeaderContentType: {"value"}},
		strings.NewReader("eudore"),
	)

	app.CancelFunc()
	app.Run()
}

type bodyError struct{}

func (bodyError) Read([]byte) (int, error) {
	return 0, fmt.Errorf("test read error")
}

func (bodyError) Close() error {
	return nil
}

func TestContextData(*testing.T) {
	app := NewApp()
	app.AddMiddleware(func(ctx Context) {
		r := ctx.Request()
		switch ctx.GetHeader("Debug") {
		case "uri":
			r.URL.RawQuery = "tag=%\007"
		case "cookie":
			r.Header.Add(HeaderCookie, "age=22; =00; tag=\007hs; aa=\"bb\"; ")
		case "body":
			r.Body = bodyError{}
		}
	})
	app.AnyFunc("/body", func(ctx Context) {
		ctx.Body()
	})
	app.AnyFunc("/* version=v0", func(ctx Context) {
		ctx.Body()
		ctx.Info(ctx.Method(), ctx.Path(), ctx.Params().String())
	})
	app.GetFunc("/params version=v0", func(ctx Context) {
		ctx.SetParam("name", "eudore")
		ctx.Info("params", ctx.Params().String(), ctx.GetParam("name"))
	})
	app.AnyFunc("/query", func(ctx Context) {
		ctx.Info("query name", ctx.GetQuery("name"))
	})
	app.AnyFunc("/querys", func(ctx Context) {
		val, err := ctx.Querys()
		ctx.Info("querys", val, err)
	})
	// cookie
	app.AnyFunc("/cookie-set", func(ctx Context) {
		ctx.SetCookie(&CookieSet{
			Name:     "set1",
			Value:    "val1",
			Path:     "/",
			HttpOnly: true,
		})
		ctx.SetCookieValue("set", "eudore", 0)
		ctx.SetCookieValue("name", "eudore", 600)
		ctx.Info("response set-cookie", ctx.Response().Header()[HeaderSetCookie])
	})
	app.AnyFunc("/cookie-get", func(ctx Context) {
		ctx.Info("cookie", ctx.GetHeader(HeaderCookie), ctx.Request().Header)
		ctx.GetCookie("name")
		for _, i := range ctx.Cookies() {
			fmt.Fprintf(ctx, "%s: %s\n", i.Name, i.Value)
		}
	})
	// form
	app.AnyFunc("/form-value", func(ctx Context) {
		ctx.Info("form value name:", ctx.FormValue("name"))
	})
	app.AnyFunc("/form-values", func(ctx Context) {
		val, err := ctx.FormValues()
		ctx.Info("form values:", val, err)
	})
	app.AnyFunc("/form-file", func(ctx Context) {
		ctx.Infof("form file: %#v", ctx.FormFile("file"))
	})
	app.AnyFunc("/form-files", func(ctx Context) {
		ctx.Infof("form values: %#v", ctx.FormFiles())
	})

	app.NewRequest("GET", "/")
	app.NewRequest("GET", "/body")
	app.NewRequest("GET", "/body", strings.NewReader("body"))
	app.NewRequest("GET", "/body", NewClientHeader("Debug", "body"))
	app.NewRequest("GET", "/body", NewClientBodyJSON(map[string]any{
		"name": "eudore", "method": "Context.Body",
	}))
	app.NewRequest("GET", "/params")
	app.NewRequest("GET", "/query?name=eudore&debug=true")
	app.NewRequest("GET", "/querys?name=eudore&debug=true")
	app.NewRequest("PUT", "/query", NewClientHeader("Debug", "uri"))
	app.NewRequest("PUT", "/querys", NewClientHeader("Debug", "uri"))
	app.NewRequest("GET", "/cookie-set")
	app.NewRequest("GET", "/cookie-get",
		http.Header{HeaderCookie: {"age=22;;;"}},
		&Cookie{"age", "22"},
		&Cookie{"name", "eudor"},
		&Cookie{"value", "a, b"},
		&Cookie{"valid", "key\x03invalid"},
	)
	app.NewRequest("GET", "/cookie-get", NewClientHeader("Debug", "cookie"))
	app.NewRequest("GET", "/form-value", NewClientBodyForm(url.Values{"name": {"eudor"}}))
	app.NewRequest("GET", "/form-value", NewClientBodyForm(url.Values{"key": {"eudor"}}))
	app.NewRequest("GET", "/form-values?name=eudore")
	app.NewRequest("GET", "/form-values", NewClientHeader("Debug", "uri"))

	body := NewClientBodyForm(nil)
	body.AddFile("file", "app.txt", strings.NewReader("eudore app"))
	app.NewRequest("GET", "/form-file", body)
	body = NewClientBodyForm(nil)
	body.AddFile("name", "app.txt", strings.NewReader("eudore app"))
	app.NewRequest("GET", "/form-file", body)
	app.NewRequest("GET", "/form-file",
		NewClientHeader(HeaderContentType, MimeText),
		strings.NewReader("body"),
	)
	body = NewClientBodyForm(nil)
	body.AddFile("file", "app.txt", strings.NewReader("eudore app"))
	app.NewRequest("GET", "/form-files", body)
	app.NewRequest("GET", "/form-files")
	app.NewRequest("GET", "/form-files",
		NewClientHeader(HeaderContentType, MimeText),
		strings.NewReader("body"),
	)

	app.NewRequest("GET", "/form-files",
		NewClientHeader("Debug", "body"),
		NewClientBodyForm(url.Values{"key": {"eudor"}}),
	)

	app.NewRequest("GET", "/form-value",
		NewClientHeader("Debug", "body"),
		NewClientHeader(HeaderContentType, MimeApplicationForm),
	)
	app.NewRequest("GET", "/form-value",
		strings.NewReader("name=%\007"),
	)
	app.NewRequest("GET", "/form-value",
		NewClientHeader(HeaderContentType, MimeApplicationForm),
		strings.NewReader("name=%\007"),
	)
	app.NewRequest("GET", "/form-value",
		NewClientHeader(HeaderContentType, MimeMultipartForm),
		strings.NewReader("body"),
	)
	app.NewRequest("GET", "/form-value",
		NewClientHeader(HeaderContentType, MimeMultipartForm+"; boundary=x"),
		strings.NewReader("body"),
	)
	app.NewRequest("GET", "/form-value",
		NewClientHeader(HeaderContentType, MimeText),
		strings.NewReader("body"),
	)

	app.CancelFunc()
	app.Run()
}

type responseError struct {
	headers http.Header
	code    int
}

func (w *responseError) Header() http.Header {
	return w.headers
}

func (w *responseError) Write([]byte) (int, error) {
	return 0, fmt.Errorf("test response Write error")
}

func (w *responseError) WriteString(string) (int, error) {
	return 0, fmt.Errorf("test response Write error")
}

func (w *responseError) WriteStatus(code int) {
	w.code = code
}

func (w *responseError) WriteHeader(code int) {
	w.code = code
}

func (w *responseError) Flush() {}

func (w *responseError) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, fmt.Errorf("test response Hijack error")
}

func (w *responseError) Push(string, *http.PushOptions) error {
	return fmt.Errorf("test response Push error")
}

func (w *responseError) Size() int {
	return 0
}

func (w *responseError) Status() int {
	return w.code
}

func TestContextResponse(*testing.T) {
	app := NewApp()
	app.SetValue(ContextKeyContextPool, NewContextBasePool(app))
	app.AddMiddleware(func(ctx Context) {
		if ctx.GetQuery("debug") != "" {
			ctx.SetResponse(&responseError{headers: make(http.Header)})
		}
	})
	app.AnyFunc("/*", func(ctx Context) {
		ctx.Info(ctx.Method(), ctx.Path(), ctx.Params().String())
	})

	app.AnyFunc("/redirect", func(ctx Context) {
		ctx.Redirect(308, "/")
	})
	app.AnyFunc("/redirect/200", func(ctx Context) {
		ctx.Redirect(200, "/")
	})
	app.AnyFunc("/redirect/err", func(ctx Context) {
		ctx.Redirect(308, "/\007")
	})
	app.AnyFunc("/ws", func(ctx Context) {
		conn, _, err := ctx.Response().Hijack()
		if err == nil {
			conn.Close()
		}
	})
	app.AnyFunc("/push", func(ctx Context) {
		ctx.Response().Push("/index.js", nil)
	})
	app.AnyFunc("/write-string", func(ctx Context) {
		ctx.WriteString("hello")
	})
	app.AnyFunc("/write-file", func(ctx Context) {
		ctx.WriteFile("/index.html")
		ctx.WriteFile("context_test.go")
	})
	app.AnyFunc("/render", func(ctx Context) {
		ctx.Render(map[string]interface{}{
			"name":  "eudore",
			"route": ctx.GetParam(ParamRoute),
		})
	})
	app.AnyFunc("/status", func(ctx Context) {
		ctx.WriteStatus(201)
		ctx.WriteStatus(0)
		ctx.WriteHeader(201)
		ctx.WriteString("hello")
		ctx.Response().Status()
	})
	app.AnyFunc("/response", func(ctx Context) {
		ctx.Response().Flush()
		ctx.Response().Hijack()
	})
	app.AnyFunc("/responseWriter", func(ctx Context) {
		ctx.Response().Size()
		unwarper, ok := ctx.Response().(interface{ Unwrap() http.ResponseWriter })
		if ok {
			unwarper.Unwrap()
		}
	})

	app.SetValue(ContextKeyClient, app.NewClient(func(rt http.RoundTripper) {
		tp, ok := rt.(*http.Transport)
		if ok {
			tp.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
	}))
	app.ListenTLS(":8088", "", "")

	app.NewRequest("GET", "/redirect")
	app.NewRequest("GET", "/redirect/200")
	app.NewRequest("GET", "/redirect/err")
	app.NewRequest("GET", "/push")
	app.NewRequest("GET", "/ws")
	app.NewRequest("GET", "/write-string")
	app.NewRequest("GET", "/write-file")
	app.NewRequest("GET", "/render", http.Header{HeaderAccept: {MimeApplicationJSON}})
	app.NewRequest("GET", "/status")
	app.NewRequest("GET", "/response")
	app.NewRequest("GET", "/responseWriter")
	app.NewRequest("GET", "https://localhost:8088/push")
	app.NewRequest("GET", "https://localhost:8088/ws")

	app.Client = app.NewClient(url.Values{"debug": {"1"}})
	app.NewRequest("GET", "/redirect")
	app.NewRequest("GET", "/push")
	app.NewRequest("GET", "/response")
	app.NewRequest("GET", "/write-json")
	app.NewRequest("GET", "/write-string")
	app.NewRequest("GET", "/write-file")
	app.NewRequest("GET", "/render", http.Header{HeaderAccept: {MimeApplicationJSON}})

	app.CancelFunc()
	app.Run()
}

func TestContextLogger(*testing.T) {
	app := NewApp()
	app.SetValue(ContextKeyLogger, NewLogger(&LoggerConfig{
		Caller: true,
	}))

	app.AnyFunc("/*", func(ctx Context) {
		ctx.Debugf("debug")
		ctx.Info("hello")
		ctx.Infof("hello path is %s", ctx.GetParam("*"))
		ctx.Warning("warning")
		ctx.Warningf("warningf")
		ctx.Error(nil)
		ctx.Error(&http.MaxBytesError{})
		ctx.Error("test error")
		ctx.Error("test error", 1)
		ctx.Errorf("test error")
		ctx.Fatal(nil)
	})
	app.AnyFunc("/field", func(ctx Context) {
		ctx.WithFields([]string{"key", "name"}, []interface{}{"ctx.WithFields", "eudore"}).Debug("hello fields")
		ctx.WithField("logger", true).Debug("hello empty fields")
		ctx.WithField("key", "test-firle").Debug("debug")
		ctx.WithField("key", "test-firle").Debugf("debugf")
		ctx.WithField("key", "test-firle").Info("hello")
		ctx.WithField("key", "test-firle").Infof("hello path is %s", ctx.GetParam("*"))
		ctx.WithField("key", "test-firle").Warning("warning")
		ctx.WithField("key", "test-firle").Warningf("warningf")
		ctx.WithField("key", "test-firle").Error(fmt.Errorf("test err"))
		ctx.WithField("key", "test-firle").Error(nil)
		ctx.WithField("key", "test-firle").Errorf("test error")
		ctx.WithField("key", "test-firle").Fatal(fmt.Errorf("test err"))
		ctx.WithField("key", "test-firle").Fatal(nil)
		ctx.WithField("key", "test-firle").WithField("hello", "haha").Fatalf("fatal logger method: %s path: %s", ctx.Method(), ctx.Path())
		ctx.WithField("method", "WithField").WithFields([]string{"key", "name"}, []interface{}{"ss", "eudore"}).Debug("hello fields")
	})
	app.AnyFunc("/ctx", func(ctx Context) {
		ctx.SetContext(context.WithValue(context.Background(), ContextKeyLogger, DefaultLoggerNull))
		ctx.Debug(1)
		ctx.SetContext(context.Background())
		ctx.Debug(2)
	})
	app.AnyFunc("/debug", func(ctx Context) {
		log := ctx.Value(ContextKeyLogger).(Logger)
		log.SetLevel(LoggerInfo)

		ctx.Request().ContentLength = -1
		ctx.Request().Body = bodyError{}
		ctx.Body()
	})
	app.AnyFunc("/err1", func(ctx Context) {
		ctx.Fatalf("fatal logger method: %s path: %s", ctx.Method(), ctx.Path())
		ctx.Debug("err:", ctx.Err())
	})
	app.AnyFunc("/err2", func(ctx Context) {
		ctx.Fatal(NewErrorWithStatusCode(fmt.Errorf("test error"), 432, 10032))
	})
	app.AnyFunc("/err3", func(ctx Context) {
		NewErrorWithStatus(fmt.Errorf("test error"), 0)
		ctx.Fatal(NewErrorWithStatus(fmt.Errorf("test error"), 432))
	})
	app.AnyFunc("/err4", func(ctx Context) {
		NewErrorWithCode(fmt.Errorf("test error"), 0)
		ctx.Fatal(NewErrorWithCode(fmt.Errorf("test error"), 10032))
	})
	app.AnyFunc("/err4", func(ctx Context) {
		ctx.Fatal(&http.MaxBytesError{})
	})
	NewErrorWithStatusCode(nil, 0, 0)
	NewErrorWithStatus(nil, 0)
	NewErrorWithCode(nil, 0)

	app.NewRequest("GET", "/index")
	app.NewRequest("GET", "/field")
	app.NewRequest("GET", "/ctx")
	app.NewRequest("GET", "/debug")
	app.NewRequest("GET", "/err1")
	app.NewRequest("GET", "/err2")
	app.NewRequest("GET", "/err3")
	app.NewRequest("GET", "/err4")
	app.NewRequest("GET", "/err5")

	app.CancelFunc()
	app.Run()
}

func TestContextValues(*testing.T) {
	app := NewApp()
	NewContextBaseFunc(app)()
	app.AnyFunc("/index", func(ctx Context) {
		ctx.GetHandlers()
	})
	app.AnyFunc("/ctx", func(ctx Context) {
		ctx.SetValue(ContextKeyLogger, ctx.Value(ContextKeyLogger))
		ctx.SetValue(ContextKeyClient, ctx.Value(ContextKeyClient))

		c := ctx.Context()
		ctx.SetValue(1, true)
		ctx.SetValue(1, true)
		ctx.Value(1)
		ctx.Value(2)
		ctx.Debug(c)
		ctx.Err()
		ctx.Fatal("err")
		ctx.Err()
		ctx.Debug(c)

		ctx.SetContext(context.Background())
		ctx.Context()
		ctx.SetValue(1, true)
	})

	app.NewRequest("GET", "/index")
	app.NewRequest("GET", "/ctx")

	app.CancelFunc()
	app.Run()
}
