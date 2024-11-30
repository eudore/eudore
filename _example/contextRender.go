package main

/*
Render会根据请求的Accept header觉得使用哪种方式写入返回数据。

设置了三种ContextKeyRender值：
	第一种设置是内置的默认渲染，根据Accept选择方法，未匹配使用eudore.DefaultHandlerRenderFunc。
	第二种设置为强制渲染JSON，HandlerDataRenderJSON的Accept为MimeApplicationJSON时渲染JSON格式，否则选择格式化JSON格式。
	第三种设置是默认渲染参数未nil时，使用的默认参数。

最后设置ContextKeyContextPool修改Pool的参数，使修改的渲染生效。
*/

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyRender, eudore.NewHandlerDataRenders(nil))
	app.SetValue(eudore.ContextKeyRender, eudore.HandlerDataRenderJSON)
	app.SetValue(eudore.ContextKeyRender, eudore.NewHandlerDataRenders(map[string]eudore.HandlerDataFunc{
		eudore.MimeText:                eudore.HandlerDataRenderText,
		eudore.MimeTextPlain:           eudore.HandlerDataRenderText,
		eudore.MimeTextHTML:            eudore.NewHandlerDataRenderTemplates(nil, nil),
		eudore.MimeApplicationJSON:     eudore.HandlerDataRenderJSON,
		eudore.MimeApplicationProtobuf: eudore.HandlerDataRenderProtobuf,
	}))
	app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))
	app.AnyFunc("/*", func(ctx eudore.Context) any {
		type renderData struct {
			Name    string `xml:"name" json:"name"`
			Message string `xml:"message" json:"message"`
		}
		return renderData{
			Name:    "eudore",
			Message: "hello eudore",
		}
	})
	// 访问首页可以看到5种不同Accept值，返回不同格式数据。
	app.GetFunc("/", func(ctx eudore.Context) {
		ctx.SetHeader(eudore.HeaderContentType, eudore.MimeTextHTML)
		ctx.WriteString(index)
	})

	app.Listen(":8088")
	app.Run()
}

var index = `
<script>"use strict";
function accept(v){
	fetch(v, {headers: {Accept:v}}).then((resp)=>resp.text()).then((txt)=>{
		console.log(txt)
		const elem = document.createElement("pre");
		elem.innerText = "accept: " + v + " message: " + txt;
		document.querySelector('body').appendChild(elem);	
	})
}
accept('undefined')
accept('text/plain')
accept('application/json')
accept('application/xml')
accept('application/protobuf')
</script>`
