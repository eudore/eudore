package eudore

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type contextKey struct {
	name string
}

// NewContextKey 定义context key。
func NewContextKey(key string) any {
	return contextKey{key}
}

func (key contextKey) String() string {
	return key.name
}

// Params 定义用于保存一些键值数据。
type Params []string

// NewParamsRoute 方法根据一个路由路径创建Params，支持路由路径块模式。
func NewParamsRoute(path string) Params {
	route := getRoutePath(path)
	args := strings.Split(path[len(route):], " ")
	if args[0] == "" {
		args = args[1:]
	}
	params := make(Params, 0, len(args)*2+2)
	params = append(params, ParamRoute, route)
	for _, str := range args {
		k, v, ok := strings.Cut(str, "=")
		if ok && v != "" {
			params = append(params, k, v)
		}
	}
	return params
}

// Clone 方法深复制一个ParamArray对象。
func (p Params) Clone() Params {
	params := make(Params, len(p))
	copy(params, p)
	return params
}

// CombineWithRoute 方法将params数据合并到p，用于路由路径合并。
func (p Params) CombineWithRoute(params Params) Params {
	p[1] += params[1]
	for i := 2; i < len(params); i += 2 {
		p = p.Set(params[i], params[i+1])
	}
	return p
}

// String 方法输出Params成字符串。
func (p Params) String() string {
	b := &bytes.Buffer{}
	for i := 0; i < len(p); i += 2 {
		if (p[i] != "" && p[i+1] != "") || i == 0 {
			if b.Len() != 0 {
				b.WriteString(" ")
			}
			fmt.Fprintf(b, "%s=%s", p[i], p[i+1])
		}
	}
	return b.String()
}

// MarshalJSON 方法设置Params json序列化显示的数据。
func (p Params) MarshalJSON() ([]byte, error) {
	data := make(map[string]string, len(p)/2)
	for i := 0; i < len(p); i += 2 {
		if p[i+1] != "" || i == 0 {
			data[p[i]] = p[i+1]
		}
	}
	return json.Marshal(data)
}

// Get 方法返回一个参数的值。
func (p Params) Get(key string) string {
	for i := 0; i < len(p); i += 2 {
		if p[i] == key {
			return p[i+1]
		}
	}
	return ""
}

// Add 方法添加一个参数。
func (p Params) Add(vals ...string) Params {
	return append(p, vals...)
}

// Set 方法设置一个参数的值。
func (p Params) Set(key, val string) Params {
	for i := 0; i < len(p); i += 2 {
		if p[i] == key {
			p[i+1] = val
			return p
		}
	}
	return append(p, key, val)
}

// Del 方法删除一个参数值。
func (p Params) Del(key string) {
	for i := 0; i < len(p); i += 2 {
		if p[i] == key {
			p[i+1] = ""
		}
	}
}

// GetWarp 对象封装Get函数提供类型转换功能。
type GetWarp func(string) any

// NewGetWarp 函数创建一个getwarp处理类型转换。
func NewGetWarp(fn func(string) any) GetWarp {
	return fn
}

// NewGetWarpWithConfig 函数使用Config.Get创建getwarp。
func NewGetWarpWithConfig(c Config) GetWarp {
	return c.Get
}

// NewGetWarpWithApp 函数使用App创建getwarp。
func NewGetWarpWithApp(app *App) GetWarp {
	return func(key string) any {
		return app.Get(key)
	}
}

// NewGetWarpWithMapString 函数使用map[string]any创建getwarp。
func NewGetWarpWithMapString(data map[string]any) GetWarp {
	return func(key string) any {
		return data[key]
	}
}

// NewGetWarpWithObject 函数使用map或创建getwarp。
func NewGetWarpWithObject(obj any) GetWarp {
	return func(key string) any {
		return GetAnyByPath(obj, key)
	}
}

// GetAny 方法获取any类型的配置值。
func (fn GetWarp) GetAny(key string) any {
	return fn(key)
}

// GetBool 方法获取bool类型的配置值。
func (fn GetWarp) GetBool(key string) bool {
	return GetAny[bool](fn(key))
}

// GetInt 方法获取int类型的配置值。
func (fn GetWarp) GetInt(key string, vals ...int) int {
	return GetAny(fn(key), vals...)
}

// GetUint 方法取获取uint类型的配置值。
func (fn GetWarp) GetUint(key string, vals ...uint) uint {
	return GetAny(fn(key), vals...)
}

// GetInt64 方法int64类型的配置值。
func (fn GetWarp) GetInt64(key string, vals ...int64) int64 {
	return GetAny(fn(key), vals...)
}

// GetUint64 方法取获取uint64类型的配置值。
func (fn GetWarp) GetUint64(key string, vals ...uint64) uint64 {
	return GetAny(fn(key), vals...)
}

// GetFloat32 方法取获取float32类型的配置值。
func (fn GetWarp) GetFloat32(key string, vals ...float32) float32 {
	return GetAny(fn(key), vals...)
}

// GetFloat64 方法取获取float64类型的配置值。
func (fn GetWarp) GetFloat64(key string, vals ...float64) float64 {
	return GetAny(fn(key), vals...)
}

// GetString 方法获取一个字符串，如果字符串为空返回其他默认非空字符串。
func (fn GetWarp) GetString(key string, vals ...string) string {
	return GetStringByAny(fn(key), vals...)
}

// TimeDuration 定义time.Duration类型处理json。
type TimeDuration time.Duration

// String 方法格式化输出时间。
func (d TimeDuration) String() string {
	return time.Duration(d).String()
}

// MarshalText 方法实现json序列化输出。
func (d TimeDuration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

// UnmarshalJSON 方法实现解析json格式时间。
func (d *TimeDuration) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	return d.UnmarshalText(b)
}

// UnmarshalText 方法实现解析时间。
func (d *TimeDuration) UnmarshalText(b []byte) error {
	str := string(b)
	// parse int64
	val, err := strconv.ParseInt(str, 10, 64)
	if err == nil {
		*d = TimeDuration(val)
		return nil
	}
	// parse string
	t, err := time.ParseDuration(str)
	if err == nil {
		*d = TimeDuration(t)
		return nil
	}
	return fmt.Errorf("invalid duration value: '%s'", b)
}

// mulitError 实现多个error组合。
type mulitError struct {
	errs []error
}

// HandleError 实现处理多个错误，如果非空则保存错误。
func (err *mulitError) HandleError(errs ...error) {
	for _, e := range errs {
		if e != nil {
			err.errs = append(err.errs, e)
		}
	}
}

// Error 方法实现error接口，返回错误描述。
func (err *mulitError) Error() string {
	return fmt.Sprint(err.errs)
}

// GetError 方法返回错误，如果没有保存的错误则返回空。
func (err *mulitError) Unwrap() error {
	switch len(err.errs) {
	case 0:
		return nil
	case 1:
		return err.errs[0]
	default:
		return err
	}
}

// NewErrorWithStatusCode 方法组合ErrorStatus和ErrorCode。
func NewErrorWithStatusCode(err error, status, code int) error {
	if code > 0 {
		err = codeError{err, code}
	}
	if status > 0 {
		err = statusError{err, status}
	}
	return err
}

// NewErrorWithStatus 方法封装error实现Status方法。
func NewErrorWithStatus(err error, status int) error {
	if status > 0 {
		return statusError{err, status}
	}
	return err
}

type statusError struct {
	err    error
	status int
}

func (err statusError) Error() string {
	return err.err.Error()
}

func (err statusError) Unwrap() error {
	return err.err
}

func (err statusError) Status() int {
	return err.status
}

// NewErrorWithCode 方法封装error实现Code方法。
func NewErrorWithCode(err error, code int) error {
	if code > 0 {
		return codeError{err, code}
	}
	return err
}

type codeError struct {
	err  error
	code int
}

func (err codeError) Error() string {
	return err.err.Error()
}

func (err codeError) Unwrap() error {
	return err.err
}

func (err codeError) Code() int {
	return err.code
}

func clearCap[T any](s []T) []T {
	l := len(s)
	return s[:l:l]
}

func cutOmit(s string) (string, bool) {
	if strings.HasSuffix(s, ",omitempty") {
		return s[:len(s)-10], true
	}
	return s, false
}

func sliceIndex[T comparable](vals []T, val T) int {
	for i := range vals {
		if val == vals[i] {
			return i
		}
	}
	return -1
}

func sliceLastIndex[T comparable](vals []T, val T) int {
	for i := len(vals) - 1; i > -1; i-- {
		if val == vals[i] {
			return i
		}
	}
	return -1
}

func sliceFilter[T any](s []T, fn func(T) bool) []T {
	size := 0
	b := make([]bool, len(s))
	for i := range s {
		b[i] = fn(s[i])
		if b[i] {
			size++
		}
	}
	if size == len(s) {
		return s
	}

	n := make([]T, 0, size)
	for i := range b {
		if b[i] {
			n = append(n, s[i])
		}
	}
	return n
}

// GetAnyDefault 函数返回非空值。
func GetAnyDefault[T comparable](arg1, arg2 T) T {
	var zero T
	if arg1 != zero {
		return arg1
	}
	return arg2
}

// GetAnyDefaults 函数返回第一个非空值。
func GetAnyDefaults[T comparable](args ...T) T {
	var zero T
	for i := range args {
		if args[i] != zero {
			return args[i]
		}
	}
	return zero
}

func SetAnyDefault[T any](arg1, arg2 *T) {
	v1 := reflect.Indirect(reflect.ValueOf(arg1))
	v2 := reflect.Indirect(reflect.ValueOf(arg2))
	if v1.Kind() == reflect.Struct && v1.Type() == v2.Type() {
		for i := 0; i < v1.NumField(); i++ {
			f1, f2 := v1.Field(i), v2.Field(i)
			if f1.CanSet() && !f2.IsZero() {
				f1.Set(f2)
			}
		}
	}
}

// TypeNumber 定义泛型数值类型集合。
type TypeNumber interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | float32 | float64 | complex64 | complex128
}

// GetAny 函数类型Value转换成另外一个类型。
func GetAny[T string | bool | TypeNumber](s any, defaults ...T) T {
	var t, zero T
	if s != nil {
		sValue := reflect.ValueOf(s)
		tType := reflect.TypeOf(t)
		switch {
		case sValue.Type() == tType:
			t = sValue.Interface().(T)
		case sValue.Kind() == tType.Kind():
			t = sValue.Convert(tType).Interface().(T)
		case sValue.Kind() == reflect.String:
			t = GetAnyByString(sValue.String(), defaults...)
		case tType.Kind() == reflect.String:
			t = any(GetStringByAny(s)).(T)
		case sValue.CanConvert(tType):
			t = sValue.Convert(tType).Interface().(T)
		}
		if t != zero {
			return t
		}
	}

	for _, value := range defaults {
		if value != zero {
			return value
		}
	}
	return t
}

// GetStringByAny 函数将any转换成string
//
//nolint:cyclop,gocyclo
func GetStringByAny(i any, strs ...string) string {
	var str string
	switch v := i.(type) {
	case string:
		str = v
	case int:
		str = strconv.FormatInt(int64(v), 10)
	case uint:
		str = strconv.FormatUint(uint64(v), 10)
	case float64:
		str = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		str = strconv.FormatBool(v)
	case []byte:
		str = string(v)
	case fmt.Stringer:
		str = v.String()
	case int64:
		str = strconv.FormatInt(v, 10)
	case int32:
		str = strconv.FormatInt(int64(v), 10)
	case int16:
		str = strconv.FormatInt(int64(v), 10)
	case int8:
		str = strconv.FormatInt(int64(v), 10)
	case uint64:
		str = strconv.FormatUint(v, 10)
	case uint32:
		str = strconv.FormatUint(uint64(v), 10)
	case uint16:
		str = strconv.FormatUint(uint64(v), 10)
	case uint8:
		str = strconv.FormatUint(uint64(v), 10)
	case float32:
		str = strconv.FormatFloat(float64(v), 'f', -1, 32)
	case complex64:
		str = strconv.FormatComplex(complex128(v), 'f', -1, 64)
	case complex128:
		str = strconv.FormatComplex(v, 'f', -1, 128)
	default:
		str = fmt.Sprint(i)
	}

	if str != "" {
		return str
	}
	for _, i := range strs {
		if i != "" {
			return i
		}
	}
	return ""
}

// GetStringRandom 函数返回指定长度随机字符串。
func GetStringRandom(length int) string {
	buf := make([]byte, length)
	io.ReadFull(rand.Reader, buf)
	return fmt.Sprintf("%x", buf)
}

// GetAnyByString 函数将字符串转换为其他值。
func GetAnyByString[T string | bool | TypeNumber | time.Time | time.Duration](str string, defaults ...T) T {
	val, _ := GetAnyByStringWithError(str, defaults...)
	return val
}

// GetAnyByStringWithError 函数将字符串转换成泛型数值。
//
//nolint:cyclop,funlen,gocyclo
func GetAnyByStringWithError[T string | bool | TypeNumber | time.Time | time.Duration](str string, defaults ...T) (T, error) {
	var zero T
	var val any
	var err error
	switch any(zero).(type) {
	case int:
		val, err = strconv.Atoi(str)
	case float64:
		val, err = strconv.ParseFloat(str, 64)
	case string:
		val = str
	case bool:
		val, err = strconv.ParseBool(str)
	case int8:
		var v int64
		v, err = strconv.ParseInt(str, 10, 8)
		val = int8(v)
	case int16:
		var v int64
		v, err = strconv.ParseInt(str, 10, 16)
		val = int16(v)
	case int32:
		var v int64
		v, err = strconv.ParseInt(str, 10, 16)
		val = int32(v)
	case int64:
		val, err = strconv.ParseInt(str, 10, 64)
	case uint:
		var v uint64
		v, err = strconv.ParseUint(str, 10, 32)
		val = uint(v)
	case uint8:
		var v uint64
		v, err = strconv.ParseUint(str, 10, 8)
		val = uint8(v)
	case uint16:
		var v uint64
		v, err = strconv.ParseUint(str, 10, 16)
		val = uint16(v)
	case uint32:
		var v uint64
		v, err = strconv.ParseUint(str, 10, 32)
		val = uint32(v)
	case uint64:
		val, err = strconv.ParseUint(str, 10, 64)
	case float32:
		var v float64
		v, err = strconv.ParseFloat(str, 32)
		val = float32(v)
	case complex64:
		var v complex128
		v, err = strconv.ParseComplex(str, 64)
		val = complex64(v)
	case complex128:
		val, err = strconv.ParseComplex(str, 128)
	case time.Duration:
		val, err = time.ParseDuration(str)
	case time.Time:
		var v time.Time
		for i, f := range DefaultValueParseTimeFormats {
			if DefaultValueParseTimeFixed[i] && len(str) != len(f) {
				continue
			}
			v, err = time.Parse(f, str)
			if err == nil {
				break
			}
		}
		val = v
	}
	if val != zero {
		return val.(T), err
	}
	for _, value := range defaults {
		if value != zero {
			return value, err
		}
	}
	return zero, err
}
