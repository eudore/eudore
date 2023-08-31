package middleware

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/eudore/eudore"
)

// NewLookFunc 函数创建一个访问对象数据处理函数。
//
// 如果参数类型为func(eudore.Context) any，可以动态返回需要渲染的数据。
//
// 获取请求路由参数"*"为object访问路径，返回object指定属性的数据，允许使用下列参数：
//
//	d=10 depth递归显时最大层数;
//	all=false 是否显时非导出属性;
//	format=html/json/text 设置数据显示格式;
//	godoc=https://golang.org 设置html格式链接的godoc服务地址;
//	width=60 设置html格式缩进宽度。
func NewLookFunc(data any) eudore.HandlerFunc {
	fn, ok := data.(func(eudore.Context) any)
	if !ok {
		fn = func(eudore.Context) any {
			return data
		}
	}
	return func(ctx eudore.Context) {
		data := fn(ctx)
		if data == nil || ctx.Response().Size() > 0 {
			return
		}
		ctx.SetHeader(eudore.HeaderXEudoreAdmin, "look")
		look := NewLookValue(ctx)
		val, err := eudore.GetAnyByPathWithValue(data, strings.ReplaceAll(ctx.GetParam("*"), "/", "."), nil, look.ShowAll)
		if err != nil {
			ctx.Fatal(err)
			return
		}
		look.Scan(val)

		switch getRequestForma(ctx) {
		case QueryFormatJSON:
			ctx.SetHeader(eudore.HeaderContentType, eudore.MimeApplicationJSONCharsetUtf8)
			_ = eudore.RenderJSON(ctx, look)
		case QueryFormatHTML:
			tmpl := getLookTemplate(strings.TrimSuffix(ctx.Path(), "/"), ctx.Querys().Encode())
			ctx.SetHeader(eudore.HeaderContentType, eudore.MimeTextHTMLCharsetUtf8)
			_ = tmpl.ExecuteTemplate(ctx, "view", viewData{ctx.GetParam("*"), eudore.GetAnyByString(ctx.GetQuery("width"), 60), look})
		default:
			tmpl := getLookTemplate(strings.TrimSuffix(ctx.Path(), "/"), ctx.Querys().Encode())
			ctx.SetHeader(eudore.HeaderContentType, eudore.MimeTextPlainCharsetUtf8)
			_ = tmpl.ExecuteTemplate(ctx, "text", look)
		}
	}
}

// NewBindLook 函数创建支持look value的eudore.NewRenders。
func NewBindLook(renders map[string]eudore.HandlerDataFunc) eudore.HandlerDataFunc {
	data := map[string]eudore.HandlerDataFunc{
		eudore.MimeApplicationJSON:     eudore.RenderJSON,
		eudore.MimeApplicationProtobuf: eudore.RenderProtobuf,
		eudore.MimeApplicationXML:      eudore.RenderXML,
		eudore.MimeTextHTML:            eudore.RenderHTML,
		eudore.MimeTextPlain:           eudore.RenderText,
		MimeValueJSON:                  RenderValueJSON,
		MimeValueHTML:                  RenderValueHTML,
		MimeValueText:                  RenderValueText,
	}
	for k, v := range renders {
		if v == nil {
			delete(data, k)
		} else {
			data[k] = v
		}
	}
	return eudore.NewRenders(data)
}

// RenderValueJSON 实现渲染Value为JSON格式。
func RenderValueJSON(ctx eudore.Context, data any) error {
	look := NewLookValue(ctx)
	look.Scan(reflect.ValueOf(data))

	header := ctx.Response().Header()
	if val := header.Get(eudore.HeaderContentType); len(val) == 0 {
		header.Add(eudore.HeaderContentType, eudore.MimeApplicationJSONCharsetUtf8)
	}
	encoder := json.NewEncoder(ctx)
	if !strings.Contains(ctx.GetHeader(eudore.HeaderAccept), eudore.MimeApplicationJSON) {
		encoder.SetIndent("", "\t")
	}
	return encoder.Encode(look)
}

// RenderValueText 实现渲染Value为Text格式。
func RenderValueText(ctx eudore.Context, data any) error {
	look := NewLookValue(ctx)
	look.Scan(reflect.ValueOf(data))

	header := ctx.Response().Header()
	if val := header.Get(eudore.HeaderContentType); len(val) == 0 {
		ctx.SetHeader(eudore.HeaderContentType, eudore.MimeTextPlainCharsetUtf8)
	}

	tmpl := getLookTemplate(strings.TrimSuffix(ctx.Path(), "/"), ctx.Querys().Encode())
	return tmpl.ExecuteTemplate(ctx, "text", look)
}

// RenderValueHTML 实现渲染Value为HTML格式。
func RenderValueHTML(ctx eudore.Context, data any) error {
	look := NewLookValue(ctx)
	look.Scan(reflect.ValueOf(data))

	header := ctx.Response().Header()
	if val := header.Get(eudore.HeaderContentType); len(val) == 0 {
		ctx.SetHeader(eudore.HeaderContentType, eudore.MimeTextHTMLCharsetUtf8)
	}
	tmpl := getLookTemplate(strings.TrimSuffix(ctx.Path(), "/"), ctx.Querys().Encode())
	return tmpl.ExecuteTemplate(ctx, "view", viewData{ctx.GetParam("*"), eudore.GetAnyByString(ctx.GetQuery("width"), 60), look})
}

func getRequestForma(ctx eudore.Context) string {
	format := ctx.GetQuery("format")
	if format != "" {
		return format
	}
	for _, accept := range strings.Split(ctx.GetHeader(eudore.HeaderAccept), ",") {
		switch strings.TrimSpace(accept) {
		case eudore.MimeApplicationJSON:
			return QueryFormatJSON
		case eudore.MimeTextHTML:
			return QueryFormatHTML
		case eudore.MimeTextPlain, eudore.MimeText:
			return QueryFormatText
		}
	}
	return ""
}

type viewData struct {
	Path  string
	Width int
	Data  *LookValue
}

func getLookTemplate(path, querys string) *template.Template {
	depth := 0
	paths := []string{path}
	if querys != "" {
		querys = "?" + querys
	}
	tpl := template.New("look").Funcs(template.FuncMap{
		"addtab":  func() string { depth++; return "" },
		"subtab":  func() string { depth--; return "" },
		"gettab":  func() string { return strings.Repeat("\t", depth) },
		"addpath": func(path string) string { paths = append(paths, path); return "" },
		"subpath": func() string { paths = paths[:len(paths)-1]; return "" },
		"getpath": func() string { return strings.Join(paths, "/") + querys },
		"isnil":   func(i any) bool { return reflect.ValueOf(i).IsNil() },
		"isline":  func(i int) bool { return i%16 == 0 },
		"showint": func(i string) string { return strings.Repeat(" ", 4-len(i)) + i },
	})
	for _, i := range lookTemplate.Templates() {
		_, _ = tpl.AddParseTree(i.Name(), i.Tree)
	}
	return tpl
}

var lookTemplate, _ = template.New("look").Funcs(template.FuncMap{
	"addtab":  getRequestForma,
	"subtab":  getRequestForma,
	"gettab":  getRequestForma,
	"addpath": getRequestForma,
	"subpath": getRequestForma,
	"getpath": getRequestForma,
	"isnil":   getRequestForma,
	"isline":  getRequestForma,
	"showint": getRequestForma,
}).Parse(`
{{- define "view" -}}
<!DOCTYPE html><html>
<head>
	<meta charset="utf-8">
	<title>Eudore Look Value {{.Path}}</title>
	<meta name="author" content="eudore">
	<meta name="referrer" content="always">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<meta name="description" content="Eudore look data all filed value">
	<style>body>div{font-family:monospace;white-space:pre;}pre{margin: 0 {{.Width}}px;}span{white-space:pre-wrap;word-wrap:break-word;overflow:hidden;}</style>
</head>
<body><div>{{- template "html" .Data -}}</div><script>
	console.log('d=10 Depth display the maximum number of layers\nall=false Whether to display the non-export attribute\nformat=html/json/text Set the data display format\ngodoc=https://golang.org Set html format linked godoc address\nwidth=60 Set html format indentation width');
	for(var i of document.getElementsByTagName('span')){
		i.addEventListener('click',(e)=>{
			var show = e.target.innerText=='-';
			e.target.innerText=show?'+':'-';
			e.target.nextSibling.style.cssText='display: '+(show?'none':'block');
		})
	}</script>
</body></html>
{{- end -}}

{{- define "html" -}}
	{{- if and (ne .Package "") (ne .Name "") -}}
		<a href="{{.Godoc}}/pkg/{{.Package}}#{{.Name}}" target="_Blank">{{.Package}}.{{.Name}}</a>
	{{- else -}}
		{{- if ne .Package ""}}{{.Package}}. {{end}}{{if ne .Name ""}}{{.Name}}{{end -}}
	{{- end -}}
	{{- if eq .Kind "bool" "int" "string" "float" "uint" "complex" -}}
		{{- if eq .String ""}}({{printf "%#v" .Value}}){{else}}("{{.String}}"){{end -}}
	{{- else if eq .Kind "struct" "map" -}}
		{{- printf "{" -}}
		{{- if ne (len .Keys ) 0 -}}
			<span>-</span><pre>
			{{- range $index, $elem := .Keys -}}
				{{- addpath (print $elem) -}}
				<a href="{{getpath}}">{{$elem}}</a>: {{template "html" index $.Vals $index -}},
				{{- printf "\n"}}{{subpath -}}
			{{- end -}}
			</pre>
		{{- end -}}
		{{- printf "}" -}}
	{{- else if eq .Kind "slice" "array" -}}
		{{- printf "[" -}}
		{{- if ne (len .Vals ) 0 -}}
			<span>-</span><pre>
			{{- range $index, $elem := .Vals -}}
				{{- addpath (print $index) -}}
				<a href="{{getpath}}">{{$index}}</a>: {{template "html" $elem -}},
				{{- printf "\n"}}{{subpath -}}
			{{- end -}}
			</pre>
		{{- end -}}
		{{- printf "]" -}}
	{{- else if eq .Kind "interface"}}{{if isnil .Elem}}(nil){{else}} {{template "html" .Elem}}{{end -}}
	{{- else if eq .Kind "func" "chan"}}{{if eq .Pointer 0}}(nil){{else}}(0x{{printf "%x" .Pointer}}){{end -}}
	{{- else -}}
		{{- if eq .Pointer 0}}(nil){{else if isnil .Elem}}(CYCLIC REFERENCE 0x{{printf "%x" .Pointer -}})
		{{- else}}{{if eq .Kind "ptr"}}&{{template "html" .Elem}}{{end}}{{end -}}
	{{end -}}
{{- end -}}

{{- define "text" -}}
	{{- if ne .Package ""}}{{.Package}}.{{end}}{{if ne .Name ""}}{{.Name}}{{end -}}
	{{- if eq .Kind "bool" "int" "string" "float" "uint" "complex"}}
		{{- if eq .String "" }}({{printf "%#v" .Value}}){{else}}("{{.String}}"){{end -}}
	{{- else if eq .Kind "struct" "map" -}}
		{{- printf "{"}}{{addtab -}}
		{{- if ne (len .Keys ) 0 -}}
			{{- range $index, $elem := .Keys -}}
				{{- printf "\n"}}{{gettab}}{{$elem}}: {{template "text" index $.Vals $index -}},
			{{- end -}}
			{{- printf "\n"}}{{subtab}}{{gettab}}{{printf "}" -}}
		{{- else -}}
			{{- subtab}}{{printf "}" -}}
		{{- end -}}
	{{- else if eq .Kind "slice" "array" -}}
		{{- printf "["}}{{addtab -}}
		{{- if ne (len .Vals ) 0}}
			{{- range $index, $elem := .Vals -}}
				{{- if eq $elem.Name "uint8"}}
					{{- if isline $index }}
						{{- printf "\n"}}{{gettab}}
					{{- end }}
					{{- showint (printf "%d" $elem.Value) }},
				{{- else}}
					{{- printf "\n"}}{{gettab}}{{$index}}: {{template "text" $elem -}},
				{{- end -}}
			{{- end -}}
			{{- printf "\n"}}{{subtab}}{{gettab}}{{printf "}" -}}
		{{- else -}}
			{{- subtab}}{{printf "]" -}}
		{{- end -}}
	{{- else if eq .Kind "interface"}}{{if isnil .Elem}}(nil){{else}} {{template "text" .Elem}}{{end -}}
	{{- else if eq .Kind "func" "chan"}}{{if eq .Pointer 0}}(nil){{else}}(0x{{ printf "%x" .Pointer}}){{end -}}
	{{- else -}}
		{{- if eq .Pointer 0 }}(nil){{else if isnil .Elem}}(CYCLIC REFERENCE 0x{{ printf "%x" .Pointer -}})
		{{- else}}{{if eq .Kind "ptr"}}&{{template "text" .Elem}}{{end}}{{end -}}
	{{- end -}}
{{- end -}}`)

// LookConfig 定义属性遍历的配置。
type LookConfig struct {
	Depth   int                  `json:"-"`
	ShowAll bool                 `json:"-"`
	Godoc   string               `json:"-"`
	Refs    map[uintptr]struct{} `json:"-"`
}

// LookValue 定义数据的每一项属性。
type LookValue struct {
	*LookConfig `json:"-"`
	Kind        string      `json:"kind"`
	Package     string      `json:"package,omitempty"`
	Name        string      `json:"name,omitempty"`
	Value       any         `json:"value,omitempty"`
	String      string      `json:"string,omitempty"`
	Pointer     uintptr     `json:"pointer,omitempty"`
	Elem        *LookValue  `json:"elem,omitempty"`
	Keys        []string    `json:"keys,omitempty"`
	Vals        []LookValue `json:"vals,omitempty"`
}

// NewLookValue 方法从请求上下文成就一个默认配置的LookValue。
func NewLookValue(ctx eudore.Context) *LookValue {
	return &LookValue{
		LookConfig: &LookConfig{
			Depth:   eudore.GetAnyByString(ctx.GetQuery("d"), 10),
			ShowAll: eudore.GetAnyByString[bool](ctx.GetQuery("all")),
			Godoc:   eudore.GetAnyByString(strings.TrimSuffix(ctx.GetQuery("godoc"), "/"), eudore.DefaultGodocServer),
			Refs:    make(map[uintptr]struct{}),
		},
	}
}

// Scan 方法扫描属性并保存。
func (look *LookValue) Scan(iValue reflect.Value) {
	look.Kind = iValue.Kind().String()
	look.Name = iValue.Type().Name()
	look.Package = iValue.Type().PkgPath()
	if look.Name == "" && iValue.Kind() != reflect.Ptr {
		look.Name = iValue.Type().String()
	}
	// check ref Chan, Func, Interface, Map, Ptr, Slice, UnsafePointer
	if look.isRef(iValue) {
		return
	}

	switch iValue.Kind() {
	case reflect.Bool:
		look.Value = iValue.Bool()
		look.String = getBasicString(iValue)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		look.Kind = "int"
		look.Value = iValue.Int()
		look.String = getBasicString(iValue)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		look.Kind = "uint"
		look.Value = iValue.Uint()
		look.String = getBasicString(iValue)
	case reflect.Float32, reflect.Float64:
		look.Kind = "float"
		look.Value = iValue.Float()
		look.String = getBasicString(iValue)
	case reflect.Complex64, reflect.Complex128:
		look.Kind = "complex"
		look.Value = iValue.Complex()
		look.String = getBasicString(iValue)
	case reflect.String:
		look.Value = iValue.String()
	case reflect.Slice, reflect.Array:
		look.scanSlice(iValue)
	case reflect.Struct:
		look.scanStruct(iValue)
	case reflect.Map:
		look.scanMap(iValue)
	case reflect.Ptr, reflect.Interface:
		look.Elem = new(LookValue)
		look.Elem.LookConfig = look.LookConfig
		look.Elem.Scan(iValue.Elem())
	case reflect.Func, reflect.Chan, reflect.UnsafePointer:
	}
}

func (look *LookValue) isRef(iValue reflect.Value) bool {
	switch iValue.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Interface, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		if iValue.IsNil() {
			return true
		}
		if iValue.Kind() != reflect.Interface {
			look.Pointer = iValue.Pointer()
			_, ok := look.Refs[look.Pointer]
			if ok {
				look.Name = iValue.Type().String()
				return true
			}
			look.Refs[look.Pointer] = struct{}{}
		}
	}
	return false
}

func (look *LookValue) scanSlice(iValue reflect.Value) {
	look.Depth--
	if look.Depth > 0 {
		look.Vals = make([]LookValue, iValue.Len())
		for i := 0; i < iValue.Len(); i++ {
			look.Vals[i].LookConfig = look.LookConfig
			look.Vals[i].Scan(iValue.Index(i))
		}
	}
	look.Depth++
}

func (look *LookValue) scanStruct(iValue reflect.Value) {
	look.Depth--
	if look.Depth > 0 {
		iType := iValue.Type()
		for i := 0; i < iValue.NumField(); i++ {
			if iValue.Field(i).CanInterface() || look.ShowAll {
				l := LookValue{LookConfig: look.LookConfig}
				l.Scan(iValue.Field(i))
				look.Keys = append(look.Keys, iType.Field(i).Name)
				look.Vals = append(look.Vals, l)
			}
		}
	}
	look.Depth++
}

func (look *LookValue) scanMap(iValue reflect.Value) {
	look.Depth--
	if look.Depth > 0 {
		look.Keys = make([]string, iValue.Len())
		look.Vals = make([]LookValue, iValue.Len())
		for i, key := range iValue.MapKeys() {
			look.Keys[i] = getKeyString(key)
			look.Vals[i].LookConfig = look.LookConfig
			look.Vals[i].Scan(iValue.MapIndex(key))
		}
	}
	look.Depth++
}

var typeStringer = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

func getBasicString(iValue reflect.Value) string {
	if iValue.CanInterface() && iValue.Type().Implements(typeStringer) {
		return iValue.MethodByName("String").Call(nil)[0].String()
	}
	return ""
}

func getKeyString(iValue reflect.Value) string {
	switch iValue.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(iValue.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(iValue.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(iValue.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(iValue.Float(), 'f', -1, 64)
	case reflect.Complex64, reflect.Complex128:
		return strconv.FormatComplex(iValue.Complex(), 'f', -1, 128)
	case reflect.String:
		return iValue.String()
	case reflect.Interface, reflect.Ptr:
		if iValue.IsNil() {
			return ""
		}
		return getKeyString(iValue.Elem())
	default:
		return "noprint(" + iValue.Type().String() + ")"
	}
}
