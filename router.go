package eudore

/*
Router

Router对象用于定义请求的路由

文件：router.go routerradix.go routerfull.go
*/

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type (
	// Router 接口，需要实现路由器方法、路由器核心两个接口。
	//
	// RouterCore实现路由匹配细节，RouterMethod调用RouterCore提供对外使用的方法。
	//
	// 任何时候请不要使用RouterCore的方法直接注册，应该使用RouterMethod的Add...方法。
	Router interface {
		RouterCore
		RouterMethod
	}
	// RouterCore interface, performs routing, middleware registration, and matches a request and returns to the handler.
	//
	// RouterCore接口，执行路由、中间件的注册和匹配一个请求并返回处理者。
	//
	// RouterCore主要实现路由匹配相关细节。
	RouterCore interface {
		RegisterMiddleware(string, HandlerFuncs)
		RegisterHandler(string, string, HandlerFuncs)
		Match(string, string, Params) HandlerFuncs
	}
	// RouterMethod the route is directly registered by default. Other methods can be directly registered using the RouterRegister interface.
	//
	// RouterMethod 路由默认直接注册的方法，其他方法可以使用RouterRegister接口直接注册。
	RouterMethod interface {
		Group(string) Router
		GetParam(string) string
		SetParam(string, string) Router
		AddHandler(string, string, ...interface{}) error
		AddMiddleware(...HandlerFunc)
		AddController(...Controller) error
		AddHandlerExtend(interface{}) error
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
		HandlerExtender
		params *ParamsArray
		Print  func(...interface{}) `set:"print"`
	}
	// RouterCoreLock 对象给RouterCore注册提供锁功能
	RouterCoreLock struct {
		RouterCore
		sync.Mutex
	}
	// trieNode 存储中间件信息的前缀树。
	//
	// 用于内存存储路由器中间件注册信息，并根据注册路由返回对应的中间件。
	trieNode struct {
		path   string
		childs []*trieNode
		vals   HandlerFuncs
	}
)

// check RouterStd has Router interface
var (
	_ Router     = &RouterStd{}
	_ RouterCore = &RouterCoreRadix{}
	_ RouterCore = &RouterCoreFull{}
	// RouterAllMethod 影响Any方法的注册。
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
		RouterCore:      core,
		HandlerExtender: NewHandlerExtendWarp(DefaultHandlerExtend),
		params: &ParamsArray{
			Keys: []string{"route"},
			Vals: []string{""},
		},
		Print: func(...interface{}) {
			// Do nothing because not print message.
		},
	}
}

// Group 方法返回一个新的组路由。新路由具有独立的参数和处理函数扩展。
//
// 新的Router将使用旧的RouterCore和Print对象；复制一份新的路由参数；
//
// 以及链式创建一个新的HandlerExtender，若HandlerExtender无法注册的类型将调用上一个Router.HandlerExtender处理。
//
// 最顶级HandlerExtender对象为defaultHandlerExtend，可以使用RegisterHandlerExtend函数和NewHandlerFuncs函数调用defaultHandlerExtend对象。
func (m *RouterStd) Group(path string) Router {
	// 构建新的路由方法配置器
	return &RouterStd{
		RouterCore:      m.RouterCore,
		HandlerExtender: NewHandlerExtendWarp(m.HandlerExtender),
		params:          m.ParamsCombine(path),
		Print:           m.Print,
	}
}

// Params 方法返回当前路由参数，路由参数值为空字符串不会被使用。
func (m *RouterStd) Params() Params {
	return m.params
}

// GetParam 方法返回一个路由器参数。
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

// SetParam 方法设置一个路由器参数。
func (m *RouterStd) SetParam(key string, val string) Router {
	m.params.Set(key, val)
	return m
}

// ParamsCombine 方法解析一个字符串路径，并合并到一个当前路由参数的副本中。
func (m *RouterStd) ParamsCombine(path string) *ParamsArray {
	args := strings.Split(path, " ")
	params := m.params.Clone()
	for _, str := range args[1:] {
		params.Set(split2byte(str, '='))
	}
	params.Set("route", params.Get("route")+args[0])
	return params
}

// registerHandlers 方法将handler转换成HandlerFuncs，并调用RouterCore对象注册路由方法。
func (m *RouterStd) registerHandlers(method, path string, hs ...interface{}) error {
	var handlers HandlerFuncs
	var errs Errors
	// 转换处理函数
	for i, h := range hs {
		handler := m.HandlerExtender.NewHandlerFuncs(h)
		if handler != nil && len(handler) > 0 {
			handlers = append(handlers, handler...)
		} else {
			err := fmt.Errorf(ErrFormatAddHandlerFuncUnregisterType, method, path, i, reflect.TypeOf(h).String())
			m.Print(newFileLineFields(3), err)
			errs.HandleError(err)
		}
	}
	if errs.GetError() != nil || handlers == nil || len(handlers) == 0 {
		return errs.GetError()
	}

	// 处理多方法
	path = m.ParamsCombine(path).String()[6:]
	for _, i := range strings.Split(method, ",") {
		if checkMethod(i) {
			m.RouterCore.RegisterHandler(i, path, handlers)
			m.Print("RegisterHandler:", i, path, handlers)
		} else {
			err := fmt.Errorf(ErrFormatAddHandlerMethodInvalid, i, method, path)
			m.Print(newFileLineFields(3), err)
			errs.HandleError(err)
		}
	}
	return errs.GetError()
}

// AddHandler 添加一个新路由, 允许添加多个方法使用','分开。
//
// AddHandler方法和RegisterHandler方法的区别在于：AddHandler方法会继承Group和Params的路径和参数信息，AddMiddleware相同。
//
// handler参数使用当前RouterStd的HandlerExtender.NewHandlerFuncs()方法处理，生成对应的HandlerFuncs。
//
// 当前Router无法处理，则调用group前的HandlerExtender或defaultHandlerExtend处理，全部无法处理则输出error日志。
func (m *RouterStd) AddHandler(method, path string, hs ...interface{}) error {
	return m.registerHandlers(method, path, hs...)
}

func checkMethod(method string) bool {
	if method == "ANY" {
		return true
	}
	for _, i := range RouterAllMethod {
		if i == method {
			return true
		}
	}
	return false
}

// AddMiddleware 给路由器添加一个中间件函数。
func (m *RouterStd) AddMiddleware(hs ...HandlerFunc) {
	if len(hs) > 0 {
		path := m.params.String()
		if len(path) > 5 && path[:6] == "route=" {
			path = path[6:]
		}
		m.RegisterMiddleware(path, hs)
		m.Print("RegisterMiddleware:", path, hs)
	}
}

// AddController 方式使用内置的控制器解析函数解析控制器获得路由配置。
//
// 如果控制器实现了RoutesInjecter接口，调用控制器自身注入路由。
func (m *RouterStd) AddController(cs ...Controller) error {
	if len(cs) == 0 {
		m.Print(newFileLineFields(2), ErrRouterAddControllerEmpty)
		return ErrRouterAddControllerEmpty
	}
	var errs Errors
	for _, c := range cs {
		err := c.Inject(c, m)
		if err != nil {
			m.Print(newFileLineFields(2), err)
			errs.HandleError(err)
		}
	}
	return errs.GetError()
}

// AddHandlerExtend 方法给当前Router添加扩展函数。
func (m *RouterStd) AddHandlerExtend(i interface{}) error {
	err := m.HandlerExtender.RegisterHandlerExtend(i)
	if err != nil {
		m.Print(newFileLineFields(2), fmt.Errorf("RouterStd AddHandlerExtend error: %v", err))
	}
	return err
}

// Set 方法允许设置Print属性，设置日志输出信息。
func (m *RouterStd) Set(key string, value interface{}) error {
	switch val := value.(type) {
	case func(...interface{}):
		m.Print = val
	case RouterCore:
		m.RouterCore = val
	case HandlerExtender:
		m.HandlerExtender = val
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

// RegisterMiddleware 方法调用路由核心注册中间件。
func (r *RouterCoreLock) RegisterMiddleware(path string, hs HandlerFuncs) {
	r.Mutex.Lock()
	r.RouterCore.RegisterMiddleware(path, hs)
	r.Mutex.Unlock()
}

// RegisterHandler 方法调用路由核心注册路由路径请求。
func (r *RouterCoreLock) RegisterHandler(method string, path string, hs HandlerFuncs) {
	r.Mutex.Lock()
	r.RouterCore.RegisterHandler(method, path, hs)
	r.Mutex.Unlock()
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
				t.childs[i].SplitNode(subStr)
			}
			t.childs[i].Insert(path[len(subStr):], vals)
			return
		}
	}
	t.childs = append(t.childs, &trieNode{path: path, vals: vals})
}

// SplitNode 方法基于路径拆分。
func (t *trieNode) SplitNode(path string) {
	t.childs = []*trieNode{
		{
			path:   t.path[len(path):],
			childs: t.childs,
			vals:   t.vals,
		},
	}
	t.path = path
	t.vals = nil
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
