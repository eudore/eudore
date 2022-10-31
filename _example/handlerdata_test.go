package eudore_test

import (
	"context"
	"fmt"
	"html/template"
	"reflect"
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

	app.NewRequest(nil, "GET", "/hello", eudore.NewClientBodyString("trace"), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "GET", "/data/header", eudore.NewClientHeader("name", "eudore"), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "GET", "/data/get-url", eudore.NewClientQuery("name", "eudore"), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "POST", "/data/post-url", eudore.NewClientQuery("name", "eudore"), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "POST", "/data/post-mime", eudore.NewClientQuery("name", "eudore"), eudore.NewClientHeader(eudore.HeaderContentType, "pb"), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "PATCH", "/data/patch-mime", eudore.NewClientQuery("name", "eudore"), eudore.NewClientHeader(eudore.HeaderContentType, "pb"), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "DELETE", "/data/detele-mime", eudore.NewClientQuery("name", "eudore"), eudore.NewClientHeader(eudore.HeaderContentType, "pb"), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "PUT", "/data/json", eudore.NewClientBodyJSONValue("name", "eudore"), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "PUT", "/data/json", eudore.NewClientHeader(eudore.HeaderContentType, eudore.MimeApplicationXML), eudore.NewClientBodyString("<Data><name>eudore</name></Data>"), eudore.NewClientCheckStatus(200))
	app.NewRequest(nil, "PUT", "/data/form", eudore.NewClientBodyFormValue("name", "eudore"), eudore.NewClientCheckStatus(200))

	app.CancelFunc()
	app.Run()
}

func TestHandlerDataRender(*testing.T) {
	type Data struct {
		Name string `json:"name" xml:"name"`
	}

	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyFilte, func(ctx eudore.Context, i interface{}) error {
		if ctx.Path() == "/err" {
			return fmt.Errorf("filte error")
		}
		return nil
	})
	app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))
	app.AnyFunc("/data/* template=data", func(ctx eudore.Context) interface{} {
		return &Data{"eudore"}
	})
	app.AnyFunc("/text/stringer", func(ctx eudore.Context) interface{} {
		return eudore.NewContextKey("text/stringer")
	})
	app.AnyFunc("/text/string template=text", func(ctx eudore.Context) interface{} {
		return "text/string"
	})

	app.NewRequest(nil, "GET", "/err", eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeTextPlain))
	app.NewRequest(nil, "GET", "/data/text", eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeTextPlain))
	app.NewRequest(nil, "GET", "/data/json", eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeApplicationJSON))
	app.NewRequest(nil, "GET", "/data/xml", eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeApplicationXML))
	app.NewRequest(nil, "GET", "/data/html", eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeTextHTML))
	eudore.DefaultRenderHTMLTemplate = nil
	app.NewRequest(nil, "GET", "/data/html", eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeTextHTML))
	app.NewRequest(nil, "GET", "/data/accept")
	app.NewRequest(nil, "GET", "/text/stringer", eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeTextPlain))
	app.NewRequest(nil, "GET", "/text/string", eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeTextPlain))
	app.NewRequest(nil, "GET", "/text/string", eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeApplicationJSON))
	app.NewRequest(nil, "GET", "/text/string", eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeApplicationJSONCharsetUtf8))

	temp, _ := template.New("").Parse(`{{- define "data" -}} Data Name is {{.Name}} {{- end -}}`)
	app.SetValue(eudore.ContextKeyTemplate, temp)

	app.NewRequest(nil, "GET", "/data/html", eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeTextHTML))
	app.NewRequest(nil, "GET", "/text/string", eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeTextHTML))
	app.NewRequest(nil, "GET", "/text/string", eudore.NewClientHeader(eudore.HeaderAccept, eudore.MimeTextHTML))

	app.CancelFunc()
	app.Run()
}

func TestFuncCreator(*testing.T) {
	type Data struct {
		Name string `json:"name" xml:"name"`
	}

	app := eudore.NewApp()

	fc := eudore.NewFuncCreator()
	app.SetValue(eudore.ContextKeyFuncCreator, fc)
	// register error
	fc.Register("test", "not func", TestFuncCreator, func(string) func(string) {
		return nil
	}, func(string) func(string) bool {
		return nil
	}, func(string) func(string) string {
		return nil
	})

	var fn interface{}
	var err error
	typeInt := reflect.TypeOf((*int)(nil)).Elem()
	typeString := reflect.TypeOf((*string)(nil)).Elem()
	typeInterface := reflect.TypeOf((*interface{})(nil)).Elem()

	fc.Create(typeInt, "nozero")
	fc.Create(typeInt, "nozero:")
	{
		// validateIntNozero
		fn, _ = fc.Create(typeInt, "nozero")
		app.Info(fn.(func(int) bool)(0))
	}
	{
		// validateStringNozero
		fn, _ = fc.Create(typeString, "nozero")
		app.Info(fn.(func(string) bool)("123456"))
	}
	{
		// validateInterfaceNozero
		fn, _ = fc.Create(typeInterface, "nozero")
		app.Info(fn.(func(interface{}) bool)("123456"))
		app.Info(fn.(func(interface{}) bool)([]int{1, 2, 3, 4, 5, 6}))
	}
	{
		// validateStringIsnum
		fn, _ = fc.Create(typeString, "isnum")
		app.Info(fn.(func(string) bool)("234"))
		app.Info(fn.(func(string) bool)("xx2"))
	}
	{
		// validateNewIntMin
		fc.Create(typeInt, "min=xx")
		fn, _ = fc.Create(typeInt, "min=033")
		app.Info(fn.(func(int) bool)(12))
		app.Info(fn.(func(int) bool)(644))
	}
	{
		// validateNewIntMax
		fc.Create(typeInt, "max")
		fc.Create(typeInt, "max=xx")
		fn, _ = fc.Create(typeInt, "max=033")
		app.Info(fn.(func(int) bool)(12))
		app.Info(fn.(func(int) bool)(644))
	}
	{
		// validateNewStringMin
		fc.Create(typeString, "min=xx")
		fn, _ = fc.Create(typeString, "min=033")
		app.Info(fn.(func(string) bool)("12"))
		app.Info(fn.(func(string) bool)("644"))
		app.Info(fn.(func(string) bool)("xx"))
	}
	{
		// validateNewStringMax
		fc.Create(typeString, "max=xx")
		fn, _ = fc.Create(typeString, "max=033")
		app.Info(fn.(func(string) bool)("12"))
		app.Info(fn.(func(string) bool)("644"))
		app.Info(fn.(func(string) bool)("xx"))
	}
	{
		// validateNewStringLen
		fc.Create(typeString, "len>x")
		fn, _ = fc.Create(typeString, "len>5")
		app.Info(fn.(func(string) bool)("8812988"))
		app.Info(fn.(func(string) bool)("123"))
		app.Info(fn.(func(string) bool)("123456"))
		fn, _ = fc.Create(typeString, "len<5")
		app.Info(fn.(func(string) bool)("8812988"))
		app.Info(fn.(func(string) bool)("123"))
		app.Info(fn.(func(string) bool)("123456"))
		fn, _ = fc.Create(typeString, "len=5")
		app.Info(fn.(func(string) bool)("8812988"))
		app.Info(fn.(func(string) bool)("123"))
		app.Info(fn.(func(string) bool)("123456"))
	}
	{
		// validateNewInterfaceLen
		fc.Create(typeInterface, "len=.")
		fn, _ = fc.Create(typeInterface, "len>4")
		app.Info(fn.(func(interface{}) bool)("123456"))
		app.Info(fn.(func(interface{}) bool)([]int{1, 2, 3, 4}))
		app.Info(fn.(func(interface{}) bool)(6))
		fn, _ = fc.Create(typeInterface, "len<4")
		app.Info(fn.(func(interface{}) bool)("123456"))
		app.Info(fn.(func(interface{}) bool)([]int{1, 2, 3, 4}))
		app.Info(fn.(func(interface{}) bool)(6))
		fn, _ = fc.Create(typeInterface, "len=4")
		app.Info(fn.(func(interface{}) bool)("123456"))
		app.Info(fn.(func(interface{}) bool)([]int{1, 2, 3, 4}))
		app.Info(fn.(func(interface{}) bool)(6))
	}
	{
		// validateNewStringRegexp
		_, err = fc.Create(typeString, "regexp^[($")
		app.Info(err)
		fn, _ = fc.Create(typeString, "regexp^\\d+$")
		app.Info(fn.(func(string) bool)("123456"))
	}

	app.CancelFunc()
	app.Run()
}

func TestHandlerDataValidateField(*testing.T) {
	eudore.NewValidateField(context.Background())
	type DataValidate01 struct {
		Name   string `json:"name" xml:"name" validate:"len>4"`
		Email  string `json:"email" xml:"email" validate:"email"`
		Phone  string `json:"phone" xml:"phone" validate:"phone"`
		Level1 string `json:"level1" xml:"level1" validate:""`
	}

	type DataValidate02 struct {
		Name   string `json:"name" xml:"name" validate:"len>4"`
		Email  string `json:"email" xml:"email" validate:"email"`
		Level1 string `json:"level1" xml:"level1" validate:"is"`
	}

	app := eudore.NewApp()
	fc := eudore.NewFuncCreator()
	app.SetValue(eudore.ContextKeyFuncCreator, fc)
	app.SetValue(eudore.ContextKeyValidate, eudore.NewValidateField(app))
	app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))

	fc.Register("email", func(email string) bool {
		return strings.HasSuffix(email, "@eudore.cn")
	})
	fc.Register("phone", func(phone string) bool {
		return len(phone) == 11
	})

	app.AnyFunc("/data/1", func(ctx eudore.Context) {
		var data DataValidate01
		ctx.Bind(&data)
	})
	app.AnyFunc("/data/2", func(ctx eudore.Context) {
		var data []DataValidate01
		ctx.Bind(&data)
	})
	app.AnyFunc("/data/3", func(ctx eudore.Context) {
		var data DataValidate02
		ctx.Bind(&data)
	})
	app.AnyFunc("/data/4", func(ctx eudore.Context) {
		var data []DataValidate02
		ctx.Bind(&data)
	})
	app.AnyFunc("/data/5", func(ctx eudore.Context) {
		var data []*DataValidate02
		ctx.Bind(&data)
	})
	app.AnyFunc("/data/7", func(ctx eudore.Context) {
		var data map[string]interface{}
		ctx.Bind(&data)
	})

	app.NewRequest(nil, "POST", "/data/1", eudore.NewClientBodyJSON(&DataValidate01{Name: "eudore", Email: "postmaster@eudore.cn", Phone: "15512344321"}))
	app.NewRequest(nil, "POST", "/data/2", eudore.NewClientBodyJSON([]DataValidate01{{Name: "eudore"}}))
	app.NewRequest(nil, "POST", "/data/3", eudore.NewClientBodyJSON(&DataValidate02{Name: "eudore"}))
	app.NewRequest(nil, "POST", "/data/4", eudore.NewClientBodyJSON([]*DataValidate02{{Name: "eudore"}}))
	app.NewRequest(nil, "POST", "/data/5", eudore.NewClientBodyJSON([]*DataValidate02{{Name: "eudore"}, {Name: "eudore"}}))
	app.NewRequest(nil, "POST", "/data/7", eudore.NewClientBodyJSON(&DataValidate02{Name: "eudore"}))

	app.CancelFunc()
	app.Run()
}
