package eudore

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"strings"
)

type (
	// Renderer 接口定义根据请求接受的数据类型来序列化数据。
	Renderer interface {
		Render(Context, interface{}) error
	}
	rendererText       struct{}
	rendererJson       struct{}
	rendererIndentJson struct{}
	rendererXml        struct{}
	rendererTemplate   struct {
		root  *template.Template
		funcs template.FuncMap
	}
	rendererDefault struct{}
)

// 定义默认Renderer对象
var (
	RenderDefault  = rendererDefault{}
	RendererText   = rendererText{}
	RenderTemplate = rendererTemplate{
		root: template.New(""),
	}
	RendererJson       = rendererJson{}
	RendererIndentJson = rendererIndentJson{}
	RendererXml        = rendererXml{}
)

func (rendererDefault) Render(ctx Context, data interface{}) error {
	for _, accept := range strings.Split(ctx.GetHeader(HeaderAccept), ",") {
		if accept != "" && accept[0] == ' ' {
			accept = accept[1:]
		}
		switch accept {
		case MimeApplicationJson:
			return RendererJson.Render(ctx, data)
		case MimeApplicationXml, MimeTextXml:
			return RendererXml.Render(ctx, data)
		case MimeTextPlain, MimeText:
			return RendererText.Render(ctx, data)
		case MimeTextHTML:
			temp := ctx.GetParam(ParamTemplate)
			if len(temp) > 0 {
				return RenderTemplate.Render(ctx, data)
			}
		}
	}
	return RendererText.Render(ctx, data)
}

func (rendererText) Render(ctx Context, data interface{}) error {
	header := ctx.Response().Header()
	if val := header.Get(HeaderContentType); len(val) == 0 {
		header.Add(HeaderContentType, MimeTextPlainCharsetUtf8)
	}
	_, err := fmt.Fprint(ctx, data)
	return err
}

func (rendererJson) Render(ctx Context, data interface{}) error {
	header := ctx.Response().Header()
	if val := header.Get(HeaderContentType); len(val) == 0 {
		header.Add(HeaderContentType, MimeApplicationJsonUtf8)
	}
	return json.NewEncoder(ctx).Encode(data)
}

func (rendererIndentJson) Render(ctx Context, data interface{}) error {
	header := ctx.Response().Header()
	if val := header.Get(HeaderContentType); len(val) == 0 {
		header.Add(HeaderContentType, MimeApplicationJsonUtf8)
	}
	en := json.NewEncoder(ctx)
	en.SetIndent("", "\t")
	return en.Encode(data)
}

func (rendererXml) Render(ctx Context, data interface{}) error {
	header := ctx.Response().Header()
	if val := header.Get(HeaderContentType); len(val) == 0 {
		header.Add(HeaderContentType, MimeApplicationxmlCharsetUtf8)
	}
	return xml.NewEncoder(ctx).Encode(data)
}

func (r rendererTemplate) Render(ctx Context, data interface{}) error {
	header := ctx.Response().Header()
	if val := header.Get(HeaderContentType); len(val) == 0 {
		header.Add(HeaderContentType, MimeTextHTMLCharsetUtf8)
	}
	path := ctx.GetParam(ParamTemplate)
	t := r.root.Lookup(path)
	if t == nil {
		var err error
		t, err = r.loadTemplate(path)
		if err != nil {
			return err
		}
	}
	fmt.Println("---h98dfh9", t)
	return t.Execute(ctx, data)
}

// loadTemplate 给根模板加载一个子模板。
func (r rendererTemplate) loadTemplate(path string) (*template.Template, error) {
	t, err := template.New(path).Funcs(r.funcs).ParseFiles(path)
	fmt.Println(err)
	if err != nil {
		return nil, err
	}
	return r.root.AddParseTree(path, t.Tree)
}
