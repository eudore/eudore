package eudore

/*
Router

Router对象用于定义请求的路由

文件：router.go routerradix.go routerfull.go
*/

import (
	"strings"
)

type (
	// RouterMethod the route is directly registered by default. Other methods can be directly registered using the RouterRegister interface.
	//
	// RouterMethod 路由默认直接注册的方法，其他方法可以使用RouterRegister接口直接注册。
	RouterMethod interface {
		Group(string) RouterMethod
		GetParam(string) string
		SetParam(string, string) RouterMethod
		AddHandler(string, string, ...interface{}) RouterMethod
		AddMiddleware(...HandlerFunc) RouterMethod
		AddController(...Controller) RouterMethod
		AnyFunc(string, ...interface{})
		GetFunc(string, ...interface{})
		PostFunc(string, ...interface{})
		PutFunc(string, ...interface{})
		DeleteFunc(string, ...interface{})
		HeadFunc(string, ...interface{})
		PatchFunc(string, ...interface{})
		OptionsFunc(string, ...interface{})
	}
	// RouterCore interface, performs routing, middleware registration, and matches a request and returns to the handler.
	//
	// RouterCore接口，执行路由、中间件的注册和匹配一个请求并返回处理者。
	RouterCore interface {
		RegisterMiddleware(string, HandlerFuncs)
		RegisterHandler(string, string, HandlerFuncs)
		Match(string, string, Params) HandlerFuncs
	}
	// Router 接口，需要实现路由器方法、路由器核心两个接口。
	Router interface {
		RouterCore
		RouterMethod
	}

	// RouterMethodStd 默认路由器方法注册实现
	RouterMethodStd struct {
		RouterCore
		params *ParamsArray
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
	_ Router       = &RouterRadix{}
	_ Router       = &RouterFull{}
	_ RouterMethod = &RouterMethodStd{}
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

// NewRouterMethodStd 方法创建一个RouterMethod对象。
func NewRouterMethodStd(core RouterCore) RouterMethod {
	return &RouterMethodStd{
		RouterCore: core,
		params: &ParamsArray{
			Keys: []string{"route"},
			Vals: []string{""},
		},
	}
}

// Group 返回一个组路由方法。
func (m *RouterMethodStd) Group(path string) RouterMethod {
	// 构建新的路由方法配置器
	return &RouterMethodStd{
		RouterCore: m.RouterCore,
		params:     m.ParamsCombine(path),
	}
}

// Params 方法返回当前路由参数，路由参数值为空字符串不会被使用。
func (m *RouterMethodStd) Params() Params {
	return m.params
}

// GetParam 方法返回一个路由器参数。
func (m *RouterMethodStd) GetParam(key string) string {
	return m.params.Get(key)
}

// SetParam 方法设置一个路由器参数。
func (m *RouterMethodStd) SetParam(key string, val string) RouterMethod {
	m.params.Set(key, val)
	return m
}

// ParamsCombine 方法解析一个字符串路径，并合并到一个当前路由参数的副本中。
func (m *RouterMethodStd) ParamsCombine(path string) *ParamsArray {
	args := strings.Split(path, " ")
	params := m.params.Clone()
	for _, str := range args[1:] {
		params.Set(split2byte(str, '='))
	}
	params.Set("route", params.Get("route")+args[0])
	return params
}

func (m *RouterMethodStd) registerHandlers(method, path string, hs ...interface{}) {
	handler := NewHandlerFuncs(hs)
	if len(handler) > 0 {
		m.RouterCore.RegisterHandler(method, m.ParamsCombine(path).String()[6:], handler)
	}
}

// AddHandler 添加一个新路由, 允许添加多个方法使用空格分开。
//
// AddHandler方法和RegisterHandler方法的区别在于：AddHandler方法会继承Group和Params的路径和参数信息，AddMiddleware相同。
func (m *RouterMethodStd) AddHandler(method, path string, hs ...interface{}) RouterMethod {
	for _, method := range strings.Split(method, " ") {
		if method != "" {
			m.registerHandlers(method, path, hs)
		}
	}
	return m
}

// AddMiddleware 给路由器添加一个中间件函数。
func (m *RouterMethodStd) AddMiddleware(hs ...HandlerFunc) RouterMethod {
	if len(hs) > 0 {
		path := m.params.String()
		if len(path) > 5 && path[:6] == "route=" {
			path = path[6:]
		}
		m.RegisterMiddleware(path, hs)
	}
	return m
}

// AddController 方式使用内置的控制器解析函数解析控制器获得路由配置。
//
// 如果控制器实现了RoutesInjecter接口，调用控制器自身注入路由。
func (m *RouterMethodStd) AddController(cs ...Controller) RouterMethod {
	for _, c := range cs {
		err := c.Inject(c, m)
		if err != nil {
			panic(err)
		}
	}
	return m
}

// AnyFunc 方法实现注册一个Any方法的http请求处理函数。
func (m *RouterMethodStd) AnyFunc(path string, h ...interface{}) {
	m.registerHandlers(MethodAny, path, h)
}

// GetFunc 方法实现注册一个Get方法的http请求处理函数。
func (m *RouterMethodStd) GetFunc(path string, h ...interface{}) {
	m.registerHandlers(MethodGet, path, h)
}

// PostFunc 方法实现注册一个Post方法的http请求处理函数。
func (m *RouterMethodStd) PostFunc(path string, h ...interface{}) {
	m.registerHandlers(MethodPost, path, h)
}

// PutFunc 方法实现注册一个Put方法的http请求处理函数。
func (m *RouterMethodStd) PutFunc(path string, h ...interface{}) {
	m.registerHandlers(MethodPut, path, h)
}

// DeleteFunc 方法实现注册一个Delete方法的http请求处理函数。
func (m *RouterMethodStd) DeleteFunc(path string, h ...interface{}) {
	m.registerHandlers(MethodDelete, path, h)
}

// HeadFunc 方法实现注册一个Head方法的http请求处理函数。
func (m *RouterMethodStd) HeadFunc(path string, h ...interface{}) {
	m.registerHandlers(MethodHead, path, h)
}

// PatchFunc 方法实现注册一个Patch方法的http请求处理函数。
func (m *RouterMethodStd) PatchFunc(path string, h ...interface{}) {
	m.registerHandlers(MethodPatch, path, h)
}

// OptionsFunc 方法实现注册一个Options方法的http请求处理函数。
func (m *RouterMethodStd) OptionsFunc(path string, h ...interface{}) {
	m.registerHandlers(MethodOptions, path, h)
}

func newTrieNode() *trieNode {
	return &trieNode{}
}

// Insert 方法实现trieNode添加一个子节点。
func (t *trieNode) Insert(path string, vals HandlerFuncs) {
	if path == "" {
		t.vals = append(t.vals, vals...)
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
		&trieNode{
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
			return append(t.vals, i.Lookup(path[len(i.path):])...)
		}
	}
	return t.vals
}
