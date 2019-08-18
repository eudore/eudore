package eudore

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

func arrayclean(names []string) (n []string) {
	for _, name := range names {
		if name != "" {
			n = append(n, name)
		}
	}
	return
}

// Each string strs handle element, if return is null, then delete this a elem.
func eachstring(strs []string, fn func(string) string) (s []string) {
	for _, i := range strs {
		i = fn(i)
		if i != "" {
			s = append(s, i)
		}
	}
	return
}

// Use sep to split str into two strings.
func split2byte(str string, b byte) (string, string) {
	pos := strings.IndexByte(str, b)
	if pos == -1 {
		return "", ""
	}
	return str[:pos], str[pos+1:]
}

// Env to Arg
func env2arg(str string) string {
	k, v := split2byte(str, '=')
	k = strings.ToLower(strings.Replace(k, "_", ".", -1))[4:]
	return fmt.Sprintf("--%s=%s", k, v)
}

// MatchStar 模式匹配对象，允许使用带'*'的模式。
func MatchStar(obj, patten string) bool {
	ps := strings.Split(patten, "*")
	if len(ps) < 2 {
		return patten == obj
	}
	if !strings.HasPrefix(obj, ps[0]) {
		return false
	}
	for _, i := range ps {
		if i == "" {
			continue
		}
		pos := strings.Index(obj, i)
		if pos == -1 {
			return false
		}
		obj = obj[pos+len(i):]
	}
	return true
}

// Json test function, json formatted output args.
//
// Json 测试函数，json格式化输出args。
func Json(args ...interface{}) {
	indent, err := json.MarshalIndent(&args, "", "\t")
	fmt.Println(string(indent), err)
}

// GetBool 使用GetDefaultBool，默认false。
func GetBool(i interface{}) bool {
	return GetDefaultBool(i, false)
}

// GetDefaultBool 函数先检查断言，再使用GetDefaultString转换字符串，使用strconv.ParseBool解析。
func GetDefaultBool(i interface{}, b bool) bool {
	if v, ok := i.(bool); ok {
		return v
	}
	if v, err := strconv.ParseBool(GetDefaultString(i, "")); err == nil {
		return v
	}
	return b
}

// GetInt 使用GetDefaultInt，默认0.
func GetInt(i interface{}) int {
	return GetDefaultInt(i, 0)
}

// GetDefaultInt 函数先检查断言，再使用GetDefaultString转换字符串，使用strconv.Atoi解析。
func GetDefaultInt(i interface{}, n int) int {
	if v, ok := i.(int); ok {
		return v
	}
	if v, err := strconv.Atoi(GetDefaultString(i, "")); err == nil {
		return v
	}
	return n
}

// GetInt64 使用GetDefaultInt64，默认0.
func GetInt64(i interface{}) int64 {
	return GetDefaultInt64(i, 0)
}

// GetDefaultInt64 函数先检查断言，再使用GetDefaultString转换字符串，使用strconv.ParseInt解析。
func GetDefaultInt64(i interface{}, n int64) int64 {
	if v, ok := i.(int64); ok {
		return v
	}
	if v, err := strconv.ParseInt(GetDefaultString(i, ""), 10, 64); err == nil {
		return v
	}
	return n
}

// GetUint 使用GetDefaultUint，默认0。
func GetUint(i interface{}) uint {
	return GetDefaultUint(i, 0)
}

// GetDefaultUint 函数先检查断言，再使用GetDefaultString转换字符串，使用strconv.ParseUint解析。
func GetDefaultUint(i interface{}, n uint) uint {
	if v, ok := i.(uint); ok {
		return v
	}
	if v, err := strconv.ParseUint(GetDefaultString(i, ""), 10, 64); err == nil {
		return uint(v)
	}
	return n
}

// GetUint64 使用GetDefaultUint64，默认0.
func GetUint64(i interface{}) uint64 {
	return GetDefaultUint64(i, 0)
}

// GetDefaultUint64 函数先检查断言，再使用GetDefaultString转换字符串，使用strconv.ParseInt解析。
func GetDefaultUint64(i interface{}, n uint64) uint64 {
	if v, ok := i.(uint64); ok {
		return v
	}
	if v, err := strconv.ParseUint(GetDefaultString(i, ""), 10, 64); err == nil {
		return v
	}
	return n
}

// GetFloat32 使用GetDefaultFloat32，默认0.
func GetFloat32(i interface{}) float32 {
	return GetDefaultFloat32(i, 0)
}

// GetDefaultFloat32 函数先检查断言，再使用GetDefaultString转换字符串，使用strconv.ParseFloat解析。
func GetDefaultFloat32(i interface{}, n float32) float32 {
	if v, ok := i.(float32); ok {
		return v
	}
	if v, err := strconv.ParseFloat(GetDefaultString(i, ""), 32); err == nil {
		return float32(v)
	}
	return n
}

// GetFloat64 使用GetDefaultFloat64，默认0.
func GetFloat64(i interface{}) float64 {
	return GetDefaultFloat64(i, 0)
}

// GetDefaultFloat64 函数先检查断言，再使用GetDefaultString转换字符串，使用strconv.ParseFloat解析。
func GetDefaultFloat64(i interface{}, n float64) float64 {
	if v, ok := i.(float64); ok {
		return v
	}
	if v, err := strconv.ParseFloat(GetDefaultString(i, ""), 64); err == nil {
		return v
	}
	return n
}

// GetString 使用GetDefaultString，默认空字符。
func GetString(i interface{}) string {
	return GetDefaultString(i, "")
}

// GetDefaultString 通过断言string类型实现转换。
func GetDefaultString(i interface{}, str string) string {
	if i == nil {
		return str
	}
	if v, ok := i.(string); ok && v != "" {
		return v
	}
	return str
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

// StringMap 定义map[string]interface{}对象的操作。
type StringMap map[string]interface{}

// NewStringMap 创建一个StringMap对象，如果参数不是map[string]interface{}，则返回空。
func NewStringMap(i interface{}) StringMap {
	v, ok := i.(map[string]interface{})
	if ok {
		return StringMap(v)
	}
	return nil
}

// Get 方法实现获得一个键值。
func (m StringMap) Get(key string) interface{} {
	return m[key]
}

// Set 方法实现设置一个值。
func (m StringMap) Set(key string, val interface{}) {
	m[key] = val
}

// Del 方法实现删除一个键值。
func (m StringMap) Del(key string) {
	delete(m, key)
}

// GetInt 方法获取对应的值并转换成int。
func (m StringMap) GetInt(key string) int {
	return GetInt(m.Get(key))
}

// GetDefultInt 方法获取对应的值并转换成int,如果无法转换返回默认值。
func (m StringMap) GetDefultInt(key string, n int) int {
	return GetDefaultInt(m.Get(key), n)
}

// GetString 方法获取对应的值并转换成string
func (m StringMap) GetString(key string) string {
	return GetString(m.Get(key))
}

// GetDefaultString 方法获取对应的值并转换成string,如果无法转换返回默认值。
func (m StringMap) GetDefaultString(key string, str string) string {
	return GetDefaultString(m.Get(key), str)
}
