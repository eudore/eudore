package eudore

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

// HandlerExtender defines the method of extending the function handler.
//
// The HandlerExtender object has three default implementations, Base, Warp, and Tree,
// which are defined in the HandlerExtender interface.
//
// HandlerExtender 定义函数扩展处理者的方法。
//
// HandlerExtender默认拥有Base、Warp、Tree三种实现，具体参数三种对象的文档。
type HandlerExtender interface {
	RegisterExtender(string, any) error
	CreateHandler(string, any) HandlerFuncs
	List() []string
}

// HandlerFuncs is a collection of HandlerFunc,
// representing multiple request processing functions.
//
// handlerExtenderBase 定义基础的函数扩展。
type handlerExtenderBase struct {
	NewType []reflect.Type
	NewFunc []reflect.Value
	AnyType []reflect.Type
	AnyFunc []reflect.Value
}

// handlerExtenderWarp 定义链式函数扩展。
type handlerExtenderWarp struct {
	data HandlerExtender
	last HandlerExtender
}

// handlerExtenderTree 定义基于路径匹配的函数扩展。
type handlerExtenderTree struct {
	data   HandlerExtender
	path   string
	childs []*handlerExtenderTree
}

type MetadataHandlerExtender struct {
	Health   bool     `alias:"health" json:"health" xml:"health" yaml:"health"`
	Name     string   `alias:"name" json:"name" xml:"name" yaml:"name"`
	Extender []string `alias:"extender" json:"extender" xml:"extender" yaml:"extender"`
}

var (
	// contextFuncName key类型一定为HandlerFunc类型，保存函数可能正确的名称。
	contextFuncName    = make(map[uintptr]string)   // 最终名称
	contextSaveName    = make(map[uintptr]string)   // 函数名称
	contextAliasName   = make(map[uintptr][]string) // 对象名称
	fineLineFieldsKeys = []string{"file", "line"}
)

// NewHandlerExtender 函数创建默认HandlerExtender并加载默认扩展函数。
func NewHandlerExtender() HandlerExtender {
	he := NewHandlerExtenderBase()
	for _, fn := range DefaultHandlerExtenderFuncs {
		_ = he.RegisterExtender("", fn)
	}
	return he
}

func NewHandlerExtenderWithContext(ctx context.Context) HandlerExtender {
	he, ok := ctx.Value(ContextKeyHandlerExtender).(HandlerExtender)
	if ok {
		return he
	}
	return DefaultHandlerExtender
}

// NewHandlerExtenderBase method returns a basic function extension processing object.
//
// The NewHandlerExtenderBase().RegisterExtender method registers
// a conversion function and ignores the path.
//
// The NewHandlerExtenderBase().CreateHandler method implementation
// creates multiple request handler functions, ignoring paths.
//
// NewHandlerExtenderBase 方法返回一个基本的函数扩展处理对象。
//
// NewHandlerExtenderBase().RegisterExtender 方法实现注册一个转换函数，忽略路径。
//
// NewHandlerExtenderBase().CreateHandler 方法实现创建多个请求处理函数，忽略路径。
func NewHandlerExtenderBase() HandlerExtender {
	return &handlerExtenderBase{}
}

// RegisterExtender function registers a request context handling conversion function.
// The parameter must be a function that takes a function, an interface, or a pointer type
// as a parameter and returns a HandlerFunc object.
//
// If multiple interface type conversions are added, the registration type is not directly the interface
// but the implementation interface, and the implementation interface will be checked
// in the order of interface registration.
//
// For example: func(func(...)) HanderFunc, func(http.Handler) HandlerFunc
//
// RegisterExtender 函数注册一个请求上下文处理转换函数，参数必须是一个函数，
// 该函数的参数必须是一个函数、接口、指针类型之一，返回值必须是返回一个HandlerFunc对象。
//
// 如果添加多个接口类型转换，注册类型不直接是接口而是实现接口，会按照接口注册顺序依次检测是否实现接口。
//
// 例如: func(func(...)) HanderFunc, func(http.Handler) HandlerFunc.
func (he *handlerExtenderBase) RegisterExtender(_ string, fn any) error {
	iType := reflect.TypeOf(fn)
	// RegisterExtender函数的参数必须是一个函数类型
	if iType.Kind() != reflect.Func {
		return ErrHandlerExtenderParamNotFunc
	}

	// 检查函数参数必须为 func(Type) 或 func(string, Type),
	// 允许使用的type值定义在DefaultHandlerExtendAllowType。
	if (iType.NumIn() != 1) && (iType.NumIn() != 2 || iType.In(0).Kind() != reflect.String) {
		return fmt.Errorf(ErrFormatHandlerExtenderInputParamError, iType.String())
	}
	_, ok := DefaultHandlerExtenderAllowType[iType.In(iType.NumIn()-1).Kind()]
	if !ok {
		return fmt.Errorf(ErrFormatHandlerExtenderInputParamError, iType.String())
	}

	// 检查函数返回值必须是HandlerFunc
	if iType.NumOut() != 1 || iType.Out(0) != typeHandlerFunc {
		return fmt.Errorf(ErrFormatHandlerExtenderOutputParamError, iType.String())
	}

	he.NewType = append(he.NewType, iType.In(iType.NumIn()-1))
	he.NewFunc = append(he.NewFunc, reflect.ValueOf(fn))
	if iType.In(iType.NumIn()-1).Kind() == reflect.Interface {
		he.AnyType = append(he.AnyType, iType.In(iType.NumIn()-1))
		he.AnyFunc = append(he.AnyFunc, reflect.ValueOf(fn))
	}
	return nil
}

// CreateHandler 函数根据参数返回一个HandlerFuncs。
func (he *handlerExtenderBase) CreateHandler(path string, i any) HandlerFuncs {
	val, ok := i.(reflect.Value)
	if !ok {
		val = reflect.ValueOf(i)
	}
	return NewHandlerFuncsFilter(he.newHandlerFuncs(path, val))
}

func (he *handlerExtenderBase) newHandlerFuncs(path string, v reflect.Value) HandlerFuncs {
	// 基础类型返回
	switch fn := v.Interface().(type) {
	case func(Context):
		SetHandlerFuncName(fn, getHandlerAliasName(v))
		return HandlerFuncs{fn}
	case HandlerFunc:
		SetHandlerFuncName(fn, getHandlerAliasName(v))
		return HandlerFuncs{fn}
	case []HandlerFunc:
		return fn
	case HandlerFuncs:
		return fn
	}
	// 尝试转换成HandlerFuncs
	fn := he.newHandlerFunc(path, v)
	if fn != nil {
		return HandlerFuncs{fn}
	}
	// 解引用数组再转换HandlerFuncs
	switch v.Type().Kind() {
	case reflect.Slice, reflect.Array:
		var fns HandlerFuncs
		for i := 0; i < v.Len(); i++ {
			hs := he.newHandlerFuncs(path, v.Index(i))
			if hs != nil {
				fns = append(fns, hs...)
			}
		}
		if len(fns) != 0 {
			return fns
		}
	case reflect.Interface, reflect.Ptr:
		return he.newHandlerFuncs(path, v.Elem())
	}
	return nil
}

// The newHandlerFunc function converts a function or interface parameter into
// a request context handler function.
//
// The parameter must be a function, the function has a parameter as an input parameter,
// and a HandlerFunc object as a return value.
//
// First check whether the object has a directly registered type extension function,
// and then check whether the object implements the registered interface type.
//
// Multiple registrations are allowed, as long as the registration return value is not empty,
// the corresponding processing function will be returned.
//
// newHandlerFunc 函数使用一个函数或接口参数转换成请求上下文处理函数。
//
// 参数必须是一个函数，函数拥有一个参数作为入参，一个HandlerFunc对象作为返回值。
//
// 先检测对象是否拥有直接注册的类型扩展函数，再检查对象是否实现其中注册的接口类型。
//
// 允许进行多次注册，只要注册返回值不为空就会返回对应的处理函数。
func (he *handlerExtenderBase) newHandlerFunc(path string, v reflect.Value) HandlerFunc {
	iType := v.Type()
	for i := range he.NewType {
		if he.NewType[i] == iType {
			h := he.createHandlerFunc(path, he.NewFunc[i], v)
			if h != nil {
				return h
			}
		}
	}
	// 判断是否实现接口类型
	for i, iface := range he.AnyType {
		if iType.Implements(iface) {
			h := he.createHandlerFunc(path, he.AnyFunc[i], v)
			if h != nil {
				return h
			}
		}
	}
	return nil
}

// The createHandlerFunc function creates a HandlerFunc using the conversion function and object,
// and saves the name of the HandlerFunc and the name of the extended function used.
// createHandlerFunc 函数使用转换函数和对象创建一个HandlerFunc，并保存HandlerFunc的名称和使用的扩展函数名称。
func (he *handlerExtenderBase) createHandlerFunc(path string, fn, v reflect.Value) (h HandlerFunc) {
	if fn.Type().NumIn() == 1 {
		h = fn.Call([]reflect.Value{v})[0].Interface().(HandlerFunc)
	} else {
		h = fn.Call([]reflect.Value{reflect.ValueOf(path), v})[0].Interface().(HandlerFunc)
	}
	if h == nil {
		return nil
	}
	// 获取新函数名称,一般来源于函数扩展返回的函数名称。
	hptr := getFuncPointer(reflect.ValueOf(h))
	name := contextSaveName[hptr]
	// 使用原值名称
	if name == "" && v.Kind() != reflect.Struct {
		name = getHandlerAliasName(v)
	}
	// 推断名称
	if name == "" {
		iType := v.Type()
		switch iType.Kind() {
		case reflect.Func:
			name = runtime.FuncForPC(v.Pointer()).Name()
		case reflect.Ptr:
			iType = iType.Elem()
			name = fmt.Sprintf("*%s.%s", iType.PkgPath(), iType.Name())
		case reflect.Struct:
			name = fmt.Sprintf("%s.%s", iType.PkgPath(), iType.Name())
		default:
			name = "any"
		}
	}
	// 获取扩展名称，eudore包移除包前缀
	extname := strings.TrimPrefix(runtime.FuncForPC(fn.Pointer()).Name(), "github.com/eudore/eudore.")
	contextFuncName[hptr] = fmt.Sprintf("%s(%s)", name, extname)
	return h
}

var formarExtendername = "%s(%s)"

// The List method returns all registered function names.
//
// List 方法返回全部注册的函数名称。
func (he *handlerExtenderBase) List() []string {
	names := make([]string, 0, len(he.NewFunc))
	for i := range he.NewType {
		if he.NewType[i].Kind() != reflect.Interface {
			name := runtime.FuncForPC(he.NewFunc[i].Pointer()).Name()
			names = append(names, fmt.Sprintf(formarExtendername, name, he.NewType[i].String()))
		}
	}
	for i, iface := range he.AnyType {
		name := runtime.FuncForPC(he.AnyFunc[i].Pointer()).Name()
		names = append(names, fmt.Sprintf(formarExtendername, name, iface.String()))
	}
	return names
}

func (he *handlerExtenderBase) Metadata() any {
	return MetadataHandlerExtender{
		Health:   true,
		Name:     "eudore.handlerExtenderBase",
		Extender: he.List(),
	}
}

// NewHandlerExtenderWarp function creates a chained HandlerExtender object.
//
// All objects are registered and created using base. If base cannot create a function handler,
// use last to create a function handler.
//
// The NewHandlerExtenderWarp(base, last).RegisterExtender method
// uses the base object to register extension functions.
//
// The NewHandlerExtenderWarp(base, last).CreateHandler method first
// uses the base object to create multiple request processing functions.
// If it returns nil, it uses the last object to create multiple request processing functions.
//
// NewHandlerExtenderWarp 函数创建一个链式HandlerExtender对象。
//
// 所有对象注册和创建均使用base，如果base无法创建函数处理者则使用last创建函数处理者。
//
// NewHandlerExtenderWarp(base, last).RegisterExtender 方法使用base对象注册扩展函数。
//
// NewHandlerExtenderWarp(base, last).CreateHandler 方法先使用base对象创建多个请求处理函数，
// 如果返回nil，则使用last对象创建多个请求处理函数。
func NewHandlerExtenderWarp(base, last HandlerExtender) HandlerExtender {
	return &handlerExtenderWarp{
		data: base,
		last: last,
	}
}

// RegisterExtender 方法基于路径注册一个扩展函数。
func (he *handlerExtenderWarp) RegisterExtender(path string, i any) error {
	return he.data.RegisterExtender(path, i)
}

// The CreateHandler method implements the CreateHandler function.
// If the current HandlerExtender cannot create HandlerFuncs,
// it calls the superior HandlerExtender to process.
//
// CreateHandler 方法实现CreateHandler函数，如果当前HandlerExtender无法创建HandlerFuncs，
// 则调用上级HandlerExtender处理。
func (he *handlerExtenderWarp) CreateHandler(path string, i any) HandlerFuncs {
	hs := he.data.CreateHandler(path, i)
	if hs != nil {
		return hs
	}
	return he.last.CreateHandler(path, i)
}

// List 方法返回全部注册的函数名称。
func (he *handlerExtenderWarp) List() []string {
	return append(he.last.List(), he.data.List()...)
}

func (he *handlerExtenderWarp) Metadata() any {
	return MetadataHandlerExtender{
		Health:   true,
		Name:     "eudore.handlerExtenderWarp",
		Extender: he.List(),
	}
}

// NewHandlerExtenderTree function creates a path-based function extender.
//
// Mainly implement path matching. All actions are processed by the node's HandlerExtender,
// and the NewHandlerExtenderBase () object is used.
//
// All registration and creation actions will be performed by matching the lowest node of the tree.
// If it cannot be created, the tree nodes will be processed upwards in order.
//
// The NewHandlerExtenderTree().RegisterExtender method registers a handler function based on the path,
// and initializes to NewHandlerExtenderBase () if the HandlerExtender is empty.
//
// The NewHandlerExtenderTree().CreateHandler method matches the child nodes of the tree based on the path,
// and then executes the CreateHandler method from the most child node up.
// If it returns non-null, it returns directly.
//
// NewHandlerExtenderTree 函数创建一个基于路径的函数扩展者。
//
// 主要实现路径匹配，所有行为使用节点的HandlerExtender处理，使用NewHandlerExtenderBase()对象。
//
// 所有注册和创建行为都会匹配树最下级节点执行，如果无法创建则在树节点依次向上处理。
//
// NewHandlerExtenderTree().RegisterExtender 方法基于路径注册一个处理函数，
// 如果HandlerExtender为空则初始化为NewHandlerExtenderBase()。
//
// NewHandlerExtenderTree().CreateHandler 方法基于路径向树子节点匹配，
// 后从最子节点依次向上执行CreateHandler方法，如果返回非空直接返回，否在会依次执行注册行为。
func NewHandlerExtenderTree() HandlerExtender {
	return &handlerExtenderTree{}
}

// RegisterExtender 方法基于路径注册一个扩展函数。
func (he *handlerExtenderTree) RegisterExtender(path string, i any) error {
	// 匹配当前节点注册
	if path == "" {
		if he.data == nil {
			he.data = NewHandlerExtenderBase()
		}
		return he.data.RegisterExtender("", i)
	}

	// 寻找对应的子节点注册
	for pos := range he.childs {
		subStr, find := getSubsetPrefix(path, he.childs[pos].path)
		if find {
			if subStr != he.childs[pos].path {
				he.childs[pos].path = strings.TrimPrefix(he.childs[pos].path, subStr)
				he.childs[pos] = &handlerExtenderTree{
					path:   subStr,
					childs: []*handlerExtenderTree{he.childs[pos]},
				}
			}
			return he.childs[pos].RegisterExtender(strings.TrimPrefix(path, subStr), i)
		}
	}

	// 追加一个新的子节点
	newnode := &handlerExtenderTree{
		path: path,
		data: NewHandlerExtenderBase(),
	}
	he.childs = append(he.childs, newnode)
	return newnode.data.RegisterExtender(path, i)
}

// CreateHandler 函数基于路径创建多个对象处理函数。
//
// 递归依次寻找子节点，然后返回时创建多个对象处理函数，如果子节点返回不为空就直接返回。
func (he *handlerExtenderTree) CreateHandler(path string, data any) HandlerFuncs {
	for _, child := range he.childs {
		if strings.HasPrefix(path, child.path) {
			hs := child.CreateHandler(path[len(child.path):], data)
			if hs != nil {
				return hs
			}
			break
		}
	}

	if he.data != nil {
		return he.data.CreateHandler(path, data)
	}
	return nil
}

// The listExtendHandlerNamesByPrefix method recursively adds path prefixes
// and returns extension function names.
//
// listExtendHandlerNamesByPrefix 方法递归添加路径前缀返回扩展函数名称。
func (he *handlerExtenderTree) listExtendHandlerNamesByPrefix(prefix string) []string {
	prefix += he.path
	var names []string
	if he.data != nil {
		names = he.data.List()
		if prefix != "" {
			for i := range names {
				names[i] = prefix + " " + names[i]
			}
		}
	}

	for i := range he.childs {
		names = append(names, he.childs[i].listExtendHandlerNamesByPrefix(prefix)...)
	}
	return names
}

// List 方法返回全部注册的函数名称。
func (he *handlerExtenderTree) List() []string {
	return he.listExtendHandlerNamesByPrefix("")
}

func (he *handlerExtenderTree) Metadata() any {
	return MetadataHandlerExtender{
		Health:   true,
		Name:     "eudore.handlerExtenderTree",
		Extender: he.List(),
	}
}

func getFileLineFieldsVals(v reflect.Value) []any {
	file, line := runtime.FuncForPC(v.Pointer()).FileLine(1)
	return []any{file, line}
}

// NewHandlerFunc 函数处理func()。
func NewHandlerFunc(fn func()) HandlerFunc {
	return func(Context) {
		fn()
	}
}

// NewHandlerFuncContextError 函数处理func(Context) error返回的error处理。
func NewHandlerFuncContextError(fn func(Context) error) HandlerFunc {
	fineLineFieldsVals := getFileLineFieldsVals(reflect.ValueOf(fn))
	return func(ctx Context) {
		err := fn(ctx)
		if err != nil {
			ctx.WithFields(fineLineFieldsKeys, fineLineFieldsVals).Fatal(err)
		}
	}
}

// NewHandlerFuncContextAnyError 函数处理func(Context) (T, error)返回数据渲染和error处理。
func NewHandlerFuncContextAnyError(fn any) HandlerFunc {
	v := reflect.ValueOf(fn)
	iType := v.Type()
	if iType.Kind() != reflect.Func || iType.NumIn() != 1 || iType.NumOut() != 2 ||
		iType.In(0) != typeContext || iType.Out(1) != typeError {
		return nil
	}

	fineLineFieldsVals := getFileLineFieldsVals(reflect.ValueOf(fn))
	return func(ctx Context) {
		vals := v.Call([]reflect.Value{reflect.ValueOf(ctx)})
		err, _ := vals[1].Interface().(error)
		if err == nil && ctx.Response().Size() == 0 {
			err = ctx.Render(vals[0].Interface())
		}
		if err != nil {
			ctx.WithFields(fineLineFieldsKeys, fineLineFieldsVals).Fatal(err)
		}
	}
}

// NewHandlerFuncContextRender 函数处理func(Context) any返回数据渲染。
func NewHandlerFuncContextRender(fn func(Context) any) HandlerFunc {
	fineLineFieldsVals := getFileLineFieldsVals(reflect.ValueOf(fn))
	return func(ctx Context) {
		data := fn(ctx)
		if ctx.Response().Size() == 0 {
			err := ctx.Render(data)
			if err != nil {
				ctx.WithFields(fineLineFieldsKeys, fineLineFieldsVals).Fatal(err)
			}
		}
	}
}

// NewHandlerFuncContextRenderError 函数处理func(Context) (any, error)返回数据渲染和error处理。
func NewHandlerFuncContextRenderError(fn func(Context) (any, error)) HandlerFunc {
	fineLineFieldsVals := getFileLineFieldsVals(reflect.ValueOf(fn))
	return func(ctx Context) {
		data, err := fn(ctx)
		if err == nil && ctx.Response().Size() == 0 {
			err = ctx.Render(data)
		}
		if err != nil {
			ctx.WithFields(fineLineFieldsKeys, fineLineFieldsVals).Fatal(err)
		}
	}
}

// NewHandlerFuncError 函数处理func() error返回的error处理。
func NewHandlerFuncError(fn func() error) HandlerFunc {
	fineLineFieldsVals := getFileLineFieldsVals(reflect.ValueOf(fn))
	return func(ctx Context) {
		err := fn()
		if err != nil {
			ctx.WithFields(fineLineFieldsKeys, fineLineFieldsVals).Fatal(err)
		}
	}
}

// NewHandlerFuncRPC function needs to pass in a function that returns a request for
// processing and is dynamically called by reflection.
//
// Function form: func (Context, Request) (Response, error)
//
// The types of Request and Response can be map or struct or pointer to struct.
// All 4 parameters need to exist, and the order cannot be changed.
//
// NewHandlerFuncRPC 函数需要传入一个函数，返回一个请求处理，通过反射来动态调用。
//
// 函数形式： func(Context, Request) (Response, error)
//
// Request和Response的类型可以为map或结构体或者结构体的指针，4个参数需要全部存在，且不可调换顺序。
func NewHandlerFuncRPC(fn any) HandlerFunc {
	iType := reflect.TypeOf(fn)
	v := reflect.ValueOf(fn)
	if iType.Kind() != reflect.Func {
		return nil
	}
	if iType.NumIn() != 2 || iType.In(0) != typeContext {
		return nil
	}
	if iType.NumOut() != 2 || iType.Out(1) != typeError {
		return nil
	}
	typeIn := iType.In(1)
	kindIn := typeIn.Kind()
	typenew := iType.In(1)
	// 检查请求类型
	switch typeIn.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Struct:
	default:
		return nil
	}
	if typenew.Kind() == reflect.Ptr {
		typenew = typenew.Elem()
	}

	fineLineFieldsVals := getFileLineFieldsVals(v)
	return func(ctx Context) {
		// 创建请求参数并初始化
		req := reflect.New(typenew)
		err := ctx.Bind(req.Interface())
		if err != nil {
			ctx.Fatal(err)
			return
		}
		if kindIn != reflect.Ptr {
			req = req.Elem()
		}

		// 反射调用执行函数。
		vals := v.Call([]reflect.Value{reflect.ValueOf(ctx), req})

		// 检查函数执行err。
		err, ok := vals[1].Interface().(error)
		if ok {
			ctx.WithFields(fineLineFieldsKeys, fineLineFieldsVals).Fatal(err)
			return
		}

		// 渲染返回的数据。
		err = ctx.Render(vals[0].Interface())
		if err != nil {
			ctx.Fatal(err)
		}
	}
}

// NewHandlerFuncRPCMap defines a fixed request and response to function processing of
// type map [string] interface {}.
//
// is a subset of NewRPCHandlerFunc and has type restrictions,
// but using map [string] interface {} to save requests does not use reflection.
//
// NewHandlerFuncRPCMap 定义了固定请求和响应为map[string]any类型的函数处理。
//
// 是NewRPCHandlerFunc的一种子集，拥有类型限制，但是使用map[string]any保存请求没用使用反射。
func NewHandlerFuncRPCMap(fn func(Context, map[string]any) (any, error)) HandlerFunc {
	fineLineFieldsVals := getFileLineFieldsVals(reflect.ValueOf(fn))
	return func(ctx Context) {
		req := make(map[string]any)
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
			ctx.WithFields(fineLineFieldsKeys, fineLineFieldsVals).Fatal(err)
		}
	}
}

// NewHandlerFuncRender 函数处理func() any。
func NewHandlerFuncRender(fn func() any) HandlerFunc {
	fineLineFieldsVals := getFileLineFieldsVals(reflect.ValueOf(fn))
	return func(ctx Context) {
		data := fn()
		if ctx.Response().Size() == 0 {
			err := ctx.Render(data)
			if err != nil {
				ctx.WithFields(fineLineFieldsKeys, fineLineFieldsVals).Fatal(err)
			}
		}
	}
}

// NewHandlerFuncRenderError 函数处理func() (any, error)返回数据渲染和error处理。
func NewHandlerFuncRenderError(fn func() (any, error)) HandlerFunc {
	fineLineFieldsVals := getFileLineFieldsVals(reflect.ValueOf(fn))
	return func(ctx Context) {
		data, err := fn()
		if err == nil && ctx.Response().Size() == 0 {
			err = ctx.Render(data)
		}
		if err != nil {
			ctx.WithFields(fineLineFieldsKeys, fineLineFieldsVals).Fatal(err)
		}
	}
}

// NewHandlerFuncString 函数处理func() string，然后指定函数生成的字符串。
func NewHandlerFuncString(fn func() string) HandlerFunc {
	return func(ctx Context) {
		ctx.WriteString(fn())
	}
}

type handlerHTTP interface {
	HandleHTTP(Context)
}

// NewHandlerHTTP 函数handlerHTTP接口转换成HandlerFunc。
func NewHandlerHTTP(h handlerHTTP) HandlerFunc {
	return h.HandleHTTP
}

// NewHandlerHTTPFunc1 函数转换处理http.HandlerFunc类型。
func NewHandlerHTTPFunc1(fn http.HandlerFunc) HandlerFunc {
	return func(ctx Context) {
		fn(ctx.Response(), ctx.Request())
	}
}

// NewHandlerHTTPFunc2 函数转换处理func(http.ResponseWriter, *http.Request)类型。
func NewHandlerHTTPFunc2(fn func(http.ResponseWriter, *http.Request)) HandlerFunc {
	return func(ctx Context) {
		fn(ctx.Response(), ctx.Request())
	}
}

// NewHandlerNetHTTP 函数转换处理http.Handler对象。
func NewHandlerHTTPHandler(h http.Handler) HandlerFunc {
	clone, ok := h.(interface{ CloneHandler() http.Handler })
	if ok {
		h = clone.CloneHandler()
	}
	return func(ctx Context) {
		h.ServeHTTP(ctx.Response(), ctx.Request())
	}
}

// NewHandlerStringer 函数处理fmt.Stringer接口类型转换成HandlerFunc。
func NewHandlerStringer(fn fmt.Stringer) HandlerFunc {
	return func(ctx Context) {
		ctx.WriteString(fn.String())
	}
}
