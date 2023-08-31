package eudore_test

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

//go:embed *.go
var root embed.FS

type fsPermission struct{}

func (fsPermission) Open(name string) (http.File, error) {
	return nil, os.ErrPermission
}

type fsHTTPDir struct{}

func (fsHTTPDir) Open(name string) (http.File, error) {
	return fsHTTPFile{}, nil
}

type fsHTTPFile struct {
	http.File
}

func (fsHTTPFile) Readdir(count int) ([]fs.FileInfo, error) {
	return nil, fmt.Errorf("test error, not dir")
}

func (fsHTTPFile) Stat() (fs.FileInfo, error) {
	return os.Stat(".")
}

func (fsHTTPFile) Close() error {
	return nil
}

func TestHandlerRoute(t *testing.T) {
	os.Mkdir("static/", 0o755)
	defer os.RemoveAll("static/")
	os.WriteFile("static/403.js", []byte("1234567890abcdef"), 0o000)
	file, _ := os.OpenFile("static/index.js", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	for i := 0; i < 10000; i++ {
		file.Write([]byte("1234567890abcdef"))
	}

	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyHandlerExtender, eudore.NewHandlerExtender())
	app.AddMiddleware("global", middleware.NewLoggerFunc(app, "route"))
	app.AddHandler("404", "", eudore.HandlerRouter404)
	app.AddHandler("405", "", eudore.HandlerRouter405)
	app.GetFunc("/403", eudore.HandlerRouter403)
	app.GetFunc("/index", eudore.HandlerEmpty)
	app.GetFunc("/meta/*", eudore.HandlerMetadata)
	app.GetFunc("/static/dir/*", eudore.NewHandlerStatic(".", "."))
	app.GetFunc("/static/index/* autoindex=true", eudore.NewHandlerStatic(".", "."))
	app.GetFunc("/static/embed/*", root)
	app.GetFunc("/static/fs1/* autoindex=true", fsPermission{})
	app.GetFunc("/static/fs2/* autoindex=true", fsHTTPDir{})

	app.NewRequest(nil, "GET", "/index")
	app.NewRequest(nil, "POST", "/index")
	app.NewRequest(nil, "GET", "/403")
	app.NewRequest(nil, "GET", "/404")
	app.NewRequest(nil, "GET", "/meta/")
	app.NewRequest(nil, "GET", "/meta/app")
	app.NewRequest(nil, "GET", "/meta/router")
	app.NewRequest(nil, "GET", "/static/dir/app_test.go")
	app.NewRequest(nil, "GET", "/static/embed/")
	app.NewRequest(nil, "GET", "/static/embed/app_test.go")
	app.NewRequest(nil, "GET", "/static/index/")
	app.NewRequest(nil, "GET", "/static/index/static/")
	app.NewRequest(nil, "GET", "/static/index/403.js")
	app.NewRequest(nil, "GET", "/static/fs1/")
	app.NewRequest(nil, "GET", "/static/fs2/")

	eudore.NewFileSystems(".", http.Dir("."), eudore.NewFileSystems(".", "."))

	app.SetValue(eudore.ContextKeyHandlerExtender, eudore.NewHandlerExtenderTree())
	app.NewRequest(nil, "GET", "/meta/")
	app.SetValue(eudore.ContextKeyHandlerExtender, eudore.NewHandlerExtenderWarp(
		eudore.NewHandlerExtender(),
		eudore.NewHandlerExtenderTree(),
	))
	app.NewRequest(nil, "GET", "/meta/")

	app.CancelFunc()
	app.Run()
}

func BindTestErr(ctx eudore.Context, i interface{}) error {
	if ctx.GetQuery("binderr") != "" {
		return errors.New("test bind error")
	}
	return eudore.NewBinds(nil)(ctx, i)
}

func RenderTestErr(ctx eudore.Context, i interface{}) error {
	if ctx.GetQuery("rendererr") != "" {
		return errors.New("test render error")
	}
	return eudore.RenderJSON(ctx, i)
}

type (
	handlerHttp1      struct{}
	handlerHttp2      struct{}
	handlerHttp3      struct{}
	handlerControler4 struct{ eudore.ControllerAutoRoute }
)

func (handlerHttp1) HandleHTTP(eudore.Context)                      {}
func (h handlerHttp2) CloneHandler() http.Handler                   { return h }
func (h handlerHttp2) ServeHTTP(http.ResponseWriter, *http.Request) {}
func (handlerHttp3) String() string                                 { return "hello" }
func (ctl handlerControler4) Get(eudore.Context)                    {}

func TestHandlerReister(t *testing.T) {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyBind, BindTestErr)
	app.SetValue(eudore.ContextKeyRender, RenderTestErr)
	app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))

	app.Info(app.AddHandlerExtend(00))
	app.Info(app.AddHandlerExtend(00, 00))
	app.Info(app.AddHandlerExtend(func(int) {}))
	app.Info(app.AddHandlerExtend(func(interface{}) {}))
	app.Info(app.AddHandlerExtend(func(interface{}, interface{}) {}))
	app.Info(app.AddHandlerExtend(func(route string, _ func(string)) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			ctx.WriteString("route: " + route)
		}
	}))
	app.AddController(new(handlerControler4))

	app.AnyFunc("/1/1", func(eudore.Context) {})
	app.AnyFunc("/1/2", eudore.HandlerFunc(eudore.HandlerEmpty))
	app.AnyFunc("/1/3", []eudore.HandlerFunc{eudore.HandlerEmpty})
	app.AnyFunc("/1/4", eudore.HandlerFuncs{eudore.HandlerEmpty})
	app.AnyFunc("/1/5", eudore.HandlerEmpty, eudore.HandlerEmpty)
	app.AnyFunc("/1/6", [3]eudore.HandlerFunc{eudore.HandlerEmpty, eudore.HandlerEmpty})
	app.AnyFunc("/1/8", eudore.LoggerDebug)
	app.AnyFunc("/1/9", func(http.ResponseWriter, *http.Request) {})
	app.AnyFunc("/1/10", http.NotFoundHandler())
	app.AnyFunc("/1/11", func(eudore.Context, int) {})
	app.AnyFunc("/1/12", func(eudore.Context, int) (interface{}, error) {
		return nil, nil
	})
	app.AnyFunc("/1/13", new(handlerHttp1))
	app.AnyFunc("/1/14", new(handlerHttp2))
	app.AnyFunc("/1/15", handlerHttp3{})
	app.AnyFunc("/1/15", handlerHttp3{})
	app.AnyFunc("/1/16", func(string) {})

	app.AnyFunc("/2/1", func(eudore.Context) error {
		return errors.New("test handler error")
	})
	app.AnyFunc("/2/2", func(eudore.Context) interface{} {
		return "test render"
	})
	app.AnyFunc("/2/3", func() string {
		return "hello"
	})
	app.AnyFunc("/2/4", func() interface{} {
		return "hello"
	})
	app.AnyFunc("/2/5", func() error {
		return errors.New("test error")
	})
	app.AnyFunc("/2/6", func() (interface{}, error) {
		return "hello", nil
	})
	app.AnyFunc("/2/7", func(eudore.Context) (interface{}, error) {
		return "hello", nil
	})
	app.AnyFunc("/2/8", func(eudore.Context) (interface{}, error) {
		return nil, errors.New("test error")
	})
	app.AnyFunc("/2/9", func(eudore.Context) (interface{}, error) {
		return "hello", nil
	})
	app.AnyFunc("/2/10", func(eudore.Context, map[string]interface{}) (interface{}, error) {
		return "hello", nil
	})
	app.AnyFunc("/2/11", func() {
	})
	app.AnyFunc("/2/12", func(*testing.T) {
	})
	app.AnyFunc("/2/13", func(eudore.Context) (*testing.T, error) {
		return t, nil
	})

	for i := 1; i < 17; i++ {
		app.NewRequest(nil, "GET", fmt.Sprintf("/1/%d", i))
	}
	for i := 1; i < 14; i++ {
		app.NewRequest(nil, "GET", fmt.Sprintf("/2/%d", i))
	}

	for i := 1; i < 14; i++ {
		app.NewRequest(nil, "GET", fmt.Sprintf("/2/%d", i), url.Values{"binderr": {"1"}})
	}
	for i := 1; i < 14; i++ {
		app.NewRequest(nil, "GET", fmt.Sprintf("/2/%d", i), url.Values{"rendererr": {"1"}})
	}

	app.CancelFunc()
	app.Run()
}

func TestHandlerList(t *testing.T) {
	app := eudore.NewApp()
	app.AddHandlerExtend("/api/user", func(interface{}) eudore.HandlerFunc {
		return eudore.HandlerEmpty
	})
	app.AddHandlerExtend("/api/icon", func(interface{}) eudore.HandlerFunc {
		return nil
	})
	api := app.Group("/api")
	api.AddHandlerExtend("/", func(interface{}) eudore.HandlerFunc {
		return nil
	})
	api.AnyFunc("/user/info", "hello")
	t.Log(strings.Join(api.(eudore.HandlerExtender).List(), "\n"))

	app.CancelFunc()
	app.Run()
}

type (
	rpcrequest struct {
		Name string
	}
	rpcresponse struct {
		Messahe string
	}
)

func TestHandlerRPC(t *testing.T) {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyBind, BindTestErr)
	app.SetValue(eudore.ContextKeyRender, RenderTestErr)
	app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))

	app.AnyFunc("/1/1", func(eudore.Context, *rpcrequest) (rpcresponse, error) {
		return rpcresponse{Messahe: "success"}, nil
	})
	app.AnyFunc("/1/2", func(eudore.Context, map[string]interface{}) (*rpcresponse, error) {
		return nil, errors.New("test rpc error")
	})

	app.NewRequest(nil, "PUT", "/1/1")
	app.NewRequest(nil, "PUT", "/1/2", http.Header{eudore.HeaderAccept: {eudore.MimeApplicationJSON}})
	app.NewRequest(nil, "PUT", "/1/2", eudore.NewClientBodyJSON(map[string]interface{}{
		"name": "eudore",
	}))
	app.NewRequest(nil, "GET", "/1/1", url.Values{"binderr": {"1"}})
	app.NewRequest(nil, "GET", "/1/1", url.Values{"rendererr": {"1"}})

	app.CancelFunc()
	app.Run()
}

func TestHandlerFunc(t *testing.T) {
	eudore.SetHandlerAliasName(new(handlerHttp1), "")
	eudore.SetHandlerAliasName(new(handlerHttp1), "handlerHttp1-test")
	defer func() {
		recover()
	}()

	hs := eudore.HandlerFuncs{eudore.HandlerEmpty}
	eudore.NewHandlerFuncsCombine(hs, nil)
	for i := 0; i < 10; i++ {
		hs = eudore.NewHandlerFuncsCombine(hs, hs)
	}
	t.Log(len(hs))
}
