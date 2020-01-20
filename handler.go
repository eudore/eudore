package eudore

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"
)

type (
	// HandlerFunc 是处理一个Context的函数
	HandlerFunc func(Context)
	// HandlerFuncs 是HandlerFunc的集合，表示多个请求处理函数。
	HandlerFuncs []HandlerFunc
	// HandlerExtender 定义函数扩展处理者的方法。
	HandlerExtender interface {
		RegisterHandlerExtend(interface{}) error
		NewHandlerFuncs(interface{}) HandlerFuncs
		ListExtendHandlerNames() []string
	}
	// handlerExtendBase 定义基础的函数扩展。
	handlerExtendBase struct {
		ExtendInterfaceType []reflect.Type
		ExtendInterfaceFunc []reflect.Value
		ExtendNewFunc       map[reflect.Type]reflect.Value
	}
	// handlerExtendWarp 定义链式函数扩展。
	handlerExtendWarp struct {
		HandlerExtender
		LastExtender HandlerExtender
	}
	handlerHTTP interface {
		HandleHTTP(Context)
	}
	handlerClone interface {
		handlerHTTP
		CloneHandler() handlerHTTP
	}
)

var (
	typeError       = reflect.TypeOf((*error)(nil)).Elem()
	typeContext     = reflect.TypeOf((*Context)(nil)).Elem()
	typeHandlerFunc = reflect.TypeOf((*HandlerFunc)(nil)).Elem()
	typeHTTPHandler = reflect.TypeOf((*http.Handler)(nil)).Elem()
	// contextFuncName key类型一定为HandlerFunc类型
	contextFuncName = make(map[uintptr]string)
	// DefaultHandlerExtend 为默认的函数扩展处理者，是RouterCoreRadix和RouterCoreFull使用的最顶级的函数扩展处理者。
	DefaultHandlerExtend = NewHandlerExtendBase()
)

// init 函数初始化内置扩展的请求上下文处理函数。
func init() {
	DefaultHandlerExtend.RegisterHandlerExtend(NewExtendContextData)
	DefaultHandlerExtend.RegisterHandlerExtend(NewExtendHandlerHTTPClone)
	DefaultHandlerExtend.RegisterHandlerExtend(NewExtendHandlerHTTP)
	DefaultHandlerExtend.RegisterHandlerExtend(NewExtendHandlerNetHTTP)
	DefaultHandlerExtend.RegisterHandlerExtend(NewExtendFuncNetHTTP1)
	DefaultHandlerExtend.RegisterHandlerExtend(NewExtendFuncNetHTTP2)
	DefaultHandlerExtend.RegisterHandlerExtend(NewExtendFuncContextError)
	DefaultHandlerExtend.RegisterHandlerExtend(NewExtendFuncContextRender)
	DefaultHandlerExtend.RegisterHandlerExtend(NewExtendFuncContextRenderError)
	DefaultHandlerExtend.RegisterHandlerExtend(NewExtendFuncRPCMap)
	DefaultHandlerExtend.RegisterHandlerExtend(NewExtendHandlerRPC)
	DefaultHandlerExtend.RegisterHandlerExtend(NewExtendFuncString)
	DefaultHandlerExtend.RegisterHandlerExtend(NewExtendHandlerStringer)
	DefaultHandlerExtend.RegisterHandlerExtend(NewExtendHandlerInterfaceRender)
}

// NewHandlerExtendBase 方法返回一个基本的函数扩展处理对象。
func NewHandlerExtendBase() HandlerExtender {
	return &handlerExtendBase{
		ExtendInterfaceType: make([]reflect.Type, 0, 6),
		ExtendInterfaceFunc: make([]reflect.Value, 0, 6),
		ExtendNewFunc:       make(map[reflect.Type]reflect.Value),
	}
}

// RegisterHandlerExtend 函数注册一个请求上下文处理转换函数，参数必须是一个函数，该函数的参数必须是一个函数h或接口，返回值必须是返回一个HandlerFunc对象。
//
// 例如: func(func(...)) HanderFunc, func(http.Handler) HandlerFunc
func (ext *handlerExtendBase) RegisterHandlerExtend(fn interface{}) error {
	iType := reflect.TypeOf(fn)
	// RegisterHandlerExtend函数的参数必须是一个函数类型
	if iType.Kind() != reflect.Func {
		return ErrRegisterNewHandlerParamNotFunc
	}
	// 检查函数参数必须是一个函数或者接口类型。
	if iType.NumIn() != 1 || (iType.In(0).Kind() != reflect.Func && iType.In(0).Kind() != reflect.Interface) {
		return fmt.Errorf(ErrFormatRegisterHandlerExtendInputParamError, iType.String())
	}
	// 检查函数返回值必须是HandlerFunc
	if iType.NumOut() != 1 || iType.Out(0) != typeHandlerFunc {
		return fmt.Errorf(ErrFormatRegisterHandlerExtendOutputParamError, iType.String())
	}

	if iType.In(0).Kind() == reflect.Interface {
		ext.ExtendInterfaceType = append(ext.ExtendInterfaceType, iType.In(0))
		ext.ExtendInterfaceFunc = append(ext.ExtendInterfaceFunc, reflect.ValueOf(fn))
	} else {
		ext.ExtendNewFunc[iType.In(0)] = reflect.ValueOf(fn)
	}
	return nil
}

// NewHandlerFuncs 函数根据参数返回一个HandlerFuncs。
func (ext *handlerExtendBase) NewHandlerFuncs(i interface{}) HandlerFuncs {
	return ext.newHandlerFuncs(reflect.ValueOf(i))
}

func (ext *handlerExtendBase) newHandlerFuncs(iValue reflect.Value) HandlerFuncs {
	// 基础类型返回
	switch fn := iValue.Interface().(type) {
	case func(Context):
		return HandlerFuncs{fn}
	case HandlerFunc:
		return HandlerFuncs{fn}
	case HandlerFuncs:
		return fn
	}
	// 尝试转换成HandlerFuncs
	fn := ext.newHandlerFunc(iValue)
	if fn != nil {
		return HandlerFuncs{fn}
	}
	// 解引用数组再转换HandlerFuncs
	switch iValue.Type().Kind() {
	case reflect.Slice, reflect.Array:
		var fns HandlerFuncs
		for i := 0; i < iValue.Len(); i++ {
			hs := ext.newHandlerFuncs(iValue.Index(i))
			if hs != nil {
				fns = append(fns, hs...)
			}
		}
		if len(fns) != 0 {
			return fns
		}
	case reflect.Interface, reflect.Ptr:
		return ext.newHandlerFuncs(iValue.Elem())
	}
	return nil
}

// NewHandlerFunc 函数使用一个函数或接口参数转换成请求上下文处理函数。
//
// 参数必须是一个函数，函数拥有一个参数作为入参，一个HandlerFunc对象作为返回值。
func (ext *handlerExtendBase) newHandlerFunc(iValue reflect.Value) HandlerFunc {
	iType := iValue.Type()
	fn, ok := ext.ExtendNewFunc[iType]
	if ok {
		h := createHandlerFunc(fn, iValue)
		if h != nil {
			return h
		}
	}
	// 判断是否实现接口类型
	for i, iface := range ext.ExtendInterfaceType {
		if iType.Implements(iface) {
			h := createHandlerFunc(ext.ExtendInterfaceFunc[i], iValue)
			if h != nil {
				return h
			}
		}
	}
	return nil
}

var formarExtendername = "%s(%s)"

// ListExtendHandlerNames 方法返回全部注册的函数名称。
func (ext *handlerExtendBase) ListExtendHandlerNames() []string {
	names := make([]string, 0, len(ext.ExtendNewFunc))
	for k, v := range ext.ExtendNewFunc {
		if k.Kind() != reflect.Interface {
			names = append(names, fmt.Sprintf(formarExtendername, runtime.FuncForPC(v.Pointer()).Name(), k.String()))
		}
	}
	for i, iface := range ext.ExtendInterfaceType {
		names = append(names, fmt.Sprintf(formarExtendername, runtime.FuncForPC(ext.ExtendInterfaceFunc[i].Pointer()).Name(), iface.String()))
	}
	return names
}

// createHandlerFunc 函数使用转换函数和对象创建一个HandlerFunc，并保存HandlerFunc的名称和使用的扩展函数名称。
func createHandlerFunc(fn, iValue reflect.Value) HandlerFunc {
	h := fn.Call([]reflect.Value{iValue})[0].Interface().(HandlerFunc)
	if h == nil {
		return nil
	}
	// 获取扩展名称
	extname := runtime.FuncForPC(fn.Pointer()).Name()
	if len(extname) > 24 && extname[:25] == "github.com/eudore/eudore." {
		extname = extname[25:]
	}
	// 保存新函数名称
	var name string
	if iValue.Kind() == reflect.Func {
		name = runtime.FuncForPC(iValue.Pointer()).Name()
	} else {
		// 其他类型输出就这样吧，不想想了。
		name = fmt.Sprintf("%#v", iValue.Interface())
	}
	contextFuncName[reflect.ValueOf(h).Pointer()] = fmt.Sprintf("%s(%s)", name, extname)
	return h
}

// NewHandlerExtendWarp 函数创建一个链式HandlerExtender对象。
func NewHandlerExtendWarp(last HandlerExtender) HandlerExtender {
	return &handlerExtendWarp{
		HandlerExtender: NewHandlerExtendBase(),
		LastExtender:    last,
	}
}

// NewHandlerFuncs 方法实现NewHandlerFuncs函数，如果当前HandlerExtender无法创建HandlerFuncs，则调用上级HandlerExtender处理。
func (ext *handlerExtendWarp) NewHandlerFuncs(i interface{}) HandlerFuncs {
	hs := ext.HandlerExtender.NewHandlerFuncs(i)
	if hs != nil {
		return hs
	}
	return ext.LastExtender.NewHandlerFuncs(i)
}

// ListExtendHandlerNames 方法返回全部注册的函数名称。
func (ext *handlerExtendWarp) ListExtendHandlerNames() []string {
	return append(ext.HandlerExtender.ListExtendHandlerNames(), ext.LastExtender.ListExtendHandlerNames()...)
}

// HandlerFuncsFilter 函数过滤掉多个请求上下文处理函数中的空对象。
func HandlerFuncsFilter(hs HandlerFuncs) HandlerFuncs {
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
	nhs := make(HandlerFuncs, 0, num)
	for _, h := range hs {
		if h != nil {
			nhs = append(nhs, h)
		}
	}
	return nhs
}

// HandlerFuncsCombine 函数将两个HandlerFuncs合并成一个，默认现在最大长度63，超过过panic
func HandlerFuncsCombine(hs1, hs2 HandlerFuncs) HandlerFuncs {
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
		panic("HandlerFuncsCombine: too many handlers")
	}
	hs := make(HandlerFuncs, finalSize)
	copy(hs, hs1)
	copy(hs[len(hs1):], hs2)
	return hs
}

// SetHandlerFuncName 实在一个请求上下文处理函数的名称。
func SetHandlerFuncName(i HandlerFunc, name string) {
	contextFuncName[reflect.ValueOf(i).Pointer()] = name
}

// String 实现fmt.Stringer接口，实现输出函数名称。
func (h HandlerFunc) String() string {
	ptr := reflect.ValueOf(h).Pointer()
	name, ok := contextFuncName[ptr]
	if ok {
		return name
	}
	return runtime.FuncForPC(ptr).Name()
}

// MarshalText 实现encoding.TextMarshaler接口。
func (h HandlerFunc) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

// NewExtendHandlerHTTP 函数handlerHTTP接口转换成HandlerFunc。
func NewExtendHandlerHTTP(h handlerHTTP) HandlerFunc {
	return h.HandleHTTP
}

// NewExtendHandlerHTTPClone 函数handlerClone接口转换成一个复制后的HandlerFunc。
func NewExtendHandlerHTTPClone(h handlerClone) HandlerFunc {
	return h.CloneHandler().HandleHTTP
}

// NewExtendHandlerNetHTTP 函数转换处理http.Handler对象。
func NewExtendHandlerNetHTTP(h http.Handler) HandlerFunc {
	clone, ok := h.(interface{ CloneHandler() http.Handler })
	if ok {
		h = clone.CloneHandler()
	}
	return func(ctx Context) {
		h.ServeHTTP(ctx.Response(), ctx.Request())
	}
}

// NewExtendFuncNetHTTP1 函数转换处理func(http.ResponseWriter, *http.Request)类型。
func NewExtendFuncNetHTTP1(fn func(http.ResponseWriter, *http.Request)) HandlerFunc {
	return func(ctx Context) {
		fn(ctx.Response(), ctx.Request())
	}
}

// NewExtendFuncNetHTTP2 函数转换处理http.HandlerFunc类型。
func NewExtendFuncNetHTTP2(fn http.HandlerFunc) HandlerFunc {
	return func(ctx Context) {
		fn(ctx.Response(), ctx.Request())
	}
}

// NewExtendFuncContextError 函数处理func(Context) error返回的error处理。
func NewExtendFuncContextError(fn func(Context) error) HandlerFunc {
	file, line := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).FileLine(0)
	return func(ctx Context) {
		err := fn(ctx)
		if err != nil {
			ctx.WithFields(Fields{"file": file, "line": line}).Fatal(err)
		}
	}
}

// NewExtendFuncContextRender 函数处理func(Context) interface{}返回数据渲染。
func NewExtendFuncContextRender(fn func(Context) interface{}) HandlerFunc {
	file, line := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).FileLine(0)
	return func(ctx Context) {
		data := fn(ctx)
		if ctx.Response().Size() == 0 {
			err := ctx.Render(data)
			if err != nil {
				ctx.WithFields(Fields{"file": file, "line": line}).Fatal(err)
			}
		}
	}
}

// NewExtendFuncContextRenderError 函数处理func(Context) (interface{}, error)返回数据渲染和error处理。
func NewExtendFuncContextRenderError(fn func(Context) (interface{}, error)) HandlerFunc {
	file, line := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).FileLine(0)
	return func(ctx Context) {
		data, err := fn(ctx)
		if err == nil && ctx.Response().Size() == 0 {
			err = ctx.Render(data)
		}
		if err != nil {
			ctx.WithFields(Fields{"file": file, "line": line}).Fatal(err)
		}
	}
}

// NewExtendFuncRPCMap 定义了固定请求和响应为map[string]interface{}类型的函数处理。
//
// 是NewRPCHandlerFunc的一种子集，拥有类型限制，但是没有使用反射。
func NewExtendFuncRPCMap(fn func(Context, map[string]interface{}) (interface{}, error)) HandlerFunc {
	file, line := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).FileLine(0)
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
			ctx.WithFields(Fields{"file": file, "line": line}).Fatal(err)
		}
	}
}

// NewExtendHandlerRPC 函数需要传入一个函数，返回一个请求处理，通过反射来动态调用。
//
// 函数形式： func(Context, Request) (Response, error)
//
// Request和Response的类型可以为map或结构体或者结构体的指针，4个参数需要全部存在，且不可调换顺序。
func NewExtendHandlerRPC(fn interface{}) HandlerFunc {
	iType := reflect.TypeOf(fn)
	iValue := reflect.ValueOf(fn)
	if iType.Kind() != reflect.Func {
		return nil
	}
	if iType.NumIn() != 2 || iType.In(0) != typeContext {
		return nil
	}
	if iType.NumOut() != 2 || iType.Out(1) != typeError {
		return nil
	}
	var typeIn = iType.In(1)
	// 检查请求类型
	switch typeIn.Kind() {
	case reflect.Map, reflect.Struct, reflect.Ptr:
	default:
		return nil
	}
	file, line := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).FileLine(0)
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
			ctx.WithFields(Fields{"file": file, "line": line}).Fatal(err)
			return
		}

		// 渲染返回的数据。
		err = ctx.Render(vals[0].Interface())
		if err != nil {
			ctx.Fatal(err)
		}
	}
}

// NewExtendHandlerStringer 函数处理fmt.Stringer接口类型转换成HandlerFunc。
func NewExtendHandlerStringer(fn fmt.Stringer) HandlerFunc {
	return func(ctx Context) {
		ctx.WriteString(fn.String())
	}
}

// NewExtendFuncString 函数处理func() string，然后指定函数生成的字符串。
func NewExtendFuncString(fn func() string) HandlerFunc {
	return func(ctx Context) {
		ctx.WriteString(fn())
	}
}

// NewExtendHandlerInterfaceRender 函数闭包一个返回指定字符串对象的HandlerFunc。
func NewExtendHandlerInterfaceRender(i interface{}) HandlerFunc {
	return func(ctx Context) {
		ctx.Render(i)
	}
}

// HandlerEmpty 函数定义一个空的请求上下文处理函数。
func HandlerEmpty(Context) {
	// Do nothing because empty handler does not process entries.
}
