package main

/*
eudore默认没有render html api，通过设置render支持。
当请求header "Accept: text/html"时，才会调用render html，否在按照Accept header顺序来选择render。
使用NewHandlerDataRenderTemplates时，必须存在Param template定义模板名称。
可以强制设置Accent header的值强制使用render html
*/

import (
	"embed"
	"html/template"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

var viewdata = map[string]interface{}{
	"name":    "eudore",
	"message": "hello eudore",
}

var tempcontent = `{{- define "index.html" -}}
name: {{.name}} message: {{.message}}
<script>
fetch("/",{headers: {Accept:"application/json"}})
.then((response) => response.text())
.then((data) => {
	const dom = document.createElement("div");
	dom.textContent = data;
	document.body.appendChild(dom);
})
</script>
{{- end -}}`

//go:embed */*
var roottemplate embed.FS

func main() {
	temp := template.Must(template.New("").Parse(tempcontent))

	app := eudore.NewApp()
	// 重新设置Render，参考eudore.DefaultHandlerDataRenders定义，修改MimeTextHTML的Render函数s使用自定义模板。
	app.SetValue(eudore.ContextKeyRender, eudore.NewHandlerDataRenders(map[string]eudore.HandlerDataFunc{
		eudore.MimeTextHTML:        eudore.NewHandlerDataRenderTemplates(temp, roottemplate, "**/*.tmpl"),
		eudore.MimeApplicationJSON: eudore.HandlerDataRenderJSON,
	}))
	app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))
	app.AddMiddleware(middleware.NewLoggerFunc(app))
	// 设置渲染使用模板名称。
	app.AnyFunc("/*path template=index.html", func(ctx eudore.Context) {
		ctx.Render(viewdata)
	})
	app.AnyFunc("/template/*", func(ctx eudore.Context) interface{} {
		ctx.SetParam("template", ctx.GetParam("*"))
		return viewdata
	})

	app.Listen(":8088")
	app.Run()
}
