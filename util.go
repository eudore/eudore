package eudore

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type contextKey struct {
	name string
}

// NewContextKey defines a custom [context.Context] key.
func NewContextKey(key string) any {
	return contextKey{key}
}

// String method output [contextKey].
func (key contextKey) String() string {
	return key.name
}

// Unmounter converts fn to [interface].
type Unmounter func(ctx context.Context)

// Unmount method implements the [anyUnmount] interface.
func (fn Unmounter) Unmount(ctx context.Context) {
	fn(ctx)
}

// Params defines [Context] and [Router] to save key-value data.
type Params []string

// NewParamsRoute method creates [Params] based on a route and supports
// routing path block mode.
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

// getRoutePath function intercepts the [ParamRoute] in the path and
// supports '{}' for block matching.
func getRoutePath(path string) string {
	var isblock bool
	var last rune
	for i, b := range path {
		if isblock {
			if b == '}' && last != '\\' {
				isblock = false
			}
			last = b
			continue
		}

		switch b {
		case '{':
			isblock = true
		case ' ':
			return path[:i]
		}
	}
	return path
}

// Clone method deeply copies a Param object.
func (p Params) Clone() Params {
	params := make(Params, len(p))
	copy(params, p)
	return params
}

// String method outputs Params as a string.
func (p Params) String() string {
	size := 0
	for i := 0; i < len(p); i += 2 {
		if (p[i] != "" && p[i+1] != "") || i == 0 {
			size += len(p[i]) + len(p[i+1]) + 2
		}
	}

	buf := make([]byte, 0, size)
	for i := 0; i < len(p); i += 2 {
		if (p[i] != "" && p[i+1] != "") || i == 0 {
			buf = append(buf, ' ')
			buf = append(buf, p[i]...)
			buf = append(buf, '=')
			buf = append(buf, p[i+1]...)
		}
	}
	return unsafe.String(unsafe.SliceData(buf[1:]), len(buf)-1)
}

// Get method returns the first value of the specified key.
func (p Params) Get(key string) string {
	for i := 0; i < len(p); i += 2 {
		if p[i] == key {
			return p[i+1]
		}
	}
	return ""
}

// Add method adds a parameter.
func (p Params) Add(vals ...string) Params {
	return append(p, vals...)
}

// Set method sets the first value of the specified key or appends it.
func (p Params) Set(key, val string) Params {
	for i := 0; i < len(p); i += 2 {
		if p[i] == key {
			p[i+1] = val
			return p
		}
	}
	return append(p, key, val)
}

// Del method clears the first value of the specified key.
func (p Params) Del(key string) {
	for i := 0; i < len(p); i += 2 {
		if p[i] == key {
			p[i+1] = ""
		}
	}
}

// GetWrap object Wrap func(string) any provides type conversion function.
type GetWrap func(string) any

// NewGetWrapWithConfig function creates [GetWrap] using [Config].Get.
func NewGetWrapWithConfig(c Config) GetWrap {
	return c.Get
}

// NewGetWrapWithApp function creates [GetWrap] using [App].
func NewGetWrapWithApp(app *App) GetWrap {
	return func(key string) any {
		return app.Get(key)
	}
}

// NewGetWrapWithMapString function creates [GetWrap] using map[string]any.
func NewGetWrapWithMapString(data map[string]any) GetWrap {
	return func(key string) any {
		return data[key]
	}
}

// NewGetWrapWithObject function uses object to create [GetWrap] and uses
// [GetAnyByPath] to get value.
func NewGetWrapWithObject(obj any) GetWrap {
	return func(key string) any {
		v, _ := GetAnyByPath(obj, key, nil)
		return v
	}
}

// GetAny method returns the any type.
func (fn GetWrap) GetAny(key string) any {
	return fn(key)
}

// GetBool method returns the bool type.
func (fn GetWrap) GetBool(key string) bool {
	return GetAny[bool](fn(key))
}

// GetInt method returns the int type.
func (fn GetWrap) GetInt(key string, vals ...int) int {
	return GetAny(fn(key), vals...)
}

// GetUint method returns the uint type.
func (fn GetWrap) GetUint(key string, vals ...uint) uint {
	return GetAny(fn(key), vals...)
}

// GetInt64 method returns the int64 type.
func (fn GetWrap) GetInt64(key string, vals ...int64) int64 {
	return GetAny(fn(key), vals...)
}

// GetUint64 method returns the uint64 type.
func (fn GetWrap) GetUint64(key string, vals ...uint64) uint64 {
	return GetAny(fn(key), vals...)
}

// GetFloat32 method returns the float32 type.
func (fn GetWrap) GetFloat32(key string, vals ...float32) float32 {
	return GetAny(fn(key), vals...)
}

// GetFloat64 method returns the float64 type.
func (fn GetWrap) GetFloat64(key string, vals ...float64) float64 {
	return GetAny(fn(key), vals...)
}

// GetString method returns the string type.
// If the string is empty, it returns another non-empty string.
func (fn GetWrap) GetString(key string, vals ...string) string {
	return GetStringByAny(fn(key), vals...)
}

// TimeDuration defines [time.Duration] and implements [json.Marshaler] and
// [json.UnmarshalJSON].
type TimeDuration time.Duration

// String method formats the output time.
func (d TimeDuration) String() string {
	return time.Duration(d).String()
}

// MarshalText method implements [encoding.MarshalText] and
// [json.Marshaler].
func (d TimeDuration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

// UnmarshalJSON method implements [json.UnmarshalJSON].
func (d *TimeDuration) UnmarshalJSON(b []byte) error {
	if len(b) > 0 && b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}
	return d.UnmarshalText(b)
}

// UnmarshalText method implements [encoding.UnmarshalText] and parse [time.Duration].
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

type radixData[T any] interface {
	*T
	Insert(vals ...any) error
}

type radixNode[P radixData[T], T any] struct {
	path  string
	data  P
	child []*radixNode[P, T]
}

func (node *radixNode[P, T]) insert(path string, data ...any) error {
	next := node.insertPath(path)
	if next.data == nil {
		next.data = new(T)
	}
	return next.data.Insert(data...)
}

func (node *radixNode[P, T]) insertPath(path string) *radixNode[P, T] {
	if path == "" {
		return node
	}
	for i := range node.child {
		prefix, find := getSubsetPrefix(path, node.child[i].path)
		if find {
			if prefix != node.child[i].path {
				node.child[i].path = node.child[i].path[len(prefix):]
				node.child[i] = &radixNode[P, T]{
					path:  prefix,
					child: []*radixNode[P, T]{node.child[i]},
				}
			}
			return node.child[i].insertPath(path[len(prefix):])
		}
	}

	next := &radixNode[P, T]{path: path}
	node.child = append(node.child, next)
	return next
}

func (node *radixNode[P, T]) lookPath(path string) []P {
	for _, child := range node.child {
		if strings.HasPrefix(path, child.path) {
			next := child.lookPath(path[len(child.path):])
			if node.data != nil {
				next = append(next, node.data)
			}
			return next
		}
	}
	if node.data != nil {
		return []P{node.data}
	}
	return nil
}

// mulitError implements multiple error combinations.
type mulitError struct {
	errs []error
}

// Handle implementation handles multiple errors and saves the errors if
// non-empty.
func (err *mulitError) Handle(errs ...error) {
	for _, e := range errs {
		if e != nil {
			err.errs = append(err.errs, e)
		}
	}
}

// Error method implements the error interface and returns an error.
func (err *mulitError) Error() string {
	errs := make([]string, len(err.errs))
	for i := range err.errs {
		errs[i] = err.errs[i].Error()
	}
	return strings.Join(errs, ", ")
}

// Unwrap method returns the error, or null if there is no saved error.
func (err *mulitError) Unwrap() []error {
	return err.errs
}

// NewErrorWithStatusCode method combines [NewErrorWithStatus] and
// [NewErrorWithCode].
func NewErrorWithStatusCode(err error, status, code int) error {
	if err == nil {
		return nil
	}
	if code > 0 {
		err = codeError{err, code}
	}
	if status > 0 {
		err = statusError{err, status}
	}
	return err
}

// NewErrorWithStatus function returns the wrap error implementation
// Status method.
func NewErrorWithStatus(err error, status int) error {
	if err == nil {
		return nil
	}
	if status > 0 {
		return statusError{err, status}
	}
	return err
}

type statusError struct {
	err    error
	status int
}

func (e statusError) Error() string {
	return e.err.Error()
}

func (e statusError) Unwrap() error {
	return e.err
}

func (e statusError) Status() int {
	return e.status
}

// NewErrorWithCode function returns the wrap error implementation Code method.
func NewErrorWithCode(err error, code int) error {
	if err == nil {
		return nil
	}
	if code > 0 {
		return codeError{err, code}
	}
	return err
}

type codeError struct {
	err  error
	code int
}

func (e codeError) Error() string {
	return e.err.Error()
}

func (e codeError) Unwrap() error {
	return e.err
}

func (e codeError) Code() int {
	return e.code
}

type warppedError struct {
	err error
	msg string
}

// NewErrorWithWrapped returns a new error that wraps the given 'err',
// using 'msg' as the primary output when Error() is called.
func NewErrorWithWrapped(err error, msg string) error {
	return &warppedError{err, msg}
}

func (e warppedError) Error() string {
	return e.msg
}

func (e warppedError) Unwrap() error {
	return e.err
}

type stackError struct {
	err   error
	stack []string
}

// NewErrorWithStack returns a new error that wraps the given 'err' and
// attaches a predefined stack trace.
func NewErrorWithStack(err error, stack []string) error {
	if err == nil {
		return nil
	}
	return &stackError{err, stack}
}

// NewErrorWithDepth returns a new error that wraps the given 'err' and
// captures the call stack starting from a specified depth.
func NewErrorWithDepth(err error, depth int) error {
	if err == nil {
		return nil
	}
	return &stackError{err, GetCallerStacks(depth)}
}

func (e stackError) Error() string {
	return e.err.Error()
}

func (e stackError) Unwrap() error {
	return e.err
}

func (e stackError) Stack() []string {
	return e.stack
}

func cutOmit(s string) (string, bool) {
	if strings.HasSuffix(s, ",omitempty") {
		return s[:len(s)-10], true
	}
	return s, false
}

func sliceClearAppend[T any](dst []T, src ...T) []T {
	if src == nil {
		return dst
	}
	dst = append(dst, src...)
	l := len(dst)
	return dst[:l:l]
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

func mapClone[K comparable, V any](d map[K]V) map[K]V {
	n := make(map[K]V, len(d))
	for k, v := range d {
		n[k] = v
	}
	return n
}

// GetAnyDefault function returns a non-NULL value.
func GetAnyDefault[T comparable](arg1, arg2 T) T {
	var zero T
	if arg1 != zero {
		return arg1
	}
	return arg2
}

// GetAnyDefaults function returns the first non-null value.
func GetAnyDefaults[T comparable](args ...T) T {
	var zero T
	for i := range args {
		if args[i] != zero {
			return args[i]
		}
	}
	return zero
}

// typeNumber defines a typeParam numeric type set.
type typeNumber interface {
	int | int8 | int16 | int32 | int64 |
		uint | uint8 | uint16 | uint32 | uint64 |
		float32 | float64 | complex64 | complex128
}

// GetAny function converts the Value any type into T type.
func GetAny[T string | bool | typeNumber](s any, defaults ...T) T {
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
		case tType.Kind() == reflect.Bool:
			t = any(getBoolByAny(s)).(T)
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

func getBoolByAny(i any) bool {
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32,
		reflect.Uint64, reflect.Uintptr:
		return v.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return v.Float() != 0
	case reflect.Ptr, reflect.Interface:
		if !v.IsNil() {
			return getBoolByAny(v.Elem().Interface())
		}
	case reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return v.Len() != 0
	}
	return false
}

// GetStringByAny function converts any into string.
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
		if i != nil {
			str = fmt.Sprint(i)
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

// GetStringRandom function returns a random string of the specified length.
func GetStringRandom(length int) string {
	buf := make([]byte, length)
	_, _ = io.ReadFull(rand.Reader, buf)
	return hex.EncodeToString(buf)
}

// GetStringDuration function get [time.Duration] friendly visualization data.
func GetStringDuration(n time.Duration) fmt.Stringer {
	var result durationString
	size := 7
	if n > 10000 {
		size = 4
		n /= 1000
	}
	for i := 0; n != 0 || i < size; i++ {
		if i == size-1 {
			result = append(result, '.')
		}
		result = append(result, byte('0'+n%10))
		n /= 10
	}
	left, right := 0, len(result)-1
	for left < right {
		result[left], result[right] = result[right], result[left]
		left++
		right--
	}
	return result
}

type durationString []byte

func (d durationString) String() string {
	return string(d)
}

func (d durationString) MarshalJSON() ([]byte, error) {
	return []byte(d), nil
}

// GetAnyByString function converts a string to T value.
//
// Refer: [GetAnyByStringWithError].
func GetAnyByString[T string | bool | time.Time | time.Duration |
	typeNumber](str string, defaults ...T) T {
	val, _ := GetAnyByStringWithError(str, defaults...)
	return val
}

// GetAnyByStringWithError function converts a string to T value.
//
// The [time.Time] type attempts to be parsed using the
// [DefaultValueParseTimeFormats] format, with the time zone
// being [DefaultValueTimeLocation].
//
//nolint:cyclop,funlen,gocyclo
func GetAnyByStringWithError[T string | bool | time.Time | time.Duration |
	typeNumber](str string, defaults ...T) (T, error) {
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
			v, err = time.ParseInLocation(f, str, DefaultValueTimeLocation)
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
