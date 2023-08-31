package eudore

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"reflect"
	"strings"
)

// HandlerDataFunc defines the request context data processing function.
//
// Define four behaviors of Bind Validater Filter Render by default.
//
// Binder object is used to request data deserialization,
// By default, data is parsed according to the request data format specified
// by the Content-Type header of the http request.
//
// The Renderer object accepts the header to select the data object serialization method.
// HandlerDataFunc 定义请求上下文数据处理函数。
//
// 默认定义Bind Validater Filter Render四种行为。
//
// Binder对象用于请求数据反序列化，
// 默认根据http请求的Content-Type header指定的请求数据格式来解析数据。
//
// Renderer对象更加Accept Header选择数据对象序列化的方法。
type HandlerDataFunc = func(Context, any) error

// The NewBinds method defines the ContentType Header mapping Bind function.
//
// NewBinds 方法定义ContentType Header映射Bind函数。
func NewBinds(binds map[string]HandlerDataFunc) HandlerDataFunc {
	if binds == nil {
		binds = map[string]HandlerDataFunc{
			MimeApplicationJSON:     BindJSON,
			MimeApplicationForm:     BindURL,
			MimeMultipartForm:       BindForm,
			MimeApplicationProtobuf: BindProtobuf,
			MimeApplicationXML:      BindXML,
		}
	}
	var mimes string
	for k := range binds {
		mimes += ", " + k
	}
	mimes = strings.TrimPrefix(mimes, ", ")
	return func(ctx Context, i any) error {
		contentType := ctx.GetHeader(HeaderContentType)
		if contentType == "" {
			return BindURL(ctx, i)
		}
		fn, ok := binds[strings.SplitN(contentType, ";", 2)[0]]
		if ok {
			return fn(ctx, i)
		}
		ctx.WriteHeader(StatusUnsupportedMediaType)
		switch ctx.Method() {
		case MethodPost:
			ctx.SetHeader(HeaderAcceptPost, mimes)
		case MethodPatch:
			ctx.SetHeader(HeaderAcceptPatch, mimes)
		}
		return fmt.Errorf(ErrFormatBindDefaultNotSupportContentType, contentType)
	}
}

// NewBindWithHeader implements Bind to additionally encapsulate bind header.
//
// NewBindWithHeader 实现Bind额外封装bind header。
func NewBindWithHeader(fn HandlerDataFunc) HandlerDataFunc {
	return func(ctx Context, i any) error {
		BindHeader(ctx, i)
		return fn(ctx, i)
	}
}

// NewBindWithURL implements Bind and also executes BindURL
// when HeaderContentType is not empty.
//
// NewBindWithURL 实现Bind在HeaderContentType非空时也执行BindURL。
func NewBindWithURL(fn HandlerDataFunc) HandlerDataFunc {
	return func(ctx Context, i any) error {
		if ctx.GetHeader(HeaderContentType) != "" {
			BindURL(ctx, i)
		}
		return fn(ctx, i)
	}
}

func bindMaps[T any](data map[string][]T, i any, tags []string) error {
	for key, vals := range data {
		for _, val := range vals {
			SetAnyByPathWithTag(i, key, val, tags, false)
		}
	}
	return nil
}

// The BindURL function uses the url parameter to parse the binding body.
//
// BindURL 函数使用url参数解析绑定body。
func BindURL(ctx Context, i any) error {
	return bindMaps(ctx.Querys(), i, DefaultHandlerBindURLTags)
}

// The BindForm function uses form to parse and bind the body.
//
// BindForm 函数使用form解析绑定body。
func BindForm(ctx Context, i any) error {
	bindMaps(ctx.FormFiles(), i, DefaultHandlerBindFormTags)
	return bindMaps(ctx.FormValues(), i, DefaultHandlerBindFormTags)
}

// The BindJSON function uses encoding/json to parse and bind the body.
//
// BindJSON 函数使用encoding/json解析绑定body。
func BindJSON(ctx Context, i any) error {
	return json.NewDecoder(ctx).Decode(i)
}

// The BindXML function uses encoding/xml to parse the bound body.
//
// BindXML 函数使用encoding/xml解析绑定body。
func BindXML(ctx Context, i any) error {
	return xml.NewDecoder(ctx).Decode(i)
}

// The BindProtobuf function uses the built-in protobu to parse the bound body.
//
// BindProtobuf 函数使用内置protobu解析绑定body。
func BindProtobuf(ctx Context, i any) error {
	return NewProtobufDecoder(ctx).Decode(i)
}

// The BindHeader function implements binding using header data.
//
// The header name prefix must be 'X-', example: X-Euduore-Name => Eudore.Name.
//
// BindHeader 函数实现使用header数据bind。
//
// header名称前缀必须是'X-'，example: X-Euduore-Name => Eudore.Name。
func BindHeader(ctx Context, i any) error {
	for key, vals := range ctx.Request().Header {
		if strings.HasPrefix(key, "X-") {
			key = strings.ReplaceAll(key[2:], "-", ".")
			for _, val := range vals {
				SetAnyByPathWithTag(i, key, val, DefaultHandlerBindHeaderTags, false)
			}
		}
	}
	return nil
}

// The NewRenders method defines the default HeaderAccept value mapping Render function.
//
// The HeaderAccept value ignores non-zero weight values, and the order takes precedence.
//
// NewRenders 方法定义默认HeaderAccept值映射Render函数。
//
// HeaderAccept值忽略非零权重值，顺序优先。
func NewRenders(renders map[string]HandlerDataFunc) HandlerDataFunc {
	if renders == nil {
		renders = map[string]HandlerDataFunc{
			MimeText:                RenderText,
			MimeTextPlain:           RenderText,
			MimeTextHTML:            RenderHTML,
			MimeApplicationJSON:     RenderJSON,
			MimeApplicationProtobuf: RenderProtobuf,
			MimeApplicationXML:      RenderXML,
		}
	}
	render, ok := renders["*"]
	if !ok {
		render = DefaultHandlerRenderFunc
	}
	return func(ctx Context, i any) error {
		for _, accept := range strings.Split(ctx.GetHeader(HeaderAccept), ",") {
			name, quality, ok := strings.Cut(strings.TrimSpace(accept), ";")
			if ok && quality == "q=0" {
				continue
			}

			fn, ok := renders[name]
			if ok && fn != nil {
				h := ctx.Response().Header()
				v := h.Values(HeaderVary)
				h.Set(HeaderVary, strings.Join(append(v, HeaderAccept), ", "))
				err := fn(ctx, i)
				if !errors.Is(err, ErrRenderHandlerSkip) {
					return err
				}
				h[HeaderVary] = v
			}
		}
		return render(ctx, i)
	}
}

func renderSetContentType(ctx Context, mime string) {
	header := ctx.Response().Header()
	if val := header.Get(HeaderContentType); len(val) == 0 {
		header.Add(HeaderContentType, mime)
	}
}

// The RenderJSON function uses the encoding/json library to implement json deserialization.
//
// If the request Accept is not "application/json", output in json indent format.
//
// RenderJSON 函数使用encoding/json库实现json反序列化。
//
// 如果请求Accept不为"application/json"，使用json indent格式输出。
func RenderJSON(ctx Context, data any) error {
	renderSetContentType(ctx, MimeApplicationJSONCharsetUtf8)
	switch reflect.Indirect(reflect.ValueOf(data)).Kind() {
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
	default:
		data = NewContextMessgae(ctx, nil, data)
	}
	encoder := json.NewEncoder(ctx)
	if !strings.Contains(ctx.GetHeader(HeaderAccept), MimeApplicationJSON) {
		encoder.SetIndent("", "\t")
	}
	return encoder.Encode(data)
}

// RenderXML function Render Xml,
// using the encoding/xml library to realize xml deserialization.
//
// RenderXML 函数Render Xml，使用encoding/xml库实现xml反序列化。
func RenderXML(ctx Context, data any) error {
	renderSetContentType(ctx, MimeApplicationXMLCharsetUtf8)
	return xml.NewEncoder(ctx).Encode(data)
}

// RenderText function Render Text, written using the fmt.Fprint function.
//
// RenderText 函数Render Text，使用fmt.Fprint函数写入。
func RenderText(ctx Context, data any) error {
	renderSetContentType(ctx, MimeTextPlainCharsetUtf8)
	if s, ok := data.(string); ok {
		ctx.WriteString(s)
		return nil
	}
	if s, ok := data.(fmt.Stringer); ok {
		ctx.WriteString(s.String())
		return nil
	}
	_, err := fmt.Fprintf(ctx, "%#v", data)
	return err
}

// RenderProtobuf function Render Protobuf,
// using the built-in protobuf encoding, invalid properties will be ignored.
//
// RenderProtobuf 函数Render Protobuf，使用内置protobuf编码，无效属性将忽略。
func RenderProtobuf(ctx Context, i any) error {
	renderSetContentType(ctx, MimeApplicationProtobuf)
	return NewProtobufEncoder(ctx).Encode(i)
}

// The RenderHTML function creates a template Renderer using a template.
//
// Load *template.Template from ctx.Value(eudore.ContextKeyTemplate),
// Load the template function from ctx.GetParam("template").
//
// RenderHTML 函数使用模板创建一个模板Renderer。
//
// 从ctx.Value(eudore.ContextKeyTemplate)加载*template.Template，
// 从ctx.GetParam("template")加载模板函数。
func RenderHTML(ctx Context, data any) error {
	tpl, ok := ctx.Value(ContextKeyTemplate).(*template.Template)
	if !ok {
		return ErrRenderHandlerSkip
	}

	name := ctx.GetParam("template")
	if name == "" {
		// 默认模板
		name = DefaultTemplateNameRenderData
		b := bytes.NewBuffer([]byte{})
		en := json.NewEncoder(b)
		en.SetEscapeHTML(false)
		en.SetIndent("", "\t")
		err := en.Encode(data)
		if err != nil {
			b.WriteString(err.Error())
		}
		data = map[string]any{
			"Method": ctx.Method(),
			"Host":   ctx.Host(),
			"Path":   ctx.Request().RequestURI,
			"Query":  ctx.Querys(),
			"Status": fmt.Sprintf("%d %s",
				ctx.Response().Status(),
				http.StatusText(ctx.Response().Status()),
			),
			"RemoteAddr":     ctx.Host(),
			"LocalAddr":      ctx.RealIP(),
			"Params":         ctx.Params(),
			"RequestHeader":  ctx.Request().Header,
			"ResponseHeader": ctx.Response().Header(),
			"Data":           b.String(),
			"GodocServer":    DefaultGodocServer,
			"TraceServer":    DefaultTraceServer,
		}
	}

	tpl = tpl.Lookup(name)
	if tpl == nil {
		return ErrRenderHandlerSkip
	}

	renderSetContentType(ctx, MimeTextHTMLCharsetUtf8)
	return tpl.Execute(ctx, data)
}
