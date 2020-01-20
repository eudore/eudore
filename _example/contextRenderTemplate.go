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
)

var viewpath = "view/index.html"
var viewdata = map[string]interface{}{
	"name":    "eudore",
	"message": "hello eudore",
}

func main() {
	app := eudore.NewCore()
	httptest.NewClient(app).Stop(0)
	app.Renderer = eudore.NewHTMLRenderWithTemplate(app.Renderer, nil)
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
	app.Listen(":8088")
	app.Run()
}
