package eudore_test

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/eudore/eudore"
)

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

type handlerHttp1 struct{}
type handlerHttp2 struct{}
type handlerHttp3 struct{}
type handlerControler4 struct{ eudore.ControllerAutoRoute }

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
		app.NewRequest(nil, "GET", fmt.Sprintf("/2/%d", i), eudore.NewClientQuery("binderr", "1"))
	}
	for i := 1; i < 14; i++ {
		app.NewRequest(nil, "GET", fmt.Sprintf("/2/%d", i), eudore.NewClientQuery("rendererr", "1"))
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
	t.Log(strings.Join(api.(eudore.HandlerExtender).ListExtendHandlerNames(), "\n"))

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
	app.NewRequest(nil, "PUT", "/1/2", eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeApplicationJSON))
	app.NewRequest(nil, "PUT", "/1/2", eudore.NewClientBodyJSON(map[string]interface{}{
		"name": "eudore",
	}))
	app.NewRequest(nil, "GET", "/1/1", eudore.NewClientQuery("binderr", "1"))
	app.NewRequest(nil, "GET", "/1/1", eudore.NewClientQuery("rendererr", "1"))

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

func TestHandlerStatic(t *testing.T) {
	app := eudore.NewApp()
	app.AnyFunc("/static/*", eudore.NewStaticHandler("", ""))

	app.NewRequest(nil, "GET", "/static/index.html")
	app.CancelFunc()
	app.Run()
}
