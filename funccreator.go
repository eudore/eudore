package eudore

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
)

const (
	FuncCreateInvalid FuncCreateKind = iota
	FuncCreateString
	FuncCreateInt
	FuncCreateUint
	FuncCreateFloat
	FuncCreateBool
	FuncCreateAny
	FuncCreateSetString
	FuncCreateSetInt
	FuncCreateSetUint
	FuncCreateSetFloat
	FuncCreateSetBool
	FuncCreateSetAny
)
const FuncCreateNumber = FuncCreateAny

// FuncCreateKind defines the types of functions that can be created by [FuncCreator].
type FuncCreateKind uint8

// FuncCreator defines a verification function constructor,
// which is used by Router, Validate, and Filter by default.
type FuncCreator interface {
	RegisterFunc(name string, funcs ...any) error
	CreateFunc(kind FuncCreateKind, name string) (any, error)
	List() []string
}

type funcTypeParams interface {
	string | int | uint | float64 | bool | any
}

type funcCreatorBase struct {
	String    typeCreator[func(string) bool]
	Int       typeCreator[func(int) bool]
	Uint      typeCreator[func(uint) bool]
	Float     typeCreator[func(float64) bool]
	Bool      typeCreator[func(bool) bool]
	Any       typeCreator[func(any) bool]
	SetString typeCreator[func(string) string]
	SetInt    typeCreator[func(int) int]
	SetUint   typeCreator[func(uint) uint]
	SetFloat  typeCreator[func(float64) float64]
	SetAny    typeCreator[func(any) any]
	Errors    map[string]string
}

// MetadataFuncCreator records all currently registered functions and errors.
type MetadataFuncCreator struct {
	Health bool     `json:"health" protobuf:"1,name=health" yaml:"health"`
	Name   string   `json:"name" protobuf:"2,name=name" yaml:"name"`
	Funcs  []string `json:"funcs" protobuf:"3,name=funcs" yaml:"funcs"`
	Exprs  []string `json:"exprs,omitempty" protobuf:"4,name=exprs" yaml:"exprs,omitempty"`
	Errors []string `json:"errors,omitempty" protobuf:"5,name=errors" yaml:"errors,omitempty"`
}

// FuncRunner defines and stores function type and address information.
type FuncRunner struct {
	Kind FuncCreateKind
	Func any
}

// NewFuncCreator function creates default [FuncCreator] and loads the
// default rules.
func NewFuncCreator() FuncCreator {
	fc := &funcCreatorBase{
		String:    newTypeCreator[func(string) bool](),
		Int:       newTypeCreator[func(int) bool](),
		Uint:      newTypeCreator[func(uint) bool](),
		Float:     newTypeCreator[func(float64) bool](),
		Bool:      newTypeCreator[func(bool) bool](),
		Any:       newTypeCreator[func(any) bool](),
		SetString: newTypeCreator[func(string) string](),
		SetInt:    newTypeCreator[func(int) int](),
		SetUint:   newTypeCreator[func(uint) uint](),
		SetFloat:  newTypeCreator[func(float64) float64](),
		SetAny:    newTypeCreator[func(any) any](),
		Errors:    make(map[string]string),
	}
	loadDefaultFuncDefine(fc)
	return fc
}

// NewFuncCreatorExpr function creates [FuncCreator] that supports the parsing
// of AND, OR, and NOT relational expressions.
func NewFuncCreatorExpr() FuncCreator {
	return &funcCreatorExpr{
		data:   NewFuncCreator().(*funcCreatorBase),
		parser: newFcExprParser(),
	}
}

// NewFuncCreatorWithContext function obtains the [ContextKeyFuncCreator]
// value from the [context.Context] as the FuncCreator,
// and defaults to returning the [DefaultFuncCreator].
func NewFuncCreatorWithContext(ctx context.Context) FuncCreator {
	fc, ok := ctx.Value(ContextKeyFuncCreator).(FuncCreator)
	if ok {
		return fc
	}
	return DefaultFuncCreator
}

// RegisterFunc 函数给一个名称注册多个类型的的ValidateFunc或ValidateNewFunc。
//
//nolint:cyclop,gocyclo
func (fc *funcCreatorBase) RegisterFunc(name string, funcs ...any) error {
	for i := range funcs {
		switch fn := funcs[i].(type) {
		case func(string) bool:
			fc.String.Register(name, fn)
		case func(int) bool:
			fc.Int.Register(name, fn)
		case func(uint) bool:
			fc.Uint.Register(name, fn)
		case func(float64) bool:
			fc.Float.Register(name, fn)
		case func(bool) bool:
			fc.Bool.Register(name, fn)
		case func(any) bool:
			fc.Any.Register(name, fn)
		case func(string) (func(string) bool, error):
			fc.String.RegisterNew(name, fn)
		case func(string) (func(uint) bool, error):
			fc.Uint.RegisterNew(name, fn)
		case func(string) (func(int) bool, error):
			fc.Int.RegisterNew(name, fn)
		case func(string) (func(float64) bool, error):
			fc.Float.RegisterNew(name, fn)
		case func(string) (func(bool) bool, error):
			fc.Bool.RegisterNew(name, fn)
		case func(string) (func(any) bool, error):
			fc.Any.RegisterNew(name, fn)
		case func(string) string:
			fc.SetString.Register(name, fn)
		case func(int) int:
			fc.SetInt.Register(name, fn)
		case func(uint) uint:
			fc.SetUint.Register(name, fn)
		case func(float64) float64:
			fc.SetFloat.Register(name, fn)
		case func(any) any:
			fc.SetAny.Register(name, fn)
		case func(string) (func(string) string, error):
			fc.SetString.RegisterNew(name, fn)
		case func(string) (func(uint) uint, error):
			fc.SetUint.RegisterNew(name, fn)
		case func(string) (func(int) int, error):
			fc.SetInt.RegisterNew(name, fn)
		case func(string) (func(float64) float64, error):
			fc.SetFloat.RegisterNew(name, fn)
		case func(string) (func(any) any, error):
			fc.SetAny.RegisterNew(name, fn)
		default:
			return fc.appendError(name, fmt.Errorf(ErrFormatFuncCreatorRegisterInvalidType, name, fn))
		}
	}
	return nil
}

// CreateFunc 方法感觉类型和名称创建校验函数。
//
// 不支持动态创建具有NOT AND OR关系表达式函数，闭包影响性能。
func (fc *funcCreatorBase) CreateFunc(kind FuncCreateKind, name string) (any, error) {
	var fn any
	var err error
	switch kind {
	case FuncCreateString:
		fn, err = fc.String.Create(name)
	case FuncCreateInt:
		fn, err = fc.Int.Create(name)
	case FuncCreateUint:
		fn, err = fc.Uint.Create(name)
	case FuncCreateFloat:
		fn, err = fc.Float.Create(name)
	case FuncCreateBool, FuncCreateSetBool:
		fn, err = fc.Bool.Create(name)
	case FuncCreateAny:
		fn, err = fc.Any.Create(name)
	case FuncCreateSetString:
		fn, err = fc.SetString.Create(name)
	case FuncCreateSetInt:
		fn, err = fc.SetInt.Create(name)
	case FuncCreateSetUint:
		fn, err = fc.SetUint.Create(name)
	case FuncCreateSetFloat:
		fn, err = fc.SetFloat.Create(name)
	case FuncCreateSetAny:
		fn, err = fc.SetAny.Create(name)
	default:
		err = fmt.Errorf("invalid func kind %d", kind)
	}
	if err != nil {
		return nil, fc.appendError(kind.String()+name, fmt.Errorf("funcCreator create kind %s func %s err: %w", kind, name, err))
	}
	return fn, nil
}

func (fc *funcCreatorBase) appendError(key string, err error) error {
	fc.Bool.Lock()
	fc.Errors[key] = err.Error()
	fc.Bool.Unlock()
	return err
}

func (fc *funcCreatorBase) List() []string {
	names := make([]string, 0, 128)
	names = fc.String.List(names)
	names = fc.Int.List(names)
	names = fc.Uint.List(names)
	names = fc.Float.List(names)
	names = fc.Bool.List(names)
	names = fc.Any.List(names)
	names = fc.SetString.List(names)
	names = fc.SetInt.List(names)
	names = fc.SetUint.List(names)
	names = fc.SetFloat.List(names)
	names = fc.SetAny.List(names)
	sort.Strings(names)
	return names
}

func (fc *funcCreatorBase) Metadata() any {
	errs := make([]string, 0, len(fc.Errors))
	fc.Bool.RLock()
	defer fc.Bool.RUnlock()
	for _, v := range fc.Errors {
		errs = append(errs, v)
	}
	return MetadataFuncCreator{
		Health: len(errs) == 0,
		Name:   "eudore.funcCreatorBase",
		Funcs:  fc.List(),
		Errors: errs,
	}
}

type typeCreator[T any] struct {
	sync.RWMutex
	Values      map[string]T
	Constructor map[string]func(string) (T, error)
}

func newTypeCreator[T any]() typeCreator[T] {
	return typeCreator[T]{
		Values:      make(map[string]T),
		Constructor: make(map[string]func(string) (T, error)),
	}
}

func (tc *typeCreator[T]) Register(name string, fn T) {
	tc.Lock()
	tc.Values[name] = fn
	tc.Unlock()
}

func (tc *typeCreator[T]) RegisterNew(name string, fn func(string) (T, error)) {
	tc.Lock()
	tc.Constructor[name] = fn
	tc.Unlock()
}

func (tc *typeCreator[T]) Get(fullname string) (T, bool) {
	tc.RLock()
	fn, ok := tc.Values[fullname]
	tc.RUnlock()
	return fn, ok
}

func (tc *typeCreator[T]) Create(fullname string) (T, error) {
	tc.RLock()
	fn, ok := tc.Values[fullname]
	tc.RUnlock()
	if ok {
		return fn, nil
	}

	name, arg := getFuncNameArg(fullname)
	if arg != "" {
		tc.RLock()
		fnnews, ok := tc.Constructor[name]
		tc.RUnlock()
		if ok {
			fn, err := fnnews(arg)
			if err == nil {
				tc.Register(fullname, fn)
			}
			return fn, err
		}
	}
	return fn, ErrFuncCreatorNotFunc
}

func (tc *typeCreator[T]) List(names []string) []string {
	tc.RLock()
	defer tc.RUnlock()
	for key, fn := range tc.Values {
		names = append(names, fmt.Sprintf("%s: %T", key, fn))
	}
	for key, fn := range tc.Constructor {
		names = append(names, fmt.Sprintf("%s: %T", key, fn))
	}
	return names
}

func getFuncNameArg(name string) (string, string) {
	for i, b := range name {
		// ! [0-9A-Za-z]
		if b < 0x30 || (0x39 < b && b < 0x41) || (0x5A < b && b < 0x61) || 0x7A < b {
			return name[:i], name[i:]
		}
	}
	return name, ""
}

type funcCreatorExpr struct {
	data   *funcCreatorBase
	parser *fcExprParser
}

func (fc *funcCreatorExpr) RegisterFunc(name string, funcs ...any) error {
	return fc.data.RegisterFunc(name, funcs...)
}

func (fc *funcCreatorExpr) CreateFunc(kind FuncCreateKind, name string) (any, error) {
	if kind < FuncCreateSetString && (strings.Contains(name, "NOT") ||
		strings.Contains(name, "AND") || strings.Contains(name, "OR")) {
		var fn any
		var err error
		switch kind {
		case FuncCreateString:
			fn, err = createFunc(&fc.data.String, name, fc.parser.parse)
		case FuncCreateInt:
			fn, err = createFunc(&fc.data.Int, name, fc.parser.parse)
		case FuncCreateUint:
			fn, err = createFunc(&fc.data.Uint, name, fc.parser.parse)
		case FuncCreateFloat:
			fn, err = createFunc(&fc.data.Float, name, fc.parser.parse)
		case FuncCreateBool, FuncCreateSetBool:
			fn, err = createFunc(&fc.data.Bool, name, fc.parser.parse)
		case FuncCreateAny:
			fn, err = createFunc(&fc.data.Any, name, fc.parser.parse)
		}
		if err != nil {
			return nil, fc.data.appendError(kind.String()+name, err)
		}
		return fn, nil
	}
	return fc.data.CreateFunc(kind, name)
}

func createFunc[T funcTypeParams](tc *typeCreator[func(T) bool], name string, parser fcExprFunc) (func(T) bool, error) {
	fn, ok := tc.Get(name)
	if ok {
		return fn, nil
	}

	expr, s := parser(name)
	if s != "" {
		return nil, fmt.Errorf("funcCreatorExpr not parse: %s, pos in %d", name, len(name)-len(s))
	}

	fn, err := createExpr(tc, expr)
	if _, isstr := expr.(string); err == nil && !isstr {
		tc.Register(name, fn)
	}
	return fn, err
}

func createExpr[T funcTypeParams](tc *typeCreator[func(T) bool], expr any) (func(T) bool, error) {
	switch e := expr.(type) {
	case fcExprNot:
		fn, err := createExpr(tc, e.Expr)
		if err != nil {
			return nil, err
		}
		return func(t T) bool {
			return !fn(t)
		}, nil
	case fcExprAnd:
		fns := make([]func(T) bool, len(e.Exprs))
		for i := range e.Exprs {
			fn, err := createExpr(tc, e.Exprs[i])
			if err != nil {
				return nil, err
			}
			fns[i] = fn
		}
		return func(t T) bool {
			for i := range fns {
				if !fns[i](t) {
					return false
				}
			}
			return true
		}, nil
	case fcExprOr:
		fns := make([]func(T) bool, len(e.Exprs))
		for i := range e.Exprs {
			fn, err := createExpr(tc, e.Exprs[i])
			if err != nil {
				return nil, err
			}
			fns[i] = fn
		}
		return func(t T) bool {
			for i := range fns {
				if fns[i](t) {
					return true
				}
			}
			return false
		}, nil
	default:
		return tc.Create(expr.(string))
	}
}

func (fc *funcCreatorExpr) List() []string {
	return fc.data.List()
}

func (fc *funcCreatorExpr) Metadata() any {
	meta := fc.data.Metadata().(MetadataFuncCreator)
	meta.Name = "eudore.funcCreatorExpr"
	var funcs, exprs []string
	for _, f := range meta.Funcs {
		if strings.Contains(f, "NOT") || strings.Contains(f, "AND") || strings.Contains(f, "OR") {
			exprs = append(exprs, f)
		} else {
			funcs = append(funcs, f)
		}
	}
	meta.Funcs = funcs
	meta.Exprs = exprs
	return meta
}

type (
	fcExprFunc   func(string) (any, string)
	fcExprParser struct {
		Parsers [][]fcExprFunc
		Handler []func([]any) any
		parse   fcExprFunc
	}
	fcExprNot struct{ Expr any }
	fcExprAnd struct{ Exprs []any }
	fcExprOr  struct{ Exprs []any }
)

func newFcExprParser() *fcExprParser {
	p := &fcExprParser{Handler: []func(data []any) any{
		func(data []any) any { return fcExprOr{fcExprData(data, false)} },
		func(data []any) any { return fcExprAnd{fcExprData(data, true)} },
		func(data []any) any { return fcExprNot{data[1]} },
		func(data []any) any { return data[1] },
		func(data []any) any { return data[0] },
	}}
	p0, p1, p2 := p.p(0), p.p(1), p.p(2)
	p.parse = p0
	p.Parsers = [][]fcExprFunc{
		{p1, fcExprMatch("OR"), p0},
		{p2, fcExprMatch("AND"), p1},
		{fcExprMatch("NOT"), p2},
		{fcExprMatch("("), p0, fcExprMatch(")")},
		{fcExprVal},
	}
	return p
}

func (p *fcExprParser) p(start int) fcExprFunc {
	return func(s string) (any, string) {
		for i, fns := range p.Parsers[start:] {
			var val any
			var vals []any
			str := s
			for _, fn := range fns {
				val, str = fn(strings.TrimSpace(str))
				if val == nil {
					break
				}
				vals = append(vals, val)
			}
			if len(vals) == len(fns) {
				return p.Handler[i+start](vals), str
			}
		}
		return nil, s
	}
}

func fcExprVal(s string) (any, string) {
	l := len(s)
	for _, sub := range []string{"NOT", "AND", "OR", ")"} {
		pos := strings.Index(s[:l], sub)
		if pos != -1 {
			l = pos
		}
	}
	s1, s2 := strings.TrimSpace(s[:l]), s[l:]
	if s1 == "" {
		return nil, s
	}
	return s1, s2
}

var exprEnd = [256]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1, '(': 1}

func fcExprMatch(str string) fcExprFunc {
	return func(s string) (any, string) {
		if strings.HasPrefix(s, str) {
			if len(str) == 1 || (str != s && exprEnd[s[len(str)]] == 1) {
				return "", s[len(str):]
			}
		}
		return nil, s
	}
}

func fcExprData(data []any, and bool) []any {
	d := make([]any, 0, len(data))
	for i := range data {
		switch val := data[i].(type) {
		case string:
			if val == "" {
				continue
			}
		case fcExprAnd:
			if and {
				d = append(d, val.Exprs...)
				continue
			}
		case fcExprOr:
			if !and {
				d = append(d, val.Exprs...)
				continue
			}
		}
		d = append(d, data[i])
	}
	return d
}

var defaultFuncCreateKindStrings = [...]string{
	"invalid", "string", "int", "uint", "float", "bool", "any",
	"setstring", "setint", "setuint", "setfloat", "setbool", "setany",
}

// NewFuncCreateKind function parses the [FuncCreateKind] string.
func NewFuncCreateKind(s string) FuncCreateKind {
	s = strings.ToLower(s)
	for i, str := range defaultFuncCreateKindStrings {
		if s == str {
			return FuncCreateKind(i)
		}
	}
	return FuncCreateInvalid
}

// NewFuncCreateKindWithType function creates [FuncCreateKind] based on [reflect.Type].
func NewFuncCreateKindWithType(t reflect.Type) FuncCreateKind {
	if t == nil {
		return FuncCreateInvalid
	}
	for {
		switch t.Kind() {
		case reflect.Ptr:
			t = t.Elem()
		case reflect.Slice, reflect.Array:
			t = t.Elem()
		case reflect.String:
			return FuncCreateString
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return FuncCreateInt
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return FuncCreateUint
		case reflect.Float32, reflect.Float64:
			return FuncCreateFloat
		case reflect.Bool:
			return FuncCreateBool
		case reflect.Struct, reflect.Map, reflect.Interface:
			return FuncCreateAny
		default:
			return FuncCreateInvalid
		}
	}
}

func (kind FuncCreateKind) String() string {
	return defaultFuncCreateKindStrings[kind]
}

// RunPtr executes any function, dereferences the Ptr type,
// and executes each value of Slice and Array.
func (fn *FuncRunner) RunPtr(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return fn.RunPtr(reflect.Zero(v.Type().Elem()))
		}
		return fn.RunPtr(v.Elem())
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if !fn.RunPtr(v.Index(i)) {
				return false
			}
		}
		return true
	default:
		return fn.Run(v)
	}
}

// Run execute any function.
func (fn *FuncRunner) Run(v reflect.Value) bool {
	switch fn.Kind {
	case FuncCreateString:
		return fn.Func.(func(string) bool)(v.String())
	case FuncCreateInt:
		return fn.Func.(func(int) bool)(int(v.Int()))
	case FuncCreateUint:
		return fn.Func.(func(uint) bool)(uint(v.Uint()))
	case FuncCreateFloat:
		return fn.Func.(func(float64) bool)(v.Float())
	case FuncCreateBool:
		return fn.Func.(func(bool) bool)(v.Bool())
	case FuncCreateAny:
		return fn.Func.(func(any) bool)(v.Interface())
	case FuncCreateSetString:
		v.SetString(fn.Func.(func(string) string)(v.String()))
	case FuncCreateSetInt:
		v.SetInt(int64((fn.Func.(func(int) int)(int(v.Int())))))
	case FuncCreateSetUint:
		v.SetUint(uint64(fn.Func.(func(uint) uint)(uint(v.Uint()))))
	case FuncCreateSetFloat:
		v.SetFloat(fn.Func.(func(float64) float64)(v.Float()))
	case FuncCreateSetBool:
		v.SetBool(fn.Func.(func(bool) bool)(v.Bool()))
	case FuncCreateSetAny:
		val := fn.Func.(func(any) any)(v.Interface())
		if val != nil {
			v.Set(reflect.ValueOf(val))
		} else {
			v.Set(reflect.Zero(v.Type()))
		}
	}
	return true
}
