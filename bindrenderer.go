package eudore

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"reflect"
	"strings"
)

/*
Binder & Renderer

Binder对象用于请求数据反序列化，默认根据http请求的Content-Type header指定的请求数据格式来解析数据。

Renderer对象更加Accept Header选择数据对象序列化的方法。
*/

// Binder 定义Bind函数处理请求。
type Binder func(Context, io.Reader, interface{}) error

// BindDefault 函数实现默认Binder。
func BindDefault(ctx Context, r io.Reader, i interface{}) error {
	if ctx.Method() == MethodGet || ctx.Method() == MethodHead {
		return BindURL(ctx, r, i)
	}
	switch strings.SplitN(ctx.GetHeader(HeaderContentType), ";", 2)[0] {
	case MimeApplicationJSON:
		return BindJSON(ctx, r, i)
	case MimeApplicationForm:
		return BindURL(ctx, r, i)
	case MimeMultipartForm:
		return BindForm(ctx, r, i)
	case MimeTextXML, MimeApplicationXML:
		return BindXML(ctx, r, i)
	default:
		return fmt.Errorf(ErrFormatBindDefaultNotSupportContentType, ctx.GetHeader(HeaderContentType))
	}
}

// BindURL 函数使用url参数实现bind。
func BindURL(ctx Context, _ io.Reader, i interface{}) error {
	return ConvertToWithTags(ctx.Querys(), i, DefaultConvertFormTags)
}

// BindForm 函数使用form格式body实现bind。
func BindForm(ctx Context, _ io.Reader, i interface{}) error {
	ConvertToWithTags(ctx.FormFiles(), i, DefaultConvertFormTags)
	return ConvertToWithTags(ctx.FormValues(), i, DefaultConvertFormTags)
}

// BindJSON 函数使用json格式body实现bind。
func BindJSON(_ Context, r io.Reader, i interface{}) error {
	return json.NewDecoder(r).Decode(i)
}

// BindXML 函数使用xml格式body实现bind。
func BindXML(_ Context, r io.Reader, i interface{}) error {
	return xml.NewDecoder(r).Decode(i)
}

// BindHeader 函数实现使用header数据bind。
func BindHeader(ctx Context, _ io.Reader, i interface{}) error {
	return ConvertToWithTags(ctx.Request().Header, i, DefaultConvertFormTags)
}

// NewBinderHeader 实现Binder额外封装bind header。
func NewBinderHeader(fn Binder) Binder {
	return func(ctx Context, r io.Reader, i interface{}) error {
		BindHeader(ctx, r, i)
		return fn(ctx, r, i)
	}
}

// NewBinderURL 实现Binder在非get和head方法下实现BindURL。
func NewBinderURL(fn Binder) Binder {
	return func(ctx Context, r io.Reader, i interface{}) error {
		if ctx.Method() != MethodGet && ctx.Method() != MethodHead {
			BindURL(ctx, r, i)
		}
		return fn(ctx, r, i)
	}
}

// NewBinderValidater 实现Binder会执行Validate。
func NewBinderValidater(fn Binder) Binder {
	return func(ctx Context, r io.Reader, i interface{}) error {
		err := fn(ctx, r, i)
		if err == nil {
			err = ctx.Validate(i)
		}
		return err
	}
}

// Renderer 接口定义根据请求接受的数据类型来序列化数据。
type Renderer func(Context, interface{}) error

// RenderDefault 函数是默认Render，更加Accent Header选择Json、Xml、Text三种Render。
func RenderDefault(ctx Context, data interface{}) error {
	for _, accept := range strings.Split(ctx.GetHeader(HeaderAccept), ",") {
		switch strings.TrimSpace(accept) {
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
	if s, ok := data.(string); ok {
		return ctx.WriteString(s)
	}
	_, err := fmt.Fprintf(ctx, "%#v", data)
	return err
}

type renderJSONData struct {
	Status     int         `json:"status"`
	XRequestID string      `json:"x-request-id,omitempty"`
	Message    interface{} `json:"message"`
}

// RenderJSON 函数使用encoding/json库实现json反序列化。
//
// 如果请求Accept不为"application/json"，使用json格式化输出。
func RenderJSON(ctx Context, data interface{}) error {
	header := ctx.Response().Header()
	if val := header.Get(HeaderContentType); len(val) == 0 {
		header.Add(HeaderContentType, MimeApplicationJSONUtf8)
	}
	switch reflect.Indirect(reflect.ValueOf(data)).Kind() {
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
	default:
		data = &renderJSONData{
			Status:     ctx.Response().Status(),
			XRequestID: ctx.GetHeader(HeaderXRequestID),
			Message:    data,
		}
	}
	encoder := json.NewEncoder(ctx)
	if !strings.Contains(ctx.GetHeader(HeaderAccept), MimeApplicationJSON) {
		encoder.SetIndent("", "\t")
	}
	return encoder.Encode(data)
}

// RenderXML 函数Render Xml，使用encoding/xml库实现xml反序列化。
func RenderXML(ctx Context, data interface{}) error {
	header := ctx.Response().Header()
	if val := header.Get(HeaderContentType); len(val) == 0 {
		header.Add(HeaderContentType, MimeApplicationxmlCharsetUtf8)
	}
	return xml.NewEncoder(ctx).Encode(data)
}

// NewRenderHTML 函数使用模板创建一个模板Renderer
func NewRenderHTML(temp *template.Template) Renderer {
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

// NewRenderHTMLWithTemplate 函数使用Renderer支持html Renderer。
func NewRenderHTMLWithTemplate(r Renderer, temp *template.Template) Renderer {
	htmlRender := NewRenderHTML(temp)
	return func(ctx Context, data interface{}) error {
		path := ctx.GetParam("template")
		if path != "" && strings.Contains(ctx.GetHeader(HeaderAccept), MimeTextHTML) {
			return htmlRender(ctx, data)
		}
		return r(ctx, data)
	}
}
