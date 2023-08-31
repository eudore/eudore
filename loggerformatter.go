package eudore

/*
在encoding/json基础上
Func/Chan/UnsafePointer类型输出指针地址
Ptr/Map/Slice类型循环引用时输出指针地址
Invalid类型输出null
Map类型不基于Key进行排序
将fmt.Stringer/errorj接口转换字符串
*/

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"unicode/utf8"
	"unsafe"
)

var (
	loggerLevelDefaultBytes = [][]byte{
		[]byte("DEBUG"), []byte("INFO"),
		[]byte("WARNING"), []byte("ERROR"), []byte("FATAL"),
	}
	loggerLevelDefaultLen = []int{5, 4, 7, 5, 5}
	loggerLevelColorBytes = [][]byte{
		[]byte("\x1b[37mDEBUG\x1b[0m"),
		[]byte("\x1b[36mINFO\x1b[0m"), []byte("\x1b[33mWARNING\x1b[0m"),
		[]byte("\x1b[31mERROR\x1b[0m"), []byte("\x1b[31mFATAL\x1b[0m"),
	}
	loggerendpart1     = []byte("\"}\r\n")
	loggerendpart2     = []byte("}\r\n")
	_hex               = "0123456789abcdef"
	storageJSONEncoder sync.Map
	storageTextEncoder sync.Map
	tableEncodeTypePtr = [...]bool{
		reflect.Chan:          true,
		reflect.Func:          true,
		reflect.Interface:     true,
		reflect.Map:           true,
		reflect.Ptr:           true,
		reflect.Slice:         true,
		reflect.UnsafePointer: true,
	}
	tableEncodeTypeValue = [...]bool{
		reflect.Bool: true, reflect.Int: true, reflect.Uint: true, reflect.String: true,
		reflect.Int8: true, reflect.Int16: true, reflect.Int32: true, reflect.Int64: true,
		reflect.Uint8: true, reflect.Uint16: true, reflect.Uint32: true, reflect.Uint64: true,
		reflect.Float32: true, reflect.Float64: true,
		reflect.Complex64: true, reflect.Complex128: true,
		reflect.Chan: true, reflect.Func: true,
		reflect.Uintptr: true, reflect.UnsafePointer: true,
	}
)

type loggerFormatterText struct {
	TimeFormat string
}

// NewLoggerFormatterText 函数创建文件行格式日志格式化。
func NewLoggerFormatterText(timeformat string) LoggerHandler {
	return &loggerFormatterText{
		TimeFormat: timeformat + " ",
	}
}

func (h *loggerFormatterText) HandlerPriority() int { return 30 }
func (h *loggerFormatterText) HandlerEntry(entry *LoggerEntry) {
	en := &loggerEncoder{
		data: entry.Buffer,
	}
	en.data = entry.Time.AppendFormat(en.data, h.TimeFormat)
	en.data = append(en.data, loggerLevelDefaultBytes[entry.Level]...)
	pos := sliceLastIndex(entry.Keys, "file")
	if pos != -1 {
		if file, ok := entry.Vals[pos].(string); ok {
			en.data = append(en.data, ' ')
			en.formatString(file)
			entry.Keys = entry.Keys[:pos+copy(entry.Keys[pos:], entry.Keys[pos+1:])]
			entry.Vals = entry.Vals[:pos+copy(entry.Vals[pos:], entry.Vals[pos+1:])]
		}
	}
	if entry.Message != "" {
		en.data = append(en.data, ' ')
		en.data = append(en.data, []byte(entry.Message)...)
	}

	for i := range entry.Keys {
		en.data = append(en.data, ' ')
		en.formatString(entry.Keys[i])
		en.data = append(en.data, '=')
		en.formatText(reflect.ValueOf(entry.Vals[i]))
	}
	en.data = append(en.data, '\r', '\n')
	entry.Buffer = en.data
}

type loggerFormatterJSON struct {
	TimeFormat string
	KeyMessage []byte
	KeyTime    []byte
	KeyLevel   []byte
}

// NewLoggerStdDataJSON 函数创建一个LoggerStd的JSON数据处理器。
//
// 如果设置EnvEudoreDaemonEnable表示为后台运行，非终端启动会自动设置Std=false；
// 在非windows系统下，仅输出到终端不输出到文件时Level关键字会设置为彩色。
func NewLoggerFormatterJSON(timeformat string) LoggerHandler {
	return &loggerFormatterJSON{
		TimeFormat: timeformat,
		KeyTime:    []byte(`{"` + DefaultLoggerFormatterKeyTime + `":"`),
		KeyLevel:   []byte(`","` + DefaultLoggerFormatterKeyLevel + `":"`),
		KeyMessage: []byte(`,"` + DefaultLoggerFormatterKeyMessage + `":"`),
	}
}

func (h *loggerFormatterJSON) HandlerPriority() int { return 30 }
func (h *loggerFormatterJSON) HandlerEntry(entry *LoggerEntry) {
	en := &loggerEncoder{
		data: entry.Buffer,
	}
	en.data = append(en.data, h.KeyTime...)
	en.data = entry.Time.AppendFormat(en.data, h.TimeFormat)
	en.data = append(en.data, h.KeyLevel...)
	en.data = append(en.data, loggerLevelDefaultBytes[entry.Level]...)
	en.data = append(en.data, '"')

	for i := range entry.Keys {
		en.data = append(en.data, ',', '"')
		en.data = append(en.data, entry.Keys[i]...)
		en.data = append(en.data, '"', ':')
		en.formatJSON(reflect.ValueOf(entry.Vals[i]))
	}

	if len(entry.Message) > 0 {
		en.data = append(en.data, h.KeyMessage...)
		en.formatString(entry.Message)
		en.data = append(en.data, loggerendpart1...)
	} else {
		en.data = append(en.data, loggerendpart2...)
	}
	entry.Buffer = en.data
}

type loggerEncoder struct {
	data     []byte
	pointers []uintptr
}

type typeEncoder func(*loggerEncoder, reflect.Value)

func (en *loggerEncoder) formatText(v reflect.Value) {
	if !v.IsValid() {
		en.WriteString("null")
		return
	}

	t := v.Type()
	if tableEncodeTypeValue[t.Kind()] && t.NumMethod() == 0 {
		valueEncoder(en, v)
		return
	} else if v.Kind() == reflect.Ptr && !v.IsNil() {
		if t.Implements(typeError) {
			errorEncoder(en, v)
			return
		} else if t.Implements(typeFmtStringer) {
			fmtStringerEncoder(en, v)
			return
		}
		v = reflect.Indirect(v)
		en.WriteBytes('&')
	}

	newTextEncoder(v.Type())(en, v)
}

func (en *loggerEncoder) formatJSON(v reflect.Value) {
	if !v.IsValid() {
		en.WriteString("null")
		return
	}

	newJSONEncoder(v.Type())(en, v)
}

func newTextEncoder(t reflect.Type) typeEncoder {
	if tableEncodeTypeValue[t.Kind()] && t.NumMethod() == 0 {
		return valueEncoder
	}
	e, ok := storageTextEncoder.Load(t)
	if ok {
		return e.(typeEncoder)
	}

	e = parseTextEncoder(t)
	storageTextEncoder.Store(t, e)
	return e.(typeEncoder)
}

func parseTextEncoder(t reflect.Type) typeEncoder {
	if t.Implements(typeError) {
		return errorEncoder
	} else if t.Implements(typeFmtStringer) {
		return fmtStringerEncoder
	}

	switch t.Kind() {
	case reflect.Struct:
		e := &structEncoder{}
		storageTextEncoder.Store(t, typeEncoder(e.encodeText))
		e.Fields = parseTextStructFields(t)
		return e.encodeText
	case reflect.Map:
		e := mapEnocder{newTextEncoder(t.Key()), newTextEncoder(t.Elem()), parseTextMapName(t)}
		e.Prefix += "{"
		return e.encode
	case reflect.Slice, reflect.Array:
		e := sliceEnocder{newTextEncoder(t.Elem())}
		return e.encode
	case reflect.Interface:
		e := anyEnocder{newTextEncoder}
		return e.encode
	default:
		return valueEncoder
	}
}

func parseTextStructFields(iType reflect.Type) []encodeJSONField {
	var fields []encodeJSONField
	for i := 0; i < iType.NumField(); i++ {
		t := iType.Field(i)
		fields = append(fields, encodeJSONField{
			Index:   i,
			Name:    t.Name,
			Encoder: newTextEncoder(t.Type),
		})
	}
	return fields
}

func parseTextMapName(iType reflect.Type) string {
	if iType.Name() == "" {
		return "map"
	}
	return iType.String()
}

func newJSONEncoder(t reflect.Type) typeEncoder {
	if tableEncodeTypeValue[t.Kind()] && t.NumMethod() == 0 {
		return valueEncoder
	}
	e, ok := storageJSONEncoder.Load(t)
	if ok {
		return e.(typeEncoder)
	}

	e = parseJSONEncoder(t)
	storageJSONEncoder.Store(t, e)
	return e.(typeEncoder)
}

func parseJSONEncoder(t reflect.Type) typeEncoder {
	switch {
	case t.Implements(typeJSONMarshaler):
		return jsonMarshalerEncoder
	case t.Implements(typeTextMarshaler):
		return textMarshalerEncoder
	case t.Implements(typeError):
		return errorEncoder
	case t.Implements(typeFmtStringer):
		return fmtStringerEncoder
	}

	switch t.Kind() {
	case reflect.Struct:
		e := &structEncoder{}
		storageJSONEncoder.Store(t, typeEncoder(e.encode))
		e.Fields = parseJSONStructFields(t)
		return e.encode
	case reflect.Map:
		e := mapEnocder{newJSONEncoder(t.Key()), newJSONEncoder(t.Elem()), ""}
		e.Prefix = "{"
		return e.encode
	case reflect.Slice, reflect.Array:
		e := sliceEnocder{newJSONEncoder(t.Elem())}
		return e.encode
	case reflect.Ptr:
		e := ptrEnocder{newJSONEncoder(t.Elem())}
		return e.encode
	case reflect.Interface:
		e := anyEnocder{newJSONEncoder}
		return e.encode
	default:
		return valueEncoder
	}
}

func parseJSONStructFields(iType reflect.Type) []encodeJSONField {
	var fields []encodeJSONField
	for i := 0; i < iType.NumField(); i++ {
		t := iType.Field(i)
		name, omit := cutOmit(t.Tag.Get("json"))
		if name == "-" || t.Name[0] < 'A' || t.Name[0] > 'Z' {
			continue
		}
		if name == "" {
			name = t.Name
		}

		field := encodeJSONField{
			Index: i,
			Name:  name,
			Omit:  omit,
		}
		if t.Anonymous && t.Type.Kind() == reflect.Struct {
			e := &structEncoder{}
			e.Fields = parseJSONStructFields(t.Type)
			field.Anonymous = true
			field.Encoder = e.encodeFields
		} else {
			field.Encoder = newJSONEncoder(t.Type)
		}
		fields = append(fields, field)
	}
	return fields
}

type ptrEnocder struct {
	Elem typeEncoder
}

func (e ptrEnocder) encode(en *loggerEncoder, v reflect.Value) {
	if en.formatVia(v) {
		return
	}
	defer en.releaseVia()
	e.Elem(en, v.Elem())
}

type anyEnocder struct {
	newEncoder func(reflect.Type) typeEncoder
}

func (e anyEnocder) encode(en *loggerEncoder, v reflect.Value) {
	if v.IsNil() {
		en.WriteString("null")
		return
	}
	v = v.Elem()
	e.newEncoder(v.Type())(en, v)
}

type mapEnocder struct {
	Key    typeEncoder
	Val    typeEncoder
	Prefix string
}

func (e mapEnocder) encode(en *loggerEncoder, v reflect.Value) {
	if en.formatVia(v) {
		return
	}
	defer en.releaseVia()
	en.WriteString(e.Prefix)
	pos := len(en.data)
	i := v.MapRange()
	for i.Next() {
		e.Key(en, i.Key())
		en.WriteBytes(':')
		e.Val(en, i.Value())
		en.WriteBytes(',')
	}
	if pos == len(en.data) {
		en.WriteBytes('}')
	} else {
		en.data[len(en.data)-1] = '}'
	}
}

type sliceEnocder struct {
	Elem typeEncoder
}

func (e sliceEnocder) encode(en *loggerEncoder, v reflect.Value) {
	if en.formatVia(v) {
		return
	}
	defer en.releaseVia()
	en.WriteBytes('[')
	pos := len(en.data)
	for i := 0; i < v.Len(); i++ {
		e.Elem(en, v.Index(i))
		en.WriteBytes(',')
	}
	if pos == len(en.data) {
		en.WriteBytes(']')
	} else {
		en.data[len(en.data)-1] = ']'
	}
}

type structEncoder struct {
	Fields []encodeJSONField
}
type encodeJSONField struct {
	Index     int
	Name      string
	Omit      bool
	Anonymous bool
	Encoder   typeEncoder
}

func (e *structEncoder) encodeText(en *loggerEncoder, v reflect.Value) {
	en.WriteBytes('{')
	pos := len(en.data)
	for _, f := range e.Fields {
		en.WriteString(f.Name)
		en.WriteBytes(':')
		f.Encoder(en, v.Field(f.Index))
		en.WriteBytes(' ')
	}
	if pos == len(en.data) {
		en.WriteBytes('}')
	} else {
		en.data[len(en.data)-1] = '}'
	}
}

func (e *structEncoder) encode(en *loggerEncoder, v reflect.Value) {
	en.WriteBytes('{')
	pos := len(en.data)
	e.encodeFields(en, v)
	if pos == len(en.data) {
		en.WriteBytes('}')
	} else {
		en.data[len(en.data)-1] = '}'
	}
}

func (e *structEncoder) encodeFields(en *loggerEncoder, v reflect.Value) {
	for _, f := range e.Fields {
		v := v.Field(f.Index)
		if f.Anonymous {
			f.Encoder(en, v)
			continue
		}

		if f.Omit && v.IsZero() {
			continue
		}
		en.WriteBytes('"')
		en.WriteString(f.Name)
		en.WriteBytes('"', ':')
		f.Encoder(en, v)
		en.WriteBytes(',')
	}
}

func valueEncoder(en *loggerEncoder, v reflect.Value) {
	// 写入类型
	switch v.Kind() {
	case reflect.Bool:
		en.data = strconv.AppendBool(en.data, v.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		en.data = strconv.AppendInt(en.data, v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		en.data = strconv.AppendUint(en.data, v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		en.data = strconv.AppendFloat(en.data, v.Float(), 'f', -1, 64)
	case reflect.Complex64, reflect.Complex128:
		val := v.Complex()
		en.WriteBytes('"')
		en.data = strconv.AppendFloat(en.data, real(val), 'f', -1, 64)
		en.WriteBytes('+')
		en.data = strconv.AppendFloat(en.data, imag(val), 'f', -1, 64)
		en.WriteBytes('i', '"')
	case reflect.String:
		en.WriteBytes('"')
		en.formatString(v.String())
		en.WriteBytes('"')
	case reflect.Ptr, reflect.Func, reflect.Chan, reflect.UnsafePointer:
		if v.IsNil() {
			en.WriteString("null")
			return
		}
		en.WriteBytes('"', '0', 'x')
		en.data = strconv.AppendUint(en.data, uint64(v.Pointer()), 16)
		en.WriteBytes('"')
	}
}

func errorEncoder(en *loggerEncoder, v reflect.Value) {
	if tableEncodeTypePtr[v.Kind()] && v.IsNil() {
		en.WriteString("null")
		return
	}
	en.WriteBytes('"')
	en.WriteString(v.Interface().(error).Error())
	en.WriteBytes('"')
}

func fmtStringerEncoder(en *loggerEncoder, v reflect.Value) {
	if tableEncodeTypePtr[v.Kind()] && v.IsNil() {
		en.WriteString("null")
		return
	}
	en.WriteBytes('"')
	en.WriteString(v.Interface().(fmt.Stringer).String())
	en.WriteBytes('"')
}

func jsonMarshalerEncoder(en *loggerEncoder, v reflect.Value) {
	if tableEncodeTypePtr[v.Kind()] && v.IsNil() {
		en.WriteString("null")
		return
	}
	body, err := v.Interface().(json.Marshaler).MarshalJSON()
	if err == nil {
		en.WriteBytes(body...)
	} else {
		en.WriteBytes('"')
		en.formatString(err.Error())
		en.WriteBytes('"')
	}
}

func textMarshalerEncoder(en *loggerEncoder, v reflect.Value) {
	if tableEncodeTypePtr[v.Kind()] && v.IsNil() {
		en.WriteString("null")
		return
	}
	body, err := v.Interface().(encoding.TextMarshaler).MarshalText()
	en.WriteBytes('"')
	if err == nil {
		en.formatString(*(*string)(unsafe.Pointer(&body)))
	} else {
		en.formatString(err.Error())
	}
	en.WriteBytes('"')
}

func (en *loggerEncoder) formatVia(v reflect.Value) bool {
	if v.IsNil() {
		en.WriteString("null")
		return true
	}
	ptr := v.Pointer()
	if en.pointers == nil {
		en.pointers = make([]uintptr, 0, 4)
	}
	for _, p := range en.pointers {
		if p == ptr {
			en.WriteBytes('"', '0', 'x')
			en.data = strconv.AppendUint(en.data, uint64(p), 16)
			en.WriteBytes('"')
			return true
		}
	}
	en.pointers = append(en.pointers, ptr)
	return false
}

func (en *loggerEncoder) releaseVia() {
	en.pointers = en.pointers[:len(en.pointers)-1]
}

// formatString 方法安全写入字符串。
func (en *loggerEncoder) formatString(s string) {
	for i := 0; i < len(s); {
		b := s[i]
		if b < utf8.RuneSelf {
			en.addRuneSelf(b)
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		switch r {
		case utf8.RuneError:
			if size == 1 {
				en.WriteString(`\ufffd`)
			}
		case '\u2028', '\u2029':
			en.WriteString(`\u202`)
			en.WriteBytes(_hex[r&0xF])
		default:
			en.WriteString(s[i : i+size])
		}
		i += size
	}
}

func (en *loggerEncoder) addRuneSelf(b byte) {
	if 0x20 <= b && b != '\\' && b != '"' {
		en.WriteBytes(b)
		return
	}
	switch b {
	case '\\', '"':
		en.WriteBytes('\\', b)
	case '\n':
		en.WriteBytes('\\', 'n')
	case '\r':
		en.WriteBytes('\\', 'r')
	case '\t':
		en.WriteBytes('\\', 't')
	default:
		en.WriteString(`\u00`)
		en.WriteBytes(_hex[b>>4], _hex[b&0xF])
	}
}

func (en *loggerEncoder) WriteBytes(b ...byte) {
	en.data = append(en.data, b...)
}

func (en *loggerEncoder) WriteString(s string) {
	b := *(*[]byte)(unsafe.Pointer(&struct {
		string
		Cap int
	}{s, len(s)}))
	en.data = append(en.data, b...)
}
