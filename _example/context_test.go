package eudore_test

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
	"github.com/google/uuid"
)

func TestContext(*testing.T) {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewRequestIDFunc(nil), middleware.NewRecoverFunc())
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Info(ctx.Method(), ctx.Path(), string(ctx.Body()))
	})
	app.AnyFunc("/context", func(ctx eudore.Context) {
		ctx.SetRequest(ctx.Request().WithContext(ctx.GetContext()))
		ctx.Info(ctx.GetHandler())
		ctx.SetValue("x-meta", "da01b314-0d7c-46ed-a086-2835b31d9133")
		ctx.Fatal("test error")
		ctx.Infof("context: %s", ctx.GetContext())
	})
	eudore.NewContextBaseFunc(app)()
	newContxt()

	app.NewRequest(nil, "GET", "/context")


	app.CancelFunc()
	app.Run()
}

func newContxt() {
	defer func() {
		recover()
	}()
	eudore.NewContextBasePool(context.Background())

}

func TestContextRequest(*testing.T) {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyValidater, func(eudore.Context, interface{}) error { return nil })
	app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))

	app.AddMiddleware("global", middleware.NewRequestIDFunc(nil))
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Info(ctx.Method(), ctx.Path(), string(ctx.Body()))
	})
	app.GetFunc("/hello", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore" + ctx.Host())
	})
	app.GetFunc("/info", func(ctx eudore.Context) {
		ctx.WriteString("host: " + ctx.Host())
		ctx.WriteString("\nmethod: " + ctx.Method())
		ctx.WriteString("\npath: " + ctx.Path())
		ctx.WriteString("\nparams: " + ctx.Params().String())
		ctx.WriteString("\ncontext type: " + ctx.ContentType())
		ctx.WriteString("\nreal ip: " + ctx.RealIP())
		ctx.WriteString("\nrequest id: " + ctx.RequestID())
		ctx.WriteString("\nistls: " + fmt.Sprint(ctx.Istls()))
		body := ctx.Body()
		if len(body) > 0 {
			ctx.WriteString("\nbody: " + string(body))
		}
	})
	app.GetFunc("/realip", func(ctx eudore.Context) {
		ctx.WriteString("real ip: " + ctx.RealIP())
	})
	type bindData struct {
		Name string
	}
	app.AnyFunc("/bind", func(ctx eudore.Context) {
		var data bindData
		ctx.Bind(&data)
	})
	app.Listen(":8088")

	app.NewRequest(nil, "GET", "/info")
	app.NewRequest(nil, "GET", "/realip")
	app.NewRequest(nil, "GET", "http://localhost:8088/realip")
	app.NewRequest(nil, "GET", "/realip", http.Header{eudore.HeaderXRealIP: {"47.11.11.11"}})
	app.NewRequest(nil, "GET", "/realip", http.Header{eudore.HeaderXForwardedFor: {"47.11.11.11"}})
	app.NewRequest(nil, "GET", "/bind", eudore.NewClientBodyJSON(bindData{"eudore"}))
	app.NewRequest(nil, "GET", "/bind",
		http.Header{eudore.HeaderContentType: {eudore.MimeApplicationJSON}},
		strings.NewReader("eudore"),
	)
	app.NewRequest(nil, "POST", "/bind",
		http.Header{eudore.HeaderContentType: {"value"}},
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
	app := eudore.NewApp()
	app.AddMiddleware(func(ctx eudore.Context) {
		d := ctx.GetHeader("Debug")
		if d == "" {
			return
		}
		r := ctx.Request()
		switch d {
		case "uri":
			r.URL.RawQuery = "tag=%\007"
		case "cookie":
			r.Header.Add(eudore.HeaderCookie, "age=22; =00; tag=\007hs; aa=\"bb\"; ")
		case "body":
			r.Body = bodyError{}
		}
	})
	app.AnyFunc("/body", func(ctx eudore.Context) {
		ctx.Body()
	})
	app.AnyFunc("/* version=v0", func(ctx eudore.Context) {
		ctx.Info(ctx.Method(), ctx.Path(), ctx.Params().String(), string(ctx.Body()))
	})
	app.GetFunc("/params version=v0", func(ctx eudore.Context) {
		ctx.SetParam("name", "eudore")
		ctx.Info("params", ctx.Params().String(), ctx.GetParam("name"))
	})
	app.AnyFunc("/query", func(ctx eudore.Context) {
		ctx.Info("query name", ctx.GetQuery("name"))
	})
	app.AnyFunc("/querys", func(ctx eudore.Context) {
		ctx.Info("querys", ctx.Querys())
	})
	// cookie
	app.AnyFunc("/cookie-set", func(ctx eudore.Context) {
		ctx.SetCookie(&eudore.CookieSet{
			Name:     "set1",
			Value:    "val1",
			Path:     "/",
			HttpOnly: true,
		})
		ctx.SetCookieValue("set", "eudore", 0)
		ctx.SetCookieValue("name", "eudore", 600)
		ctx.Info("response set-cookie", ctx.Response().Header()[eudore.HeaderSetCookie])
	})
	app.AnyFunc("/cookie-get", func(ctx eudore.Context) {
		ctx.Info("cookie", ctx.GetHeader(eudore.HeaderCookie))
		ctx.GetCookie("name")
		for _, i := range ctx.Cookies() {
			fmt.Fprintf(ctx, "%s: %s\n", i.Name, i.Value)
		}
	})
	// form
	app.AnyFunc("/form-value", func(ctx eudore.Context) {
		ctx.Info("form value name:", ctx.FormValue("name"))
	})
	app.AnyFunc("/form-values", func(ctx eudore.Context) {
		ctx.Info("form values:", ctx.FormValues())
	})
	app.AnyFunc("/form-file", func(ctx eudore.Context) {
		ctx.Infof("form file: %#v", ctx.FormFile("file"))
	})
	app.AnyFunc("/form-files", func(ctx eudore.Context) {
		ctx.Infof("form values: %#v", ctx.FormFiles())
	})

	app.NewRequest(nil, "GET", "/")
	app.NewRequest(nil, "GET", "/body")
	app.NewRequest(nil, "GET", "/body", eudore.NewClientOptionHeader("Debug", "body"))
	app.NewRequest(nil, "GET", "/params")
	app.NewRequest(nil, "GET", "/query?name=eudore&debug=true")
	app.NewRequest(nil, "GET", "/querys?name=eudore&debug=true")
	app.NewRequest(nil, "PUT", "/query", eudore.NewClientOptionHeader("Debug", "uri"))
	app.NewRequest(nil, "PUT", "/querys", eudore.NewClientOptionHeader("Debug", "uri"))
	app.NewRequest(nil, "GET", "/cookie-set")
	app.NewRequest(nil, "GET", "/cookie-get",
		eudore.Cookie{"age", "22"},
		eudore.Cookie{"name", "a, b"},
		eudore.Cookie{"valid", "key\x03invalid"},
		http.Header{eudore.HeaderCookie: {"age=22;;;"}},
	)
	app.NewRequest(nil, "GET", "/cookie-get", eudore.NewClientOptionHeader("Debug", "cookie"))
	app.NewRequest(nil, "GET", "/form-value", eudore.NewClientBodyForm(url.Values{"name": {"eudor"}}))
	app.NewRequest(nil, "GET", "/form-value", eudore.NewClientBodyForm(url.Values{"key": {"eudor"}}))
	app.NewRequest(nil, "GET", "/form-values?name=eudore")
	app.NewRequest(nil, "GET", "/form-values", eudore.NewClientOptionHeader("Debug", "uri"))

	body := eudore.NewClientBodyForm(nil)
	body.AddFile("file", "app.txt", strings.NewReader("eudore app"))
	app.NewRequest(nil, "GET", "/form-file", body)
	body = eudore.NewClientBodyForm(nil)
	body.AddFile("name", "app.txt", strings.NewReader("eudore app"))
	app.NewRequest(nil, "GET", "/form-file", body)
	app.NewRequest(nil, "GET", "/form-file",
		eudore.NewClientOptionHeader(eudore.HeaderContentType, eudore.MimeText),
		strings.NewReader("body"),
	)
	body = eudore.NewClientBodyForm(nil)
	body.AddFile("file", "app.txt", strings.NewReader("eudore app"))
	app.NewRequest(nil, "GET", "/form-files", body)
	app.NewRequest(nil, "GET", "/form-files")
	app.NewRequest(nil, "GET", "/form-files",
		eudore.NewClientOptionHeader(eudore.HeaderContentType, eudore.MimeText),
		strings.NewReader("body"),
	)

	app.NewRequest(nil, "GET", "/form-value",
		eudore.NewClientOptionHeader("Debug", "body"),
		eudore.NewClientOptionHeader(eudore.HeaderContentType, eudore.MimeApplicationForm),
	)
	app.NewRequest(nil, "GET", "/form-value",
		strings.NewReader("name=%\007"),
	)
	app.NewRequest(nil, "GET", "/form-value",
		eudore.NewClientOptionHeader(eudore.HeaderContentType, eudore.MimeApplicationForm),
		strings.NewReader("name=%\007"),
	)
	app.NewRequest(nil, "GET", "/form-value",
		eudore.NewClientOptionHeader(eudore.HeaderContentType, eudore.MimeMultipartForm),
		strings.NewReader("body"),
	)
	app.NewRequest(nil, "GET", "/form-value",
		eudore.NewClientOptionHeader(eudore.HeaderContentType, eudore.MimeMultipartForm+"; boundary=x"),
		strings.NewReader("body"),
	)
	app.NewRequest(nil, "GET", "/form-value",
		eudore.NewClientOptionHeader(eudore.HeaderContentType, eudore.MimeText),
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
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyFilter, func(eudore.Context, interface{}) error { return nil })
	app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))
	app.AddMiddleware(func(ctx eudore.Context) {
		unwarper, ok := ctx.Response().(interface{ Unwrap() http.ResponseWriter })
		if ok {
			unwarper.Unwrap()
		}
		if ctx.GetQuery("debug") != "" {
			ctx.SetResponse(&responseError{headers: make(http.Header)})
		}
	})
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Info(ctx.Method(), ctx.Path(), ctx.Params().String(), string(ctx.Body()))
	})

	app.AnyFunc("/redirect", func(ctx eudore.Context) {
		ctx.Redirect(308, "/")
	})
	app.AnyFunc("/redirect200", func(ctx eudore.Context) {
		ctx.Redirect(200, "/")
	})
	app.AnyFunc("/ws", func(ctx eudore.Context) {
		conn, _, err := ctx.Response().Hijack()
		if err == nil {
			conn.Close()
		}
	})
	app.AnyFunc("/push", func(ctx eudore.Context) {
		ctx.Push("/index.js", nil)
	})
	app.AnyFunc("/write-string", func(ctx eudore.Context) {
		ctx.WriteString("hello")
	})
	app.AnyFunc("/write-file", func(ctx eudore.Context) {
		ctx.WriteFile("/index.html")
	})
	app.AnyFunc("/render", func(ctx eudore.Context) {
		ctx.Render(map[string]interface{}{
			"name":  "eudore",
			"route": ctx.GetParam(eudore.ParamRoute),
		})
	})
	app.AnyFunc("/status", func(ctx eudore.Context) {
		ctx.WriteHeader(201)
		ctx.WriteString("hello")
		ctx.Response().Status()
	})
	app.AnyFunc("/response", func(ctx eudore.Context) {
		ctx.Response().Flush()
		ctx.Response().Hijack()
	})

	client := app.GetClient()
	tp, ok := client.Transport.(*http.Transport)
	if ok {
		tp.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	app.ListenTLS(":8089", "", "")

	app.NewRequest(nil, "GET", "/redirect")
	app.NewRequest(nil, "GET", "/redirect200")
	app.NewRequest(nil, "GET", "/push")
	app.NewRequest(nil, "GET", "/ws")
	app.NewRequest(nil, "GET", "/write-string")
	app.NewRequest(nil, "GET", "/write-file")
	app.NewRequest(nil, "GET", "/render", http.Header{eudore.HeaderAccept: {eudore.MimeApplicationJSON}})
	app.NewRequest(nil, "GET", "/status")
	app.NewRequest(nil, "GET", "/response")
	app.NewRequest(nil, "GET", "https://localhost:8089/push")
	app.NewRequest(nil, "GET", "https://localhost:8089/ws")

	app.Client = app.WithClient(url.Values{"debug": {"1"}})
	app.NewRequest(nil, "GET", "/redirect")
	app.NewRequest(nil, "GET", "/push")
	app.NewRequest(nil, "GET", "/response")
	app.NewRequest(nil, "GET", "/write-json")
	app.NewRequest(nil, "GET", "/write-string")
	app.NewRequest(nil, "GET", "/write-file")
	app.NewRequest(nil, "GET", "/render", http.Header{eudore.HeaderAccept: {eudore.MimeApplicationJSON}})

	app.CancelFunc()
	app.Run()
}

func TestContextLogger(*testing.T) {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyLogger, eudore.NewLogger(&eudore.LoggerConfig{
		Caller: true,
	}))
	app.AddMiddleware("global", middleware.NewRequestIDFunc(func(eudore.Context) string {
		return uuid.New().String()
	}))

	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Debugf("debug")
		ctx.Info("hello")
		ctx.Infof("hello path is %s", ctx.GetParam("*"))
		ctx.Warning("warning")
		ctx.Warningf("warningf")
		ctx.Error(nil)
		ctx.Error("test error")
		ctx.Errorf("test error")
		ctx.Fatal(nil)
	})
	app.AnyFunc("/field", func(ctx eudore.Context) {
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
	app.AnyFunc("/err1", func(ctx eudore.Context) {
		ctx.Fatalf("fatal logger method: %s path: %s", ctx.Method(), ctx.Path())
		ctx.Debug("err:", ctx.Err())
	})
	app.AnyFunc("/err2", func(ctx eudore.Context) {
		ctx.Fatal(eudore.NewErrorWithStatusCode(fmt.Errorf("test error"), 432, 10032))
	})
	app.AnyFunc("/err3", func(ctx eudore.Context) {
		eudore.NewErrorWithStatus(fmt.Errorf("test error"), 0)
		ctx.Fatal(eudore.NewErrorWithStatus(fmt.Errorf("test error"), 432))
	})
	app.AnyFunc("/err4", func(ctx eudore.Context) {
		eudore.NewErrorWithCode(fmt.Errorf("test error"), 0)
		ctx.Fatal(eudore.NewErrorWithCode(fmt.Errorf("test error"), 10032))
	})

	app.NewRequest(nil, "GET", "/ffile")
	app.NewRequest(nil, "GET", "/field")
	app.NewRequest(nil, "GET", "/err1")
	app.NewRequest(nil, "GET", "/err2")
	app.NewRequest(nil, "GET", "/err3")
	app.NewRequest(nil, "GET", "/err4")

	app.CancelFunc()
	app.Run()
}

func TestContextValue(*testing.T) {
	type Data struct {
		Name string `json:"name" xml:"name"`
	}
	app := eudore.NewApp()
	app.Set("name", "eudore")

	temp, _ := template.New("").Parse(`{{- define "data" -}}Data Name is {{.Name}}{{- end -}}`)
	app.AnyFunc("/index", func(ctx eudore.Context) interface{} {
		ctx.SetValue(eudore.ContextKeyDatabase, ctx.Value(eudore.ContextKeyDatabase))
		ctx.SetValue(eudore.ContextKeyTemplate, temp)
		ctx.SetValue(eudore.ContextKeyTemplate, temp)
		return &Data{ctx.Value(eudore.ContextKeyConfig).(eudore.Config).Get("name").(string)}
	})
	app.AnyFunc("/cannel", func(ctx eudore.Context) error {
		c, fn := context.WithCancel(ctx.GetContext())
		fn()
		ctx.SetContext(c)
		return ctx.Err()
	})

	app.NewRequest(nil, "GET", "/index", http.Header{eudore.HeaderAccept: {eudore.MimeApplicationJSON}})
	app.NewRequest(nil, "GET", "/index", http.Header{eudore.HeaderAccept: {eudore.MimeTextHTML}})
	app.NewRequest(nil, "GET", "/cannel")

	app.CancelFunc()
	app.Run()
}
