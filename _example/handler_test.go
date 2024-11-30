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
	NewHandlerExtenderWithContext(app)
	NewHandlerExtenderWithContext(context.Background())

	app.AddHandler("404", "", HandlerRouter404)
	app.AddHandler("405", "", HandlerRouter405)
	app.GetFunc("/403", HandlerRouter403)
	app.GetFunc("/index", HandlerEmpty)
	app.GetFunc("/static/dir/*", NewHandlerFileSystems(".", "."))
	app.GetFunc("/static/index/* autoindex=true", NewHandlerFileSystems(".", "."))
	app.GetFunc("/static/embed/*", root)
	app.GetFunc("/static/fs1/* autoindex=true", fsPermission{})
	app.GetFunc("/static/fs2/* autoindex=true", fsHTTPDir{})

	app.NewRequest("GET", "/index")
	app.NewRequest("POST", "/index")
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
	if ctx.GetHeader("Debug") == "rendererr" {
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
		[]HandlerFunc{HandlerEmpty},
		[3]HandlerFunc{HandlerEmpty, HandlerEmpty},
		http.NotFoundHandler(),
		http.RedirectHandler("/", 308),
		func(http.ResponseWriter, *http.Request) {},
		func(string) {},
		func(*testing.T) {},
		func(Context, int) {},
		func(Context, int) (any, error) {
			return nil, nil
		},
	}
	for i := range exts {
		app.AnyFunc("/1/"+strconv.Itoa(i+1), exts[i])
		app.NewRequest("GET", "/1/"+strconv.Itoa(i+1))
	}

	funcs := []any{
		func() {},
		func() any {
			return "hello"
		},
		func() error {
			return errors.New("test error")
		},
		func() (any, error) {
			return "hello", nil
		},
		func(Context) any {
			return "test render"
		},
		func(Context) error {
			return errors.New("test handler error")
		},
		func(Context) (any, error) {
			return "hello", nil
		},
		func(Context) (any, error) {
			return nil, errors.New("test error")
		},
		func(Context, map[string]any) (any, error) {
			return "hello", nil
		},
		func(Context, *request017) {},
		func(Context, *request017) any { return nil },
		func(Context, *request017) error {
			return errors.New("test error")
		},
		func(Context, *request017) (any, error) {
			return t, nil
		},
	}
	for i := range funcs {
		app.AnyFunc("/2/"+strconv.Itoa(i+1), funcs[i])
	}

	bindh := http.Header{"Debug": []string{"binderr"}}
	renderh := http.Header{"Debug": []string{"rendererr"}}
	for i := 0; i < len(funcs); i++ {
		app.NewRequest("GET", "/2/"+strconv.Itoa(i+1))
	}
	for i := 0; i < len(funcs); i++ {
		app.NewRequest("GET", "/2/"+strconv.Itoa(i+1), bindh)
	}
	for i := 0; i < len(funcs); i++ {
		app.NewRequest("GET", "/2/"+strconv.Itoa(i+1), renderh)
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

	app.AnyFunc("/1/0", 0)
	app.AnyFunc("/1/1", func(Context, *rpcrequest) (rpcresponse, error) {
		return rpcresponse{Messahe: "success"}, nil
	})
	app.AnyFunc("/1/2", func(Context, map[string]any) (*rpcresponse, error) {
		return nil, errors.New("test rpc error")
	})

	app.NewRequest("PUT", "/1/1")
	app.NewRequest("PUT", "/1/2", http.Header{HeaderAccept: {MimeApplicationJSON}})
	app.NewRequest("PUT", "/1/2", NewClientBodyJSON(map[string]any{
		"name": "eudore",
	}))

	app.NewRequest("GET", "/1/1", http.Header{"Debug": []string{"binderr"}})
	app.NewRequest("GET", "/1/1", http.Header{"Debug": []string{"rendererr"}})

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
