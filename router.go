package eudore

/*
Router

Router对象用于定义请求的路由器

文件：router.go routerradix.go routerfull.go
*/

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"sync"
)

type (
	// Router interface needs to implement the router method and the router core two interfaces.
	//
	// RouterCore implements route matching details. RouterMethod calls RouterCore to provide methods for external use.
	//
	// Do not use the RouterCore method to register routes directly at any time. You should use the Add ... method of RouterMethod.
	//
	// Router 接口，需要实现路由器方法、路由器核心两个接口。
	//
	// RouterCore实现路由匹配细节，RouterMethod调用RouterCore提供对外使用的方法。
	//
	// 任何时候请不要使用RouterCore的方法直接注册路由，应该使用RouterMethod的Add...方法。
	Router interface {
		RouterCore
		RouterMethod
	}
	// The RouterCore interface performs registration of the route and matches a request and returns the handler.
	//
	// RouterCore mainly implements routing matching related details.
	//
	// RouterCore接口，执行路由的注册和匹配一个请求并返回处理者。
	//
	// RouterCore主要实现路由匹配相关细节。
	RouterCore interface {
		HandleFunc(string, string, HandlerFuncs)
		Match(string, string, Params) HandlerFuncs
	}
	// RouterMethod The default directly registered interface of the route. Set the routing parameters, group routing, middleware, function extensions, controllers and other behaviors.
	//
	// RouterMethod 路由默认直接注册的接口，设置路由参数、组路由、中间件、函数扩展、控制器等行为。
	RouterMethod interface {
		Group(string) Router
		GetParam(string) string
		SetParam(string, string) Router
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
		OptionsFunc(string, ...interface{})
	}
	// RouterStd 默认路由器注册实现。
	//
	// 需要指定一个路由核心，处理函数扩展者默认为DefaultHandlerExtend。
	RouterStd struct {
		RouterCore
		params          *ParamsArray
		HandlerExtender HandlerExtender
		Middlewares     *trieNode
		Print           func(...interface{}) `set:"print"`
	}
	// RouterCoreLock object provides read-write lock function for RouterCore registration
	//
	// RouterCoreLock 对象给RouterCore注册提供读写锁功能
	RouterCoreLock struct {
		RouterCore
		sync.RWMutex
	}
	// trieNode 存储中间件信息的前缀树。
	//
	// 用于内存存储路由器中间件注册信息，并根据注册路由返回对应的中间件。
	trieNode struct {
		path   string
		vals   HandlerFuncs
		childs []*trieNode
	}
)

// check RouterStd has Router interface
var (
	_ Router     = &RouterStd{}
	_ RouterCore = &RouterCoreRadix{}
	_ RouterCore = &RouterCoreFull{}
	// RouterAllMethod 定义全部的方法，影响Any方法的注册。
	RouterAllMethod = []string{MethodGet, MethodPost, MethodPut, MethodDelete, MethodHead, MethodPatch, MethodOptions}
)

// HandlerRouter405 函数定义默认405处理
func HandlerRouter405(ctx Context) {
	const page405 string = "405 method not allowed\n"
	ctx.Response().Header().Add("Allow", "HEAD, GET, POST, PUT, DELETE, PATCH")
	ctx.WriteHeader(405)
	ctx.WriteString(page405)
}

// HandlerRouter404 函数定义默认404处理
func HandlerRouter404(ctx Context) {
	const page404 string = "404 page not found\n"
	ctx.WriteHeader(404)
	ctx.WriteString(page404)
}

// NewRouterStd 方法创建使用RouterCore一个Router对象。
func NewRouterStd(core RouterCore) Router {
	return &RouterStd{
		RouterCore: core,
		params: &ParamsArray{
			Keys: []string{ParamRoute},
			Vals: []string{""},
		},
		HandlerExtender: NewHandlerExtendWarp(NewHandlerExtendTree(), DefaultHandlerExtend),
		Middlewares:     newTrieNode(),
		Print: func(...interface{}) {
			// Do nothing because default router not print message.
		},
	}
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
func (m *RouterStd) Group(path string) Router {
	// 构建新的路由方法配置器
	return &RouterStd{
		RouterCore:      m.RouterCore,
		params:          m.ParamsCombine(path),
		HandlerExtender: NewHandlerExtendWarp(NewHandlerExtendTree(), m.HandlerExtender),
		Middlewares:     m.Middlewares.clone(),
		Print:           m.Print,
	}
}

// PrintError 方法输出一个err，附加错误的函数名称和文件位置。
func (m *RouterStd) PrintError(depth int, err error) {
	// 兼容添加控制器错误输出
	name, _, _ := logFormatNameFileLine(depth + 5)
	if name == "github.com/eudore/eudore.(*RouterStd).AddController" {
		depth += 3
	}

	name, file, line := logFormatNameFileLine(depth + 3)
	m.Print(Fields{"func": name, "file": file, "line": line}, err)
}

// PrintPanic 方法输出一个err，附加当前stack。
func (m *RouterStd) PrintPanic(err error) {
	pc := make([]uintptr, 20)
	n := runtime.Callers(0, pc)
	if n == 0 {
		m.Print(err)
	}

	pc = pc[:n] // pass only valid pcs to runtime.CallersFrames
	frames := runtime.CallersFrames(pc)
	stack := make([]string, 0, 20)

	frame, more := frames.Next()
	for more {
		pos := strings.Index(frame.File, "src")
		if pos >= 0 {
			frame.File = frame.File[pos+4:]
		}
		pos = strings.LastIndex(frame.Function, "/")
		if pos >= 0 {
			frame.Function = frame.Function[pos+1:]
		}
		stack = append(stack, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))

		frame, more = frames.Next()
	}
	m.Print(Fields{"stack": stack}, err)
}

// Params 方法返回当前路由参数，路由参数值为空字符串不会被使用。
func (m *RouterStd) Params() Params {
	return m.params
}

// GetParam method returns a router parameter.
//
// If the key is eudore.ParamAllKeys / eudore.ParamAllVals and the value is empty, all keys / values are returned, separated by spaces between multiple values.
//
// GetParam 方法返回一个路由器参数。
//
// 如果key为eudore.ParamAllKeys/eudore.ParamAllVals且值为空，则返回全部的键/值，多值间空格分割。
func (m *RouterStd) GetParam(key string) string {
	val := m.params.Get(key)
	// 返回params全部key/val
	if val == "" {
		switch key {
		case ParamAllKeys:
			val = strings.Join(m.params.Keys, " ")
		case ParamAllVals:
			val = strings.Join(m.params.Vals, " ")
		}
	}
	return val
}

// SetParam 方法给当前路由器设置一个参数。
func (m *RouterStd) SetParam(key string, val string) Router {
	m.params.Set(key, val)
	return m
}

// ParamsCombine method parses a string path and merges it into a copy of the current routing parameters.
//
// For example, the path format is: / user action = user
//
// ParamsCombine 方法解析一个字符串路径，并合并到一个当前路由参数的副本中。
//
// 例如路径格式为：/user action=user
func (m *RouterStd) ParamsCombine(path string) *ParamsArray {
	args := strings.Split(path, " ")
	params := m.params.Clone()
	key, val := split2byte(args[0], '=')
	switch key {
	case "", "route":
		params.Set("route", params.Get("route")+val)
		args = args[1:]
	}
	for _, str := range args {
		params.Set(split2byte(str, '='))
	}
	return params
}

// AddHandler adds a new route, allowing multiple request methods to be separated using ','.
//
// The handler parameter is processed using the current HandlerExtender.NewHandlerFuncs () method of RouterStd to generate the corresponding HandlerFuncs.
//
// The current Router cannot process, then call the HandlerExtender or defaultHandlerExtend before the group, and output all error logs if it cannot process all.
//
// will match the aligned request middleware and append to the request based on the current routing path.
//
// AddHandler 添加一个新路由, 允许添加多个请求方法使用','分开。
//
// handler参数使用当前RouterStd的HandlerExtender.NewHandlerFuncs()方法处理，生成对应的HandlerFuncs。
//
// 当前Router无法处理，则调用group前的HandlerExtender或defaultHandlerExtend处理，全部无法处理则输出error日志。
//
// 会根据当前路由路径匹配到对齐的请求中间件并附加到请求中。
func (m *RouterStd) AddHandler(method, path string, hs ...interface{}) error {
	return m.registerHandlers(method, path, hs...)
}

// registerHandlers 方法将handler转换成HandlerFuncs，添加路由路径对应的请求中间件，并调用RouterCore对象注册路由方法。
func (m *RouterStd) registerHandlers(method, path string, hs ...interface{}) (err error) {
	defer func() {
		// RouterCoreFull 注册未知校验规则存在panic
		if rerr := recover(); rerr != nil {
			err = fmt.Errorf(ErrFormatRouterStdRegisterHandlersRecover, method, path, rerr)
			m.PrintPanic(err)
		}
	}()

	params := m.ParamsCombine(path)
	path = params.Get("route")
	fullpath := params.String()
	if len(fullpath) > 6 && fullpath[:6] == "route=" {
		fullpath = fullpath[6:]
	}

	handlers, err := m.newHandlerFuncs(path, hs)
	if err != nil {
		m.PrintError(1, err)
		return err
	}
	m.Print("RegisterHandler:", method, fullpath, handlers)
	handlers = HandlerFuncsCombine(m.Middlewares.Lookup(path), handlers)

	// 处理多方法
	var errs Errors
	for _, i := range strings.Split(method, ",") {
		if checkMethod(i) {
			m.RouterCore.HandleFunc(i, fullpath, handlers)
		} else {
			err := fmt.Errorf(ErrFormatRouterStdRegisterHandlersMethodInvalid, i, method, fullpath)
			errs.HandleError(err)
			m.PrintError(1, err)
		}
	}
	return errs.GetError()
}

// The newHandlerFuncs method creates HandlerFuncs based on the path and multiple parameters.
//
// RouterStd first calls the current HandlerExtender.NewHandlerFuncs to create multiple function handlers. If it returns null, it will be created from the superior HandlerExtender.
//
// newHandlerFuncs 方法根据路径和多个参数创建HandlerFuncs。
//
// RouterStd先调用当前HandlerExtender.NewHandlerFuncs创建多个函数处理者，如果返回空会从上级HandlerExtender创建。
func (m *RouterStd) newHandlerFuncs(path string, hs []interface{}) (HandlerFuncs, error) {
	var handlers HandlerFuncs
	var errs Errors
	// 转换处理函数
	for i, h := range hs {
		handler := m.HandlerExtender.NewHandlerFuncs(path, h)
		if handler != nil && len(handler) > 0 {
			handlers = HandlerFuncsCombine(handlers, handler)
		} else {
			errs.HandleError(fmt.Errorf(ErrFormatRouterStdNewHandlerFuncsUnregisterType, path, i, reflect.TypeOf(h).String()))
		}
	}
	return handlers, errs.GetError()
}

func checkMethod(method string) bool {
	switch method {
	case "ANY", "404", "405", "NotFound", "MethodNotAllowed":
		return true
	}
	for _, i := range RouterAllMethod {
		if i == method {
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
func (m *RouterStd) AddController(cs ...Controller) error {
	var errs Errors
	for _, c := range cs {
		err := c.Inject(c, m)
		if err != nil {
			err = fmt.Errorf(ErrFormatRouterStdAddController, err)
			errs.HandleError(err)
			m.PrintError(0, err)
		}
	}
	return errs.GetError()
}

// AddMiddleware adds multiple middleware functions to the router, which will use HandlerExtender to convert parameters.
//
// If the number of parameters is greater than 1 and the first parameter is a string type, the first string type parameter is used as the path to add the middleware.
//
// AddMiddleware 给路由器添加多个中间件函数，会使用HandlerExtender转换参数。
//
// 如果参数数量大于1且第一个参数为字符串类型，会将第一个字符串类型参数作为添加中间件的路径。
func (m *RouterStd) AddMiddleware(hs ...interface{}) error {
	if len(hs) == 0 {
		return nil
	}

	path := m.GetParam("route")
	if len(hs) > 1 {
		perfix, ok := hs[0].(string)
		if ok {
			path = perfix + path
			hs = hs[1:]
		}
	}

	handlers, err := m.newHandlerFuncs(path, hs)
	if err != nil {
		m.PrintError(0, err)
		return err
	}

	m.Middlewares.Insert(path, handlers)
	m.RouterCore.HandleFunc("Middlewares", path, handlers)
	m.Print("RegisterMiddleware:", path, handlers)
	return nil
}

// AddHandlerExtend method adds an extension function to the current Router.
//
// If the number of parameters is greater than 1 and the first parameter is a string type, the first string type parameter is used as the path to add the extension function.
//
// AddHandlerExtend 方法给当前Router添加扩展函数。
//
// 如果参数数量大于1且第一个参数为字符串类型，会将第一个字符串类型参数作为添加扩展函数的路径。
func (m *RouterStd) AddHandlerExtend(hs ...interface{}) error {
	if len(hs) == 0 {
		return nil
	}

	path := m.GetParam("route")
	if len(hs) > 1 {
		perfix, ok := hs[0].(string)
		if ok {
			path = perfix + path
			hs = hs[1:]
		}
	}

	var errs Errors
	for _, h := range hs {
		err := m.HandlerExtender.RegisterHandlerExtend(path, h)
		if err != nil {
			err = fmt.Errorf(ErrFormatRouterStdAddHandlerExtend, path, err)
			errs.HandleError(err)
			m.PrintError(0, err)
		}
	}
	return errs.GetError()
}

// Set 方法允许设置Print属性，设置日志输出信息。
func (m *RouterStd) Set(key string, value interface{}) error {
	switch val := value.(type) {
	case RouterCore:
		m.RouterCore = val
	case *ParamsArray:
		m.params = val
	case HandlerExtender:
		m.HandlerExtender = val
	case func(...interface{}):
		m.Print = val
	default:
		return ErrSeterNotSupportField
	}
	return nil
}

// AnyFunc 方法实现注册一个Any方法的http请求处理函数。
func (m *RouterStd) AnyFunc(path string, h ...interface{}) {
	m.registerHandlers(MethodAny, path, h...)
}

// GetFunc 方法实现注册一个Get方法的http请求处理函数。
func (m *RouterStd) GetFunc(path string, h ...interface{}) {
	m.registerHandlers(MethodGet, path, h...)
}

// PostFunc 方法实现注册一个Post方法的http请求处理函数。
func (m *RouterStd) PostFunc(path string, h ...interface{}) {
	m.registerHandlers(MethodPost, path, h...)
}

// PutFunc 方法实现注册一个Put方法的http请求处理函数。
func (m *RouterStd) PutFunc(path string, h ...interface{}) {
	m.registerHandlers(MethodPut, path, h...)
}

// DeleteFunc 方法实现注册一个Delete方法的http请求处理函数。
func (m *RouterStd) DeleteFunc(path string, h ...interface{}) {
	m.registerHandlers(MethodDelete, path, h...)
}

// HeadFunc 方法实现注册一个Head方法的http请求处理函数。
func (m *RouterStd) HeadFunc(path string, h ...interface{}) {
	m.registerHandlers(MethodHead, path, h...)
}

// PatchFunc 方法实现注册一个Patch方法的http请求处理函数。
func (m *RouterStd) PatchFunc(path string, h ...interface{}) {
	m.registerHandlers(MethodPatch, path, h...)
}

// OptionsFunc 方法实现注册一个Options方法的http请求处理函数。
func (m *RouterStd) OptionsFunc(path string, h ...interface{}) {
	m.registerHandlers(MethodOptions, path, h...)
}

// NewRouterCoreLock 函数创建一个带锁的路由核心，通常路由不需要使用到锁。
func NewRouterCoreLock(core RouterCore) RouterCore {
	return &RouterCoreLock{
		RouterCore: core,
	}
}

// HandleFunc 方法调用路由核心注册路由路径请求。
func (r *RouterCoreLock) HandleFunc(method string, path string, hs HandlerFuncs) {
	r.RWMutex.Lock()
	r.RouterCore.HandleFunc(method, path, hs)
	r.RWMutex.Unlock()
}

// Match 方法调用路由核心匹配路由路径请求。
func (r *RouterCoreLock) Match(method string, path string, params Params) HandlerFuncs {
	r.RWMutex.RLock()
	hs := r.RouterCore.Match(method, path, params)
	r.RWMutex.RUnlock()
	return hs
}

func newTrieNode() *trieNode {
	return &trieNode{}
}

// Insert 方法实现trieNode添加一个子节点。
func (t *trieNode) Insert(path string, vals HandlerFuncs) {
	if path == "" {
		t.vals = HandlerFuncsCombine(t.vals, vals)
		return
	}
	for i := range t.childs {
		subStr, find := getSubsetPrefix(path, t.childs[i].path)
		if find {
			if subStr != t.childs[i].path {
				t.childs[i].path = strings.TrimPrefix(t.childs[i].path, subStr)
				t.childs[i] = &trieNode{
					path:   subStr,
					childs: []*trieNode{t.childs[i]},
				}
			}
			t.childs[i].Insert(path[len(subStr):], vals)
			return
		}
	}
	t.childs = append(t.childs, &trieNode{path: path, vals: vals})
}

// Lookup Find if seachKey exist in current trie tree and return its value
func (t *trieNode) Lookup(path string) HandlerFuncs {
	for _, i := range t.childs {
		if strings.HasPrefix(path, i.path) {
			return HandlerFuncsCombine(t.vals, i.Lookup(path[len(i.path):]))
		}
	}
	return t.vals
}

func (t *trieNode) clone() *trieNode {
	nt := *t
	for i := range nt.childs {
		nt.childs[i] = nt.childs[i].clone()
	}
	return &nt
}
