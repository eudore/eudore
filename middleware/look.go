package middleware

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"

	"github.com/eudore/eudore"
)

// NewLookFunc 函数创建一个访问对象数据处理函数。
func NewLookFunc(data interface{}) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		data := eudore.Get(data, strings.Replace(ctx.GetParam("*"), "/", ".", -1))
		p := ctx.Path()
		if p[len(p)-1] == '/' {
			p = p[:len(p)-1]
		}
		fmt.Println(data)
		fm := &formatter{
			buffer:   bytes.NewBuffer(nil),
			visited:  make(map[uintptr]string),
			depth:    0,
			maxDepth: eudore.GetStringInt(ctx.GetQuery("d"), 10),
			godoc:    eudore.GetString(ctx.GetParam("godoc"), "https://golang.org") + p + "/",
			args:     ctx.Querys().Encode(),
			showall:  ctx.GetQuery("all") != "",
			format:   eudore.GetString(ctx.GetQuery("format"), "html"),
			tmp:      getTmp(),
		}

		tmp := fm.tmp.Lookup(fm.format)
		if tmp == nil {
			ctx.WriteHeader(404)
			ctx.WriteString("not found look format: " + fm.format)
			return
		}

		fm.handleValue(reflect.ValueOf(data), " ")
		ctx.SetHeader(eudore.HeaderContentType, eudore.MimeTextHTMLCharsetUtf8)
		ctx.SetHeader("X-Eudore-Admin", "look")
		ctx.WriteString(`<style>
			 span{white-space:pre-wrap;word-wrap : break-word ;overflow: hidden ;}
			</style><script>console.log("d=10 depth递归显时层数\nall=false 是否显时非导出属性")</script><pre>`)
		ctx.Write(fm.buffer.Bytes())
		ctx.WriteString("</pre>")
	}
}

// LookValue 定义渲染数据的每一项属性。
type LookValue struct {
	Kind    string `json:"kind"`
	Type    string `json:",omitempty"`
	Pkg     string `json:",omitempty"`
	PkgPath string `json:",omitempty"`
	Public  bool   `json:",omitempty"`
	Keypath string `json:",omitempty"`
	Value   interface{}
}

type formatter struct {
	buffer   *bytes.Buffer
	visited  map[uintptr]string
	depth    int
	maxDepth int
	godoc    string
	args     string
	showall  bool
	format   string
	tmp      *template.Template
}

func getTmp() *template.Template {
	str := strings.Replace(tmpdefault, "\n", "", -1)
	str = strings.Replace(str, "\t", "", -1)
	str = strings.Replace(str, "GODOC", "https://godoc.org", -1)
	tmpl, err := template.New("name").Parse(str)
	fmt.Println(err)
	return tmpl
}

var tmpdefault string = `
{{define "txt"}}
	{{if ne .PkgPath ""}}
		{{.PkgPath}}.
	{{end}}
	{{if ne .Type ""}}
		{{.Type}}
	{{end}}
	{{ if eq .Kind "bool" "int" "string" "float" "uint" "complex"}}
		{{if .Keypath }}
			{{if .Value}}
				{{.Value}}
			{{else}}
				jump to {{.Keypath}}
			{{end}}
		{{else}}
			{{ if eq .Kind "string"}}
				"{{.Value}}"
			{{else}}
				{{.Value}}
			{{end}}
		{{end}}
	{{else if eq .Kind "func" "chan"}}
		(0x{{ printf "%x" .Value}})
	{{else }}
		{{if .Value }}
			{{.Value}}
		{{end}}
	{{end}}
{{end}}

{{define "html"}}
	{{if ne .Type ""}}
		{{if .Public}}
			<a href="GODOC/pkg/{{.PkgPath}}#{{.Type}}">{{.Pkg}}.{{.Type}}</a>
		{{else}}
			{{if .Pkg}}
				{{.Pkg}}.
			{{end}}
			{{.Type}}
		{{end}}
	{{end}}
	{{ if eq .Kind "bool" "int" "string" "float" "uint" "complex"}}
		{{if .Keypath }}
			{{if .Value}}
				<a href="{{.Keypath}}?" id="{{.Keypath}}">{{.Value}}</a>
			{{else}}
				jump to <a href="#{{.Keypath}}"">{{.Keypath}}</a>
			{{end}}
		{{else}}
			{{ if eq .Kind "string"}}
				"{{.Value}}"
			{{else}}
				{{.Value}}
			{{end}}
		{{end}}
	{{else if eq .Kind "func" "chan"}}
		{{ if .Value }}
			(0x{{ printf "%x" .Value}})
		{{else}}
			(nil)
		{{end}}
	{{else }}
		{{if .Value }}
			{{.Value}}
		{{end}}
	{{end}}
{{end}}
`

func (fm *formatter) WriteKey(look LookValue) {
	fm.WriteValue(look)
}

func (fm *formatter) WriteValue(look LookValue) {
	if look.PkgPath != "" {
		look.Pkg = filepath.Base(look.PkgPath)
	}
	if look.Type != "" && 'A' <= look.Type[0] && look.Type[0] <= 'Z' {
		look.Public = true
	}

	fm.tmp.ExecuteTemplate(fm.buffer, fm.format, look)
}

func (fm *formatter) WriteString(args ...interface{}) {
	fmt.Fprint(fm.buffer, args...)
}
func (fm *formatter) WriteLine() {
	fmt.Fprint(fm.buffer, strings.Repeat("\t", fm.depth))
}

var typeStringer = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

func (fm *formatter) handleValue(iValue reflect.Value, path string) {
	if fm.depth > fm.maxDepth || fm.handleElemString(iValue, path) {
		return
	}

	switch iValue.Kind() {
	case reflect.Bool:
		fm.WriteValue(LookValue{Kind: "bool", Value: iValue.Bool()})
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fm.WriteValue(LookValue{Kind: "int", Value: iValue.Int()})
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		fm.WriteValue(LookValue{Kind: "uint", Value: iValue.Uint()})
	case reflect.Float32, reflect.Float64:
		fm.WriteValue(LookValue{Kind: "float", Value: iValue.Float()})
	case reflect.Complex64, reflect.Complex128:
		fm.WriteValue(LookValue{Kind: "complex", Value: iValue.Complex()})
	case reflect.String:
		fm.WriteValue(LookValue{Kind: "string", Value: iValue.String()})
	case reflect.Func:
		fm.WriteValue(LookValue{Kind: "func", Type: getTypeName(iValue.Type()), PkgPath: iValue.Type().PkgPath(), Value: iValue.Pointer()})
	case reflect.Ptr:
		fm.WriteString("&")
		fm.handleValue(iValue.Elem(), path)
	case reflect.Interface:
		fm.WriteValue(LookValue{
			Kind:    "interface",
			Type:    getTypeName(iValue.Type()),
			PkgPath: iValue.Type().PkgPath(),
		})
		fm.WriteString(" ")
		fm.handleValue(iValue.Elem(), path)
	case reflect.Slice, reflect.Array:
		fm.handleSliceValue(iValue, path)
	case reflect.Struct:
		fm.handleStructValue(iValue, path)
	case reflect.Map:
		fm.handleMapValue(iValue, path)
	}
}

func (fm *formatter) handleElemString(iValue reflect.Value, path string) bool {
	switch iValue.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Interface, reflect.Ptr, reflect.Slice, reflect.UnsafePointer:
		if iValue.IsNil() {
			fm.WriteString(iValue.Type(), "(nil)")
			return true
		} else if iValue.Kind() != reflect.Interface {
			line, ok := fm.visited[iValue.Pointer()]
			if ok {
				fm.WriteValue(LookValue{
					Kind:    "string",
					Keypath: line,
				})
				return true
			}
			fm.visited[iValue.Pointer()] = path
		}
	}

	if iValue.CanSet() && iValue.Type().Implements(typeStringer) {
		val := iValue.MethodByName("String").Call([]reflect.Value{})[0]
		str := val.String()
		if str != "" {
			fm.WriteString(htmltemplate.HTMLEscapeString(str))
			return true
		}
	}
	return false
}

func (fm *formatter) handleSliceValue(iValue reflect.Value, path string) {
	iType := iValue.Type().Elem()
	fmt.Fprint(fm.buffer, "[]")
	fm.WriteValue(LookValue{
		Kind:    iValue.Type().Kind().String(),
		Type:    getTypeName(iType),
		PkgPath: iType.PkgPath(),
		Keypath: path,
	})
	fmt.Fprint(fm.buffer, "{")

	isbase := getBaseType(iType)
	if iValue.Len() > 0 {
		fm.depth++
		//			fmt.Fprint(fm.buffer, "\n")
		for i := 0; i < iValue.Len(); i++ {
			if !isbase || (i%16 == 0 && iValue.Len() > 16) {
				fmt.Fprint(fm.buffer, "\n")
				fm.WriteLine()
				fm.handleValue(iValue.Index(i), path+"/"+fmt.Sprint(i))
				fmt.Fprint(fm.buffer, ", ")
			} else {
				fm.handleValue(iValue.Index(i), path+"/"+fmt.Sprint(i))
				fmt.Fprint(fm.buffer, ", ")
			}
		}
		fm.depth--
		if !isbase || iValue.Len() > 16 {
			fmt.Fprint(fm.buffer, "\n")
			fm.WriteLine()
		}
	}
	fmt.Fprint(fm.buffer, "}")
}

func (fm *formatter) handleStructValue(iValue reflect.Value, path string) {
	iType := iValue.Type()
	fm.WriteValue(LookValue{
		Kind:    "struct",
		Type:    getTypeName(iValue.Type()),
		PkgPath: iValue.Type().PkgPath(),
		Keypath: path,
	})
	fmt.Fprintln(fm.buffer, "{")

	fm.depth++
	var isdata bool
	for i := 0; i < iType.NumField(); i++ {
		if iValue.Field(i).CanSet() || fm.showall {
			fm.WriteLine()
			fm.WriteKey(LookValue{
				Kind:    "string",
				PkgPath: iType.Field(i).PkgPath,
				Value:   iType.Field(i).Name,
				Keypath: path + "/" + iType.Field(i).Name,
			})
			isdata = true
			fmt.Fprint(fm.buffer, ": ")
			fm.handleValue(iValue.Field(i), path+"/"+iType.Field(i).Name)
			fmt.Fprintln(fm.buffer, ",")
		}
	}
	fm.depth--
	if isdata {
		fm.WriteLine()
		fmt.Fprint(fm.buffer, "}")
	} else {
		fm.buffer.Bytes()[fm.buffer.Len()-1] = '}'
	}
}

func (fm *formatter) handleMapValue(iValue reflect.Value, path string) {
	fm.WriteValue(LookValue{
		Kind:    "map",
		Type:    getTypeName(iValue.Type()),
		PkgPath: iValue.Type().PkgPath(),
		Keypath: path,
	})
	fmt.Fprintln(fm.buffer, "{")

	fm.depth++
	var isdata bool
	for _, key := range iValue.MapKeys() {
		fm.WriteLine()
		newpath := path + "/" + getKeyString(key)
		if getBaseType(reflect.Indirect(key).Type()) {
			fm.WriteKey(LookValue{
				Kind:    "string",
				Value:   getKeyString(key),
				Keypath: newpath,
			})
		} else {
			fm.handleValue(key, newpath)
		}
		isdata = true
		fmt.Fprint(fm.buffer, ": ")
		fm.handleValue(iValue.MapIndex(key), newpath)
		fmt.Fprintln(fm.buffer, ",")
	}
	fm.depth--
	if isdata {
		fm.WriteLine()
		fmt.Fprint(fm.buffer, "}")
	} else {
		fm.buffer.Bytes()[fm.buffer.Len()-1] = '}'
	}
}

func getTypeName(iType reflect.Type) string {
	name := iType.Name()
	if name != "" {
		return name
	}
	return iType.String()
}

func getBaseType(iType reflect.Type) bool {
	switch iType.Kind() {
	case reflect.String, reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return true
	default:
		return false
	}
}

func getKeyString(iValue reflect.Value) string {
	switch iValue.Kind() {
	case reflect.Bool:
		return fmt.Sprint(iValue.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprint(iValue.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return fmt.Sprint(iValue.Uint())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprint(iValue.Float())
	case reflect.Complex64, reflect.Complex128:
		return fmt.Sprint(iValue.Complex())
	case reflect.String:
		return iValue.String()
	case reflect.Interface, reflect.Ptr:
		if iValue.IsNil() {
			return ""
		}
		return getKeyString(iValue.Elem())
	default:
		return ""
	}
}
