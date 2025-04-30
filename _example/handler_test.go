package eudore_test

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	. "github.com/eudore/eudore"
	. "github.com/eudore/eudore/middleware"
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

	app := NewApp()
	app.SetValue(ContextKeyHandlerExtender, NewHandlerExtender())
	app.AddMiddleware(NewLoggerLevelFunc(func(ctx Context) int {
		return int(LoggerFatal)
	}))
	NewHandlerExtenderWithContext(app)
	NewHandlerExtenderWithContext(context.Background())

	app.AddHandler("404", "", HandlerRouter404)
	app.AddHandler("405", "", HandlerRouter405)
	app.GetFunc("/403", HandlerRouter403)
	app.GetFunc("/index", HandlerEmpty)
	app.GetFunc("/trace", HandlerMethodTrace)
	app.GetFunc("/static/dir/*", NewHandlerFileSystems(".", "."))
	app.GetFunc("/static/index/* autoindex=true", NewHandlerFileSystems(".", "."))
	app.GetFunc("/static/embed/*", root)
	app.GetFunc("/static/fs1/* autoindex=true", fsPermission{})
	app.GetFunc("/static/fs2/* autoindex=true", fsHTTPDir{})

	app.NewRequest("GET", "/index")
	app.NewRequest("POST", "/index")
	app.NewRequest("GET", "/trace")
	app.NewRequest("GET", "/403")
	app.NewRequest("GET", "/404")
	app.NewRequest("GET", "/static/dir/app_test.go")
	app.NewRequest("GET", "/static/embed/")
	app.NewRequest("GET", "/static/embed/app_test.go")
	app.NewRequest("GET", "/static/index/")
	app.NewRequest("GET", "/static/index/", NewClientHeader(HeaderAccept, MimeTextHTML))
	app.NewRequest("GET", "/static/index/static/")
	app.NewRequest("GET", "/static/index/403.js")
	app.NewRequest("GET", "/static/fs1/")
	app.NewRequest("GET", "/static/fs2/")
	NewFileSystems(".", http.Dir("."), NewFileSystems(".", "."))

	app.CancelFunc()
	app.Run()
}

func BindTestErr(ctx Context, i any) error {
	if ctx.GetHeader("Debug") == "binderr" {
		return errors.New("test bind error")
	}
	return nil
}

func RenderTestErr(ctx Context, i any) error {
	if ctx.GetHeader("Debug") == "rendererr" && fmt.Sprintf("%T", i) != "eudore.contextMessage" {
		return errors.New("test render error")
	}
	return HandlerDataRenderJSON(ctx, i)
}

type request017 struct {
	Name string
}
type handlerControler4 struct{ ControllerAutoRoute }

func (ctl handlerControler4) Get(Context) {}

func TestHandlerRegister(t *testing.T) {
	app := NewApp()
	app.SetValue(ContextKeyBind, BindTestErr)
	app.SetValue(ContextKeyRender, RenderTestErr)
	app.SetValue(ContextKeyContextPool, NewContextBasePool(app))
	app.SetValue(ContextKeyRouter, NewRouter(nil).Group(" loggerkind=~all"))

	app.AddController(new(handlerControler4))
	app.AddHandlerExtend(
		00,
		00, 00,
		func(any) {},
		func() HandlerFunc { return nil },
		func(int) HandlerFunc { return nil },
		func(r string, _ func(string)) HandlerFunc {
			return func(ctx Context) {
				ctx.WriteString("route: " + r)
			}
		},
		NewHandlerFuncContextType[*request017],
		NewHandlerFuncContextTypeAny[*request017],
		NewHandlerFuncContextTypeError[*request017],
		NewHandlerFuncContextTypeAnyError[*request017],
	)

	exts := []any{
		0,
		"Router: newHandlerFuncs path is '/1/1', 0th handler parameter type is 'int', this is the unregistered handler type",
		[]HandlerFunc{HandlerEmpty},
		"",
		[3]HandlerFunc{HandlerEmpty, HandlerEmpty},
		"",
		http.NotFoundHandler(),
		"",
		http.RedirectHandler("/", 308),
		"",
		func(http.ResponseWriter, *http.Request) {},
		"",
		func(string) {},
		"",
		func(*testing.T) {},
		"Router: newHandlerFuncs path is '/1/15', 0th handler parameter type is 'func(*testing.T)', this is the unregistered handler type",
		func(Context, int) {},
		"Router: newHandlerFuncs path is '/1/17', 0th handler parameter type is 'func(eudore.Context, int)', this is the unregistered handler type",
		func(Context, int) (any, error) {
			return nil, nil
		},
		"Router: newHandlerFuncs path is '/1/19', 0th handler parameter type is 'func(eudore.Context, int) (interface {}, error)', this is the unregistered handler type",
	}

	for i := 0; i < len(exts); i += 2 {
		err := app.AddHandler("GET", "/1/"+strconv.Itoa(i+1), exts[i])
		if err != nil && err.Error() != exts[i+1].(string) {
			t.Log(i+1, err)
		}
	}

	funcs := []any{
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}),
		"",
		"",
		"",
		func(http.ResponseWriter, *http.Request) {},
		"",
		"",
		"",
		http.FileServer(http.Dir(".")),
		"Client: parse not suppert Content-Type: text/plain",
		"Client: parse not suppert Content-Type: text/plain",
		"Client: parse not suppert Content-Type: text/plain",
		func() {},
		"",
		"",
		"",
		func() any {
			return "hello"
		},
		"",
		"",
		"client request status is 500, error: test render error",
		func() error {
			return errors.New("test handler error")
		},
		"client request status is 500, error: test handler error",
		"client request status is 500, error: test handler error",
		"client request status is 500, error: test handler error",
		func() (any, error) {
			return "hello", nil
		},
		"",
		"",
		"client request status is 500, error: test render error",
		func(Context) any {
			return "test render"
		},
		"",
		"",
		"client request status is 500, error: test render error",
		func(Context) error {
			return errors.New("test handler error")
		},
		"client request status is 500, error: test handler error",
		"client request status is 500, error: test handler error",
		"client request status is 500, error: test handler error",
		func(Context) (any, error) {
			return "hello", nil
		},
		"",
		"",
		"client request status is 500, error: test render error",
		func(Context) (any, error) {
			return nil, errors.New("test handler error")
		},
		"client request status is 500, error: test handler error",
		"client request status is 500, error: test handler error",
		"client request status is 500, error: test handler error",
		func(Context, map[string]any) (any, error) {
			return "hello", nil
		},
		"",
		"client request status is 500, error: test bind error",
		"client request status is 500, error: test render error",
		func(Context, *request017) {},
		"",
		"client request status is 500, error: test bind error",
		"",
		func(Context, *request017) any { return nil },
		"",
		"client request status is 500, error: test bind error",
		"client request status is 500, error: test render error",
		func(Context, *request017) error {
			return errors.New("test error")
		},
		"client request status is 500, error: test error",
		"client request status is 500, error: test bind error",
		"client request status is 500, error: test error",
		func(Context, *request017) (any, error) {
			return t, nil
		},
		"",
		"client request status is 500, error: test bind error",
		"client request status is 500, error: test render error",
	}

	bindh := http.Header{"Debug": []string{"binderr"}}
	renderh := http.Header{"Debug": []string{"rendererr"}, HeaderAccept: []string{MimeApplicationJSON}}
	options := []any{NewClientParseErr(), context.WithValue(app, ContextKeyLogger, DefaultLoggerNull)}
	app.AddMiddleware(NewLoggerLevelFunc(func(ctx Context) int {
		return int(LoggerFatal)
	}))

	for i := 0; i < len(funcs); i += 4 {
		app.GetFunc("/2/"+strconv.Itoa(i+1), funcs[i])
		err := app.GetRequest("/2/"+strconv.Itoa(i+1), options)
		if err != nil && err.Error() != funcs[i+1].(string) {
			t.Log(i+1, err)
		}
		err = app.GetRequest("/2/"+strconv.Itoa(i+1), bindh, options)
		if err != nil && err.Error() != funcs[i+2].(string) {
			t.Log(i+2, err)
		}
		err = app.GetRequest("/2/"+strconv.Itoa(i+1), renderh, options)
		if err != nil && err.Error() != funcs[i+3].(string) {
			t.Log(i+3, err)
		}
	}

	hes := []HandlerExtender{
		NewHandlerExtender(),
		NewHandlerExtenderBase(),
		NewHandlerExtenderTree(),
		NewHandlerExtenderWithContext(context.Background()),
		NewHandlerExtenderWrap(DefaultHandlerExtender, DefaultHandlerExtender),
	}
	for _, he := range hes {
		he.(interface{ Metadata() any }).Metadata()
	}

	app.CancelFunc()
	app.Run()
}

func TestHandlerList(t *testing.T) {
	app := NewApp()
	app.SetValue(ContextKeyLogger, DefaultLoggerNull)
	app.AddHandlerExtend("/", func(any) HandlerFunc {
		return nil
	})
	app.AddHandlerExtend("/api/user", func(any) HandlerFunc {
		return HandlerEmpty
	})
	app.AddHandlerExtend("/api/icon", func(any) HandlerFunc {
		return nil
	})
	api := app.Group("/api")
	api.AddHandlerExtend("/", func(any) HandlerFunc {
		return nil
	})
	api.AnyFunc("/user/info", "hello")
	t.Log(strings.Join(api.(HandlerExtender).List(), "\n"))

	app.CancelFunc()
	app.Run()
}

type rpcrequest struct {
	Name string
}
type rpcresponse struct {
	Messahe string
}

func TestHandlerRPC(t *testing.T) {
	app := NewApp()
	app.SetValue(ContextKeyBind, BindTestErr)
	app.SetValue(ContextKeyRender, RenderTestErr)
	app.SetValue(ContextKeyContextPool, NewContextBasePool(app))

	api := app.Group(" loggerkind=~handler|~middleware")
	api.AddMiddleware(NewLoggerLevelFunc(func(ctx Context) int {
		return int(LoggerFatal)
	}))
	api.AnyFunc("/1/1", func(Context, *rpcrequest) (rpcresponse, error) {
		return rpcresponse{Messahe: "success"}, nil
	})
	api.AnyFunc("/1/2", func(Context, map[string]any) (*rpcresponse, error) {
		return nil, errors.New("test rpc error")
	})

	app.NewRequest("PUT", "/1/1", NewClientCheckStatus(200))
	app.NewRequest("PUT", "/1/2", http.Header{HeaderAccept: {MimeApplicationJSON}}, NewClientCheckStatus(500))
	app.NewRequest("PUT", "/1/2", NewClientBodyJSON(map[string]any{
		"name": "eudore",
	}), NewClientCheckStatus(500))

	app.NewRequest("GET", "/1/1", http.Header{"Debug": []string{"binderr"}}, NewClientCheckStatus(500))
	app.NewRequest("GET", "/1/1", http.Header{"Debug": []string{"rendererr"}}, NewClientCheckStatus(500))

	app.CancelFunc()
	app.Run()
}

func TestHandlerFunc(t *testing.T) {
	SetHandlerAliasName(new(request017), "")
	SetHandlerAliasName(new(request017), "handlerHttp1-test")
	defer func() {
		recover()
	}()

	hs := HandlerFuncs{HandlerEmpty}
	NewHandlerFuncsCombine(hs, nil)
	for i := 0; i < 10; i++ {
		hs = NewHandlerFuncsCombine(hs, hs)
	}
	t.Log(len(hs))
}
