package eudore

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// HandlerDataFunc 定义请求上下文数据出来函数。
//
// 默认定义Bind Validate Filte Render四种行为。
//
// Binder对象用于请求数据反序列化，默认根据http请求的Content-Type header指定的请求数据格式来解析数据。
//
// Renderer对象更加Accept Header选择数据对象序列化的方法。
type HandlerDataFunc = func(Context, interface{}) error

func init() {
	DefaultRenderHTMLTemplate, _ = template.New("render").Parse(renderHTMLTempdate)
}

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
	return func(ctx Context, i interface{}) error {
		contentType := ctx.GetHeader(HeaderContentType)
		if ctx.Method() == MethodGet || ctx.Method() == MethodHead || contentType == "" {
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

// NewBindWithHeader 实现Binder额外封装bind header。
func NewBindWithHeader(fn HandlerDataFunc) HandlerDataFunc {
	return func(ctx Context, i interface{}) error {
		BindHeader(ctx, i)
		return fn(ctx, i)
	}
}

// NewBindWithURL 实现Binder在非get和head方法下实现BindURL。
func NewBindWithURL(fn HandlerDataFunc) HandlerDataFunc {
	return func(ctx Context, i interface{}) error {
		if ctx.Method() != MethodGet && ctx.Method() != MethodHead {
			BindURL(ctx, i)
		}
		return fn(ctx, i)
	}
}

// BindURL 函数使用url参数解析绑定body。
func BindURL(ctx Context, i interface{}) error {
	return ConvertToWithTags(ctx.Querys(), i, DefaultBindURLTags)
}

// BindForm 函数使用form解析绑定body。
func BindForm(ctx Context, i interface{}) error {
	ConvertToWithTags(ctx.FormFiles(), i, DefaultBindFormTags)
	return ConvertToWithTags(ctx.FormValues(), i, DefaultBindFormTags)
}

// BindJSON 函数使用json解析绑定body。
func BindJSON(ctx Context, i interface{}) error {
	return json.NewDecoder(ctx).Decode(i)
}

// BindXML 函数使用xml解析绑定body。
func BindXML(ctx Context, i interface{}) error {
	return xml.NewDecoder(ctx).Decode(i)
}

// BindProtobuf 函数使用内置protobu解析绑定body。
func BindProtobuf(ctx Context, i interface{}) error {
	return NewProtobufDecoder(ctx).Decode(i)
}

// BindHeader 函数实现使用header数据bind。
func BindHeader(ctx Context, i interface{}) error {
	return ConvertToWithTags(ctx.Request().Header, i, DefaultBindHeaderTags)
}

// NewRenders 方法定义默认和Accepte Header映射Render函数。
func NewRenders(renders map[string]HandlerDataFunc) HandlerDataFunc {
	if renders == nil {
		renders = map[string]HandlerDataFunc{
			MimeApplicationJSON:     RenderJSON,
			MimeApplicationProtobuf: RenderProtobuf,
			MimeApplicationXML:      RenderXML,
			MimeTextHTML:            RenderHTML,
			MimeTextPlain:           RenderText,
		}
	}
	return func(ctx Context, i interface{}) error {
		for _, accept := range strings.Split(ctx.GetHeader(HeaderAccept), ",") {
			pos := strings.IndexByte(accept, ';')
			if pos != -1 {
				accept = accept[:pos]
			}
			fn, ok := renders[strings.TrimSpace(accept)]
			if ok && fn != nil {
				err := fn(ctx, i)
				if err != ErrRenderHandlerSkip {
					return err
				}
			}
		}
		return DefaultRenderFunc(ctx, i)
	}
}

func renderSetContentType(ctx Context, mime string) {
	header := ctx.Response().Header()
	if val := header.Get(HeaderContentType); len(val) == 0 {
		header.Add(HeaderContentType, mime)
	}
}

// RenderJSON 函数使用encoding/json库实现json反序列化。
//
// 如果请求Accept不为"application/json"，使用json indent格式输出。
func RenderJSON(ctx Context, data interface{}) error {
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

// RenderXML 函数Render Xml，使用encoding/xml库实现xml反序列化。
func RenderXML(ctx Context, data interface{}) error {
	renderSetContentType(ctx, MimeApplicationXMLCharsetUtf8)
	return xml.NewEncoder(ctx).Encode(data)
}

// RenderText 函数Render Text，使用fmt.Fprint函数写入。
func RenderText(ctx Context, data interface{}) error {
	renderSetContentType(ctx, MimeTextPlainCharsetUtf8)
	if s, ok := data.(string); ok {
		return ctx.WriteString(s)
	}
	if s, ok := data.(fmt.Stringer); ok {
		return ctx.WriteString(s.String())
	}
	_, err := fmt.Fprintf(ctx, "%#v", data)
	return err
}

// RenderProtobuf 函数Render Protobuf，使用内置protobuf编码，无效属性将忽略。
func RenderProtobuf(ctx Context, i interface{}) error {
	renderSetContentType(ctx, MimeApplicationProtobuf)
	return NewProtobufEncoder(ctx).Encode(i)
}

// RenderHTML 函数使用模板创建一个模板Renderer。
//
// 从ctx.Value(eudore.ContextKeyTemplate)加载*template.Template，
// 从ctx.GetParam("template")加载模板函数。
func RenderHTML(ctx Context, data interface{}) error {
	t, ok := ctx.Value(ContextKeyTemplate).(*template.Template)
	if ok {
		// 模板必须加载name，防止渲染空模板
		name := ctx.GetParam("template")
		if name != "" {
			t = t.Lookup(name)
			if t != nil {
				renderSetContentType(ctx, MimeTextHTMLCharsetUtf8)
				return t.Execute(ctx, data)
			}
		}
	}

	if DefaultRenderHTMLTemplate != nil {
		b, _ := json.MarshalIndent(data, "", "\t")
		renderSetContentType(ctx, MimeTextHTMLCharsetUtf8)
		return DefaultRenderHTMLTemplate.Execute(ctx, map[string]interface{}{
			"Method":         ctx.Method(),
			"Host":           ctx.Host(),
			"Path":           ctx.Request().RequestURI,
			"Query":          ctx.Querys(),
			"Status":         fmt.Sprintf("%d %s", ctx.Response().Status(), http.StatusText(ctx.Response().Status())),
			"RemoteAddr":     ctx.RealIP(),
			"Params":         ctx.Params(),
			"RequestHeader":  ctx.Request().Header,
			"ResponseHeader": ctx.Response().Header(),
			"Data":           string(b),
			"GodocServer":    DefaultGodocServer,
			"TraceServer":    DefaultTraceServer,
		})
	}
	return ErrRenderHandlerSkip
}

var renderHTMLTempdate = `
<html>
<head>
	<meta charset="utf-8">
	<title>Eudore Look Value </title>
	<meta name="author" content="eudore">
	<meta name="referrer" content="always">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<meta name="description" content="Eudore render text/html">
	<style>
		*{margin:0;padding:0}
		main {display: flex;flex-direction: column;padding: 10px;}
		fieldset {border: 0;}
		.title {font-weight: 900;}
		.name {color: #5f6368;font-weight: 900;margin-left: 20px;}
	</style>
</head>
<body>
	<main>
        <h1>Eudore default render html</h1>
		<fieldset>
			<legend class="title">General</legend>
			<div><span class="name">Request URL: </span><span>{{.Host}}{{.Path}}</span></div>
			<div><span class="name">Request Method: </span><span>{{.Method}}</span></div>
			<div><span class="name">Status Code: </span><span>{{.Status}}</span></div>
			<div><span class="name">Remote Address: </span><span>{{.RemoteAddr}}</span></div>
		</fieldset>
		{{- if ne (len .Query) 0 }}
		<fieldset>
			<legend class="title">Requesst Querys</legend>
			{{- range $key, $vals := .Query -}}
			{{- range $i, $val := $vals }}
			<div><span class="name">{{$key}}: </span><span>{{$val}}</span></div>
			{{- end }}
			{{- end }}
		</fieldset>
		{{- end }}
		<fieldset>
			<legend class="title">Requesst Params</legend>
			{{- $iskey := true }}
			{{- range $i,$val := .Params}}
			{{- if $iskey}}
			<div><span class="name">{{$val}}: </span>{{- else}}<span>{{$val}}</span></div>{{end}}
			{{- $iskey = not $iskey}}
			{{- end}}
		</fieldset>
		<fieldset>
			<legend class="title">Request Headers</legend>
			{{- range $key, $vals := .RequestHeader -}}
			{{- range $i, $val := $vals }}
			<div><span class="name">{{$key}}: </span><span>{{$val}}</span></div>
			{{- end }}
			{{- end }}
		</fieldset>
		<fieldset>
			{{- $trace := .TraceServer }}
			<legend class="title">Response Headers</legend>
			{{- range $key, $vals := .ResponseHeader -}}
			{{- range $i, $val := $vals }}
			{{- if and (eq $key "X-Trace-Id") (ne $trace "")}}
			<div><span class="name">{{$key}}: </span><span><a href="{{$trace}}/trace/{{$val}}">{{$val}}</a></span></div>
			{{- else }}
			<div><span class="name">{{$key}}: </span><span>{{$val}}</span></div>
			{{- end }}
			{{- end }}
			{{- end }}
		</fieldset>
		<fieldset>
			<legend class="title">Response Data</legend>
			<pre class="name">{{.Data}}</pre>
		</fieldset>
	</main>
</body></html>
`

// NewValidateField 方法创建结构体属性校验器。
//
// 使用结构体tag validate从FuncCreator获取校验函数。
// 获取ContextKeyFuncCreator.(FuncCreator)用于创建校验函数。
func NewValidateField(ctx context.Context) HandlerDataFunc {
	fc, ok := ctx.Value(ContextKeyFuncCreator).(FuncCreator)
	if !ok {
		fc = DefaultFuncCreator
	}
	validater := &validateField{
		ValidateTypes: make(map[reflect.Type][]validateFieldValue),
		FuncCreator:   fc,
	}
	return func(_ Context, i interface{}) error {
		return validater.Validate(i)
	}
}

type validateField struct {
	sync.Map
	ValidateTypes map[reflect.Type][]validateFieldValue
	FuncCreator   FuncCreator
}

type validateFieldValue struct {
	Index  int
	Value  reflect.Value
	Format string
}

// Validate 方法校验一个对象属性。
//
// 允许类型为struct []struct []*struct []interface
func (v *validateField) Validate(i interface{}) error {
	iValue := reflect.Indirect(reflect.ValueOf(i))
	switch iValue.Kind() {
	case reflect.Struct:
		return v.validate(iValue, nil)
	case reflect.Slice, reflect.Array:
		switch iValue.Type().Elem().Kind() {
		case reflect.Struct:
			// []struct
			vfs, err := v.parseStructFields(iValue.Type().Elem())
			if err != nil || len(vfs) == 0 {
				return err
			}
			for i := 0; i < iValue.Len(); i++ {
				err = v.validate(iValue.Index(i), vfs)
				if err != nil {
					return err
				}
			}
		case reflect.Interface, reflect.Ptr:
			// []*struct
			// []interface{}{*structA}
			for i := 0; i < iValue.Len(); i++ {
				field := reflect.Indirect(iValue.Index(i))
				if field.Kind() == reflect.Struct {
					err := v.validate(field, nil)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (v *validateField) validate(iValue reflect.Value, vfs []validateFieldValue) error {
	if vfs == nil {
		var err error
		vfs, err = v.parseStructFields(iValue.Type())
		if err != nil {
			return err
		}
	}

	// 匹配验证器规则
	for _, i := range vfs {
		field := iValue.Field(i.Index)
		// 反射调用Validater检测函数
		out := i.Value.Call([]reflect.Value{field})
		if !out[0].Bool() {
			return fmt.Errorf(i.Format, field.Interface())
		}
	}
	return nil
}

func (v *validateField) parseStructFields(iType reflect.Type) ([]validateFieldValue, error) {
	data, ok := v.Load(iType)
	if ok {
		return data.([]validateFieldValue), nil
	}

	var vfs []validateFieldValue
	for i := 0; i < iType.NumField(); i++ {
		field := iType.Field(i)
		tags := field.Tag.Get(DefaultValidateTag)
		// 解析tags
		for _, tag := range strings.Split(tags, " ") {
			if tag == "" {
				continue
			}
			fn, err := v.FuncCreator.Create(field.Type, tag)
			if err != nil {
				return nil, fmt.Errorf(ErrFormatParseValidateFieldError, iType.PkgPath(), iType.Name(), field.Name, tag, err.Error())
			}
			vfs = append(vfs, validateFieldValue{
				Index:  i,
				Value:  reflect.ValueOf(fn),
				Format: fmt.Sprintf("validate %s.%s field %s check %s rule fatal, value: %%#v", iType.PkgPath(), iType.Name(), field.Name, tag),
			})
		}
	}

	v.Store(iType, vfs)
	return vfs, nil
}

// FuncCreator 定义校验函数构造器，默认由RouterStd和validateField使用。
type FuncCreator interface {
	Register(string, ...interface{}) error
	Create(reflect.Type, string) (interface{}, error)
}

type funcCreator struct {
	sync.RWMutex
	Logger Logger
	// 验证规则 - 验证类型 - 验证函数
	FuncValues map[string]map[reflect.Type]interface{}
	// 验证规则 - 验证生成函数
	FuncNews map[string]map[reflect.Type]reflect.Value
}

// NewFuncCreator 函数创建默认校验函数构造器。
func NewFuncCreator() FuncCreator {
	fc := &funcCreator{
		Logger:     NewLoggerNull(),
		FuncValues: make(map[string]map[reflect.Type]interface{}),
		FuncNews:   make(map[string]map[reflect.Type]reflect.Value),
	}
	fc.initFunc()
	return fc
}

// Mount 方法获取ContextKeyApp.(Logger)作为默认日志输出。
func (fc *funcCreator) Mount(ctx context.Context) {
	logger, ok := ctx.Value(ContextKeyApp).(Logger)
	if ok {
		fc.Logger = logger.WithField("creator", "funcCreator").WithField("depth", 2).WithField("logger", true)
		fc.initFunc()
	}
}

func (fc *funcCreator) initFunc() {
	fc.Register("nozero", validateIntNozero, validateStringNozero, validateInterfaceNozero)
	fc.Register("isnum", validateStringIsnum)
	fc.Register("min", validateNewIntMin, validateNewStringMin)
	fc.Register("max", validateNewIntMax, validateNewStringMax)
	fc.Register("len", validateNewStringLen, validateNewInterfaceLen)
	fc.Register("regexp", validateNewStringRegexp)
}

// Register 函数给一个名称注册多个类型的的ValidateFunc或ValidateNewFunc。
//
// ValidateFunc func(T) bool
//
// ValidateNewFunc func(string) (func(T) bool, error)
func (fc *funcCreator) Register(name string, fns ...interface{}) error {
	fc.Lock()
	defer fc.Unlock()
	var errs errormulit
	for _, fn := range fns {
		errs.HandleError(fc.registerFunc(name, fn))
	}
	return errs.Unwrap()
}

// registerFunc 函数注册一个ValidateFunc或ValidateNewFunc
func (fc *funcCreator) registerFunc(name string, fn interface{}) error {
	iType := reflect.TypeOf(fn)

	if checkValidateFunc(iType) {
		if fc.FuncValues[name] == nil {
			fc.FuncValues[name] = make(map[reflect.Type]interface{})
		}
		fc.FuncValues[name][iType.In(0)] = fn
		fc.Logger.Debugf("Register func %s %T", name, fn)
		return nil
	}

	if iType.Kind() == reflect.Func && iType.NumIn() == 1 && iType.NumOut() == 2 && iType.In(0) == typeString && iType.Out(1) == typeError {
		fType := iType.Out(0)
		if checkValidateFunc(fType) {
			if fc.FuncNews[name] == nil {
				fc.FuncNews[name] = make(map[reflect.Type]reflect.Value)
			}
			fc.FuncNews[name][fType.In(0)] = reflect.ValueOf(fn)
			return nil
		}
	}

	err := fmt.Errorf(ErrFormatFuncCreatorRegisterInvalidType, name, fn)
	fc.Logger.Error(err)
	return err
}

// Create 方法获取或创建一个校验函数。
// func(Type) bool/ func(interface{}) bool/ error/ func(Type) Func
func (fc *funcCreator) Create(iType reflect.Type, fullname string) (interface{}, error) {
	fc.RLock()
	fvs, ok := fc.FuncValues[fullname]
	if ok {
		fn, ok := fvs[iType]
		if ok {
			fc.RUnlock()
			return fn, nil
		}
	}
	fc.RUnlock()

	// 升级锁
	fc.Lock()
	defer fc.Unlock()

	name, arg := getValidateNameArg(fullname)
	if arg != "" {
		fns, ok := fc.FuncNews[name]
		if ok {
			fn, ok := fns[iType]
			if ok {
				vals := fn.Call([]reflect.Value{reflect.ValueOf(arg)})
				fn, err := vals[0].Interface(), vals[1].Interface()
				if err != nil {
					fc.Logger.Errorf("Create func %s error: %v", fullname, err)
					return nil, err.(error)
				}
				fc.registerFunc(fullname, fn)
				return fn, nil
			}
		}
	}

	return nil, fmt.Errorf(ErrFormatFuncCreatorNotFunc, fullname)
}

func checkValidateFunc(iType reflect.Type) bool {
	if iType.Kind() != reflect.Func {
		return false
	}
	if iType.NumIn() != 1 || iType.NumOut() != 1 {
		return false
	}
	if iType.Out(0) != typeBool {
		return false
	}
	return true
}

func getValidateNameArg(name string) (string, string) {
	for i, b := range name {
		// ! [0-9A-Za-z]
		if b < 0x30 || (0x39 < b && b < 0x41) || (0x5A < b && b < 0x61) || 0x7A < b {
			return name[:i], name[i:]
		}
	}
	return name, ""
}
func getValidateNameNumber(name string) string {
	var number string
	for i, b := range name {
		if 0x2F < b && b < 0x3A {
			number += name[i : i+1]
		}
	}
	return number
}

// validateIntNozero 函数验证一个int是否为零
func validateIntNozero(num int) bool {
	return num != 0
}

// validateStringNozero 函数验证一个字符串是否为空
func validateStringNozero(str string) bool {
	return str != ""
}

// validateInterfaceNozero 函数验证一个对象是否为零值，使用reflect.Value.IsZero函数实现。
func validateInterfaceNozero(i interface{}) bool {
	return !reflect.ValueOf(i).IsZero()
}

// validateStringIsnum 函数验证一个字符串是否为数字。
func validateStringIsnum(str string) bool {
	_, err := strconv.Atoi(str)
	return err == nil
}

// validateNewIntMin 函数生成一个验证int最小值的验证函数。
func validateNewIntMin(str string) (func(int) bool, error) {
	str = getValidateNameNumber(str)
	min, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return nil, err
	}
	intmin := int(min)
	return func(num int) bool {
		return num >= intmin
	}, nil
}

// validateNewIntMax 函数生成一个验证int最大值的验证函数。
func validateNewIntMax(str string) (func(int) bool, error) {
	str = getValidateNameNumber(str)
	max, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return nil, err
	}
	intmax := int(max)
	return func(num int) bool {
		return num <= intmax
	}, nil
}

// validateNewStringMin 函数生成一个验证string最小值的验证函数。
func validateNewStringMin(str string) (func(string) bool, error) {
	str = getValidateNameNumber(str)
	min, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return nil, err
	}
	intmin := int(min)
	return func(arg string) bool {
		num, err := strconv.Atoi(arg)
		if err != nil {
			return false
		}
		return num >= intmin
	}, nil
}

// validateNewStringMax 函数生成一个验证string最大值的验证函数。
func validateNewStringMax(str string) (func(string) bool, error) {
	str = getValidateNameNumber(str)
	max, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return nil, err
	}
	intmax := int(max)
	return func(arg string) bool {
		num, err := strconv.Atoi(arg)
		if err != nil {
			return false
		}
		return num <= intmax
	}, nil
}

// validateNewStringLen 函数生一个验证字符串长度'>','<','='指定长度的验证函数。
func validateNewStringLen(str string) (func(s string) bool, error) {
	var flag string
	for _, i := range []string{">", "<", "=", ""} {
		if strings.HasPrefix(str, i) {
			flag = i
			str = str[len(i):]
			break
		}
	}

	length, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return nil, err
	}
	intlength := int(length)
	switch flag {
	case ">":
		return func(s string) bool {
			return len(s) > intlength
		}, nil
	case "<":
		return func(s string) bool {
			return len(s) < intlength
		}, nil
	default:
		return func(s string) bool {
			return len(s) == intlength
		}, nil
	}
}

// validateNewInterfaceLen 函数生一个验证字符串长度'>','<','='指定长度的验证函数。
func validateNewInterfaceLen(str string) (func(i interface{}) bool, error) {
	var flag string
	for _, i := range []string{">", "<", "=", ""} {
		if strings.HasPrefix(str, i) {
			flag = i
			str = str[len(i):]
			break
		}
	}

	length, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return nil, err
	}
	intlength := int(length)
	switch flag {
	case ">":
		return func(i interface{}) bool {
			iValue := reflect.Indirect(reflect.ValueOf(i))
			switch iValue.Kind() {
			case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
				return iValue.Len() > intlength
			default:
				return false
			}
		}, nil
	case "<":
		return func(i interface{}) bool {
			iValue := reflect.Indirect(reflect.ValueOf(i))
			switch iValue.Kind() {
			case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
				return iValue.Len() < intlength
			default:
				return false
			}
		}, nil
	default:
		return func(i interface{}) bool {
			iValue := reflect.Indirect(reflect.ValueOf(i))
			switch iValue.Kind() {
			case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
				return iValue.Len() == intlength
			default:
				return false
			}
		}, nil
	}
}

// validateNewStringRegexp 函数生成一个正则检测字符串的验证函数。
func validateNewStringRegexp(str string) (func(arg string) bool, error) {
	re, err := regexp.Compile(str)
	if err != nil {
		return nil, err
	}
	// 返回正则匹配校验函数
	return func(arg string) bool {
		return re.MatchString(arg)
	}, nil
}
