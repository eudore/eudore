package eudore_test

import (
	"context"
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

	client := eudore.NewClientWarp()
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyClient, client)
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

	client.NewRequest("GET", "/hello").BodyString("trace").Do().Callback(eudore.NewResponseReaderCheckStatus(200))
	client.NewRequest("GET", "/data/header").AddHeader("name", "eudore").Do().Callback(eudore.NewResponseReaderCheckStatus(200))
	client.NewRequest("GET", "/data/get-url").AddQuery("name", "eudore").Do().Callback(eudore.NewResponseReaderCheckStatus(200))
	client.NewRequest("POST", "/data/post-url").AddQuery("name", "eudore").Do().Callback(eudore.NewResponseReaderCheckStatus(200))
	client.NewRequest("PUT", "/data/json").BodyJSONValue("name", "eudore").Do().Callback(eudore.NewResponseReaderCheckStatus(200))
	client.NewRequest("PUT", "/data/json").AddHeader(eudore.HeaderContentType, eudore.MimeTextXML).BodyString("<Data><name>eudore</name></Data>").Do().Callback(eudore.NewResponseReaderCheckStatus(200))
	client.NewRequest("PUT", "/data/form").BodyFormValue("name", "eudore").Do().Callback(eudore.NewResponseReaderCheckStatus(200))

	app.CancelFunc()
	app.Run()
}

func TestHandlerDataRender(*testing.T) {
	type Data struct {
		Name string `json:"name" xml:"name"`
	}

	client := eudore.NewClientWarp()
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyClient, client)
	app.AnyFunc("/data/* template=data", func(ctx eudore.Context) interface{} {
		return &Data{"eudore"}
	})
	app.AnyFunc("/text/stringer", func(ctx eudore.Context) interface{} {
		return eudore.NewContextKey("text/stringer")
	})
	app.AnyFunc("/text/string template=text", func(ctx eudore.Context) interface{} {
		return "text/string"
	})

	client.NewRequest("GET", "/data/text").AddHeader(eudore.HeaderAccept, eudore.MimeText).Do()
	client.NewRequest("GET", "/data/json").AddHeader(eudore.HeaderAccept, eudore.MimeApplicationJSON).Do()
	client.NewRequest("GET", "/data/xml").AddHeader(eudore.HeaderAccept, eudore.MimeApplicationXML).Do()
	client.NewRequest("GET", "/data/html").AddHeader(eudore.HeaderAccept, eudore.MimeTextHTML).Do()
	client.NewRequest("GET", "/text/stringer").AddHeader(eudore.HeaderAccept, eudore.MimeText).Do()
	client.NewRequest("GET", "/text/string").AddHeader(eudore.HeaderAccept, eudore.MimeText).Do()
	client.NewRequest("GET", "/text/string").AddHeader(eudore.HeaderAccept, eudore.MimeApplicationJSON).Do()

	temp, _ := template.New("").Parse(`{{- define "data" -}} Data Name is {{.Name}} {{- end -}}`)
	app.SetValue(eudore.ContextKeyTempldate, temp)

	client.NewRequest("GET", "/data/html").AddHeader(eudore.HeaderAccept, eudore.MimeTextHTML).Do()
	client.NewRequest("GET", "/text/string").AddHeader(eudore.HeaderAccept, eudore.MimeTextHTML).Do()
	client.NewRequest("GET", "/text/string").AddHeader(eudore.HeaderAccept, eudore.MimeTextHTML).Do()

	app.CancelFunc()
	app.Run()
}

func TestFuncCreator(*testing.T) {
	type Data struct {
		Name string `json:"name" xml:"name"`
	}

	client := eudore.NewClientWarp()
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyClient, client)

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
	client := eudore.NewClientWarp()
	fc := eudore.NewFuncCreator()
	app.SetValue(eudore.ContextKeyClient, client)
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

	client.NewRequest("POST", "/data/1").BodyJSON(&DataValidate01{Name: "eudore", Email: "postmaster@eudore.cn", Phone: "15512344321"}).Do()
	client.NewRequest("POST", "/data/2").BodyJSON([]DataValidate01{{Name: "eudore"}}).Do()
	client.NewRequest("POST", "/data/3").BodyJSON(&DataValidate02{Name: "eudore"}).Do()
	client.NewRequest("POST", "/data/4").BodyJSON([]*DataValidate02{{Name: "eudore"}}).Do()
	client.NewRequest("POST", "/data/5").BodyJSON([]*DataValidate02{{Name: "eudore"}, {Name: "eudore"}}).Do()
	client.NewRequest("POST", "/data/7").BodyJSON(&DataValidate02{Name: "eudore"}).Do()

	app.CancelFunc()
	app.Run()
}
