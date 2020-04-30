package eudore

import (
	"fmt"
	"net/http"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

// HandlerFunc 是处理一个Context的函数
type HandlerFunc func(Context)

// HandlerFuncs 是HandlerFunc的集合，表示多个请求处理函数。
type HandlerFuncs []HandlerFunc

// HandlerExtender 定义函数扩展处理者的方法。
//
// HandlerExtender默认拥有Base、Warp、Tree三种实现，具体参数三种对象的文档。
type HandlerExtender interface {
	RegisterHandlerExtend(string, interface{}) error
	NewHandlerFuncs(string, interface{}) HandlerFuncs
	ListExtendHandlerNames() []string
}

// handlerExtendBase 定义基础的函数扩展。
type handlerExtendBase struct {
	ExtendNewType       []reflect.Type
	ExtendNewFunc       []reflect.Value
	ExtendInterfaceType []reflect.Type
	ExtendInterfaceFunc []reflect.Value
	// ExtendNewFunc       map[reflect.Type]reflect.Value
}

// handlerExtendWarp 定义链式函数扩展。
type handlerExtendWarp struct {
	HandlerExtender
	LastExtender HandlerExtender
}

// handlerExtendTree 定义基于路径匹配的函数扩展。
type handlerExtendTree struct {
	HandlerExtender
	path   string
	childs []*handlerExtendTree
}

type handlerHTTP interface {
	HandleHTTP(Context)
}

var (
	// contextFuncName key类型一定为HandlerFunc类型，保存函数可能正确的名称。
	contextFuncName  = make(map[uintptr]string)
	contextSaveName  = make(map[uintptr]string)
	contextAliasName = make(map[uintptr]string)
)

// init 函数初始化内置扩展的请求上下文处理函数。
func init() {
	// 路由方法扩展
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendContextData)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendHandlerHTTP)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendHandlerNetHTTP)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendFuncNetHTTP1)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendFuncNetHTTP2)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendFunc)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendFuncRender)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendFuncError)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendFuncContextError)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendFuncContextRender)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendFuncContextRenderError)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendFuncRPCMap)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendHandlerRPC)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendFuncString)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendHandlerStringer)

	// 控制器方法扩展
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendControllerFunc)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendControllerFuncRender)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendControllerFuncError)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendControllerFuncRenderError)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendControllerFuncContext)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendControllerFuncContextRender)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendControllerFuncContextError)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendControllerFuncContextRenderError)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendControllerFuncMapString)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendControllerFuncMapStringRender)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendControllerFuncMapStringError)
	DefaultHandlerExtend.RegisterHandlerExtend("", NewExtendControllerFuncMapStringRenderError)
}

// NewHandlerExtendBase method returns a basic function extension processing object.
//
// The NewHandlerExtendBase().RegisterHandlerExtend method registers a conversion function and ignores the path.
//
// The NewHandlerExtendBase().NewHandlerFuncs method implementation creates multiple request handler functions, ignoring paths.
//
// NewHandlerExtendBase 方法返回一个基本的函数扩展处理对象。
//
// NewHandlerExtendBase().RegisterHandlerExtend 方法实现注册一个转换函数，忽略路径。
//
// NewHandlerExtendBase().NewHandlerFuncs 方法实现创建多个请求处理函数，忽略路径。
func NewHandlerExtendBase() HandlerExtender {
	return &handlerExtendBase{}
}

// RegisterHandlerExtend 函数注册一个请求上下文处理转换函数，参数必须是一个函数，该函数的参数必须是一个函数、接口、指针、结构体类型之一，返回值必须是返回一个HandlerFunc对象。
//
// 如果添加多个接口类型转换，注册类型不直接是接口而是实现接口，会按照接口注册顺序依次检测是否实现接口。
//
// 例如: func(func(...)) HanderFunc, func(http.Handler) HandlerFunc
func (ext *handlerExtendBase) RegisterHandlerExtend(_ string, fn interface{}) error {
	iType := reflect.TypeOf(fn)
	// RegisterHandlerExtend函数的参数必须是一个函数类型
	if iType.Kind() != reflect.Func {
		return ErrRegisterNewHandlerParamNotFunc
	}
	// 检查函数参数必须是一个函数、接口、指针、结构体类型之一类型。
	if iType.NumIn() != 1 || (iType.In(0).Kind() != reflect.Func &&
		iType.In(0).Kind() != reflect.Interface &&
		iType.In(0).Kind() != reflect.Ptr &&
		iType.In(0).Kind() != reflect.Struct) {
		return fmt.Errorf(ErrFormatRegisterHandlerExtendInputParamError, iType.String())
	}
	// 检查函数返回值必须是HandlerFunc
	if iType.NumOut() != 1 || iType.Out(0) != typeHandlerFunc {
		return fmt.Errorf(ErrFormatRegisterHandlerExtendOutputParamError, iType.String())
	}

	ext.ExtendNewType = append(ext.ExtendNewType, iType.In(0))
	ext.ExtendNewFunc = append(ext.ExtendNewFunc, reflect.ValueOf(fn))
	if iType.In(0).Kind() == reflect.Interface {
		ext.ExtendInterfaceType = append(ext.ExtendInterfaceType, iType.In(0))
		ext.ExtendInterfaceFunc = append(ext.ExtendInterfaceFunc, reflect.ValueOf(fn))
	}
	return nil
}

// NewHandlerFuncs 函数根据参数返回一个HandlerFuncs。
func (ext *handlerExtendBase) NewHandlerFuncs(_ string, i interface{}) HandlerFuncs {
	return HandlerFuncsFilter(ext.newHandlerFuncs(reflect.ValueOf(i)))
}

func (ext *handlerExtendBase) newHandlerFuncs(iValue reflect.Value) HandlerFuncs {
	// 基础类型返回
	switch fn := iValue.Interface().(type) {
	case func(Context):
		return HandlerFuncs{fn}
	case HandlerFunc:
		return HandlerFuncs{fn}
	case []HandlerFunc:
		return fn
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

// newHandlerFunc 函数使用一个函数或接口参数转换成请求上下文处理函数。
//
// 参数必须是一个函数，函数拥有一个参数作为入参，一个HandlerFunc对象作为返回值。
//
// 先检测对象是否拥有直接注册的类型扩展函数，再检查对象是否实现其中注册的接口类型。
//
// 允许进行多次注册，只要注册返回值不为空就会返回对应的处理函数。
func (ext *handlerExtendBase) newHandlerFunc(iValue reflect.Value) HandlerFunc {
	iType := iValue.Type()
	for i := range ext.ExtendNewType {
		if ext.ExtendNewType[i] == iType {
			h := ext.createHandlerFunc(ext.ExtendNewFunc[i], iValue)
			if h != nil {
				return h
			}
		}
	}
	// 判断是否实现接口类型
	for i, iface := range ext.ExtendInterfaceType {
		if iType.Implements(iface) {
			h := ext.createHandlerFunc(ext.ExtendInterfaceFunc[i], iValue)
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
	for i := range ext.ExtendNewType {
		if ext.ExtendNewType[i].Kind() != reflect.Interface {
			names = append(names, fmt.Sprintf(formarExtendername, runtime.FuncForPC(ext.ExtendNewFunc[i].Pointer()).Name(), ext.ExtendNewType[i].String()))
		}
	}
	for i, iface := range ext.ExtendInterfaceType {
		names = append(names, fmt.Sprintf(formarExtendername, runtime.FuncForPC(ext.ExtendInterfaceFunc[i].Pointer()).Name(), iface.String()))
	}
	return names
}

// createHandlerFunc 函数使用转换函数和对象创建一个HandlerFunc，并保存HandlerFunc的名称和使用的扩展函数名称。
func (ext *handlerExtendBase) createHandlerFunc(fn, iValue reflect.Value) HandlerFunc {
	h := fn.Call([]reflect.Value{iValue})[0].Interface().(HandlerFunc)
	if h == nil {
		return nil
	}
	// 获取扩展名称，eudore包移除包前缀
	extname := runtime.FuncForPC(fn.Pointer()).Name()
	if len(extname) > 24 && extname[:25] == "github.com/eudore/eudore." {
		extname = extname[25:]
	}
	// 获取新函数名称,一般来源于函数扩展返回的函数名称。
	name := contextSaveName[reflect.ValueOf(h).Pointer()]
	// 使用原值名称
	if name == "" && iValue.Kind() != reflect.Struct {
		name = contextAliasName[iValue.Pointer()]
	}
	// 推断名称
	if name == "" {
		iType := iValue.Type()
		switch iType.Kind() {
		case reflect.Func:
			name = runtime.FuncForPC(iValue.Pointer()).Name()
		case reflect.Ptr:
			iType = iType.Elem()
			name = fmt.Sprintf("*%s.%s", iType.PkgPath(), iType.Name())
		case reflect.Struct:
			name = fmt.Sprintf("%s.%s", iType.PkgPath(), iType.Name())
		}
	}
	contextFuncName[reflect.ValueOf(h).Pointer()] = fmt.Sprintf("%s(%s)", name, extname)
	return h
}

// NewHandlerExtendWarp function creates a chained HandlerExtender object.
//
// All objects are registered and created using base. If base cannot create a function handler, use last to create a function handler.
//
// NewHandlerExtendWarp 函数创建一个链式HandlerExtender对象。
//
// The NewHandlerExtendWarp(base, last).RegisterHandlerExtend method uses the base object to register extension functions.
//
// The NewHandlerExtendWarp(base, last).NewHandlerFuncs method first uses the base object to create multiple request processing functions. If it returns nil, it uses the last object to create multiple request processing functions.
//
// 所有对象注册和创建均使用base，如果base无法创建函数处理者则使用last创建函数处理者。
//
// NewHandlerExtendWarp(base, last).RegisterHandlerExtend 方法使用base对象注册扩展函数。
//
// NewHandlerExtendWarp(base, last).NewHandlerFuncs 方法先使用base对象创建多个请求处理函数，如果返回nil，则使用last对象创建多个请求处理函数。
func NewHandlerExtendWarp(base, last HandlerExtender) HandlerExtender {
	return &handlerExtendWarp{
		HandlerExtender: base,
		LastExtender:    last,
	}
}

// The NewHandlerFuncs method implements the NewHandlerFuncs function. If the current HandlerExtender cannot create HandlerFuncs, it calls the superior HandlerExtender to process.
//
// NewHandlerFuncs 方法实现NewHandlerFuncs函数，如果当前HandlerExtender无法创建HandlerFuncs，则调用上级HandlerExtender处理。
func (ext *handlerExtendWarp) NewHandlerFuncs(path string, i interface{}) HandlerFuncs {
	hs := ext.HandlerExtender.NewHandlerFuncs(path, i)
	if hs != nil {
		return hs
	}
	return ext.LastExtender.NewHandlerFuncs(path, i)
}

// ListExtendHandlerNames 方法返回全部注册的函数名称。
func (ext *handlerExtendWarp) ListExtendHandlerNames() []string {
	return append(ext.LastExtender.ListExtendHandlerNames(), ext.HandlerExtender.ListExtendHandlerNames()...)
}

// NewHandlerExtendTree function creates a path-based function extender.
//
// Mainly implement path matching. All actions are processed by the node's HandlerExtender, and the NewHandlerExtendBase () object is used.
//
// All registration and creation actions will be performed by matching the lowest node of the tree. If it cannot be created, the tree nodes will be processed upwards in order.
//
// The NewHandlerExtendTree().RegisterHandlerExtend method registers a handler function based on the path, and initializes to NewHandlerExtendBase () if the HandlerExtender is empty.
//
// The NewHandlerExtendTree().NewHandlerFuncs method matches the child nodes of the tree based on the path, and then executes the NewHandlerFuncs method from the most child node up. If it returns non-null, it returns directly.
//
// NewHandlerExtendTree 函数创建一个基于路径的函数扩展者。
//
// 主要实现路径匹配，所有行为使用节点的HandlerExtender处理，使用NewHandlerExtendBase()对象。
//
// 所有注册和创建行为都会匹配树最下级节点执行，如果无法创建则在树节点依次向上处理。
//
// NewHandlerExtendTree().RegisterHandlerExtend 方法基于路径注册一个处理函数，如果HandlerExtender为空则初始化为NewHandlerExtendBase()。
//
// NewHandlerExtendTree().NewHandlerFuncs 方法基于路径向树子节点匹配，后从最子节点依次向上执行NewHandlerFuncs方法，如果返回非空直接返回，否在会依次执行注册行为。
func NewHandlerExtendTree() HandlerExtender {
	return &handlerExtendTree{}
}

// RegisterHandlerExtend 方法基于路径注册一个扩展函数。
func (ext *handlerExtendTree) RegisterHandlerExtend(path string, i interface{}) error {
	// 匹配当前节点注册
	if path == "" {
		if ext.HandlerExtender == nil {
			ext.HandlerExtender = NewHandlerExtendBase()
		}
		return ext.HandlerExtender.RegisterHandlerExtend("", i)
	}

	// 寻找对应的子节点注册
	for pos := range ext.childs {
		subStr, find := getSubsetPrefix(path, ext.childs[pos].path)
		if find {
			if subStr != ext.childs[pos].path {
				ext.childs[pos].path = strings.TrimPrefix(ext.childs[pos].path, subStr)
				ext.childs[pos] = &handlerExtendTree{
					path:   subStr,
					childs: []*handlerExtendTree{ext.childs[pos]},
				}
			}
			return ext.childs[pos].RegisterHandlerExtend(strings.TrimPrefix(path, subStr), i)
		}
	}

	// 追加一个新的子节点
	newnode := &handlerExtendTree{
		path:            path,
		HandlerExtender: NewHandlerExtendBase(),
	}
	ext.childs = append(ext.childs, newnode)
	return newnode.HandlerExtender.RegisterHandlerExtend(path, i)
}

// NewHandlerFuncs 函数基于路径创建多个对象处理函数。
//
// 递归依次寻找子节点，然后返回时创建多个对象处理函数，如果子节点返回不为空就直接返回。
func (ext *handlerExtendTree) NewHandlerFuncs(path string, i interface{}) HandlerFuncs {
	for _, i := range ext.childs {
		if strings.HasPrefix(path, i.path) {
			hs := i.NewHandlerFuncs(path[len(i.path):], i)
			if hs != nil {
				return hs
			}
			break
		}
	}

	if ext.HandlerExtender != nil {
		return ext.HandlerExtender.NewHandlerFuncs(path, i)
	}
	return nil
}

// listExtendHandlerNamesByPrefix 方法递归添加路径前缀返回扩展函数名称。
func (ext *handlerExtendTree) listExtendHandlerNamesByPrefix(prefix string) []string {
	prefix += ext.path
	var names []string
	if ext.HandlerExtender != nil {
		names = ext.HandlerExtender.ListExtendHandlerNames()
		if prefix != "" {
			for i := range names {
				names[i] = prefix + " " + names[i]
			}
		}
	}

	for i := range ext.childs {
		names = append(names, ext.childs[i].listExtendHandlerNamesByPrefix(prefix)...)
	}
	return names
}

// ListExtendHandlerNames 方法返回全部注册的函数名称。
func (ext *handlerExtendTree) ListExtendHandlerNames() []string {
	return ext.listExtendHandlerNamesByPrefix("")
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

// HandlerFuncsCombine function merges two HandlerFuncs into one. The default maximum length is now 63, which exceeds panic.
//
// Used to reconstruct the slice and prevent the appended data from being confused.
//
// HandlerFuncsCombine 函数将两个HandlerFuncs合并成一个，默认现在最大长度63，超过过panic。
//
// 用于重构切片，防止切片append数据混乱。
func HandlerFuncsCombine(hs1, hs2 HandlerFuncs) HandlerFuncs {
	// if nil
	if len(hs1) == 0 {
		return hs2
	}
	if len(hs2) == 0 {
		return hs1
	}
	// combine
	finalSize := len(hs1) + len(hs2)
	if finalSize >= 127 {
		panic("HandlerFuncsCombine: too many handlers")
	}
	hs := make(HandlerFuncs, finalSize)
	copy(hs, hs1)
	copy(hs[len(hs1):], hs2)
	return hs
}

// SetHandlerFuncName function sets the name of a request context handler.
//
// Note: functions are not comparable, the method names of objects are overwritten by other method names.
//
// SetHandlerFuncName 函数设置一个请求上下文处理函数的名称。
//
// 注意：函数不具有可比性，对象的方法的名称会被其他方法名称覆盖。
func SetHandlerFuncName(i HandlerFunc, name string) {
	contextSaveName[reflect.ValueOf(i).Pointer()] = name
}

// SetHandlerAliasName 函数设置一个函数处理对象原始名称，如果扩展未生成名称，使用此值。
//
// 在handlerExtendBase对象和ControllerInjectSingleton函数中使用到，用于传递控制器函数名称。
func SetHandlerAliasName(i interface{}, name string) {
	iValue, ok := i.(reflect.Value)
	if !ok {
		iValue = reflect.ValueOf(i)
	}
	if iValue.Kind() == reflect.Func || iValue.Kind() == reflect.Ptr {
		contextAliasName[iValue.Pointer()] = name
	}
}

// String method implements the fmt.Stringer interface and implements the output function name.
//
// If the processing function has set the function name, use the set value, or use the runtime to get the default value. Method names may be confusing.
//
// String 方法实现fmt.Stringer接口，实现输出函数名称。
//
// 如果处理函数设置过函数名称，使用设置的值，否在使用runtime获取默认值，方法名称可能混乱。
func (h HandlerFunc) String() string {
	ptr := reflect.ValueOf(h).Pointer()
	name, ok := contextFuncName[ptr]
	if ok {
		return name
	}
	name, ok = contextSaveName[ptr]
	if ok {
		return name
	}
	return runtime.FuncForPC(ptr).Name()
}

// NewExtendHandlerHTTP 函数handlerHTTP接口转换成HandlerFunc。
func NewExtendHandlerHTTP(h handlerHTTP) HandlerFunc {
	return h.HandleHTTP
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

// NewExtendFunc 函数处理func()。
func NewExtendFunc(fn func()) HandlerFunc {
	return func(Context) {
		fn()
	}
}

// NewExtendFuncRender 函数处理func() interface{}。
func NewExtendFuncRender(fn func() interface{}) HandlerFunc {
	file, line := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).FileLine(0)
	return func(ctx Context) {
		data := fn()
		if ctx.Response().Size() == 0 {
			err := ctx.Render(data)
			if err != nil {
				ctx.WithFields(Fields{"file": file, "line": line}).Fatal(err)
			}
		}
	}
}

// NewExtendFuncError 函数处理func() error返回的error处理。
func NewExtendFuncError(fn func() error) HandlerFunc {
	file, line := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).FileLine(0)
	return func(ctx Context) {
		err := fn()
		if err != nil {
			ctx.WithFields(Fields{"file": file, "line": line}).Fatal(err)
		}
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

// NewExtendFuncRPCMap defines a fixed request and response to function processing of type map [string] interface {}.
//
// is a subset of NewRPCHandlerFunc and has type restrictions, but using map [string] interface {} to save requests does not use reflection.
//
// NewExtendFuncRPCMap 定义了固定请求和响应为map[string]interface{}类型的函数处理。
//
// 是NewRPCHandlerFunc的一种子集，拥有类型限制，但是使用map[string]interface{}保存请求没用使用反射。
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

// NewExtendHandlerRPC function needs to pass in a function that returns a request for processing and is dynamically called by reflection.
//
// Function form: func (Context, Request) (Response, error)
//
// The types of Request and Response can be map or struct or pointer to struct. All 4 parameters need to exist, and the order cannot be changed.
//
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

// NewStaticHandler 函数更加目标创建一个静态文件处理函数。
func NewStaticHandler(dir string) HandlerFunc {
	if dir == "" {
		dir = "."
	}
	return func(ctx Context) {
		upath := ctx.GetParam("path")
		if upath == "" {
			upath = ctx.Path()
		}
		ctx.WriteFile(filepath.Join(dir, filepath.Clean("/"+upath)))
	}
}

// HandlerEmpty 函数定义一个空的请求上下文处理函数。
func HandlerEmpty(Context) {
	// Do nothing because empty handler does not process entries.
}
