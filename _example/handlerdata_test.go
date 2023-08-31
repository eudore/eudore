package eudore_test

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func TestHandlerDataBind(*testing.T) {
	type Data struct {
		Name string `json:"name" xml:"name"`
	}

	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyBind, eudore.NewBindWithHeader(eudore.NewBindWithURL(eudore.NewBinds(nil))))
	app.SetValue(eudore.ContextKeyRender, eudore.RenderJSON)
	app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))
	app.AddMiddleware("global", middleware.NewLoggerFunc(app, "route"))
	app.GetFunc("/hello", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore")
	})
	app.AnyFunc("/data/*", func(ctx eudore.Context) (interface{}, error) {
		var data Data
		err := ctx.Bind(&data)
		if err != nil {
			return nil, err
		}
		return &data, nil
	})

	form := eudore.NewClientBodyForm(nil)
	form.AddFile("file", "app", []byte("form body"))

	app.NewRequest(nil, "GET", "/hello", strings.NewReader("trace"))
	app.NewRequest(nil, "GET", "/data/header", http.Header{"X-Name": {"eudore"}})
	app.NewRequest(nil, "GET", "/data/get-url", url.Values{"name": {"eudore"}})
	app.NewRequest(nil, "POST", "/data/post-url", url.Values{"name": {"eudore"}})
	app.NewRequest(nil, "POST", "/data/post-mime", url.Values{"name": {"eudore"}}, http.Header{eudore.HeaderContentType: {"pb"}})
	app.NewRequest(nil, "PATCH", "/data/patch-mime", url.Values{"name": {"eudore"}}, http.Header{eudore.HeaderContentType: {"pb"}})
	app.NewRequest(nil, "DELETE", "/data/detele-mime", url.Values{"name": {"eudore"}}, http.Header{eudore.HeaderContentType: {"pb"}})
	app.NewRequest(nil, "PUT", "/data/json", eudore.NewClientBodyJSON(url.Values{"name": {"eudore"}}))
	app.NewRequest(nil, "PUT", "/data/xml", eudore.NewClientBodyXML(&Data{"eudore"}))
	app.NewRequest(nil, "PUT", "/data/url", eudore.NewClientBodyForm(url.Values{"name": {"eudore"}}))
	app.NewRequest(nil, "PUT", "/data/form", form)
	app.NewRequest(nil, "PUT", "/data/protobuf", http.Header{eudore.HeaderContentType: {eudore.MimeApplicationProtobuf}})

	app.CancelFunc()
	app.Run()
}

func TestHandlerDataRender(*testing.T) {
	type Data struct {
		Name string `json:"name" xml:"name"`
	}

	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyFilter, func(ctx eudore.Context, i interface{}) error {
		if ctx.Path() == "/err" {
			return fmt.Errorf("filte error")
		}
		return nil
	})
	app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))
	app.AnyFunc("/data/*", func(ctx eudore.Context) interface{} {
		return &Data{"eudore"}
	})
	app.AnyFunc("/html/err", func(ctx eudore.Context) interface{} {
		return &struct{ Name func() }{}
	})
	app.AnyFunc("/html/* template=data", func(ctx eudore.Context) interface{} {
		return &Data{"eudore"}
	})
	app.AnyFunc("/text/stringer", func(ctx eudore.Context) interface{} {
		return eudore.NewContextKey("text/stringer")
	})
	app.AnyFunc("/text/string template=text", func(ctx eudore.Context) interface{} {
		return "text/string"
	})

	accept := func(val string) http.Header {
		return http.Header{eudore.HeaderAccept: {val}}
	}

	app.NewRequest(nil, "GET", "/err", accept(eudore.MimeTextPlain))
	app.NewRequest(nil, "GET", "/data/quality", accept(eudore.MimeTextPlain+";q=0"))
	app.NewRequest(nil, "GET", "/data/text", accept(eudore.MimeTextPlain))
	app.NewRequest(nil, "GET", "/data/json", accept(eudore.MimeApplicationJSON))
	app.NewRequest(nil, "GET", "/data/xml", accept(eudore.MimeApplicationXML))
	app.NewRequest(nil, "GET", "/data/html", accept(eudore.MimeTextHTML))
	app.NewRequest(nil, "GET", "/data/protobuf", accept(eudore.MimeApplicationProtobuf))
	app.NewRequest(nil, "GET", "/html/err", accept(eudore.MimeTextHTML))
	app.NewRequest(nil, "GET", "/html/html", accept(eudore.MimeTextHTML))
	app.NewRequest(nil, "GET", "/data/accept")

	app.NewRequest(nil, "GET", "/text/stringer", accept(eudore.MimeTextPlain))
	app.NewRequest(nil, "GET", "/text/string", accept(eudore.MimeTextPlain))
	app.NewRequest(nil, "GET", "/text/string", accept(eudore.MimeApplicationJSON))
	app.NewRequest(nil, "GET", "/text/string", accept(eudore.MimeApplicationJSONCharsetUtf8))

	temp, _ := template.New("").Parse(`{{- define "data" -}} Data Name is {{.Name}} {{- end -}}`)
	app.SetValue(eudore.ContextKeyTemplate, temp)
	app.NewRequest(nil, "GET", "/data/html", accept(eudore.MimeTextHTML))
	app.NewRequest(nil, "GET", "/text/string", accept(eudore.MimeTextHTML))
	app.NewRequest(nil, "GET", "/text/string", accept(eudore.MimeTextHTML))

	app.SetValue(eudore.ContextKeyTemplate, nil)
	app.NewRequest(nil, "GET", "/data/html", accept(eudore.MimeTextHTML))

	app.CancelFunc()
	app.Run()
}

type dataValidate01 struct {
	ID     *int   `json:"id" xml:"id" validate:"nozero,omitempty"`
	Child  []int  `json:"child" xml:"child" validate:"nozero,omitempty"`
	Name   string `json:"name" xml:"name" validate:"nozero,len>4"`
	Level1 string `json:"level1" xml:"level1" validate:"-"`
}
type dataValidate02 struct {
	ID     *int   `json:"id" xml:"id" validate:"nozero"`
	Name   string `json:"name" xml:"name" validate:"len>4"`
	Level1 string `json:"level1" xml:"level1"`
}
type dataValidate03 struct {
	ID int `json:"id" xml:"id" validate:"(nozero),,"`
}
type dataValidate04 struct {
	ID int `json:"id" xml:"id" validate:"not"`
}
type dataValidate05 struct {
	dataValidate03
	*dataValidate04
}

func (dataValidate03) Validate(context.Context) error {
	return fmt.Errorf("test error validate")
}

func TestHandlerDataValidateField(*testing.T) {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyValidater, eudore.NewValidateField(app))
	app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))

	app.AddMiddleware(middleware.NewLoggerFunc(app))
	app.AnyFunc("/data/struct1", func(ctx eudore.Context) {
		var data dataValidate01
		ctx.Bind(&data)
	})
	app.AnyFunc("/data/slice1", func(ctx eudore.Context) {
		var data []dataValidate01
		ctx.Bind(&data)
	})
	app.AnyFunc("/data/ptr1", func(ctx eudore.Context) {
		var data []*dataValidate01
		ctx.Bind(&data)
	})
	app.AnyFunc("/data/any1", func(ctx eudore.Context) {
		var data []any = []any{new(dataValidate01)}
		ctx.Bind(&data)
	})
	app.AnyFunc("/data/struct2", func(ctx eudore.Context) {
		var data dataValidate02
		ctx.Bind(&data)
	})
	app.AnyFunc("/data/struct3", func(ctx eudore.Context) {
		var data dataValidate03
		ctx.Bind(&data)
	})
	app.AnyFunc("/data/struct4", func(ctx eudore.Context) {
		var data dataValidate04
		ctx.Bind(&data)
	})
	app.AnyFunc("/data/struct5", func(ctx eudore.Context) {
		var data dataValidate05
		ctx.Bind(&data)
	})

	fn := func(name string, val any) {
		app.NewRequest(nil, "POST", "/data/"+name, eudore.NewClientBodyJSON(val))
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

	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyRender, eudore.HandlerDataFunc(func(ctx eudore.Context, i any) error {
		ctx.Debugf("%#v", i)
		return nil
	}))
	fc := eudore.NewFuncCreator()
	app.SetValue(eudore.ContextKeyFuncCreator, fc)
	app.SetValue(eudore.ContextKeyFilter, eudore.NewFilterRules(app))
	app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))
	app.AddMiddleware(middleware.NewLoggerFunc(app))

	app.GetFunc("/data", func(ctx eudore.Context) {
		data := []any{
			[]string{"path=zero"},
			&LoggerConfig{Stdout: true},
			&eudore.FilterData{
				Name:    "*",
				Checks:  []string{"path=zero"},
				Modifys: []string{"stdout=value:true"},
			},
			&LoggerConfig{},
			&eudore.FilterData{Name: "LoggerConfig", Package: "eudore", Checks: []string{"path=zero"}},
			&LoggerConfig{},
			&eudore.FilterData{Name: "Logger*", Checks: []string{"path=zero"}},
			&LoggerConfig{},
			&eudore.FilterData{Name: "App", Checks: []string{"path=zero"}},
			&LoggerConfig{},
			&eudore.FilterData{Name: "App*", Checks: []string{"path=zero"}},
			&LoggerConfig{},
			&eudore.FilterData{Name: "Logger*1", Checks: []string{"path=zero"}},
			&LoggerConfig{},
			[]eudore.FilterData{{Checks: []string{"path=zero"}}},
			&LoggerConfig{},
			&eudore.FilterData{Checks: []string{"path=zero"}},
			[]*LoggerConfig{{}, {Path: "app.log"}},
			&eudore.FilterData{Checks: []string{"path=zero"}},
			[]any{LoggerConfig{}, &LoggerConfig{Path: "app.log"}},
			[]string{"link=zero"},
			&LoggerConfig{},
			[]string{"Chan=k"},
			&LoggerConfig{},
			[]string{"handlers=k"},
			&LoggerConfig{},
			&eudore.FilterData{
				Checks:  []string{"string=zero", "int=zero", "uint=zero", "float=zero", "bool=zero", "any=zero"},
				Modifys: []string{"string=now:20060102", "int=value:4", "uint=value:4", "float=value:4", "bool=value:true", "any=now"},
			},
			&FilterType{},
			&eudore.FilterData{
				Modifys: []string{"any=default"},
			},
			&FilterType{},
		}

		for i := 0; i < len(data); i += 2 {
			ctx.SetValue(eudore.ContextKeyFilterRules, data[i])
			ctx.Render(data[i+1])
		}
	})
	app.NewRequest(nil, "GET", "/data")

	meta, ok := fc.(interface{ Metadata() any }).Metadata().(eudore.MetadataFuncCreator)
	if ok {
		for _, err := range meta.Errors {
			app.Debug("err:", err)
		}
	}
	app.CancelFunc()
	app.Run()
}
