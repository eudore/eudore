package eudore

import (
	"fmt"
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

// GetBool 函数转换bool、int、uint、float、string成bool。
func GetBool(i interface{}) bool {
	if v, ok := i.(bool); ok {
		return v
	}

	i = getNumber(i)
	if i != nil {
		if v, err := strconv.ParseBool(fmt.Sprint(i)); err == nil {
			return v
		}
	}
	return false
}

// GetInt 函数转换一个bool、int、uint、float、string类型成int,或者返回第一个非零值。
func GetInt(i interface{}, nums ...int) int {
	i = getNumber(i)
	if i != nil {
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
	}
	for _, i := range nums {
		if i != 0 {
			return i
		}
	}
	return 0
}

// GetInt64 函数转换一个bool、int、uint、float、string类型成int64,或者返回第一个非零值。
func GetInt64(i interface{}, nums ...int64) int64 {
	i = getNumber(i)
	if i != nil {
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
	}
	for _, i := range nums {
		if i != 0 {
			return i
		}
	}
	return 0
}

// GetUint 函数转换一个bool、int、uint、float、string类型成uint,或者返回第一个非零值。
func GetUint(i interface{}, nums ...uint) uint {
	i = getNumber(i)
	if i != nil {
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
	}
	for _, i := range nums {
		if i != 0 {
			return i
		}
	}
	return 0
}

// GetUint64 函数转换一个bool、int、uint、float、string类型成uint64,或者返回第一个非零值。
func GetUint64(i interface{}, nums ...uint64) uint64 {
	i = getNumber(i)
	if i != nil {
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
	}
	for _, i := range nums {
		if i != 0 {
			return i
		}
	}
	return 0
}

// GetFloat32 函数转换一个bool、int、uint、float、string类型成float32,或者返回第一个非零值。
func GetFloat32(i interface{}, nums ...float32) float32 {
	i = getNumber(i)
	if i != nil {
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
	}
	for _, i := range nums {
		if i != 0 {
			return i
		}
	}
	return 0
}

// GetFloat64 函数转换一个bool、int、uint、float、string类型成float64,或者返回第一个非零值。
func GetFloat64(i interface{}, nums ...float64) float64 {
	i = getNumber(i)
	if i != nil {
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
	}
	for _, i := range nums {
		if i != 0 {
			return i
		}
	}
	return 0
}

// GetString 方法转换一个bool、int、uint、float、string成string类型,或者返回第一个非零值，如果参数类型是string必须是非空才会作为返回值。
func GetString(i interface{}, strs ...string) string {
	if ster, ok := i.(fmt.Stringer); ok {
		return ster.String()
	}
	if v, ok := i.(string); ok {
		if v != "" {
			return v
		}
	} else {
		if v, ok := i.(bool); ok {
			return fmt.Sprint(v)
		}
		i = getNumber(i)
		if i != nil {
			return fmt.Sprint(i)
		}
	}
	for _, i := range strs {
		if i != "" {
			return i
		}
	}
	return ""
}

func getNumber(i interface{}) interface{} {
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

// GetBytes 方法断言[]byte类型或使用GetString方法转换成string类型
func GetBytes(i interface{}) []byte {
	body, ok := i.([]byte)
	if ok {
		return body
	}

	str := GetString(i)
	if str != "" {
		return []byte(str)
	}

	return nil
}

// GetStrings 转换string、[]strng、[]interface{}成[]string。
func GetStrings(i interface{}) []string {
	if i == nil {
		return nil
	}
	switch val := i.(type) {
	case string:
		return []string{val}
	case []string:
		return val
	case []interface{}:
		strs := make([]string, len(val))
		for i := range val {
			strs[i] = GetString(val[i])
		}
		return strs
	}
	return nil
}

// GetStringBool 使用strconv.ParseBool解析。
func GetStringBool(str string) bool {
	if v, err := strconv.ParseBool(str); err == nil {
		return v
	}
	return false
}

// GetStringInt 使用strconv.Atoi解析返回数据,如果解析返回错误使用第一个非零值。
func GetStringInt(str string, nums ...int) int {
	if v, err := strconv.Atoi(str); err == nil {
		return v
	}
	for _, i := range nums {
		if i != 0 {
			return i
		}
	}
	return 0
}

// GetStringInt64 使用strconv.ParseInt解析返回数据,如果解析返回错误使用第一个非零值。
func GetStringInt64(str string, nums ...int64) int64 {
	if v, err := strconv.ParseInt(str, 10, 64); err == nil {
		return v
	}
	for _, i := range nums {
		if i != 0 {
			return i
		}
	}
	return 0
}

// GetStringUint 使用strconv.ParseUint解析返回数据,如果解析返回错误使用第一个非零值。
func GetStringUint(str string, nums ...uint) uint {
	if v, err := strconv.ParseUint(str, 10, 64); err == nil {
		return uint(v)
	}
	for _, i := range nums {
		if i != 0 {
			return i
		}
	}
	return 0
}

// GetStringUint64 使用strconv.ParseUint解析返回数据,如果解析返回错误使用第一个非零值。
func GetStringUint64(str string, nums ...uint64) uint64 {
	if v, err := strconv.ParseUint(str, 10, 64); err == nil {
		return v
	}
	for _, i := range nums {
		if i != 0 {
			return i
		}
	}
	return 0
}

// GetStringFloat32 使用strconv.ParseFloa解析数据,如果解析返回错误使用第一个第一个非零值。
func GetStringFloat32(str string, nums ...float32) float32 {
	if v, err := strconv.ParseFloat(str, 32); err == nil {
		return float32(v)
	}
	for _, i := range nums {
		if i != 0 {
			return i
		}
	}
	return 0
}

// GetStringFloat64 使用strconv.ParseFloa解析数据,如果解析返回错误使用第一个第一个非零值。
func GetStringFloat64(str string, nums ...float64) float64 {
	if v, err := strconv.ParseFloat(str, 64); err == nil {
		return v
	}
	for _, i := range nums {
		if i != 0 {
			return i
		}
	}
	return 0
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

// NewGetWarpWithMapString 函数使用map[string]interface{}创建getwarp。
func NewGetWarpWithMapString(data map[string]interface{}) GetWarp {
	return func(key string) interface{} {
		return data[key]
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
func (fn GetWarp) GetBool(key string) bool {
	return GetBool(fn(key))
}

// GetInt 方法获取int类型的配置值。
func (fn GetWarp) GetInt(key string, vals ...int) int {
	return GetInt(fn(key), vals...)
}

// GetUint 方法取获取uint类型的配置值。
func (fn GetWarp) GetUint(key string, vals ...uint) uint {
	return GetUint(fn(key), vals...)
}

// GetInt64 方法int64类型的配置值。
func (fn GetWarp) GetInt64(key string, vals ...int64) int64 {
	return GetInt64(fn(key), vals...)
}

// GetUint64 方法取获取uint64类型的配置值。
func (fn GetWarp) GetUint64(key string, vals ...uint64) uint64 {
	return GetUint64(fn(key), vals...)
}

// GetFloat32 方法取获取float32类型的配置值。
func (fn GetWarp) GetFloat32(key string, vals ...float32) float32 {
	return GetFloat32(fn(key), vals...)
}

// GetFloat64 方法取获取float64类型的配置值。
func (fn GetWarp) GetFloat64(key string, vals ...float64) float64 {
	return GetFloat64(fn(key), vals...)
}

// GetString 方法获取一个字符串，如果字符串为空返回其他默认非空字符串，
func (fn GetWarp) GetString(key string, vals ...string) string {
	return GetString(fn(key), vals...)
}

// GetBytes 方法获取[]byte类型的配置值，如果是字符串类型会转换成[]byte。
func (fn GetWarp) GetBytes(key string) []byte {
	return GetBytes(fn(key))
}

// GetStrings 方法获取[]string值
func (fn GetWarp) GetStrings(key string) []string {
	return GetStrings(fn(key))
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
