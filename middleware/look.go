package middleware

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"text/template"

	"github.com/eudore/eudore"
)

// The NewHandlerMetadata function creates [eudore.HandlerFunc] to access object
// all data.
//
// If the data type is func(eudore.Context) any, the data to be rendered
// can be returned dynamically.
//
// Use the [eudore.Params] "*" to return the data of the specified attribute
// of the object.
//
// The following URI parameters are allowed:
//
//	d=10 Depth display the maximum number of layers.
//	all=false Whether to display the non-export attribute.
//	format=html/json/text Set the data display format.
//	godoc=https://golang.org Set html format linked godoc address.
//	width=60 Set html format indentation width.
//
//go:noinline
func NewLookFunc(data any) Middleware {
	fn, ok := data.(func(eudore.Context) any)
	if !ok {
		fn = func(eudore.Context) any {
			return data
		}
	}
	return func(ctx eudore.Context) {
		doc := strings.TrimSuffix(ctx.GetQuery("godoc"), "/")
		look := &lookValue{
			lookConfig: &lookConfig{
				Depth:   eudore.GetAnyByString(ctx.GetQuery("d"), 10),
				ShowAll: eudore.GetAnyByString(ctx.GetQuery("all"), false),
				Godoc:   eudore.GetAnyByString(doc, eudore.DefaultGodocServer),
				Refs:    make(map[uintptr]struct{}),
			},
		}

		v := fn(ctx)
		if look.ShowAll && v != nil {
			v = reflect.ValueOf(v)
		}
		val, err := eudore.GetValueByPath(v,
			strings.ReplaceAll(ctx.GetParam("*"), "/", "."),
			nil,
		)
		if err != nil {
			ctx.Fatal(err)
			return
		}
		look.Scan(val)

		path := strings.TrimSuffix(ctx.Path(), "/")
		raw := ctx.Request().URL.RawQuery
		switch getRequestFormat(ctx) {
		case QueryFormatJSON:
			_ = eudore.HandlerDataRenderJSON(ctx, look)
		case QueryFormatHTML:
			data := viewData{
				ctx.GetParam("*"),
				eudore.GetAnyByString(ctx.GetQuery("width"), 60),
				look,
			}
			tmpl := getLookTemplate(path, raw)
			ctx.SetHeader(headerContentType, eudore.MimeTextHTMLCharsetUtf8)
			_ = tmpl.ExecuteTemplate(ctx, "view", data)
		default:
			tmpl := getLookTemplate(path, raw)
			ctx.SetHeader(headerContentType, eudore.MimeTextPlainCharsetUtf8)
			_ = tmpl.ExecuteTemplate(ctx, "text", look)
		}
	}
}

func getRequestFormat(ctx eudore.Context) string {
	format := ctx.GetQuery("format")
	if format != "" {
		return format
	}

	accepts := strings.Split(ctx.GetHeader(eudore.HeaderAccept), ",")
	for _, accept := range accepts {
		switch strings.TrimSpace(accept) {
		case eudore.MimeApplicationJSON:
			return QueryFormatJSON
		case eudore.MimeTextHTML:
			return QueryFormatHTML
		}
	}
	return ""
}

type viewData struct {
	Path  string
	Width int
	Data  *lookValue
}

func getLookTemplate(path, raw string) *template.Template {
	depth := 0
	paths := []string{path}
	if raw != "" {
		raw = "?" + raw
	}
	tpl := template.New("look").Funcs(template.FuncMap{
		"addtab": func() string { depth++; return "" },
		"subtab": func() string { depth--; return "" },
		"gettab": func() string { return strings.Repeat("\t", depth) },
		"addpath": func(path string) string {
			paths = append(paths, path)
			return ""
		},
		"subpath": func() string { paths = paths[:len(paths)-1]; return "" },
		"getpath": func() string { return strings.Join(paths, "/") + raw },
		"isnil":   func(i any) bool { return reflect.ValueOf(i).IsNil() },
		"isline":  func(i int) bool { return i%16 == 0 },
		"showint": func(i string) string {
			return strings.Repeat(" ", 4-len(i)) + i
		},
	})
	for _, i := range lookTemplate.Templates() {
		_, _ = tpl.AddParseTree(i.Name(), i.Tree)
	}
	return tpl
}

var lookTemplate, _ = template.New("look").Funcs(template.FuncMap{
	"addtab":  getRequestFormat,
	"subtab":  getRequestFormat,
	"gettab":  getRequestFormat,
	"addpath": getRequestFormat,
	"subpath": getRequestFormat,
	"getpath": getRequestFormat,
	"isnil":   getRequestFormat,
	"isline":  getRequestFormat,
	"showint": getRequestFormat,
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
	<style>
	body>div{font-family:monospace;white-space:pre;}
	pre{margin: 0 {{.Width}}px;}
	span{white-space:pre-wrap;word-wrap:break-word;overflow:hidden;}
	</style>
</head>
<body><div>{{- template "html" .Data -}}</div><script>
console.log('d=10 Depth display the maximum number of layers\n' +
	'all=false Whether to display the non-export attribute\n' +
	'format=html/json/text Set the data display format\n' +
	'godoc=https://golang.org Set html format linked godoc address\n' +
	'width=60 Set html format indentation width');
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

// lookConfig defines the configuration for attribute traversal.
type lookConfig struct {
	Depth   int
	ShowAll bool
	Godoc   string
	Refs    map[uintptr]struct{}
}

// lookValue defines each attribute of the data.
type lookValue struct {
	*lookConfig `json:"-"`
	Kind        string      `json:"kind"`
	Package     string      `json:"package,omitempty"`
	Name        string      `json:"name,omitempty"`
	Value       any         `json:"value,omitempty"`
	String      string      `json:"string,omitempty"`
	Pointer     uintptr     `json:"pointer,omitempty"`
	Elem        *lookValue  `json:"elem,omitempty"`
	Keys        []string    `json:"keys,omitempty"`
	Vals        []lookValue `json:"vals,omitempty"`
}

// The Scan method scans the attributes and saves them.
func (look *lookValue) Scan(iValue reflect.Value) {
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
	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
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
		look.Elem = new(lookValue)
		look.Elem.lookConfig = look.lookConfig
		look.Elem.Scan(iValue.Elem())
	case reflect.Func, reflect.Chan, reflect.UnsafePointer:
	}
}

func (look *lookValue) isRef(iValue reflect.Value) bool {
	switch iValue.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Interface,
		reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
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

func (look *lookValue) scanSlice(iValue reflect.Value) {
	look.Depth--
	if look.Depth > 0 {
		look.Vals = make([]lookValue, iValue.Len())
		for i := 0; i < iValue.Len(); i++ {
			look.Vals[i].lookConfig = look.lookConfig
			look.Vals[i].Scan(iValue.Index(i))
		}
	}
	look.Depth++
}

func (look *lookValue) scanStruct(iValue reflect.Value) {
	look.Depth--
	if look.Depth > 0 {
		iType := iValue.Type()
		for i := 0; i < iValue.NumField(); i++ {
			if iValue.Field(i).CanInterface() || look.ShowAll {
				l := lookValue{lookConfig: look.lookConfig}
				l.Scan(iValue.Field(i))
				look.Keys = append(look.Keys, iType.Field(i).Name)
				look.Vals = append(look.Vals, l)
			}
		}
	}
	look.Depth++
}

func (look *lookValue) scanMap(iValue reflect.Value) {
	look.Depth--
	if look.Depth > 0 {
		look.Keys = make([]string, iValue.Len())
		look.Vals = make([]lookValue, iValue.Len())
		for i, key := range iValue.MapKeys() {
			look.Keys[i] = getKeyString(key)
			look.Vals[i].lookConfig = look.lookConfig
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
	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
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
