package eudore

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// HandlerDataFunc defines the [Context] data processing function.
//
// Define behaviors such as Bind and Render.
//
// Bind is used to request data parsing.
// By default, [HeaderContentType] is used to select the data parsing method.
//
// Render is used to return response data.
// By default, [HeaderAccept] is used to select the data rendering method.
type HandlerDataFunc = func(Context, any) error

// The NewHandlerDataFuncs function combines multiple [HandlerDataFunc] to
// process data in sequence.
func NewHandlerDataFuncs(handlers ...HandlerDataFunc) HandlerDataFunc {
	switch len(handlers) {
	case 0:
		return nil
	case 1:
		return handlers[0]
	default:
		return func(ctx Context, data any) error {
			for i := range handlers {
				err := handlers[i](ctx, data)
				if err != nil {
					return err
				}
			}
			return nil
		}
	}
}

// NewHandlerDataStatusCode function wraps the response status or code
// when [HandlerDataFunc] returns an error.
//
// If err is [http.MaxBytesError],
// it may override [StatusRequestEntityTooLarge].
func NewHandlerDataStatusCode(handler HandlerDataFunc, status, code int,
) HandlerDataFunc {
	if handler == nil || (status == 0 && code == 0) {
		return handler
	}
	return func(ctx Context, data any) error {
		err := handler(ctx, data)
		if err != nil {
			return NewErrorWithStatusCode(err, status, code)
		}
		return nil
	}
}

// The NewHandlerDataBinds method defines the [HeaderContentType] mapping
// Bind function.
//
// [DefaultHandlerDataBinds] is used by default.
// [HandlerDataBindURL] is used when [HeaderContentType] is empty.
//
// If there is no matching [HandlerDataFunc],
// return [StatusUnsupportedMediaType].
func NewHandlerDataBinds(binds map[string]HandlerDataFunc) HandlerDataFunc {
	if binds == nil {
		binds = mapClone(DefaultHandlerDataBinds)
	}
	var mimes string
	for k := range binds {
		if k != "" && k != MimeApplicationOctetStream {
			mimes += ", " + k
		}
	}
	mimes = strings.TrimPrefix(mimes, ", ")
	return func(ctx Context, data any) error {
		contentType := ctx.GetHeader(HeaderContentType)
		fn, ok := binds[strings.SplitN(contentType, ";", 2)[0]]
		if ok {
			return fn(ctx, data)
		}

		switch ctx.Method() {
		case MethodPost:
			ctx.SetHeader(HeaderAcceptPost, mimes)
		case MethodPatch:
			ctx.SetHeader(HeaderAcceptPatch, mimes)
		}

		err := fmt.Errorf(ErrHandlerDataBindNotSupportContentType, contentType)
		return NewErrorWithStatus(err, StatusUnsupportedMediaType)
	}
}

func bindMaps[T any](source map[string][]T, target any, tags []string) error {
	v := reflect.Indirect(reflect.ValueOf(target))
	switch v.Kind() {
	case reflect.Struct, reflect.Map:
		for key, vals := range source {
			for _, val := range vals {
				err := SetAnyByPath(target, key, val, tags)
				if err != nil && !errors.Is(err, ErrValueNotFound) {
					return err
				}
			}
		}
		return nil
	default:
		// map data is unordered and cannot be bound to an array.
		t := reflect.TypeOf(target).String()
		return fmt.Errorf(ErrHandlerDataBindMustSturct, t)
	}
}

// The HandlerDataBindURL function uses the url parameter to Bind data.
//
// Using tag [DefaultHandlerDataBindURLTags],
// Use the [SetAnyByPathWithTag] function to bind data.
func HandlerDataBindURL(ctx Context, data any) error {
	vals, err := ctx.Querys()
	if err != nil {
		return err
	}
	if len(vals) > 0 {
		return bindMaps(vals, data, DefaultHandlerDataBindURLTags)
	}
	return nil
}

// The HandlerDataBindForm function uses form data to Bind data.
//
// If the request body is empty, use the url parameter.
//
// Using tag [DefaultHandlerDataBindFormTags],
// Use the [SetAnyByPathWithTag] function to bind data.
func HandlerDataBindForm(ctx Context, data any) error {
	vals, err := ctx.FormValues()
	if err != nil {
		return err
	}
	files := ctx.FormFiles()
	if len(files) > 0 {
		_ = bindMaps(files, data, DefaultHandlerDataBindFormTags)
	}
	if len(vals) > 0 {
		return bindMaps(vals, data, DefaultHandlerDataBindFormTags)
	}
	return nil
}

// The HandlerDataBindJSON function uses [json.NewDecoder] to Bind data.
func HandlerDataBindJSON(ctx Context, data any) error {
	return json.NewDecoder(ctx).Decode(data)
}

// The BindXML function uses [xml.NewDecoder] to Bind data.
func HandlerDataBindXML(ctx Context, data any) error {
	return xml.NewDecoder(ctx).Decode(data)
}

// The NewHandlerDataRenders method uses [HeaderAccept] to matching for
// Render functions in renders.
// [DefaultHandlerDataRenders] is used by default.
//
// If Render fails and [ResponseWriter].Size=0, ignore this Render.
func NewHandlerDataRenders(renders map[string]HandlerDataFunc) HandlerDataFunc {
	if renders == nil {
		renders = mapClone(DefaultHandlerDataRenders)
	}
	render, ok := renders[MimeAll]
	if !ok {
		render = HandlerDataRenderNotAcceptable
	}
	return func(ctx Context, data any) error {
		w := ctx.Response()
		h := w.Header()
		contentType, vary := h[HeaderContentType], h[HeaderVary]

		for _, accept := range strings.Split(ctx.GetHeader(HeaderAccept), ",") {
			name, quality, ok := strings.Cut(strings.TrimSpace(accept), ";")
			if ok && quality == "q=0" {
				continue
			}

			fn, ok := renders[name]
			if ok && fn != nil {
				h.Set(HeaderVary,
					strings.Join(append(vary, HeaderAccept), ", "),
				)
				err := fn(ctx, data)
				// Render is successful if return nil
				// Render is irrevocable if size > 0
				if err == nil || w.Size() > 0 {
					return err
				}
				h[HeaderContentType], h[HeaderVary] = contentType, vary
			}
		}
		return render(ctx, data)
	}
}

func renderSetContentType(ctx Context, mime string) {
	h := ctx.Response().Header()
	if val := h.Get(HeaderContentType); len(val) == 0 {
		h.Add(HeaderContentType, mime)
	}
}

func HandlerDataRenderNotAcceptable(ctx Context, _ any) error {
	ctx.WriteHeader(StatusNotAcceptable)
	return nil
}

// RenderText function Render Text, written using the [fmt.Fprint] function.
func HandlerDataRenderText(ctx Context, data any) error {
	renderSetContentType(ctx, MimeTextPlainCharsetUtf8)
	var err error
	switch v := data.(type) {
	case []byte:
		_, err = ctx.Write(v)
	case string:
		_, err = ctx.WriteString(v)
	case fmt.Stringer:
		_, err = ctx.WriteString(v.String())
	default:
		_, err = fmt.Fprintf(ctx, "%#v", data)
	}
	return err
}

// The HandlerDataRenderJSON function uses [json.NewEncoder] to Render data.
//
// If [HeaderAccept] is not [MimeApplicationJSON], use json indent for output.
func HandlerDataRenderJSON(ctx Context, data any) error {
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

// The HandlerDataRenderHTML function creates Render using [template.Template].
//
// patterns will load templates from both [template.ParseFS] and
// [template.ParseFiles].
//
// If [fs.FS] is not empty, the [embed] template will be loaded.
//
// If patterns loads the template from the [os]; each request is reloaded,
// you can use [DefaultHandlerDataTemplateReload] to
// turn off the template automatic reload feature.
//
// When returning an HTML response,
// append [DefaultHandlerDataRenderTemplateHeaders].
//
// If go run is started, workdir may be in a temp directory and the file cannot
// be read.
func NewHandlerDataRenderTemplates(temp *template.Template,
	fs fs.FS, patterns ...string,
) HandlerDataFunc {
	if temp == nil {
		temp = template.New("")
	}
	t, reload, err := parseTemplates(temp, fs, patterns)
	if err != nil {
		return func(ctx Context, _ any) error {
			return renderTemplatesError(ctx, err)
		}
	}
	h := templateHeaders()
	if reload && DefaultHandlerDataTemplateReload {
		return func(ctx Context, data any) error {
			t, _, err := parseTemplates(temp, nil, patterns)
			if err != nil {
				return renderTemplatesError(ctx, err)
			}

			return renderTemplatesData(ctx, data, t, h)
		}
	}
	return func(ctx Context, data any) error {
		return renderTemplatesData(ctx, data, t, h)
	}
}

func templateHeaders() http.Header {
	h := make(http.Header)
	h[HeaderContentType] = []string{MimeTextHTMLCharsetUtf8}
	for k, v := range DefaultHandlerDataRenderTemplateHeaders {
		h.Set(k, v)
	}
	return h
}

func parseTemplates(temp *template.Template, fs fs.FS, patterns []string,
) (*template.Template, bool, error) {
	temp, err := temp.Clone()
	if err != nil {
		return nil, false, err
	}
	size0 := len(temp.Templates())
	if fs != nil && len(patterns) > 0 {
		_, err := temp.ParseFS(fs, patterns...)
		if err != nil {
			return nil, false, err
		}
	}

	size1 := len(temp.Templates())
	for i := range patterns {
		names, err := filepath.Glob(patterns[i])
		if err != nil {
			return nil, false, err
		}
		if len(names) > 0 {
			_, err = temp.ParseFiles(names...)
			if err != nil {
				return nil, false, err
			}
		}
	}
	size2 := len(temp.Templates())

	if patterns != nil && size0 == size2 {
		dir, _ := os.Getwd()
		return nil, false, fmt.Errorf(ErrHandlerDataRenderTemplateNotLoad,
			dir, patterns,
		)
	}

	// append default template
	for _, t := range DefaultHandlerDataRenderTemplateAppend.Templates() {
		if temp.Lookup(t.Name()) == nil {
			_, _ = temp.AddParseTree(t.Name(), t.Tree)
		}
	}
	return temp, size1 != size2, nil
}

func renderTemplatesError(ctx Context, err error) error {
	ctx.Error(err)
	renderSetContentType(ctx, MimeTextPlainCharsetUtf8)
	ctx.WriteHeader(StatusInternalServerError)
	_, _ = ctx.WriteString(err.Error())
	return nil
}

func renderTemplatesData(ctx Context, data any,
	temp *template.Template, hr http.Header,
) error {
	name := ctx.GetParam(ParamTemplate)
	if name == "" {
		var err error
		switch v := data.(type) {
		case []byte:
			renderSetContentType(ctx, MimeTextPlainCharsetUtf8)
			_, err = ctx.Write(v)
		case string:
			renderSetContentType(ctx, MimeTextPlainCharsetUtf8)
			_, err = ctx.WriteString(v)
		case fmt.Stringer:
			renderSetContentType(ctx, MimeTextPlainCharsetUtf8)
			_, err = ctx.WriteString(v.String())
		default:
			return ErrHandlerDataRenderTemplateNeedName
		}
		return err
	}

	t := temp.Lookup(name)
	if t == nil {
		return fmt.Errorf(ErrHandlerDataRenderTemplateNotFound, name)
	}

	hw := ctx.Response().Header()
	for k, v := range hr {
		if hw.Values(k) == nil {
			hw[k] = v
		}
	}
	return t.Execute(ctx, data)
}
