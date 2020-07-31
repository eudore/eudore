package pprof

import (
	"bytes"
	"fmt"
	"html/template"
	"reflect"
	"strings"

	"github.com/eudore/eudore"
)

// Look 函数用于显示app对象及其属性。
//
// 当前暂时使用github.com/kr/pretty格式化app对象。
func Look(ctx eudore.Context) {
	data := ctx.GetContext().Value(eudore.AppContextKey)
	if data == nil {
		ctx.SetHeader(eudore.HeaderContentType, "text/plain; charset=utf-8")
		ctx.WriteString(`pprof look not found *eudore.App object.

go1.13+
// 确保app.Context.Value(eudore.AppContextKey)可以获得app对象，并设置go 1.13 net/htpp.Server.BaseContext属性返回app.Context。
eudore.Set(app.Server, "BaseContext", func(net.Listener) context.Context {
	return app.Context
})

go1.9-go1.12
// 自定义路由处理函数输出app对象。
app.AnyFunc("/eudore/debug/pprof/look/* godoc=/eudore/debug/pprof/godoc", pprof.NewLook(app))
			`)
		return
	}
	NewLook(data)(ctx)
}

// NewLook 函数创建一个对象look处理函数。
func NewLook(data interface{}) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		data := eudore.Get(data, strings.Replace(ctx.GetParam("*"), "/", ".", -1))
		p := ctx.Path()
		if p[len(p)-1] == '/' {
			p = p[:len(p)-1]
		}
		fm := &formatter{
			buffer:    bytes.NewBuffer(nil),
			visited:   make(map[uintptr]int),
			pathfield: []string{p},
			depth:     1,
			maxDepth:  eudore.GetStringInt(ctx.GetQuery("d"), 10),
			godoc:     eudore.GetString(ctx.GetParam("godoc"), "https://golang.org"),
			args:      ctx.Querys().Encode(),
			showall:   ctx.GetQuery("all") == "",
		}
		fm.indent()
		fm.handle(reflect.ValueOf(data))
		ctx.SetHeader(eudore.HeaderContentType, eudore.MimeTextHTMLCharsetUtf8)
		ctx.SetHeader("X-Eudore-Admin", "look")
		ctx.WriteString(`<style>
			 span{white-space:pre-wrap;word-wrap : break-word ;overflow: hidden ;}
			</style><script>console.log("d=10 depth递归显时层数\nall=false 是否显时非导出属性")</script><pre>`)
		ctx.Write(fm.buffer.Bytes())
		ctx.WriteString("</pre>")
	}
}

type formatter struct {
	buffer    *bytes.Buffer
	visited   map[uintptr]int
	line      int
	pathfield []string
	depth     int
	maxDepth  int
	godoc     string
	args      string
	showall   bool
}

func (fm *formatter) print(text string) {
	fmt.Fprint(fm.buffer, text)
}
func (fm *formatter) printf(format string, args ...interface{}) {
	fmt.Fprintf(fm.buffer, format, args...)
}
func (fm *formatter) indent() {
	fm.line++
	fm.buffer.WriteString(strings.Repeat("\t", fm.depth-1))
}

func (fm *formatter) handle(iValue reflect.Value) {
	if fm.handleStringer(iValue) {
		return
	}
	switch iValue.Kind() {
	case reflect.Bool:
		fm.printf("%t", iValue.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fm.printf("%d", iValue.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		fm.printf("%d", iValue.Uint())
	case reflect.Float32, reflect.Float64:
		fm.printf("%f", iValue.Float())
	case reflect.Complex64, reflect.Complex128:
		fm.printf("%g", iValue.Complex())
	case reflect.String:
		fm.print("\"" + iValue.String() + "\"")
	case reflect.UnsafePointer:
		fm.printf("%s(%d)", iValue.Type().String(), iValue.Pointer())
	case reflect.Invalid:
		fm.print("nil")
	default:
		fm.handleOther(iValue)
	}
}

func (fm *formatter) handleOther(iValue reflect.Value) {
	switch iValue.Kind() {
	case reflect.Ptr:
		fm.handlePtr(iValue)
	case reflect.Interface:
		fm.handleInterface(iValue)
	case reflect.Struct:
		fm.handleStruct(iValue)
	case reflect.Slice, reflect.Array:
		fm.handleSlice(iValue)
	case reflect.Map:
		fm.handleMap(iValue)
	case reflect.Func, reflect.Chan:
		fm.handleFunc(iValue)
	default:
		fmt.Printf("not handle kind %s type %s\n", iValue.Kind().String(), iValue.Type().String())
	}
}

var typeStringer = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

func (fm *formatter) handleStringer(iValue reflect.Value) bool {
	if iValue.CanSet() && iValue.Type().Implements(typeStringer) {
		val := iValue.MethodByName("String").Call([]reflect.Value{})[0]
		str := val.String()
		if str != "" {
			fm.print(template.HTMLEscapeString(str))
			return true
		}
	}
	return false
}

func (fm *formatter) handlePtr(iValue reflect.Value) {
	if !iValue.IsNil() {
		line, ok := fm.visited[iValue.Pointer()]
		if ok {
			fm.printf("jump: <a href='#L%d'>%s</a>", line, iValue.Type().String())
		} else {
			fm.print("&")
			fm.visited[iValue.Pointer()] = fm.line
			fm.handle(iValue.Elem())
		}
	} else {
		fm.printf("%s(nil)", iValue.Type().String())
	}
}

func (fm *formatter) handleInterface(iValue reflect.Value) {
	if iValue.Elem().Kind() != reflect.Invalid {
		iType := iValue.Type()
		if iType.PkgPath() != "" {
			fm.printf("<a href='%s/pkg/%s#%s'>%s</a> ", fm.godoc, iType.PkgPath(), iType.Name(), iType.String())
		} else {
			fm.print(iType.String() + " ")
		}
		fm.handle(iValue.Elem())
	} else {
		fm.printf("%s(nil)", iValue.Type().String())
	}
}

func (fm *formatter) handleSlice(iValue reflect.Value) {
	if iValue.Kind() == reflect.Slice {
		if iValue.IsNil() {
			fm.printf("%s (nil)", iValue.Type().String())
			return
		}
		line, ok := fm.visited[iValue.Pointer()]
		if ok {
			fm.printf("jump: <a href='#L%d'>%s</a>", line, iValue.Type().String())
			return
		}
		fm.visited[iValue.Pointer()] = fm.line
	}

	st := simpleType(iValue.Type().Elem())
	fm.printf("%s{", iValue.Type().String())
	if st {
		fm.print("<span>")
	} else {
		fm.print("\n")
	}

	if fm.depth <= fm.maxDepth {
		fm.depth += 1
		fm.pathfield = append(fm.pathfield, "")
		for i := 0; i < iValue.Len(); i++ {
			if st {
				fm.handle(iValue.Index(i))
				fm.print(",")
			} else {
				fm.indent()
				fm.pathfield[fm.depth-1] = fmt.Sprint(i)
				fm.handle(iValue.Index(i))
				fm.print(",\r\n")
			}
		}
		fm.depth -= 1
		fm.pathfield = fm.pathfield[0:fm.depth]
	}

	if st {
		fm.buffer.Bytes()[fm.buffer.Len()-1] = '}'
		fm.print("</span>")
	} else {
		fm.indent()
		fm.print("}")
	}
}

func simpleType(iType reflect.Type) bool {
	switch iType.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr, reflect.Float32, reflect.Float64, reflect.String, reflect.Func:
		return true
	default:
		return false
	}
}

func (fm *formatter) handleMap(iValue reflect.Value) {
	if iValue.IsNil() {
		fm.printf("%s{},", iValue.Type().String())
		return
	}
	line, ok := fm.visited[iValue.Pointer()]
	if ok {
		fm.printf("jump: <a href='#L%d'>%s</a>", line, iValue.Type().String())
		return
	}
	fm.visited[iValue.Pointer()] = fm.line

	fm.print(iValue.Type().String() + "{\n")
	if fm.depth <= fm.maxDepth {
		fm.depth += 1
		fm.pathfield = append(fm.pathfield, "")
		for _, key := range iValue.MapKeys() {
			fm.indent()
			l1 := fm.buffer.Len()
			fm.handle(key)
			l2 := fm.buffer.Len()
			fm.pathfield[fm.depth-1] = getMapKey(fm.buffer.Bytes()[l1:l2])
			fm.print(": ")
			fm.handle(iValue.MapIndex(key))
			fm.print(",\r\n")
		}
		fm.depth -= 1
		fm.pathfield = fm.pathfield[0:fm.depth]
	}
	fm.indent()
	fm.print("}")
}

func getMapKey(key []byte) string {
	if len(key) > 1 && key[0] == '"' {
		return string(key[1 : len(key)-1])
	}
	return string(key)
}

func (fm *formatter) handleStruct(iValue reflect.Value) {
	iType := iValue.Type()
	if iType.NumField() == 0 {
		fm.printf("<a href='%s/pkg/%s#%s'>%s</a>{},", fm.godoc, iType.PkgPath(), iType.Name(), iType.String())
		return
	}
	f := iType.Name()[0]
	if f < 'A' || f > 'Z' {
		fm.print(iType.String() + "{\r\n")
	} else {
		fm.printf("<a href='%s/pkg/%s#%s'>%s</a>{\r\n", fm.godoc, iType.PkgPath(), iType.Name(), iType.String())
	}

	if fm.depth <= fm.maxDepth {
		fm.depth += 1
		fm.pathfield = append(fm.pathfield, "")
		for i := 0; i < iType.NumField(); i++ {
			if !iValue.Field(i).CanSet() && fm.showall {
				continue
			}
			fm.indent()
			fm.pathfield[fm.depth-1] = iType.Field(i).Name
			fm.printf("<a href='%s?%s' id='L%d'>%s</a>", strings.Join(fm.pathfield[:fm.depth], "/"), fm.args, fm.line, iType.Field(i).Name)
			fm.print(": ")
			fm.handle(iValue.Field(i))
			fm.print(",\n")
		}
		fm.pathfield = fm.pathfield[0:fm.depth]
		fm.depth -= 1
	}
	fm.indent()
	fm.print("}")
}

func (fm *formatter) handleFunc(iValue reflect.Value) {
	iType := iValue.Type()
	if iType.PkgPath() != "" {
		fm.printf("<a href='%s/pkg/%s#%s'>%s</a>", fm.godoc, iType.PkgPath(), iType.Name(), iType.String())
	} else {
		fm.print(iValue.Type().String())
	}
	if iValue.IsNil() {
		fm.print("(nil)")
	} else {
		fm.print(" {...}")
	}
}
