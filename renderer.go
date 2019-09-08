package eudore

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"path/filepath"
	"strings"
)

// Renderer 接口定义根据请求接受的数据类型来序列化数据。
type Renderer func(Context, interface{}) error

// RenderDefault 函数是默认Render，更加Accent Header选择Json、Xml、Text三种Render。
func RenderDefault(ctx Context, data interface{}) error {
	for _, accept := range strings.Split(ctx.GetHeader(HeaderAccept), ",") {
		if accept != "" && accept[0] == ' ' {
			accept = accept[1:]
		}
		switch accept {
		case MimeApplicationJSON:
			return RenderJSON(ctx, data)
		case MimeApplicationXML, MimeTextXML:
			return RenderXML(ctx, data)
		case MimeTextPlain, MimeText:
			return RenderText(ctx, data)
		}
	}
	return RenderText(ctx, data)
}

// RenderText 函数Render Text，使用fmt.Fprint函数写入。
func RenderText(ctx Context, data interface{}) error {
	header := ctx.Response().Header()
	if val := header.Get(HeaderContentType); len(val) == 0 {
		header.Add(HeaderContentType, MimeTextPlainCharsetUtf8)
	}
	_, err := fmt.Fprint(ctx, data)
	return err
}

// RenderJSON 函数Render Json，使用encoding/json库实现json反序列化。
func RenderJSON(ctx Context, data interface{}) error {
	header := ctx.Response().Header()
	if val := header.Get(HeaderContentType); len(val) == 0 {
		header.Add(HeaderContentType, MimeApplicationJSONUtf8)
	}
	return json.NewEncoder(ctx).Encode(data)
}

// RenderIndentJSON 函数Render Indent Json，使用encoding/json库实现json反序列化。
func RenderIndentJSON(ctx Context, data interface{}) error {
	header := ctx.Response().Header()
	if val := header.Get(HeaderContentType); len(val) == 0 {
		header.Add(HeaderContentType, MimeApplicationJSONUtf8)
	}
	en := json.NewEncoder(ctx)
	en.SetIndent("", "\t")
	return en.Encode(data)
}

// RenderXML 函数Render Xml，使用encoding/xml库实现xml反序列化。
func RenderXML(ctx Context, data interface{}) error {
	header := ctx.Response().Header()
	if val := header.Get(HeaderContentType); len(val) == 0 {
		header.Add(HeaderContentType, MimeApplicationxmlCharsetUtf8)
	}
	return xml.NewEncoder(ctx).Encode(data)
}

// NewHTMLRender 函数使用模板创建一个模板Renderer
func NewHTMLRender(temp *template.Template) Renderer {
	if temp == nil {
		temp = template.Must(template.New("").Parse(""))
	}
	return func(ctx Context, data interface{}) error {
		path := ctx.GetParam("template")
		t, err := template.Must(temp.Clone()).New(filepath.Base(path)).ParseFiles(path)
		if err != nil {
			return err
		}

		// 添加header
		header := ctx.Response().Header()
		if val := header.Get(HeaderContentType); len(val) == 0 {
			header.Add(HeaderContentType, MimeTextHTMLCharsetUtf8)
		}

		return t.Execute(ctx, data)
	}
}

// NewHTMLWithRender 函数使用Renderer支持html Renderer。
func NewHTMLWithRender(r Renderer, temp *template.Template) Renderer {
	htmlRender := NewHTMLRender(temp)
	return func(ctx Context, data interface{}) error {
		path := ctx.GetParam("template")
		if path != "" && strings.Contains(ctx.GetHeader(HeaderAccept), MimeTextHTML) {
			return htmlRender(ctx, data)
		}
		return r(ctx, data)
	}
}
