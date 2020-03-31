package eudore_test

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

type handlerHttp1 struct{}
type handlerHttp2 struct{}

func (handlerHttp1) HandleHTTP(eudore.Context)                      {}
func (h handlerHttp2) CloneHandler() http.Handler                   { return h }
func (h handlerHttp2) ServeHTTP(http.ResponseWriter, *http.Request) {}

func TestHandlerReister2(t *testing.T) {
	app := eudore.NewCore()
	app.AddHandlerExtend(00)
	app.AddHandlerExtend(func(int) {})
	app.AddHandlerExtend(func(interface{}) {})
	app.AddHandlerExtend(func(eudore.Errors) eudore.HandlerFunc {
		return eudore.HandlerEmpty
	})

	app.AnyFunc("/1/1", func(eudore.Context) {})
	app.AnyFunc("/1/2", eudore.HandlerFunc(eudore.HandlerEmpty))
	app.AnyFunc("/1/3", []eudore.HandlerFunc{eudore.HandlerEmpty})
	app.AnyFunc("/1/4", eudore.HandlerFuncs{eudore.HandlerEmpty})
	app.AnyFunc("/1/5", eudore.HandlerEmpty, eudore.HandlerEmpty)
	app.AnyFunc("/1/6", [3]eudore.HandlerFunc{eudore.HandlerEmpty, eudore.HandlerEmpty})
	app.AnyFunc("/1/7", error(&eudore.Errors{}))
	app.AnyFunc("/1/8", eudore.LogDebug)
	app.AnyFunc("/1/9", func(http.ResponseWriter, *http.Request) {})
	app.AnyFunc("/1/10", http.NotFoundHandler())
	app.AnyFunc("/1/11", func(eudore.Context, int) {})
	app.AnyFunc("/1/12", func(eudore.Context, int) (interface{}, error) {
		return nil, nil
	})
	app.AnyFunc("/1/13", new(handlerHttp1))
	app.AnyFunc("/1/13", new(handlerHttp2))

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
	app.AnyFunc("/2/6", func(eudore.Context) (interface{}, error) {
		return "hello", nil
	})
	app.AnyFunc("/2/7", func(eudore.Context, map[string]interface{}) (interface{}, error) {
		return "hello", nil
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/1/1").Do()
	client.NewRequest("GET", "/1/2").Do()
	client.NewRequest("GET", "/1/3").Do()
	client.NewRequest("GET", "/1/4").Do()
	client.NewRequest("GET", "/1/8").Do()
	client.NewRequest("GET", "/1/9").Do()
	client.NewRequest("GET", "/1/10").Do()
	client.NewRequest("GET", "/2/1").Do()
	client.NewRequest("GET", "/2/2").Do()
	client.NewRequest("GET", "/2/3").Do()
	client.NewRequest("GET", "/2/4").Do()
	client.NewRequest("GET", "/2/5").Do()
	client.NewRequest("GET", "/2/6").Do()
	client.NewRequest("GET", "/2/7").Do()

	app.Renderer = func(eudore.Context, interface{}) error {
		return errors.New("test render error")
	}
	client.NewRequest("GET", "/2/2").Do()
	client.NewRequest("GET", "/2/4").Do()
	client.NewRequest("GET", "/2/6").Do()
	client.NewRequest("GET", "/2/7").Do()
	app.Binder = func(eudore.Context, io.Reader, interface{}) error {
		return errors.New("test binder error")
	}
	client.NewRequest("GET", "/2/7").Do()

	for client.Next() {
		app.Error(client.Error())
	}
	app.Run()
}

func TestHandlerList2(t *testing.T) {
	app := eudore.NewCore()
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

func TestHandlerRPC2(t *testing.T) {
	app := eudore.NewCore()
	app.AnyFunc("/1/1", func(eudore.Context, *rpcrequest) (rpcresponse, error) {
		return rpcresponse{Messahe: "success"}, nil
	})
	app.AnyFunc("/1/2", func(eudore.Context, map[string]interface{}) (*rpcresponse, error) {
		return nil, errors.New("test rpc error")
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/1/1").Do()
	client.NewRequest("GET", "/1/2").WithHeaderValue(eudore.HeaderAccept, eudore.MimeApplicationJSON).Do()

	app.Renderer = func(eudore.Context, interface{}) error {
		return errors.New("test render error")
	}
	client.NewRequest("GET", "/1/1").Do()
	app.Binder = func(eudore.Context, io.Reader, interface{}) error {
		return errors.New("test binder error")
	}
	client.NewRequest("GET", "/1/1").Do()
	for client.Next() {
		app.Error(client.Error())
	}
	app.Run()
}

func TestHandlerFunc2(t *testing.T) {
	eudore.SetHandlerAliasName(new(handlerHttp1), "handlerHttp1-test")
	defer func() {
		recover()
	}()

	hs := eudore.HandlerFuncs{eudore.HandlerEmpty}
	eudore.HandlerFuncsCombine(hs, nil)
	for i := 0; i < 10; i++ {
		hs = eudore.HandlerFuncsCombine(hs, hs)
	}
	t.Log(len(hs))
}
