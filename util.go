package eudore

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type contextKey struct {
	name string
}

// NewContextKey 定义context key。
func NewContextKey(key string) interface{} {
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
		k, v := split2byte(str, '=')
		if v != "" {
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
	p[1] = p[1] + params[1]
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

// Del 方法删除一个参数值
func (p Params) Del(key string) {
	for i := 0; i < len(p); i += 2 {
		if p[i] == key {
			p[i+1] = ""
		}
	}
}

// TimeDuration 定义time.Duration类型处理json
type TimeDuration time.Duration

// String 方法格式化输出时间。
func (d TimeDuration) String() string {
	return time.Duration(d).String()
}

// MarshalJSON 方法实现json序列化输出。
func (d TimeDuration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// UnmarshalJSON 方法实现解析json格式时间。
func (d *TimeDuration) UnmarshalJSON(b []byte) error {
	str := string(b)
	if str != "" && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}
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
	return fmt.Errorf("invalid duration type %T, value: '%s'", b, b)
}

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
	iValue := reflect.ValueOf(i)
	switch iValue.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		return iValue.Int() != 0
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		return iValue.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return iValue.Float() != 0
	case reflect.String:
		str := iValue.String()
		return str != "" && str != "true" && str != "1"
	default:
		return false
	}
}

// GetInt 函数转换一个bool、int、uint、float、string类型成int,或者返回第一个非零值。
func GetInt(i interface{}, nums ...int) int {
	var number int
	iValue := reflect.ValueOf(i)
	switch iValue.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		number = int(iValue.Int())
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		number = int(iValue.Uint())
	case reflect.Float32, reflect.Float64:
		number = int(iValue.Float())
	case reflect.String:
		if v, err := strconv.Atoi(iValue.String()); err == nil {
			number = v
		}
	}
	if number != 0 {
		return number
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
	var number int64
	iValue := reflect.ValueOf(i)
	switch iValue.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		number = iValue.Int()
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		number = int64(iValue.Uint())
	case reflect.Float32, reflect.Float64:
		number = int64(iValue.Float())
	case reflect.String:
		if v, err := strconv.ParseInt(iValue.String(), 10, 64); err == nil {
			number = v
		}
	}
	if number != 0 {
		return number
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
	var number uint
	iValue := reflect.ValueOf(i)
	switch iValue.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		number = uint(iValue.Int())
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		number = uint(iValue.Uint())
	case reflect.Float32, reflect.Float64:
		number = uint(iValue.Float())
	case reflect.String:
		if v, err := strconv.ParseUint(iValue.String(), 10, 64); err == nil {
			number = uint(v)
		}
	}
	if number != 0 {
		return number
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
	var number uint64
	iValue := reflect.ValueOf(i)
	switch iValue.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		number = uint64(iValue.Int())
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		number = iValue.Uint()
	case reflect.Float32, reflect.Float64:
		number = uint64(iValue.Float())
	case reflect.String:
		if v, err := strconv.ParseUint(iValue.String(), 10, 64); err == nil {
			number = v
		}
	}
	if number != 0 {
		return number
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
	var number float32
	iValue := reflect.ValueOf(i)
	switch iValue.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		number = float32(iValue.Int())
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		number = float32(iValue.Uint())
	case reflect.Float32, reflect.Float64:
		number = float32(iValue.Float())
	case reflect.String:
		if v, err := strconv.ParseFloat(iValue.String(), 32); err == nil {
			return float32(v)
		}
	}
	if number != 0 {
		return number
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
	var number float64
	iValue := reflect.ValueOf(i)
	switch iValue.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		number = float64(iValue.Int())
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		number = float64(iValue.Uint())
	case reflect.Float32, reflect.Float64:
		number = iValue.Float()
	case reflect.String:
		if v, err := strconv.ParseFloat(iValue.String(), 64); err == nil {
			return v
		}
	}
	if number != 0 {
		return number
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
	var str string
	iValue := reflect.ValueOf(i)
	switch iValue.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		str = strconv.FormatInt(iValue.Int(), 10)
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		str = strconv.FormatUint(iValue.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		str = strconv.FormatFloat(iValue.Float(), 'f', -1, 64)
	case reflect.String:
		str = iValue.String()
	case reflect.Bool:
		str = strconv.FormatBool(iValue.Bool())
	default:
		switch val := i.(type) {
		case fmt.Stringer:
			str = val.String()
		case []byte:
			str = string(val)
		}
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
	if v, err := strconv.Atoi(str); err == nil && v != 0 {
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
	if v, err := strconv.ParseInt(str, 10, 64); err == nil && v != 0 {
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
	if v, err := strconv.ParseUint(str, 10, 64); err == nil && v != 0 {
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
	if v, err := strconv.ParseUint(str, 10, 64); err == nil && v != 0 {
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
	if v, err := strconv.ParseFloat(str, 32); err == nil && v != 0 {
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
	if v, err := strconv.ParseFloat(str, 64); err == nil && v != 0 {
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

// errormulit 实现多个error组合。
type errormulit struct {
	errs []error
}

// HandleError 实现处理多个错误，如果非空则保存错误。
func (err *errormulit) HandleError(errs ...error) {
	for _, e := range errs {
		if e != nil {
			err.errs = append(err.errs, e)
		}
	}
}

// Error 方法实现error接口，返回错误描述。
func (err *errormulit) Error() string {
	return fmt.Sprint(err.errs)
}

// GetError 方法返回错误，如果没有保存的错误则返回空。
func (err *errormulit) Unwrap() error {
	switch len(err.errs) {
	case 0:
		return nil
	case 1:
		return err.errs[0]
	default:
		return err
	}
}

// NewErrorStatusCode 方法组合ErrorStatus和ErrorCode。
func NewErrorStatusCode(err error, status, code int) error {
	if code > 0 {
		err = errorCode{err, code}
	}
	if status > 0 {
		err = errorStatus{err, status}
	}
	return err
}

// NewErrorStatus 方法封装error实现Status方法。
func NewErrorStatus(err error, status int) error {
	if status > 0 {
		return errorStatus{err, status}
	}
	return err
}

type errorStatus struct {
	err    error
	status int
}

func (err errorStatus) Error() string {
	return err.err.Error()
}

func (err errorStatus) Unwrap() error {
	return err.err
}

func (err errorStatus) Status() int {
	return err.status
}

// NewErrorCode 方法封装error实现Code方法。
func NewErrorCode(err error, code int) error {
	if code > 0 {
		return errorCode{err, code}
	}
	return err
}

type errorCode struct {
	err  error
	code int
}

func (err errorCode) Error() string {
	return err.err.Error()
}
func (err errorCode) Unwrap() error {
	return err.err
}

func (err errorCode) Code() int {
	return err.code
}
