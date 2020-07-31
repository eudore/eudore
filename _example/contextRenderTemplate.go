package main

/*
eudore默认不支持render html，需要设置render支持。
当请求header "Accept: text/html"时，才会调用render html，否在按照accept header来进行render。
可以强制设置Accent header的值强制使用render html
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"html/template"
	"os"
)

var viewpath = "index.html"
var viewdata = map[string]interface{}{
	"name":    "eudore",
	"message": "hello eudore",
}

func main() {
	content := []byte(`name: {{.name}} message: {{.message}}`)
	tmpfile, _ := os.Create(viewpath)
	defer os.Remove(tmpfile.Name())
	tmpfile.Write(content)

	app := eudore.NewApp()
	app.Renderer = eudore.NewRenderHTMLWithTemplate(app.Renderer, nil)
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.SetParam("template", viewpath)
		ctx.Render(viewdata)
	})
	app.AnyFunc("/2/*path", func(ctx eudore.Context) interface{} {
		ctx.SetParam("template", viewpath)
		return viewdata
	})
	app.AnyFunc("/template/*", func(ctx eudore.Context) error {
		t, err := template.ParseFiles(viewpath)
		if err != nil {
			return err
		}
		return t.Execute(ctx, viewdata)
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/1").WithHeaderValue("Accept", "application/json").Do().Out()
	client.NewRequest("GET", "/1").WithHeaderValue("Accept", eudore.MimeTextHTML).Do().Out()
	client.NewRequest("GET", "/1").WithHeaderValue("Accept", eudore.MimeTextPlain).Do().Out()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
