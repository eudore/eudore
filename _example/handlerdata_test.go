package eudore_test

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	. "github.com/eudore/eudore"
)

func TestHandlerDataFuncs(*testing.T) {
	NewHandlerDataFuncs()
	NewHandlerDataFuncs(HandlerDataBindJSON)
	NewHandlerDataStatusCode(nil, 400, 1000)
}

func TestHandlerDataBind(*testing.T) {
	type Data struct {
		Name string `alias:"name" json:"name" xml:"name"`
		Int  int    `alias:"int" json:"int" xml:"int"`
	}

	app := NewApp()
	app.SetValue(ContextKeyBind, NewHandlerDataStatusCode(
		NewHandlerDataBinds(nil), 400, 1000,
	))
	app.SetValue(ContextKeyRender, HandlerDataRenderJSON)
	app.SetValue(ContextKeyContextPool, NewContextBasePool(app))
	app.GetFunc("/hello", func(ctx Context) {
		ctx.WriteString("hello eudore")
	})
	app.GetFunc("/bind/err", func(ctx Context) {
		ctx.Request().URL.RawQuery = "tag=%\007"
		var data Data
		ctx.Bind(&data)
	})
	app.AnyFunc("/data/*", func(ctx Context) (interface{}, error) {
		var datas []Data
		ctx.Bind(&datas)
		var data Data
		err := ctx.Bind(&data)
		if err != nil {
			return nil, err
		}
		return &data, nil
	})

	form := NewClientBodyForm(nil)
	form.AddFile("file", "app", []byte("form body"))
	pb := "\u0010\u0006eudore"

	app.NewRequest("GET", "/hello", strings.NewReader("trace"))
	app.NewRequest("GET", "/bind/err")
	app.NewRequest("GET", "/bind/err",
		http.Header{HeaderContentType: {MimeApplicationForm}},
	)
	app.NewRequest("GET", "/data/header", http.Header{"X-Name": {"eudore"}})
	app.NewRequest("GET", "/data/get-url",
		url.Values{"name": {"eudore"}, "int": {"str"}},
	)
	app.NewRequest("POST", "/data/post-url", url.Values{"name": {"eudore"}})
	app.NewRequest("POST", "/data/post-mime", url.Values{"name": {"eudore"}},
		http.Header{HeaderContentType: {"pb"}}, strings.NewReader(pb),
	)
	app.NewRequest("PATCH", "/data/patch-mime", url.Values{"name": {"eudore"}},
		http.Header{HeaderContentType: {"pb"}}, strings.NewReader(pb),
	)
	app.NewRequest("DELETE", "/data/detele-mime", url.Values{"name": {"eudore"}},
		http.Header{HeaderContentType: {"pb"}}, strings.NewReader(pb),
	)
	app.NewRequest("PUT", "/data/json",
		NewClientBodyJSON(url.Values{"name": {"eudore"}}),
	)
	app.NewRequest("PUT", "/data/xml",
		NewClientBodyXML(&Data{"eudore", 0}),
	)
	app.NewRequest("PUT", "/data/url",
		NewClientBodyForm(url.Values{"name": {"eudore"}}),
	)
	app.NewRequest("PUT", "/data/form", form)
	app.NewRequest("PUT", "/data/protobuf", strings.NewReader(pb))
	app.NewRequest("PUT", "/data/protobuf",
		http.Header{HeaderContentType: {MimeApplicationProtobuf}},
		strings.NewReader(pb),
	)

	app.CancelFunc()
	app.Run()
}

func TestHandlerDataRender(*testing.T) {
	type Data struct {
		Name string `json:"name" xml:"name"`
	}

	app := NewApp()
	app.SetValue(ContextKeyRender, NewHandlerDataFuncs(
		func(ctx Context, i any) error {
			if ctx.Path() == "/err" {
				return fmt.Errorf("filte error")
			}
			return nil
		},
		NewHandlerDataRenders(nil),
	))
	app.SetValue(ContextKeyContextPool, NewContextBasePool(app))
	app.AnyFunc("/data/*", func(ctx Context) interface{} {
		return &Data{"eudore"}
	})
	app.AnyFunc("/html/err", func(ctx Context) interface{} {
		return &struct{ Name func() }{}
	})
	app.AnyFunc("/html/* template=data", func(ctx Context) interface{} {
		return &Data{"eudore"}
	})
	app.AnyFunc("/text/stringer", func(ctx Context) interface{} {
		return NewContextKey("text/stringer")
	})
	app.AnyFunc("/text/bytes template=text", func(ctx Context) interface{} {
		return []byte("text/bytes")
	})
	app.AnyFunc("/text/string template=text", func(ctx Context) interface{} {
		return "text/string"
	})

	accept := func(val string) http.Header {
		return http.Header{HeaderAccept: {val}}
	}

	app.NewRequest("GET", "/err", accept(MimeTextPlain))
	app.NewRequest("GET", "/data/quality", accept(MimeTextPlain+";q=0"))
	app.NewRequest("GET", "/data/text", accept(MimeTextPlain))
	app.NewRequest("GET", "/data/json", accept(MimeApplicationJSON))
	app.NewRequest("GET", "/data/xml", accept(MimeApplicationXML))
	app.NewRequest("GET", "/data/html", accept(MimeTextHTML))
	app.NewRequest("GET", "/data/protobuf", accept(MimeApplicationProtobuf))
	app.NewRequest("GET", "/html/err", accept(MimeTextHTML))
	app.NewRequest("GET", "/html/html", accept(MimeTextHTML))
	app.NewRequest("GET", "/data/accept")

	app.NewRequest("GET", "/text/stringer", accept(MimeTextPlain))
	app.NewRequest("GET", "/text/bytes", accept(MimeTextPlain))
	app.NewRequest("GET", "/text/string", accept(MimeTextPlain))
	app.NewRequest("GET", "/text/string", accept(MimeApplicationJSON))
	app.NewRequest("GET", "/text/string", accept(MimeApplicationJSONCharsetUtf8))

	app.CancelFunc()
	app.Run()
}

//go:embed handlerdata_test.go
var handlerdatafile embed.FS

func TestHandlerDataRenderTemplates(*testing.T) {
	tt, _ := template.New("").Parse("")
	tt.Execute(os.Stdout, nil)
	NewHandlerDataRenderTemplates(nil, nil)
	NewHandlerDataRenderTemplates(nil, handlerdatafile, "none")
	NewHandlerDataRenderTemplates(nil, handlerdatafile, "handlerdata_test.go")
	NewHandlerDataRenderTemplates(nil, nil, "none")
	NewHandlerDataRenderTemplates(nil, nil, "[invalid-pattern")
	NewHandlerDataRenderTemplates(nil, nil, "handlerdata_test.go")
	NewHandlerDataRenderTemplates(tt, nil)

	filepath := "handlerdata.tmpl"
	tmpfile, _ := os.Create(filepath)
	defer os.Remove(tmpfile.Name())
	tmpfile.WriteString(`{{- define "index.html" -}}name: {{.name}} message: {{.message}}{{end}}`)
	renders := [...]HandlerDataFunc{
		NewHandlerDataRenderTemplates(nil, handlerdatafile, "handlerdata_test.go"),
		NewHandlerDataRenderTemplates(nil, nil, filepath),
		NewHandlerDataRenderTemplates(nil, nil),
	}

	app := NewApp()
	app.GetFunc("/renders/:index", func(ctx Context) {
		index := GetAny[int](ctx.GetParam("index")) % 3
		renders[index](ctx, "body")
		if index == 1 {
			tmpfile.WriteString(`{{end}}`)
		}
	})
	app.GetFunc("/string/:index", func(ctx Context) {
		index := GetAny[int](ctx.GetParam("index")) % 3
		data := []any{[]byte("not found"), "not allow", time.Now()}
		renders[2](ctx, data[index])
	})

	app.NewRequest("GET", "/renders/0")
	app.NewRequest("GET", "/renders/1")
	app.NewRequest("GET", "/renders/1")
	app.NewRequest("GET", "/renders/1")
	app.NewRequest("GET", "/renders/2")
	app.NewRequest("GET", "/string/0")
	app.NewRequest("GET", "/string/1")
	app.NewRequest("GET", "/string/2")

	app.CancelFunc()
	app.Run()
}

type dataValidate01 struct {
	ID     *int   `json:"id" xml:"id" valid:"nozero,omitempty"`
	Child  []int  `json:"child" xml:"child" valid:"nozero,omitempty"`
	Name   string `json:"name" xml:"name" valid:"nozero,len>4"`
	Level1 string `json:"level1" xml:"level1" valid:"-"`
}
type dataValidate02 struct {
	ID     *int   `json:"id" xml:"id" valid:"nozero"`
	Name   string `json:"name" xml:"name" valid:"len>4"`
	Level1 string `json:"level1" xml:"level1"`
}
type dataValidate03 struct {
	ID int `json:"id" xml:"id" valid:"(nozero),,"`
}
type dataValidate04 struct {
	ID int `json:"id" xml:"id" valid:"not"`
}
type dataValidate05 struct {
	dataValidate03
	*dataValidate04
}

func (dataValidate03) Validate(context.Context) error {
	return fmt.Errorf("test error validate")
}

func TestHandlerDataValidate(*testing.T) {
	app := NewApp()
	app.SetValue(ContextKeyBind,
		NewHandlerDataValidate(),
	)
	app.SetValue(ContextKeyContextPool, NewContextBasePool(app))

	app.AnyFunc("/data/struct", func(ctx Context) {
		var data dataValidate03
		ctx.Bind(&data)
	})
	app.AnyFunc("/data/slice", func(ctx Context) {
		data := []dataValidate03{}
		ctx.Bind(&data)
		data = []dataValidate03{{}, {}, {}}
		ctx.Bind(&data)
	})

	app.NewRequest("GET", "/data/struct")
	app.NewRequest("GET", "/data/slice")

	app.CancelFunc()
	app.Run()
}

func TestHandlerDataValidateStruct(*testing.T) {
	app := NewApp()
	app.SetValue(ContextKeyBind, NewHandlerDataFuncs(
		NewHandlerDataBinds(nil),
		NewHandlerDataValidateStruct(app),
	))
	app.SetValue(ContextKeyContextPool, NewContextBasePool(app))

	app.AnyFunc("/data/struct1", func(ctx Context) {
		var data dataValidate01
		ctx.Bind(&data)
	})
	app.AnyFunc("/data/slice1", func(ctx Context) {
		var data []dataValidate01
		ctx.Bind(&data)
	})
	app.AnyFunc("/data/ptr1", func(ctx Context) {
		var data []*dataValidate01
		ctx.Bind(&data)
	})
	app.AnyFunc("/data/any1", func(ctx Context) {
		var data []any = []any{new(dataValidate01)}
		ctx.Bind(&data)
	})
	app.AnyFunc("/data/struct2", func(ctx Context) {
		var data dataValidate02
		ctx.Bind(&data)
	})
	app.AnyFunc("/data/struct3", func(ctx Context) {
		var data dataValidate03
		ctx.Bind(&data)
	})
	app.AnyFunc("/data/struct4", func(ctx Context) {
		var data dataValidate04
		ctx.Bind(&data)
	})
	app.AnyFunc("/data/struct5", func(ctx Context) {
		var data dataValidate05
		ctx.Bind(&data)
	})

	fn := func(name string, val any) {
		app.NewRequest("POST", "/data/"+name, NewClientBodyJSON(val))
	}

	id := 4
	fn("struct1", dataValidate01{})
	fn("struct1", dataValidate01{ID: &id})
	fn("slice1", []dataValidate01{{Name: "A1"}})
	fn("slice1", []dataValidate01{{Name: "eudore", Child: []int{0, 0, 0}}})
	fn("slice1", []dataValidate01{{Name: "eudore", Child: []int{1, 2, 3}}})
	fn("ptr1", []dataValidate01{{Name: "eudore"}})
	fn("any1", []dataValidate01{{Name: "eudore"}})
	fn("struct2", dataValidate02{})
	fn("struct5", dataValidate03{ID: 32})
	fn("struct4", dataValidate04{})
	fn("struct4", dataValidate04{})
	fn("struct3", dataValidate05{})

	app.CancelFunc()
	app.Run()
}

func TestHandlerDataFilterRule(*testing.T) {
	type LoggerConfig struct {
		Stdout   bool   `json:"stdout" xml:"stdout" alias:"stdout"`
		Path     string `json:"path" xml:"path" alias:"path"`
		Handlers []any  `json:"-" xml:"-" alias:"handlers"`
		Chan     chan int
	}
	type FilterType struct {
		String string  `alias:"string"`
		Int    int     `alias:"int"`
		Uint   uint    `alias:"uint"`
		Float  float64 `alias:"float"`
		Bool   bool    `alias:"bool"`
		Any    any     `alias:"any"`
	}

	app := NewApp()
	fc := NewFuncCreator()
	app.SetValue(ContextKeyFuncCreator, fc)
	app.SetValue(ContextKeyRender, NewHandlerDataFuncs(
		NewHandlerDataFilter(app),
		func(ctx Context, i any) error {
			ctx.Debugf("%#v", i)
			return nil
		},
	))
	app.SetValue(ContextKeyContextPool, NewContextBasePool(app))

	app.GetFunc("/data", func(ctx Context) {
		data := []any{
			[]string{"path=zero"},
			&LoggerConfig{Stdout: true},
			&FilterRule{
				Name:    "*",
				Checks:  []string{"path=zero"},
				Modifys: []string{"stdout=value:true"},
			},
			&LoggerConfig{},
			&FilterRule{
				Name:    "LoggerConfig",
				Package: "eudore",
				Checks:  []string{"path=zero"},
			},
			&LoggerConfig{},
			&FilterRule{Name: "Logger*", Checks: []string{"path=zero"}},
			&LoggerConfig{},
			&FilterRule{Name: "App", Checks: []string{"path=zero"}},
			&LoggerConfig{},
			&FilterRule{Name: "App*", Checks: []string{"path=zero"}},
			&LoggerConfig{},
			&FilterRule{Name: "Logger*1", Checks: []string{"path=zero"}},
			&LoggerConfig{},
			[]FilterRule{{Checks: []string{"path=zero"}}},
			&LoggerConfig{},
			&FilterRule{Checks: []string{"path=zero"}},
			[]*LoggerConfig{{}, {Path: "app.log"}},
			&FilterRule{Checks: []string{"path=zero"}},
			[]any{LoggerConfig{}, &LoggerConfig{Path: "app.log"}},
			[]string{"link=zero"},
			&LoggerConfig{},
			[]string{"Chan=k"},
			&LoggerConfig{},
			[]string{"handlers=k"},
			&LoggerConfig{},
			&FilterRule{
				Checks: []string{
					"string=zero", "int=zero",
					"uint=zero", "float=zero",
					"bool=zero", "any=zero",
				},
				Modifys: []string{
					"string=now:20060102", "int=value:4",
					"uint=value:4", "float=value:4",
					"bool=value:true", "any=now",
				},
			},
			&FilterType{},
			&FilterRule{
				Modifys: []string{"any=default"},
			},
			&FilterType{},
		}

		for i := 0; i < len(data); i += 2 {
			ctx.SetValue(ContextKeyFilterRules, data[i])
			ctx.Render(data[i+1])
		}
	})
	app.NewRequest("GET", "/data")

	meta, ok := fc.(interface{ Metadata() any }).Metadata().(MetadataFuncCreator)
	if ok {
		for _, err := range meta.Errors {
			app.Debug("err:", err)
		}
	}
	app.CancelFunc()
	app.Run()
}
