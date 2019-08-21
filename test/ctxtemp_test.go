package test

/*
import (
	"github.com/eudore/eudore"
	"html/template"
	"testing"
)

type (
	Request1 struct {
		Field1 int `set:"field1"`
		Field2 string
		Field3 bool `set:"field3"`
	}
)

func TestCacheMap(*testing.T) {
	// create template tree
	const tpl = `webpage: {{.}}{{define "T1"}}ONE{{end}}{{Func1 "Version"}}
`
	var funcs = make(template.FuncMap)
	funcs["Func1"] = func(a string) string {
		return "func1: " + a
	}

	t, _ := template.New("webpage").Funcs(funcs).Parse(tpl)
	t.AddParseTree("ss", template.Must(template.New("ss").Parse(`ss: {{.}}{{template "T1"}}`)).Tree)

	eudore.RegisterHandlerFunc(func(fn func(ctx eudore.Context, tmp *template.Template)) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			fn(ctx, t)
		}
	})

	app := eudore.NewCore()
	app.GetFunc("/ template:webpage", ContextTemplatre)
	app.GetFunc("/get template:ss", ContextTemplatre)
	app.GetFunc("/render template:/tmp/05.tmp", render)
	// eudore.TestAppRequest(app, "GET", "/", nil).Show()
	// eudore.TestAppRequest(app, "GET", "/get", nil).Show()
	eudore.TestAppRequest(app, "GET", "/render", nil).Show()
	app.Listen(":8084")
	app.Run()
}

func ContextTemplatre(ctx eudore.Context, tmp *template.Template) {
	data := struct {
		Title string
		Items []string
	}{
		Title: "My page",
		Items: []string{
			"My photos",
			"My blog",
		},
	}
	tmp.Lookup(ctx.GetParam("template")).Execute(ctx, data)
}

func render(ctx eudore.Context) {
	data := struct {
		Title string
		Items []string
	}{
		Title: "My page",
		Items: []string{
			"My photos",
			"My blog",
		},
	}
	ctx.WriteRender(data)
}
*/
