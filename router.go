package eudore

// Router对象用于定义请求的路由器

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
)

/*
Router interface is divided into RouterCore and RouterMethod. RouterCore implements router matching algorithm and logic, and RouterMethod implements the encapsulation of routing rule registration.

RouterCore implements route matching details. RouterMethod calls RouterCore to provide methods for external use.

RouterMethod The default directly registered interface of the route. Set the routing parameters, group routing, middleware, function extensions, controllers and other behaviors.

Do not use the RouterCore method to register routes directly at any time. You should use the Add ... method of RouterMethod.

RouterMethod implements the following functions:
    Group routing
    The middleware or function extension is registered in the local scope/global scope
    Add controller
    Display routing registration debug information

RouterCore has four router cores to implement the following functions:
    High performance (70%-90% of httprouter performance, using less memory)
    Low code complexity (RouterCoreStd supports 5 levels of priority, a code complexity of 19 is not satisfied)
    Request for additional default parameters (including current routing matching rules)
    Extend custom routing methods
    Variable and wildcard matching
    Matching priority Constant > Variable verification > Variable > Wildcard verification > Wildcard (RouterCoreStd five-level priority)
    Method priority Specify method > Any method (The specified method will override the Any method, and vice versa)
    Variables and wildcards support regular and custom functions to verify data
    Variables and wildcards support constant prefix
    Get all registered routing rule information (RouterCoreBebug implementation)
    Routing rule matching based on Host (implemented by RouterCoreHost)
    Allows dynamic addition and deletion of router rules at runtime (RouterCoreStd implementation, the outer layer requires RouterCoreLock packaging layer)

Router 接口分为RouterCore和RouterMethod，RouterCore实现路由器匹配算法和逻辑，RouterMethod实现路由规则注册的封装。

RouterCore实现路由匹配细节，RouterMethod调用RouterCore提供对外使用的方法。

RouterMethod 路由默认直接注册的接口，设置路由参数、组路由、中间件、函数扩展、控制器等行为。

任何时候请不要使用RouterCore的方法直接注册路由，应该使用RouterMethod的Add...方法。

RouterMethod实现下列功能：
    组路由
    中间件或函数扩展注册在局部作用域/全局作用域
    添加控制器
    显示路由注册debug信息

RouterCore拥有四种路由器核心实现下列功能：
    高性能(httprouter性能的70%-90%，使用更少的内存)
    低代码复杂度(RouterCoreStd支持5级优先级 一处代码复杂度19不满足)
    请求获取额外的默认参数(包含当前路由匹配规则)
    扩展自定义路由方法
    变量和通配符匹配
    匹配优先级 常量 > 变量校验 > 变量 > 通配符校验 > 通配符(RouterCoreStd五级优先级)
    方法优先级 指定方法 > Any方法(指定方法会覆盖Any方法，反之不行)
    变量和通配符支持正则和自定义函数进行校验数据
    变量和通配符支持常量前缀
    获取注册的全部路由规则信息(RouterCoreBebug实现)
    基于Host进行路由规则匹配(RouterCoreHost实现)
    允许运行时进行动态增删路由器规则(RouterCoreStd实现，外层需要RouterCoreLock包装一层)
*/
type Router interface {
	RouterCore
	// RouterMethod method
	Group(string) Router
	Params() *Params
	AddHandler(string, string, ...interface{}) error
	AddController(...Controller) error
	AddMiddleware(...interface{}) error
	AddHandlerExtend(...interface{}) error
	AnyFunc(string, ...interface{})
	GetFunc(string, ...interface{})
	PostFunc(string, ...interface{})
	PutFunc(string, ...interface{})
	DeleteFunc(string, ...interface{})
	HeadFunc(string, ...interface{})
	PatchFunc(string, ...interface{})
}

// The RouterCore interface performs registration of the route and matches a request and returns the handler.
//
// RouterCore mainly implements routing matching related details.
//
// RouterCore接口，执行路由的注册和匹配一个请求并返回处理者。
//
// RouterCore主要实现路由匹配相关细节。
type RouterCore interface {
	HandleFunc(string, string, HandlerFuncs)
	Match(string, string, *Params) HandlerFuncs
}

// RouterStd default router registration implementation.
//
// Need to specify a routing core, the default handler function extension is DefaultHandlerExtend.
// As a public attribute, it is only used by godoc to display the documentation of the relevant method.
//
// RouterStd 默认路由器注册实现。
//
// 需要指定一个路由核心，处理函数扩展者默认为DefaultHandlerExtend。
// 作为公开属性仅用于godoc展示相关方法文档说明。
type RouterStd struct {
	RouterCore      `alias:"routercore"`
	HandlerExtender `alias:"handlerextender"`
	Middlewares     *middlewareTree      `alias:"middlewares"`
	Print           func(...interface{}) `alias:"print"`
	params          Params               `alias:"params"`
}

// HandlerRouter405 function defines the default 405 processing and returns Allow and X-Match-Route Header.
//
// HandlerRouter405 函数定义默认405处理,返回Allow和X-Match-Route Header。
func HandlerRouter405(ctx Context) {
	const page405 string = "405 method not allowed"
	ctx.SetHeader(HeaderAllow, ctx.GetParam(ParamAllow))
	ctx.SetHeader(HeaderXMatchRoute, ctx.GetParam(ParamRoute))
	ctx.WriteHeader(405)
	ctx.Render(page405)
}

// HandlerRouter404 function defines the default 404 processing.
//
// HandlerRouter404 函数定义默认404处理。
func HandlerRouter404(ctx Context) {
	const page404 string = "404 page not found"
	ctx.WriteHeader(404)
	ctx.Render(page404)
}

// NewRouterStd method uses a RouterCore to create a Router object.
//
// RouterStd implements RouterMethod interface registration related details, and routing matching is implemented by RouterCore.
//
// NewRouterStd 方法使用一个RouterCore创建Router对象。
//
// RouterStd实现RouterMethod接口注册相关细节，路由匹配由RouterCore实现。
func NewRouterStd(core RouterCore) Router {
	if core == nil {
		core = NewRouterCoreStd()
	}
	return &RouterStd{
		RouterCore:      core,
		params:          Params{ParamRoute, ""},
		HandlerExtender: NewHandlerExtendWarp(NewHandlerExtendTree(), DefaultHandlerExtend),
		Middlewares:     newMiddlewareTree(),
		Print:           printEmpty,
	}
}

// Mount 方法使RouterStd挂载上下文，上下文传递给RouterCore。
//
// 并从ctx.Value(ContextKeyApp)获取Logger，初始化RouterStd日志输出函数。
func (r *RouterStd) Mount(ctx context.Context) {
	r.Print = NewPrintFunc(ctx.Value(ContextKeyApp).(Logger))
	withMount(ctx, r.RouterCore)
}

// Unmount 方法使RouterStd卸载上下文，上下文传递给RouterCore。
func (r *RouterStd) Unmount(ctx context.Context) {
	withUnmount(ctx, r.RouterCore)
	r.Print = printEmpty
}

// Metadata 方法返回RouterCore的Metadata。
func (r *RouterStd) Metadata() interface{} {
	return withMetadata(r.RouterCore)
}

// Group method returns a new group router.
//
// The parameters, middleware, and function extensions of each Group group route registration will not affect the superior, but the subordinate will inherit the superior data.
//
// The new Router will use the old RouterCore and Print objects; the middleware information and routing parameters are deep copied from the superior, while processing the Group parameters.
//
// And create a new HandlerExtender in chain, if the type that HandlerExtender cannot register will call the previous Router.HandlerExtender to process.
//
// The top-level HandlerExtender object is defaultHandlerExtend. You can use the RegisterHandlerExtend function and the NewHandlerFuncs function to call the defaultHandlerExtend object.
//
// Group 方法返回一个新的组路由器。
//
// 每个Group组路由注册的参数、中间件、函数扩展都不会影响上级，但是下级会继承上级数据。
//
// 新的Router将使用旧的RouterCore和Print对象；中间件信息和路由参数从上级深拷贝一份，同时处理Group参数。
//
// 以及链式创建一个新的HandlerExtender，若HandlerExtender无法注册的类型将调用上一个Router.HandlerExtender处理。
//
// 最顶级HandlerExtender对象为defaultHandlerExtend，可以使用RegisterHandlerExtend函数和NewHandlerFuncs函数调用defaultHandlerExtend对象。
func (r *RouterStd) Group(path string) Router {
	// 构建新的路由方法配置器
	return &RouterStd{
		RouterCore:      r.RouterCore,
		params:          r.params.Clone().CombineWithRoute(NewParamsRoute(path)),
		HandlerExtender: NewHandlerExtendWarp(NewHandlerExtendTree(), r.HandlerExtender),
		Middlewares:     r.Middlewares.clone(),
		Print:           r.Print,
	}
}

// Params method returns the current route parameters, and the route parameter value is an empty string will not be used.
//
// Params 方法返回当前路由参数，路由参数值为空字符串不会被使用。
func (r *RouterStd) Params() *Params {
	return &r.params
}

// printError 方法输出一个err，附加错误的函数名称和文件位置。
func (r *RouterStd) printError(depth int, err error) {
	name, file, line := logFormatNameFileLine(depth + 3)
	r.Print([]string{"params", "func", "file", "line"}, []interface{}{r.params, name, file, line}, err)
}

// printPanic 方法输出一个err，附加当前stack。
func (r *RouterStd) printPanic(err error) {
	r.Print([]string{"params", "stack"}, []interface{}{r.params, GetPanicStack(4)}, err)
}

// getRoutePath 函数截取到路径中的route，支持'{}'进行块匹配。
func getRoutePath(path string) string {
	var depth = 0
	var str = ""
	for i := range path {
		switch path[i] {
		case '{':
			depth++
		case '}':
			depth--
		case ' ':
			if depth == 0 {
				return str
			}
		}
		str += path[i : i+1]
	}
	return path
}

// getRouteParam 函数截取到路径中的指定参数，支持对route部分使用'{}'进行块匹配。
func getRouteParam(path, key string) string {
	key += "="
	for _, i := range strings.Split(path[len(getRoutePath(path)):], " ") {
		if strings.HasPrefix(i, key) {
			return i[len(key):]
		}
	}
	return ""
}

// AddHandler method adds a new route, allowing multiple request methods to be added separately using','.
//
// You can register 9 methods defined by http (three of the Router interfaces do not provide direct registration),
// or you can register the method as: ANY TEST 404 405 NotFound MethodNotAllowed, register Any, TEST, 404, 405 routing rules.
// the registration method is ANY to register all methods, the ANY method route will be covered by the same path non-ANY method,
// and vice versa; the registration method is TEST will output the debug information related to the route registration,
// but the registration behavior will not be performed;
//
// The handler parameter is processed using the HandlerExtender.NewHandlerFuncs() method of the current RouterStd to generate the corresponding HandlerFuncs.
//
// If the current Router cannot be processed, call the HandlerExtender or defaultHandlerExtend of the upper-level group for processing,
// and output the error log if all of them cannot be processed.
//
// The middleware data will be matched from the data according to the current routing path, and then the request processing function will be appended before the processing function.
//
// AddHandler 方法添加一条新路由, 允许添加多个请求方法使用','分开。
//
// 可以注册http定义的9种方法(其中三种Router接口未提供直接注册),也可以注册方法为：ANY TEST 404 405 NotFound MethodNotAllowed，注册Any、TEST、404、405路由规则。注册方法为ANY注册全部方法，ANY方法路由会被同路径非ANY方法覆盖，反之不行；注册方法为TEST会输出路由注册相关debug信息，但不执行注册行为;
//
// handler参数使用当前RouterStd的HandlerExtender.NewHandlerFuncs()方法处理，生成对应的HandlerFuncs。
//
// 如果当前Router无法处理，则调用上一级group的HandlerExtender或defaultHandlerExtend处理，全部无法处理则输出error日志。
//
// 中间件数据会根据当前路由路径从数据中匹配，然后将请求处理函数附加到处理函数之前。
//
func (r *RouterStd) AddHandler(method, path string, hs ...interface{}) error {
	return r.registerHandlers(method, path, hs...)
}

// registerHandlers 方法将handler转换成HandlerFuncs，添加路由路径对应的请求中间件，并调用RouterCore对象注册路由方法。
func (r *RouterStd) registerHandlers(method, path string, hs ...interface{}) (err error) {
	defer func() {
		// RouterCoreStd 注册未知校验规则存在panic,或者其他自定义路由注册出现panic。
		if rerr := recover(); rerr != nil {
			err = fmt.Errorf(ErrFormatRouterStdRegisterHandlersRecover, method, path, rerr)
			r.printPanic(err)
		}
	}()

	params := r.params.Clone().CombineWithRoute(NewParamsRoute(path))
	path = params.Get("route")
	fullpath := params.String()
	// 如果方法为404、405方法，route为空
	if len(fullpath) > 6 && fullpath[:6] == "route=" {
		fullpath = fullpath[6:]
	}
	method = strings.ToUpper(method)

	handlers, err := r.newHandlerFuncs(path, hs)
	if err != nil {
		r.printError(1, err)
		return err
	}
	// 如果注册方法是TEST则输出RouterStd debug信息
	if method == "TEST" {
		r.Print(fmt.Sprintf("Test handlers params is %s, split path to: ['%s'], match middlewares is: %v, register handlers is: %v.", params.String(), strings.Join(getSplitPath(path), "', '"), r.Middlewares.Lookup(path), handlers))
		return
	}
	r.Print("Register handler:", method, strings.TrimPrefix(params.String(), "route="), handlers)
	if handlers != nil {
		handlers = NewHandlerFuncsCombine(r.Middlewares.Lookup(path), handlers)
	}

	// 处理多方法
	var errs errormulit
	for _, method := range strings.Split(method, ",") {
		method = strings.TrimSpace(method)
		if checkMethod(method) {
			r.RouterCore.HandleFunc(method, fullpath, handlers)
		} else {
			err := fmt.Errorf(ErrFormatRouterStdRegisterHandlersMethodInvalid, method, method, fullpath)
			errs.HandleError(err)
			r.printError(1, err)
		}
	}
	return errs.Unwrap()
}

// The newHandlerFuncs method creates HandlerFuncs based on the path and multiple parameters.
//
// RouterStd first calls the current HandlerExtender.NewHandlerFuncs to create multiple function handlers. If it returns null, it will be created from the superior HandlerExtender.
//
// newHandlerFuncs 方法根据路径和多个参数创建HandlerFuncs。
//
// RouterStd先调用当前HandlerExtender.NewHandlerFuncs创建多个函数处理者，如果返回空会从上级HandlerExtender创建。
func (r *RouterStd) newHandlerFuncs(path string, handlers []interface{}) (HandlerFuncs, error) {
	var hs HandlerFuncs
	var errs errormulit
	// 转换处理函数
	for i, fn := range handlers {
		handler := r.HandlerExtender.NewHandlerFuncs(path, fn)
		if handler != nil && len(handler) > 0 {
			hs = NewHandlerFuncsCombine(hs, handler)
		} else {
			errs.HandleError(fmt.Errorf(ErrFormatRouterStdNewHandlerFuncsUnregisterType, path, i, reflect.TypeOf(fn).String()))
		}
	}
	return hs, errs.Unwrap()
}

func checkMethod(method string) bool {
	switch method {
	case "ANY", "404", "405", "NotFound", "MethodNotAllowed":
		return true
	}
	for _, allMethod := range RouterAllMethod {
		if allMethod == method {
			return true
		}
	}
	return false
}

// AddController method uses the built-in controller parsing function to resolve the controller to obtain the routing configuration.
//
// If the controller implements the RoutesInjecter interface, call the controller to inject the route itself.
//
// AddController 方式使用内置的控制器解析函数解析控制器获得路由配置。
//
// 如果控制器实现了RoutesInjecter接口，调用控制器自身注入路由。
func (r *RouterStd) AddController(controllers ...Controller) error {
	var errs errormulit
	for _, controller := range controllers {
		name := getControllerPathName(controller)
		r.Print("Register controller:", r.params.String(), name)
		err := controller.Inject(controller, r)
		if err != nil {
			err = fmt.Errorf(ErrFormatRouterStdAddController, name, err)
			errs.HandleError(err)
			r.printError(0, err)
		}
	}
	return errs.Unwrap()
}

// getControllerPathName 函数获取控制器的名称
func getControllerPathName(ctl Controller) string {
	ster, ok := ctl.(controllerName)
	if ok {
		return ster.ControllerName()
	}
	cType := reflect.Indirect(reflect.ValueOf(ctl)).Type()
	return fmt.Sprintf("%s.%s", cType.PkgPath(), cType.Name())
}

// AddMiddleware adds multiple middleware functions to the router, which will use HandlerExtender to convert parameters.
//
// If the number of parameters is greater than 1 and the first parameter is a string type, the first string type parameter is used as the path to add the middleware.
//
// AddMiddleware 给路由器添加多个中间件函数，会使用HandlerExtender转换参数。
//
// 如果参数数量大于1且第一个参数为字符串类型，会将第一个字符串类型参数作为添加中间件的路径。
func (r *RouterStd) AddMiddleware(hs ...interface{}) error {
	if len(hs) == 0 {
		return nil
	}

	path := r.params.Get("route")
	if len(hs) > 1 {
		route, ok := hs[0].(string)
		if ok {
			path = path + route
			hs = hs[1:]
		}
	}

	handlers, err := r.newHandlerFuncs(path, hs)
	if err != nil {
		r.printError(0, err)
		return err
	}

	r.Middlewares.Insert(path, handlers)
	r.RouterCore.HandleFunc("Middlewares", path, handlers)
	r.Print("Register middleware:", path, handlers)
	return nil
}

// AddHandlerExtend method adds an extension function to the current Router.
//
// If the number of parameters is greater than 1 and the first parameter is a string type, the first string type parameter is used as the path to add the extension function.
//
// AddHandlerExtend 方法给当前Router添加扩展函数。
//
// 如果参数数量大于1且第一个参数为字符串类型，会将第一个字符串类型参数作为添加扩展函数的路径。
func (r *RouterStd) AddHandlerExtend(handlers ...interface{}) error {
	if len(handlers) == 0 {
		return nil
	}

	path := r.params.Get("route")
	if len(handlers) > 1 {
		route, ok := handlers[0].(string)
		if ok {
			path = path + route
			handlers = handlers[1:]
		}
	}

	var errs errormulit
	for _, handler := range handlers {
		err := r.HandlerExtender.RegisterHandlerExtend(path, handler)
		if err != nil {
			err = fmt.Errorf(ErrFormatRouterStdAddHandlerExtend, path, err)
			errs.HandleError(err)
			r.printError(0, err)
		} else {
			iValue := reflect.ValueOf(handler)
			if iValue.Kind() == reflect.Func {
				r.Print("Register extend:", runtime.FuncForPC(iValue.Pointer()).Name(), iValue.Type().In(0).String())
			}
		}
	}
	return errs.Unwrap()
}

// AnyFunc method realizes the http request processing function that registers an Any method.
//
// The routing rules registered by the Any method will be overwritten by the specified method registration, and vice versa.
// Any default registration method includes six types of Get Post Put Delete Head Patch,
// which are defined in the global variable RouterAnyMethod.
//
// AnyFunc 方法实现注册一个Any方法的http请求处理函数。
//
// Any方法注册的路由规则会被指定方法注册覆盖，反之不行。
// Any默认注册方法包含Get Post Put Delete Head Patch六种，定义在全局变量RouterAnyMethod。
func (r *RouterStd) AnyFunc(path string, h ...interface{}) {
	r.registerHandlers(MethodAny, path, h...)
}

// GetFunc 方法实现注册一个Get方法的http请求处理函数。
func (r *RouterStd) GetFunc(path string, h ...interface{}) {
	r.registerHandlers(MethodGet, path, h...)
}

// PostFunc 方法实现注册一个Post方法的http请求处理函数。
func (r *RouterStd) PostFunc(path string, h ...interface{}) {
	r.registerHandlers(MethodPost, path, h...)
}

// PutFunc 方法实现注册一个Put方法的http请求处理函数。
func (r *RouterStd) PutFunc(path string, h ...interface{}) {
	r.registerHandlers(MethodPut, path, h...)
}

// DeleteFunc 方法实现注册一个Delete方法的http请求处理函数。
func (r *RouterStd) DeleteFunc(path string, h ...interface{}) {
	r.registerHandlers(MethodDelete, path, h...)
}

// HeadFunc 方法实现注册一个Head方法的http请求处理函数。
func (r *RouterStd) HeadFunc(path string, h ...interface{}) {
	r.registerHandlers(MethodHead, path, h...)
}

// PatchFunc 方法实现注册一个Patch方法的http请求处理函数。
func (r *RouterStd) PatchFunc(path string, h ...interface{}) {
	r.registerHandlers(MethodPatch, path, h...)
}

// middlewareTree 定义中间件信息存储树
type middlewareTree struct {
	index int
	node  *middlewareNode
}

func newMiddlewareTree() *middlewareTree {
	return &middlewareTree{node: new(middlewareNode)}
}

func (t *middlewareTree) Insert(path string, val HandlerFuncs) {
	t.index++
	indexs := make([]int, len(val))
	for i := range indexs {
		indexs[i] = t.index
	}
	t.node.Insert(path, indexs, val)
}

// Lookup 方法查找路径对应的处理函数，并安装索引进行排序。
func (t *middlewareTree) Lookup(path string) HandlerFuncs {
	indexs, vals := t.node.Lookup(path)
	length := len(vals)
	for i := 0; i < length; i++ {
		for j := i; j < length; j++ {
			if indexs[i] > indexs[j] {
				indexs[i], indexs[j] = indexs[j], indexs[i]
				vals[i], vals[j] = vals[j], vals[i]
			}
		}
	}
	return vals
}

func (t *middlewareTree) clone() *middlewareTree {
	return &middlewareTree{
		index: t.index,
		node:  t.node.clone(),
	}
}

// middlewareNode 存储中间件信息的前缀树。
//
// 用于内存存储路由器中间件注册信息，并根据注册路由返回对应的中间件。
type middlewareNode struct {
	path   string
	vals   HandlerFuncs
	indexs []int
	childs []*middlewareNode
}

// Insert 方法实现middlewareNode添加一个子节点。
func (t *middlewareNode) Insert(path string, indexs []int, vals HandlerFuncs) {
	if path == "" {
		t.indexs = indexsCombine(t.indexs, indexs)
		t.vals = NewHandlerFuncsCombine(t.vals, vals)
		return
	}
	for i := range t.childs {
		subStr, find := getSubsetPrefix(path, t.childs[i].path)
		if find {
			if subStr != t.childs[i].path {
				t.childs[i].path = strings.TrimPrefix(t.childs[i].path, subStr)
				t.childs[i] = &middlewareNode{
					path:   subStr,
					childs: []*middlewareNode{t.childs[i]},
				}
			}
			t.childs[i].Insert(path[len(subStr):], indexs, vals)
			return
		}
	}
	t.childs = append(t.childs, &middlewareNode{path: path, indexs: indexs, vals: vals})
}

// Lookup Find if seachKey exist in current trie tree and return its value
func (t *middlewareNode) Lookup(path string) ([]int, HandlerFuncs) {
	for _, i := range t.childs {
		if strings.HasPrefix(path, i.path) {
			indexs, val := i.Lookup(path[len(i.path):])
			return indexsCombine(t.indexs, indexs), NewHandlerFuncsCombine(t.vals, val)
		}
	}
	return t.indexs, t.vals
}

// clone 方法深拷贝这个中间件存储节点
func (t *middlewareNode) clone() *middlewareNode {
	nt := *t
	for i := range nt.childs {
		nt.childs[i] = nt.childs[i].clone()
	}
	return &nt
}

// indexsCombine 函数合并两个int切片
func indexsCombine(hs1, hs2 []int) []int {
	// if nil
	if len(hs1) == 0 {
		return hs2
	}
	hs := make([]int, len(hs1)+len(hs2))
	copy(hs, hs1)
	copy(hs[len(hs1):], hs2)
	return hs
}

// routerCoreLock allows reading and writing of RouterCore to be locked, which is used to dynamically add and delete routing rules at runtime.
//
// routerCoreLock 允许对RouterCore读写进行加锁，用于运行时动态增删路由规则。
type routerCoreLock struct {
	sync.RWMutex
	RouterCore
}

// NewRouterCoreLock function creates a router core with a read-write lock,
// and other router cores use the Lock core package when they need to dynamically modify the rules.
//
// NewRouterCoreLock 函数创建一个带读写锁的路由器核心，其他路由器核心在需要动态修改规则时使用Lock核心包装。
func NewRouterCoreLock(core RouterCore) RouterCore {
	if core == nil {
		core = NewRouterCoreStd()
	}
	return &routerCoreLock{RouterCore: core}
}

// The HandleFunc method adds a write lock to the router core to register routing rules, and defer prevents panic from being unable to unlock.
//
// HandleFunc 方法对路由器核心加写锁进行注册路由规则, defer 防止panic导致无法解锁。
func (r *routerCoreLock) HandleFunc(method, path string, hs HandlerFuncs) {
	r.Lock()
	defer r.Unlock()
	r.RouterCore.HandleFunc(method, path, hs)
}

// Match 方法对路由器加读锁进行匹配请求。
func (r *routerCoreLock) Match(method, path string, params *Params) (hs HandlerFuncs) {
	r.RLock()
	hs = r.RouterCore.Match(method, path, params)
	r.RUnlock()
	return
}

// routerCoreDebug 定义debug路由器。
type routerCoreDebug struct {
	RouterCore   `json:"-" xml:"-"`
	Methods      []string   `json:"methods" xml:"methods"`
	Paths        []string   `json:"paths" xml:"paths"`
	Params       []Params   `json:"params" xml:"params"`
	HandlerNames [][]string `json:"handlernames" xml:"handlernames"`
}

type routerCoreMetadata struct {
	Health       bool       `json:"health" xml:"health"`
	Name         string     `json:"name" xml:"name"`
	Methods      []string   `json:"methods" xml:"methods"`
	Paths        []string   `json:"paths" xml:"paths"`
	Params       []Params   `json:"params" xml:"params"`
	HandlerNames [][]string `json:"handlernames" xml:"handlernames"`
}

// NewRouterCoreDebug function specifies the routing core to create a debug core, using eudore.RouterCoreStd as the core by default.
//
// Visit GET /eudore/debug/router/data to get router registration information.
//
// NewRouterCoreDebug 函数指定路由核心创建一个debug核心,默认使用eudore.RouterCoreStd为核心。
//
// 访问 GET /eudore/debug/router/data 可以获取路由器注册信息。
func NewRouterCoreDebug(core RouterCore) RouterCore {
	if core == nil {
		core = NewRouterCoreStd()
	}
	r := &routerCoreDebug{
		RouterCore: core,
	}
	r.HandleFunc("GET", "/eudore/debug/router/data", HandlerFuncs{r.HandleHTTP})
	return r
}

// HandleFunc implements the eudore.RouterCore interface and records all routing information.
//
// HandleFunc 实现eudore.RouterCore接口，记录全部路由信息。
func (r *routerCoreDebug) HandleFunc(method, path string, handlers HandlerFuncs) {
	r.RouterCore.HandleFunc(method, path, handlers)
	// 删除记录的路由信息
	if getRouteParam(path, ParamRegister) == "off" || handlers == nil {
		path = getRoutePath(path)
		for i := range r.Methods {
			if r.Paths[i] == path && r.Methods[i] == method {
				r.Methods = r.Methods[:i+copy(r.Methods[i:], r.Methods[i+1:])]
				r.Paths = r.Paths[:i+copy(r.Paths[i:], r.Paths[i+1:])]
				r.Params = r.Params[:i+copy(r.Params[i:], r.Params[i+1:])]
				r.HandlerNames = r.HandlerNames[:i+copy(r.HandlerNames[i:], r.HandlerNames[i+1:])]
				break
			}
		}
		return
	}

	names := make([]string, len(handlers))
	for i := range handlers {
		names[i] = fmt.Sprint(handlers[i])
	}
	r.Methods = append(r.Methods, method)
	r.Paths = append(r.Paths, getRoutePath(path))
	r.Params = append(r.Params, NewParamsRoute(path))
	r.HandlerNames = append(r.HandlerNames, names)
}

// Metadata 方法返回routerCoreDebug记录的路由信息
func (r *routerCoreDebug) Metadata() interface{} {
	return routerCoreMetadata{
		Health:       true,
		Name:         "eudore.routerCoreDebug",
		Methods:      r.Methods,
		Paths:        r.Paths,
		Params:       r.Params,
		HandlerNames: r.HandlerNames,
	}
}

// HandleHTTP 方法返回debug路由信息数据。
func (r *routerCoreDebug) HandleHTTP(ctx Context) {
	ctx.SetHeader(HeaderXEudoreAdmin, "router-debug")
	ctx.Render(r.Metadata())
}

// routerCoreHost 实现基于host进行路由匹配
type routerCoreHost struct {
	routertree   routerHostNode
	routers      map[string]RouterCore
	newRouteCore func(string) RouterCore
}

// NewRouterCoreHost function creates a Host routing core, and a function that creates a routing core based on the host value needs to be given.
//
// If the parameter is empty, each route Host will create NewRouterCoreStd by default.
//
// NewRouterCoreHost 函数创建一个Host路由核心，需要给定一个根据host值创建路由核心的函数。
//
// 如果参数为空默认每个路由Host都创建NewRouterCoreStd。
func NewRouterCoreHost(fn func(string) RouterCore) RouterCore {
	if fn == nil {
		fn = func(string) RouterCore {
			return NewRouterCoreStd()
		}
	}
	r := &routerCoreHost{
		newRouteCore: fn,
		routers:      make(map[string]RouterCore),
	}
	r.getRouterCore("*")
	return r
}

// The HandleFunc method looks for the host parameter from the path to select the router registration match
//
// The host value is a host mode, and * is allowed, which means any character from the current to the next'.' or the end.
//
// If the host value is'*', the registration will be added to all current router cores.
// If the host value is empty and registered to the router core of'*', multiple hosts are allowed to use',' to divide the registration to multiple hosts at once.
//
// HandleFunc 方法从path中寻找host参数选择路由器注册匹配
//
// host值为一个host模式，允许存在*，表示当前任意字符到下一个'.'或结尾。
//
// 如果host值为'*'将注册添加给当前全部路由器核心，如果host值为空注册给'*'的路由器核心，允许多个host使用','分割一次注册给多host。
func (r *routerCoreHost) HandleFunc(method, path string, hs HandlerFuncs) {
	host := getRouteParam(path, "host")
	switch host {
	case "*":
		for _, core := range r.routers {
			core.HandleFunc(method, path, hs)
		}
	case "":
		r.getRouterCore("*").HandleFunc(method, path, hs)
	default:
		for _, host := range strings.Split(host, ",") {
			r.getRouterCore(host).HandleFunc(method, path, hs)
		}
	}
}

// getRouterCore 方法寻找参数对应的路由器核心，如果不存在则调用函数创建并存储。
func (r *routerCoreHost) getRouterCore(host string) RouterCore {
	core, ok := r.routers[host]
	if ok {
		return core
	}
	core = r.newRouteCore(host)
	r.routers[host] = core
	r.routertree.insert(host, core)
	return core
}

// Match 方法返回routerCoreHost.matchHost函数处理请求，在matchHost函数中使用host值进行二次匹配并拼接请求处理函数。
func (r *routerCoreHost) Match(method, path string, params *Params) HandlerFuncs {
	return HandlerFuncs{r.matchHost}
}

func (r *routerCoreHost) matchHost(ctx Context) {
	hs := r.routertree.matchNode(split2byte(ctx.Host(), ':')).Match(ctx.Method(), ctx.Path(), ctx.Params())
	index, handlers := ctx.GetHandler()
	ctx.SetHandler(index, NewHandlerFuncsCombine(NewHandlerFuncsCombine(handlers[:index+1], hs), handlers[index+1:]))
}

type routerHostNode struct {
	path     string
	wildcard *routerHostNode
	children []*routerHostNode
	any      RouterCore
	ports    map[string]RouterCore
}

func (node *routerHostNode) setRouter(port string, router RouterCore) {
	if port == "" {
		node.any = router
		return
	}
	if node.ports == nil {
		node.ports = make(map[string]RouterCore)
	}
	node.ports[port] = router
}

func (node *routerHostNode) getRouter(port string) RouterCore {
	router, ok := node.ports[port]
	if ok {
		return router
	}
	return node.any
}

func (node *routerHostNode) insert(path string, val RouterCore) {
	host, port := split2byte(path, ':')
	paths := strings.Split(host, "*")
	newpaths := make([]string, 1, len(paths)*2-1)
	newpaths[0] = paths[0]
	for _, path := range paths[1:] {
		newpaths = append(newpaths, "*")
		if path != "" {
			newpaths = append(newpaths, path)
		}
	}
	for _, p := range newpaths {
		node = node.insertNode(p)
	}
	node.setRouter(port, val)
}

func (node *routerHostNode) insertNode(path string) *routerHostNode {
	if path == "*" {
		if node.wildcard == nil {
			node.wildcard = &routerHostNode{path: path}
		}
		return node.wildcard
	}
	if path == "" {
		return node
	}

	for i := range node.children {
		subStr, find := getSubsetPrefix(path, node.children[i].path)
		if find {
			if subStr != node.children[i].path {
				node.children[i].path = strings.TrimPrefix(node.children[i].path, subStr)
				node.children[i] = &routerHostNode{
					path:     subStr,
					children: []*routerHostNode{node.children[i]},
				}
			}
			return node.children[i].insertNode(strings.TrimPrefix(path, subStr))
		}
	}
	newnode := &routerHostNode{path: path}
	node.children = append(node.children, newnode)
	// 常量node按照首字母排序。
	for i := len(node.children) - 1; i > 0; i-- {
		if node.children[i].path[0] < node.children[i-1].path[0] {
			node.children[i], node.children[i-1] = node.children[i-1], node.children[i]
		}
	}

	return newnode
}

func (node *routerHostNode) matchNode(path, port string) RouterCore {
	if path == "" {
		core := node.getRouter(port)
		if core != nil {
			return core
		}
	}
	for _, current := range node.children {
		if strings.HasPrefix(path, current.path) {
			if result := current.matchNode(path[len(current.path):], port); result != nil {
				return result
			}
		}
	}
	if node.wildcard != nil {
		if node.wildcard.children != nil {
			pos := strings.IndexByte(path, '.')
			if pos == -1 {
				pos = len(path)
			}
			if result := node.wildcard.matchNode(path[pos:], port); result != nil {
				return result
			}
		}
		router := node.wildcard.getRouter(port)
		if router != nil {
			return router
		}
	}
	return nil
}
