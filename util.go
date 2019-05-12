package eudore

import (
	"fmt"
	"strings"
	"strconv"
	"math/rand"
	"encoding/json"
)

func arrayclean(names []string) (n []string){
	for _, name := range names {
		if name != "" {
			n = append(n, name)
		}
	}
	return
}

// Each string strs handle element, if return is null, then delete this a elem.
func eachstring(strs []string, fn func(string) string) (s []string){
	for _, i := range strs {
		i = fn(i)
		if i != "" {
			s = append(s, i)
		}
	}
	return
}

// Use sep to split str into two strings.
func split2byte(str string,b byte) (string, string) {
	pos := strings.IndexByte(str, b)
	if pos == -1 {
		return "", ""
	}
	return str[:pos], str[pos + 1:]
}

/*
func split2(str string,sep string) (string, string) {
	pos := strings.IndexByte(str, sep[0])
	if pos == -1 {
		return "", ""
	}
	return str[:pos], str[pos + len(sep):]
}
*/

// Env to Arg
func env2arg(str string) string {
	k, v := split2byte(str, '=')
	k = strings.ToLower(strings.Replace(k, "_", ".", -1))[4:]
	return fmt.Sprintf("--%s=%s", k, v)
}

func GetRandomString() string {
	const letters = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXY"
	result := make([]byte, 16)
	for i := range result {
		result[i] = letters[rand.Intn(61)]
	}
	return string(result)
}

// 模式匹配对象，运行试验带'*'的模式
func MatchStar(obj, patten string) bool {
	ps := strings.Split(patten,"*")
	if len(ps) == 0 {
		return patten == obj
	}
	if !strings.HasPrefix(obj, ps[0]) {
		return false
	}
	for _,i := range ps {
		if i == "" {
			continue
		}
		pos := strings.Index(obj, i)
		if pos == -1 {
			return false
		}
		obj = obj[pos + len(i):]
	}
	return true
}

// test function, json formatted output args.
//
// 测试函数，json格式化输出args。
func Json(args ...interface{}) {
	indent, err := json.MarshalIndent(&args, "", "\t")
	fmt.Println(string(indent), err)
}



func GetBool(i interface{}) bool {
	return GetDefaultBool(i, false)
}

func GetDefaultBool(i interface{}, b bool) bool {
	if v, ok := i.(bool); ok {
		return v
	}
	if v, err := strconv.ParseBool(GetDefaultString(i, "")); err == nil {
		return v
	}
	return b
}

func GetInt(i interface{}) int {
	return GetDefaultInt(i, 0)
}

func GetDefaultInt(i interface{},n int) int {
	if v, ok := i.(int); ok {
		return v
	}	
	if v, err := strconv.Atoi(GetDefaultString(i, "")); err == nil {
		return v
	}
	return n
}

func GetInt64(i interface{}) int64 {
	return GetDefaultInt64(i, 0)
}

func GetDefaultInt64(i interface{},n int64) int64 {
	if v, ok := i.(int64); ok {
		return v
	}	
	if v, err := strconv.ParseInt(GetDefaultString(i, ""), 10, 64); err == nil {
		return v
	}
	return n
}

func GetUint(i interface{}) uint {
	return GetDefaultUint(i, 0)
}

func GetDefaultUint(i interface{},n uint) uint {
	if v, ok := i.(uint); ok {
		return v
	}	
	if v, err := strconv.ParseUint(GetDefaultString(i, ""), 10, 64); err == nil {
		return uint(v)
	}
	return n
}

func GetUint64(i interface{}) uint64 {
	return GetDefaultUint64(i, 0)
}

func GetDefaultUint64(i interface{},n uint64) uint64 {
	if v, ok := i.(uint64); ok {
		return v
	}	
	if v, err := strconv.ParseUint(GetDefaultString(i, ""), 10, 64); err == nil {
		return v
	}
	return n
}

func GetFloat32(i interface{}) float32 {
	return GetDefaultFloat32(i, 0)
}

func GetDefaultFloat32(i interface{},n float32) float32 {
	if v, ok := i.(float32); ok {
		return v
	}	
	if v, err := strconv.ParseFloat(GetDefaultString(i, ""), 32); err == nil {
		return float32(v)
	}
	return n
}

func GetFloat64(i interface{}) float64 {
	return GetDefaultFloat64(i, 0)
}

func GetDefaultFloat64(i interface{},n float64) float64 {
	if v, ok := i.(float64); ok {
		return v
	}	
	if v, err := strconv.ParseFloat(GetDefaultString(i, ""), 64); err == nil {
		return v
	}
	return n
}

func GetString(i interface{}) string {
	return GetDefaultString(i, "")
}

func GetDefaultString(i interface{}, str string) string {
	if i == nil {
		return str
	}
	if v, ok := i.(string); ok && v != "" {
		return v
	}
	return str
}



func GetStringBool(str string) bool {
	return GetStringDefaultBool(str, false)
}

func GetStringDefaultBool(str string, b bool) bool {
	if v, err := strconv.ParseBool(str); err == nil {
		return v
	}
	return b
}

func GetStringInt(str string) int {
	return GetStringDefaultInt(str, 0)
}

func GetStringDefaultInt(str string,n int) int {
	if v, err := strconv.Atoi(str); err == nil {
		return v
	}
	return n
}

func GetStringInt64(str string) int64 {
	return GetStringDefaultInt64(str, 0)
}

func GetStringDefaultInt64(str string,n int64) int64 {
	if v, err := strconv.ParseInt(str, 10, 64); err == nil {
		return v
	}
	return n
}

func GetStringUint(str string) uint {
	return GetStringDefaultUint(str, 0)
}

func GetStringDefaultUint(str string, n uint) uint {
	if v, err := strconv.ParseUint(str, 10, 64); err == nil {
		return uint(v)
	}
	return n
}

func GetStringUint64(str string) uint64 {
	return GetStringDefaultUint64(str, 0)
}

func GetStringDefaultUint64(str string, n uint64) uint64 {
	if v, err := strconv.ParseUint(str, 10, 64); err == nil {
		return v
	}
	return n
}

func GetStringFloat32(str string) float32 {
	return GetStringDefaultFloat32(str, 0)
}

func GetStringDefaultFloat32(str string,n float32) float32 {
	if v, err := strconv.ParseFloat(str, 32); err == nil {
		return float32(v)
	}
	return n
}

func GetStringFloat64(str string) float64 {
	return GetStringDefaultFloat64(str, 0)
}

func GetStringDefaultFloat64(str string,n float64) float64 {
	if v, err := strconv.ParseFloat(str, 64); err == nil {
		return v
	}
	return n
}

func GetStringDefault(s1 , s2 string) string {
	if len(s1) == 0 {
		return s2
	}
	return s1
}


type (
	StringMap  map[string]interface{}
)

func NewStringMap(i interface{}) StringMap {
	v, ok := i.(map[string]interface{})
	if ok {
		return StringMap(v)
	}
	return nil
}

func (m StringMap) Get(key string) interface{} {
	if len(key) == 0 {
		return m
	}
	v, ok := m[key]
	if ok {
		return v
	}
	return nil
}

func (m StringMap) Set(key string, val interface{}) {
	if len(key) == 0 {
		v, ok := val.(map[string]interface{})
		if ok {
			m = v
		}
	}else {
		m[key] = val
	}
}

func (m StringMap) Del(key string) {
	delete(m, key)
}

func (m StringMap) GetInt(key string) int {
	return GetInt(m.Get(key))
}

func (m StringMap) GetDefultInt(key string,n int) int {
	return GetDefaultInt(m.Get(key), n)
}

func (m StringMap) GetString(key string) string {
	return GetString(m.Get(key))
}

func (m StringMap) GetDefaultString(key string,str string) string {
	return GetDefaultString(m.Get(key), str)
}
