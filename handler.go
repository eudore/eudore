package eudore

import (
	"fmt"
	"reflect"
	"runtime"
)

type (
	// HandlerFunc 是处理一个Context的函数
	HandlerFunc func(Context)
	// HandlerFuncs 是HandlerFunc的集合，表示多个请求处理函数。
	HandlerFuncs []HandlerFunc
)

var (
	typeHandlerFunc = reflect.TypeOf((*HandlerFunc)(nil)).Elem()
	typeContext     = reflect.TypeOf((*Context)(nil)).Elem()
	typeError       = reflect.TypeOf((*error)(nil)).Elem()
	contextNewFunc  = make(map[reflect.Type]reflect.Value)
	contextFuncName = make(map[reflect.Value]string)
)

// init 函数初始化内置四种扩展的请求上下文处理函数。
func init() {
	RegisterHandlerFunc(NewContextErrorHanderFunc)
	RegisterHandlerFunc(NewContextRenderHanderFunc)
	RegisterHandlerFunc(NewContextRenderErrorHanderFunc)
	RegisterHandlerFunc(NewContextDataHanderFunc)
	RegisterHandlerFunc(NewRPCMapHandlerFunc)
}

// NewHandlerFuncs 函数根据参数返回一个HandlerFuncs。
func NewHandlerFuncs(i interface{}) HandlerFuncs {
	return newHandlerFuncs(reflect.ValueOf(i))
}

func newHandlerFuncs(iValue reflect.Value) HandlerFuncs {
	switch iValue.Type().Kind() {
	case reflect.Slice:
		var fns HandlerFuncs
		for i := 0; i < iValue.Len(); i++ {
			fns = append(fns, newHandlerFuncs(iValue.Index(i))...)
		}
		return fns
	case reflect.Func:
		return HandlerFuncs{NewHandlerFunc(iValue.Interface())}
	case reflect.Interface:
		return newHandlerFuncs(iValue.Elem())
	}
	panic(fmt.Errorf(ErrFormatNewHandlerFuncsTypeNotFunc, iValue.Type().String()))
}

// NewHandlerFunc 函数使用一个函数参数转换成请求上下文处理函数。
//
// 参数必须是一个函数，函数拥有一个参数作为入参，一个HandlerFunc对象作为返回值。
//
// 例如: func(func(...)) HanderFunc
func NewHandlerFunc(i interface{}) HandlerFunc {
	if reflect.TypeOf(i).Kind() != reflect.Func {
		panic(ErrNewHandlerFuncParamNotFunc)
	}
	switch val := i.(type) {
	case func(Context):
		return val
	case HandlerFunc:
		return val
	default:
		fn, ok := contextNewFunc[reflect.TypeOf(i)]
		if ok {
			h := fn.Call([]reflect.Value{reflect.ValueOf(i)})[0].Interface().(HandlerFunc)
			contextFuncName[reflect.ValueOf(h)] = runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
			return h
		}
	}
	panic(fmt.Errorf(ErrFormatNewHandlerFuncsUnregisterType, reflect.TypeOf(i).String()))
}

// RegisterHandlerFunc 函数注册一个请求上下文处理转换函数，参数必须是一个函数，该函数的参数必须是一个函数，返回值必须是返回一个HandlerFunc对象。
func RegisterHandlerFunc(fn interface{}) {
	iType := reflect.TypeOf(fn)
	if iType.Kind() != reflect.Func {
		panic(ErrRegisterNewHandlerParamNotFunc)
	}
	if iType.NumIn() != 1 || iType.In(0).Kind() != reflect.Func {
		panic(fmt.Errorf(ErrFormatRegisterHandlerFuncInputParamError, iType.String()))
	}
	if iType.NumOut() != 1 || iType.Out(0) != typeHandlerFunc {
		panic(fmt.Errorf(ErrFormatRegisterHandlerFuncOutputParamError, iType.String()))

	}
	contextNewFunc[iType.In(0)] = reflect.ValueOf(fn)
}

// FilterHandlerFuncs 函数过滤掉多个请求上下文处理函数中的空对象。
func FilterHandlerFuncs(hs HandlerFuncs) HandlerFuncs {
	var num int
	for _, h := range hs {
		if h != nil {
			num++
		}
	}
	if num == len(hs) {
		return hs
	}

	// 返回新过滤空的处理函数。
	nhs := make(HandlerFuncs, num)
	for _, h := range hs {
		if h != nil {
			nhs = append(nhs, h)
		}
	}
	return nhs
}

// CombineHandlerFuncs 函数将两个HandlerFuncs合并成一个，默认现在最大长度63，超过过panic
func CombineHandlerFuncs(hs1, hs2 HandlerFuncs) HandlerFuncs {
	// if nil
	if len(hs1) == 0 {
		return hs2
	}
	if len(hs2) == 0 {
		return hs1
	}
	// combine
	const abortIndex int8 = 63
	finalSize := len(hs1) + len(hs2)
	if finalSize >= int(abortIndex) {
		panic("CombineHandlerFuncs: too many handlers")
	}
	hs := make(HandlerFuncs, finalSize)
	copy(hs, hs1)
	copy(hs[len(hs1):], hs2)
	return hs
}

// SetHandlerFuncName 实在一个请求上下文处理函数的名称。
func SetHandlerFuncName(i HandlerFunc, name string) {
	contextFuncName[reflect.ValueOf(i)] = name
}

// String 实现fmt.Stringer接口，实现输出函数名称。
func (h HandlerFunc) String() string {
	name, ok := contextFuncName[reflect.ValueOf(h)]
	if ok {
		return name
	}
	return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
}

// MarshalText 实现encoding.TextMarshaler接口。
func (h HandlerFunc) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

// ListExtendHandlerFunc 函数返回注册的扩展执行函数一定类型名称。
func ListExtendHandlerFunc() []string {
	strs := make([]string, 0, len(contextNewFunc))
	for i := range contextNewFunc {
		strs = append(strs, i.String())
	}
	return strs
}

// NewContextErrorHanderFunc 函数处理func(Context) error返回的error处理。
func NewContextErrorHanderFunc(fn func(Context) error) HandlerFunc {
	return func(ctx Context) {
		err := fn(ctx)
		if err != nil {
			ctx.Fatal(err)
		}
	}
}

// NewContextRenderHanderFunc 函数处理func(Context) interface{}返回数据渲染。
func NewContextRenderHanderFunc(fn func(Context) interface{}) HandlerFunc {
	return func(ctx Context) {
		data := fn(ctx)
		if ctx.Response().Size() == 0 {
			err := ctx.Render(data)
			if err != nil {
				ctx.Fatal(err)
			}
		}
	}
}

// NewContextRenderErrorHanderFunc 函数处理func(Context) (interface{}, error)返回数据渲染和error处理。
func NewContextRenderErrorHanderFunc(fn func(Context) (interface{}, error)) HandlerFunc {
	return func(ctx Context) {
		data, err := fn(ctx)
		if err == nil && ctx.Response().Size() == 0 {
			err = ctx.Render(data)
		}
		if err != nil {
			ctx.Fatal(err)
		}
	}
}

// NewRPCMapHandlerFunc 定义了固定请求和响应为map[string]interface{}类型的函数处理。
//
// 是NewRPCHandlerFunc的一种子集，拥有类型限制，但是没有使用反射。
func NewRPCMapHandlerFunc(fn func(Context, map[string]interface{}) (map[string]interface{}, error)) HandlerFunc {
	return func(ctx Context) {
		req := make(map[string]interface{})
		err := ctx.Bind(&req)
		if err != nil {
			ctx.Fatal(err)
			return
		}
		resp, err := fn(ctx, req)
		if err == nil && ctx.Response().Size() == 0 {
			err = ctx.Render(resp)
		}
		if err != nil {
			ctx.Fatal(err)
		}
	}
}

// NewRPCHandlerFunc 函数需要传入一个函数，返回一个请求处理，通过反射来动态调用。
//
// 函数形式： func(Context, Request) (Response, error)
//
// Request和Response的类型可以为map或结构体或者结构体的指针，4个参数需要全部存在，但是不可调换顺序。
func NewRPCHandlerFunc(fn interface{}) HandlerFunc {
	iType := reflect.TypeOf(fn)
	iValue := reflect.ValueOf(fn)
	if iType.Kind() != reflect.Func {
		panic("func is invalid")
	}
	if iType.NumIn() != 2 || iType.In(0) != typeContext {
		panic("--")
	}
	if iType.NumOut() != 2 || iType.Out(1) != typeError {

	}
	var typeIn = iType.In(1)
	// 检查请求类型
	switch typeIn.Kind() {
	case reflect.Map, reflect.Struct, reflect.Ptr:
	default:
		panic("func request not is map, struct, ptr.")
	}
	return func(ctx Context) {
		// 创建请求参数并初始化
		var req reflect.Value
		switch typeIn.Kind() {
		case reflect.Ptr:
			req = reflect.New(typeIn.Elem())
		case reflect.Struct, reflect.Map:
			req = reflect.New(typeIn)
		}

		err := ctx.Bind(req.Interface())
		if err != nil {
			ctx.Fatal(err)
			return
		}
		if typeIn.Kind() != reflect.Ptr {
			req = req.Elem()
		}

		// 反射调用执行函数。
		vals := iValue.Call([]reflect.Value{reflect.ValueOf(ctx), req})

		// 检查函数执行err。
		err, ok := vals[1].Interface().(error)
		if ok {
			ctx.Fatal(err)
			return
		}

		// 渲染返回的数据。
		err = ctx.Render(vals[0].Interface())
		if err != nil {
			ctx.Fatal(err)
		}
	}
}

// HandlerEmpty 函数定义一个空的请求上下文处理函数。
func HandlerEmpty(Context) {
	// Do nothing because empty handler does not process entries.
}
