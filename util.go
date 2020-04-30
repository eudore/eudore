package eudore

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// split2byte internal function, splits two strings into two segments using the first specified byte, and returns "", str if there is no split symbol.
//
// split2byte 内部函数，使用第一个指定byte两字符串分割成两段，如果不存在分割符号，返回"", str。
func split2byte(str string, b byte) (string, string) {
	pos := strings.IndexByte(str, b)
	if pos == -1 {
		return str, ""
	}
	return str[:pos], str[pos+1:]
}

// GetBool 使用GetDefaultBool，默认false。
func GetBool(i interface{}) bool {
	return GetDefaultBool(i, false)
}

// GetDefaultBool 函数转换bool、int、uint、float、string成bool。
func GetDefaultBool(i interface{}, b bool) bool {
	if v, ok := i.(bool); ok {
		return v
	}

	i = isNum(i)
	if i == nil {
		return b
	}
	if v, err := strconv.ParseBool(fmt.Sprint(i)); err == nil {
		return v
	}
	return b
}

// GetInt 使用GetDefaultInt，默认0.
func GetInt(i interface{}) int {
	return GetDefaultInt(i, 0)
}

// GetDefaultInt 函数转换一个int、uint、float、string类型成int。
func GetDefaultInt(i interface{}, n int) int {
	i = isNum(i)
	if i == nil {
		return n
	}
	if v, ok := i.(int64); ok {
		return int(v)
	}
	if v, ok := i.(uint64); ok {
		return int(v)
	}
	if v, ok := i.(float64); ok {
		return int(v)
	}
	if s, ok := i.(string); ok {
		if v, err := strconv.Atoi(s); err == nil {
			return int(v)
		}
	}
	return n
}

// GetInt64 使用GetDefaultInt64，默认0.
func GetInt64(i interface{}) int64 {
	return GetDefaultInt64(i, 0)
}

// GetDefaultInt64 函数转换一个int、uint、float、string类型成int64。
func GetDefaultInt64(i interface{}, n int64) int64 {
	i = isNum(i)
	if i == nil {
		return n
	}
	if v, ok := i.(int64); ok {
		return v
	}
	if v, ok := i.(uint64); ok {
		return int64(v)
	}
	if v, ok := i.(float64); ok {
		return int64(v)
	}
	if s, ok := i.(string); ok {
		if v, err := strconv.ParseInt(s, 10, 64); err == nil {
			return v
		}
	}
	return n
}

// GetUint 使用GetDefaultUint，默认0。
func GetUint(i interface{}) uint {
	return GetDefaultUint(i, 0)
}

// GetDefaultUint 函数转换一个int、uint、float、string类型成uint。
func GetDefaultUint(i interface{}, n uint) uint {
	i = isNum(i)
	if i == nil {
		return n
	}
	if v, ok := i.(uint64); ok {
		return uint(v)
	}
	if v, ok := i.(int64); ok {
		return uint(v)
	}
	if v, ok := i.(float64); ok {
		return uint(v)
	}
	if s, ok := i.(string); ok {
		if v, err := strconv.ParseUint(s, 10, 64); err == nil {
			return uint(v)
		}
	}
	return n
}

// GetUint64 使用GetDefaultUint64，默认0.
func GetUint64(i interface{}) uint64 {
	return GetDefaultUint64(i, 0)
}

// GetDefaultUint64 函数转换一个int、uint、float、string类型成uint64。
func GetDefaultUint64(i interface{}, n uint64) uint64 {
	i = isNum(i)
	if i == nil {
		return n
	}
	if v, ok := i.(uint64); ok {
		return v
	}
	if v, ok := i.(int64); ok {
		return uint64(v)
	}
	if v, ok := i.(float64); ok {
		return uint64(v)
	}
	if s, ok := i.(string); ok {
		if v, err := strconv.ParseUint(s, 10, 64); err == nil {
			return v
		}
	}
	return n
}

// GetFloat32 使用GetDefaultFloat32，默认0.
func GetFloat32(i interface{}) float32 {
	return GetDefaultFloat32(i, 0)
}

// GetDefaultFloat32 函数转换一个int、uint、float、string类型成float32。
func GetDefaultFloat32(i interface{}, n float32) float32 {
	i = isNum(i)
	if i == nil {
		return n
	}
	if v, ok := i.(float64); ok {
		return float32(v)
	}
	if v, ok := i.(int64); ok {
		return float32(v)
	}
	if v, ok := i.(uint64); ok {
		return float32(v)
	}
	if s, ok := i.(string); ok {
		if v, err := strconv.ParseFloat(s, 32); err == nil {
			return float32(v)
		}
	}
	return n
}

// GetFloat64 使用GetDefaultFloat64，默认0.
func GetFloat64(i interface{}) float64 {
	return GetDefaultFloat64(i, 0)
}

// GetDefaultFloat64 函数转换一个int、uint、float、string类型成float64。
func GetDefaultFloat64(i interface{}, n float64) float64 {
	i = isNum(i)
	if i == nil {
		return n
	}
	if v, ok := i.(float64); ok {
		return v
	}
	if v, ok := i.(int64); ok {
		return float64(v)
	}
	if v, ok := i.(uint64); ok {
		return float64(v)
	}

	if s, ok := i.(string); ok {
		if v, err := strconv.ParseFloat(s, 64); err == nil {
			return v
		}
	}
	return n
}

// GetString 使用GetDefaultString，默认空字符。
func GetString(i interface{}) string {
	return GetDefaultString(i, "")
}

// GetDefaultString 通过断言string类型实现转换。
func GetDefaultString(i interface{}, str string) string {
	if v, ok := i.(bool); ok {
		return fmt.Sprint(v)
	}
	i = isNum(i)
	if i != nil {
		return fmt.Sprint(i)
	}
	return str
}

func isNum(i interface{}) interface{} {
	switch val := i.(type) {
	case string, int64, uint64, float64:
		return val
	case int:
		return int64(val)
	case int8:
		return int64(val)
	case int16:
		return int64(val)
	case int32:
		return int64(val)
	case uint:
		return uint64(val)
	case uint8:
		return uint64(val)
	case uint16:
		return uint64(val)
	case uint32:
		return uint64(val)
	case []byte:
		return string(val)
	case float32:
		return float64(val)
	case bool:
		if val {
			return int64(1)
		}
		return int64(0)
	}
	return nil
}

// GetArrayString 转换成[]string
func GetArrayString(i interface{}) []string {
	if i == nil {
		return nil
	}
	str, ok := i.(string)
	if ok {
		return []string{str}
	}
	strs, ok := i.([]string)
	if ok {
		return strs
	}
	iType := reflect.TypeOf(i)
	if iType.Kind() == reflect.Slice || iType.Kind() == reflect.Array {
		iValue := reflect.ValueOf(i)
		strs = make([]string, iValue.Len())
		for i := 0; i < iValue.Len(); i++ {
			strs[i] = fmt.Sprintf("%#v", iValue.Index(i))
		}
		return strs
	}
	return nil
}

// GetStringBool 使用GetStringDefaultBool，默认false。
func GetStringBool(str string) bool {
	return GetStringDefaultBool(str, false)
}

// GetStringDefaultBool 使用strconv.ParseBool解析。
func GetStringDefaultBool(str string, b bool) bool {
	if v, err := strconv.ParseBool(str); err == nil {
		return v
	}
	return b
}

// GetStringInt 使用GetStringDefaultInt，默认0.
func GetStringInt(str string) int {
	return GetStringDefaultInt(str, 0)
}

// GetStringDefaultInt 使用strconv.Atoi解析。
func GetStringDefaultInt(str string, n int) int {
	if v, err := strconv.Atoi(str); err == nil {
		return v
	}
	return n
}

// GetStringInt64 使用GetStringDefaultInt64，默认0.
func GetStringInt64(str string) int64 {
	return GetStringDefaultInt64(str, 0)
}

// GetStringDefaultInt64 使用strconv.ParseInt解析。
func GetStringDefaultInt64(str string, n int64) int64 {
	if v, err := strconv.ParseInt(str, 10, 64); err == nil {
		return v
	}
	return n
}

// GetStringUint 使用GetStringDefaultUint，默认0.
func GetStringUint(str string) uint {
	return GetStringDefaultUint(str, 0)
}

// GetStringDefaultUint 使用strconv.ParseUint解析。
func GetStringDefaultUint(str string, n uint) uint {
	if v, err := strconv.ParseUint(str, 10, 64); err == nil {
		return uint(v)
	}
	return n
}

// GetStringUint64 使用GetStringDefaultUint64，默认0.
func GetStringUint64(str string) uint64 {
	return GetStringDefaultUint64(str, 0)
}

// GetStringDefaultUint64 使用strconv.ParseUint解析。
func GetStringDefaultUint64(str string, n uint64) uint64 {
	if v, err := strconv.ParseUint(str, 10, 64); err == nil {
		return v
	}
	return n
}

// GetStringFloat32 使用GetStringDefaultFloat32，默认0。
func GetStringFloat32(str string) float32 {
	return GetStringDefaultFloat32(str, 0)
}

// GetStringDefaultFloat32 使用strconv.ParseFloat解析。
func GetStringDefaultFloat32(str string, n float32) float32 {
	if v, err := strconv.ParseFloat(str, 32); err == nil {
		return float32(v)
	}
	return n
}

// GetStringFloat64 使用GetStringDefaultFloat64，默认0。
func GetStringFloat64(str string) float64 {
	return GetStringDefaultFloat64(str, 0)
}

// GetStringDefaultFloat64 使用strconv.ParseFloat解析。
func GetStringDefaultFloat64(str string, n float64) float64 {
	if v, err := strconv.ParseFloat(str, 64); err == nil {
		return v
	}
	return n
}

// GetStringDefault 如果s1为空，返回s2
func GetStringDefault(s1, s2 string) string {
	if len(s1) == 0 {
		return s2
	}
	return s1
}

// GetStringsDefault 函数返回第一个非空的字符串。
func GetStringsDefault(strs ...string) string {
	for _, i := range strs {
		if i != "" {
			return i
		}
	}
	return ""
}

// GetWarp 对象封装Get函数提供类型转换功能。
type GetWarp func(string) interface{}

// NewGetWarp 函数创建一个getwarp处理类型转换。
func NewGetWarp(fn func(string) interface{}) GetWarp {
	return fn
}

// NewGetWarpWithConfig 函数使用Config.Get创建getwarp。
func NewGetWarpWithConfig(c Config) GetWarp {
	return c.Get
}

// NewGetWarpWithApp 函数使用App创建getwarp。
func NewGetWarpWithApp(app *App) GetWarp {
	return func(key string) interface{} {
		return app.Get(key)
	}
}

// NewGetWarpWithObject 函数使用map或创建getwarp。
func NewGetWarpWithObject(obj interface{}) GetWarp {
	return func(key string) interface{} {
		return Get(obj, key)
	}
}

// GetInterface 方法获取interface类型的配置值。
func (fn GetWarp) GetInterface(key string) interface{} {
	return fn(key)
}

// GetBool 方法获取bool类型的配置值。
func (fn GetWarp) GetBool(key string, vals ...bool) bool {
	return GetBool(fn(key))
}

// GetInt 方法获取int类型的配置值。
func (fn GetWarp) GetInt(key string, vals ...int) int {
	num := GetInt(fn(key))
	if num != 0 {
		return num
	}
	for _, i := range vals {
		if i != 0 {
			return i
		}
	}
	return 0
}

// GetUint 方法取获取uint类型的配置值。
func (fn GetWarp) GetUint(key string, vals ...uint) uint {
	num := GetUint(fn(key))
	if num != 0 {
		return num
	}
	for _, i := range vals {
		if i != 0 {
			return i
		}
	}
	return 0
}

// GetInt64 方法int64类型的配置值。
func (fn GetWarp) GetInt64(key string, vals ...int64) int64 {
	num := GetInt64(fn(key))
	if num != 0 {
		return num
	}
	for _, i := range vals {
		if i != 0 {
			return i
		}
	}
	return 0
}

// GetUint64 方法取获取uint64类型的配置值。
func (fn GetWarp) GetUint64(key string, vals ...uint64) uint64 {
	num := GetUint64(fn(key))
	if num != 0 {
		return num
	}
	for _, i := range vals {
		if i != 0 {
			return i
		}
	}
	return 0
}

// GetFloat32 方法取获取float32类型的配置值。
func (fn GetWarp) GetFloat32(key string, vals ...float32) float32 {
	num := GetFloat32(fn(key))
	if num != 0 {
		return num
	}
	for _, i := range vals {
		if i != 0 {
			return i
		}
	}
	return 0.0
}

// GetFloat64 方法取获取float64类型的配置值。
func (fn GetWarp) GetFloat64(key string, vals ...float64) float64 {
	num := GetFloat64(fn(key))
	if num != 0 {
		return num
	}
	for _, i := range vals {
		if i != 0 {
			return i
		}
	}
	return 0.0
}

// GetString 方法获取一个字符串，如果字符串为空返回其他默认非空字符串，
func (fn GetWarp) GetString(key string, vals ...string) string {
	str := GetString(fn(key))
	if str != "" {
		return str
	}
	for _, i := range vals {
		if i != "" {
			return i
		}
	}
	return ""
}

// GetBytes 方法获取[]byte类型的配置值，如果是字符串类型会转换成[]byte。
func (fn GetWarp) GetBytes(key string) []byte {
	val := fn(key)
	body, ok := val.([]byte)
	if ok {
		return body
	}

	str := GetString(val)
	if str != "" {
		return []byte(str)
	}

	return nil
}

// GetStrings 方法获取[]string值
func (fn GetWarp) GetStrings(key string) []string {
	return GetArrayString(fn(key))
}

// muliterror 实现多个error组合。
type muliterror struct {
	errs []error
}

// HandleError 实现处理多个错误，如果非空则保存错误。
func (err *muliterror) HandleError(errs ...error) {
	for _, e := range errs {
		if e != nil {
			err.errs = append(err.errs, e)
		}
	}
}

// Error 方法实现error接口，返回错误描述。
func (err *muliterror) Error() string {
	return fmt.Sprint(err.errs)
}

// GetError 方法返回错误，如果没有保存的错误则返回空。
func (err *muliterror) GetError() error {
	switch len(err.errs) {
	case 0:
		return nil
	case 1:
		return err.errs[0]
	default:
		return err
	}
}
