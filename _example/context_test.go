package eudore_test

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"testing"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
	"github.com/google/uuid"
)

func TestContext(*testing.T) {
	app := eudore.NewApp()
	client := eudore.NewClientWarp()
	app.SetValue(eudore.ContextKeyClient, client)

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

	client.NewRequest("GET", "/context").Do()

	app.CancelFunc()
	app.Run()
}

func TestContextRequest(*testing.T) {
	app := eudore.NewApp()
	client := eudore.NewClientWarp()
	app.SetValue(eudore.ContextKeyClient, client)
	app.SetValue(eudore.ContextKeyValidate, func(eudore.Context, interface{}) error { return nil })
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

	client.NewRequest("GET", "/info").Do()
	client.NewRequest("GET", "/realip").Do()
	client.NewRequest("GET", "/realip").AddHeader(eudore.HeaderXRealIP, "47.11.11.11").Do()
	client.NewRequest("GET", "/realip").AddHeader(eudore.HeaderXForwardedFor, "47.11.11.11").Do()
	client.NewRequest("GET", "/bind").BodyJSON(bindData{"eudore"}).Do()
	client.NewRequest("GET", "/bind").AddHeader(eudore.HeaderContentType, eudore.MimeApplicationJSON).BodyString("eudore").Do()
	client.NewRequest("POST", "/bind").AddHeader(eudore.HeaderContentType, "value").BodyString("eudore").Do()

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
	client := eudore.NewClientWarp()
	app.SetValue(eudore.ContextKeyClient, client)

	app.AnyFunc("/* version=v0", func(ctx eudore.Context) {
		ctx.Info(ctx.Method(), ctx.Path(), ctx.Params().String(), string(ctx.Body()))
	})
	app.GetFunc("/params version=v0", func(ctx eudore.Context) {
		ctx.SetParam("name", "eudore")
		ctx.Info("params", ctx.Params().String(), ctx.GetParam("name"))
	})
	app.AnyFunc("/querys", func(ctx eudore.Context) {
		ctx.Debug(string(ctx.Body()), ctx.Request().RequestURI)
		ctx.Info("querys", ctx.Querys())
		ctx.Info("query name", ctx.GetQuery("name"))
	})
	app.AnyFunc("/querys-err1", func(ctx eudore.Context) {
		ctx.Request().URL.RawQuery = "tag=%\007"
		ctx.Info("querys", ctx.Querys())
	})
	app.AnyFunc("/querys-err2", func(ctx eudore.Context) {
		ctx.Request().URL.RawQuery = "tag=%\007"
		ctx.Info("query name", ctx.GetQuery("name"))
	})
	// cookie
	app.AnyFunc("/cookie-set", func(ctx eudore.Context) {
		ctx.SetCookie(&eudore.SetCookie{
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
		ctx.Infof("cookie name value is: %s", ctx.GetCookie("name"))
		ctx.Infof("cookie age value is: %s", ctx.GetCookie("age"))
		for _, i := range ctx.Cookies() {
			fmt.Fprintf(ctx, "%s: %s\n", i.Name, i.Value)
		}
	})
	app.AnyFunc("/cookie-err", func(ctx eudore.Context) {
		ctx.Request().Header.Add(eudore.HeaderCookie, "age=22; =00; tag=\007hs; aa=\"bb\"; ")
		ctx.Info("cookies", ctx.Cookies())
	})
	// form
	app.AnyFunc("/form-value", func(ctx eudore.Context) {
		ctx.Info("form value name:", ctx.FormValue("name"))
		ctx.Info("form value group:", ctx.FormValue("group"))
		ctx.Info("form values:", ctx.FormValues())
	})
	app.AnyFunc("/form-file", func(ctx eudore.Context) {
		ctx.Infof("%s", ctx.Body())
		ctx.Infof("form value name: %#v", ctx.FormFile("file"))
		ctx.Infof("form value group: %#v", ctx.FormFile("name"))
		ctx.Infof("form values: %#v", ctx.FormFiles())
	})
	app.AnyFunc("/form-err", func(ctx eudore.Context) {
		ctx.FormValue("name")
		ctx.FormValues()
		ctx.FormFile("file")
		ctx.FormFiles()
	})
	app.AnyFunc("/body", func(ctx eudore.Context) {
		ctx.Request().Body = bodyError{}
		ctx.Body()
	})
	app.AnyFunc("/read", func(ctx eudore.Context) {
		body := make([]byte, 4096)
		ctx.Read(body)
	})

	client.NewRequest("GET", "/").Do()
	client.NewRequest("GET", "/params").Do()
	client.NewRequest("GET", "/querys").AddQuery("name", "eudore").AddQuery("debug", "true").Do()
	client.NewRequest("PUT", "/querys-err1").Do()
	client.NewRequest("PUT", "/querys-err2").Do()
	client.NewRequest("GET", "/cookie-get").Do()
	client.NewRequest("GET", "/cookie-set").Do()
	client.NewRequest("GET", "/cookie-get").Do()
	client.NewRequest("GET", "/cookie-get").AddHeader("Cookie", "age=22;").Do()
	client.NewRequest("GET", "/cookie-err").Do()
	client.NewRequest("GET", "/form-value").BodyFormValue("name", "eudore").Do()
	client.NewRequest("GET", "/form-file").BodyFormFile("file", "app.txt", "eudore app").Do()
	client.NewRequest("GET", "/form-err").BodyString("name=eudore").Do()
	client.NewRequest("GET", "/body").Do()
	client.NewRequest("GET", "/read").Do()

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
	client := eudore.NewClientWarp()
	app.SetValue(eudore.ContextKeyClient, client)
	app.SetValue(eudore.ContextKeyFilte, func(eudore.Context, interface{}) error { return nil })
	app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))
	app.AddMiddleware(func(ctx eudore.Context) {
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
	app.AnyFunc("/ws", func(ctx eudore.Context) {
		conn, _, err := ctx.Response().Hijack()
		if err == nil {
			conn.Close()
		}
	})
	app.AnyFunc("/push", func(ctx eudore.Context) {
		ctx.Push("/index.js", nil)
	})
	app.AnyFunc("/write-json", func(ctx eudore.Context) {
		ctx.WriteJSON(map[string]interface{}{
			"name":  "eudore",
			"route": ctx.GetParam(eudore.ParamRoute),
		})
	})
	app.AnyFunc("/write-string", func(ctx eudore.Context) {
		ctx.WriteString("hello")
	})
	app.AnyFunc("/write-file", func(ctx eudore.Context) {
		ctx.WriteFile("/index.html")
	})
	app.AnyFunc("/write-error1", func(ctx eudore.Context) {
		ctx.WriteError(eudore.NewErrorCode(fmt.Errorf("test err"), 1002))
	})
	app.AnyFunc("/write-error2", func(ctx eudore.Context) {
		ctx.WriteError(eudore.NewErrorStatus(fmt.Errorf("test err"), 499))
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

	c, ok := client.Client.(*http.Client)
	if ok {
		tp, ok := c.Transport.(*http.Transport)
		if ok {
			tp.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
	}
	app.ListenTLS(":8089", "", "")

	client.NewRequest("GET", "/redirect").Do()
	client.NewRequest("GET", "/push").Do()
	client.NewRequest("GET", "/ws").Do()
	client.NewRequest("GET", "/write-json").Do()
	client.NewRequest("GET", "/write-string").Do()
	client.NewRequest("GET", "/write-file").Do()
	client.NewRequest("GET", "/write-error1").Do()
	client.NewRequest("GET", "/write-error2").Do()
	client.NewRequest("GET", "/render").AddHeader(eudore.HeaderAccept, eudore.MimeApplicationJSON).Do()
	client.NewRequest("GET", "/status").Do()
	client.NewRequest("GET", "/response").Do()
	client.NewRequest("GET", "https://localhost:8089/push").Do()
	client.NewRequest("GET", "https://localhost:8089/ws").Do()

	client.AddQuery("debug", "1")
	client.NewRequest("GET", "/redirect").Do()
	client.NewRequest("GET", "/push").Do()
	client.NewRequest("GET", "/write-json").Do()
	client.NewRequest("GET", "/write-string").Do()
	client.NewRequest("GET", "/write-file").Do()
	client.NewRequest("GET", "/render").AddHeader(eudore.HeaderAccept, eudore.MimeApplicationJSON).Do()

	app.CancelFunc()
	app.Run()
}

func TestContextLogger(*testing.T) {
	app := eudore.NewApp()
	client := eudore.NewClientWarp()
	app.SetValue(eudore.ContextKeyClient, client)

	app.SetValue(eudore.ContextKeyLogger, eudore.NewLoggerStd(map[string]interface{}{"FileLine": true}))
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
	app.AnyFunc("/err", func(ctx eudore.Context) {
		ctx.Fatalf("fatal logger method: %s path: %s", ctx.Method(), ctx.Path())
		ctx.Debug("err:", ctx.Err())
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

	client.NewRequest("GET", "/ffile").Do()
	client.NewRequest("GET", "/err").Do()
	client.NewRequest("GET", "/field").Do()

	app.CancelFunc()
	app.Run()
}

func TestContextValue(*testing.T) {
	type Data struct {
		Name string `json:"name" xml:"name"`
	}
	app := eudore.NewApp()
	client := eudore.NewClientWarp()
	app.SetValue(eudore.ContextKeyClient, client)
	app.Set("name", "eudore")

	temp, _ := template.New("").Parse(`{{- define "data" -}}Data Name is {{.Name}}{{- end -}}`)
	app.AnyFunc("/index", func(ctx eudore.Context) interface{} {
		ctx.SetValue(eudore.ContextKeyTempldate, temp)
		ctx.SetValue(eudore.ContextKeyTempldate, temp)
		return &Data{ctx.Value(eudore.ContextKeyConfig).(eudore.Config).Get("name").(string)}
	})
	app.AnyFunc("/cannel", func(ctx eudore.Context) error {
		c, fn := context.WithCancel(ctx.GetContext())
		fn()
		ctx.SetContext(c)
		return ctx.Err()
	})

	client.NewRequest("GET", "/index").AddHeader(eudore.HeaderAccept, eudore.MimeApplicationJSON).Do()
	client.NewRequest("GET", "/index").AddHeader(eudore.HeaderAccept, eudore.MimeTextHTML).Do()
	client.NewRequest("GET", "/cannel").Do()

	app.CancelFunc()
	app.Run()
}
