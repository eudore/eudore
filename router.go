package eudore

/*
Router

Router对象用于定义请求的路由

文件：router.go routerradix.go routerfull.go
*/

import (
	"fmt"
	"strings"
)

// 默认http请求方法
const (
	MethodAny     = "ANY"
	MethodGet     = "GET"
	MethodPost    = "POST"
	MethodPut     = "PUT"
	MethodDelete  = "DELETE"
	MethodHead    = "HEAD"
	MethodPatch   = "PATCH"
	MethodOptions = "OPTIONS"
	MethodConnect = "CONNECT"
	MethodTrace   = "TRACE"
)

type (
	// RouterMethod the route is directly registered by default. Other methods can be directly registered using the RouterRegister interface.
	//
	// RouterMethod 路由默认直接注册的方法，其他方法可以使用RouterRegister接口直接注册。
	RouterMethod interface {
		Group(string) RouterMethod
		AddHandler(string, string, ...interface{}) RouterMethod
		AddMiddleware(string, string, ...HandlerFunc) RouterMethod
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
		RegisterMiddleware(string, string, HandlerFuncs)
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
		ControllerParseFunc
		prefix string
		tags   string
	}
	// 存储中间件信息的基数树。
	//
	// 用于内存存储路由器中间件注册信息，并根据注册路由返回对应的中间件。
	middTree struct {
		root middNode
	}
	middNode struct {
		path     string
		children []*middNode
		key      string
		val      HandlerFuncs
	}
	// RoutesInjecter 定义路由注入接口，允许调用路由器方法注入自身路由信息。
	RoutesInjecter interface {
		RoutesInject(RouterMethod)
	}
	// RouterConfig storage router configuration for constructing routers.
	//
	// RouterConfig 存储路由器配置，用于构造路由器。
	RouterConfig struct {
		Path       string          `json:",omitempty"`
		Method     string          `json:",omitempty"`
		Handler    HandlerFuncs    `json:",omitempty"`
		Middleware HandlerFuncs    `json:",omitempty"`
		Routes     []*RouterConfig `json:",omitempty"`
	}
)

// check RouterStd has Router interface
var (
	_               Router       = &RouterRadix{}
	_               Router       = &RouterFull{}
	_               RouterMethod = &RouterMethodStd{}
	RouterAllMethod              = []string{MethodGet, MethodPost, MethodPut, MethodDelete, MethodHead, MethodPatch, MethodOptions}
)

// DefaultRouter405Func 函数定义默认405处理
func DefaultRouter405Func(ctx Context) {
	const page405 string = "405 method not allowed\n"
	ctx.Response().Header().Add("Allow", "HEAD, GET, POST, PUT, DELETE, PATCH")
	ctx.WriteHeader(405)
	ctx.WriteString(page405)
}

// DefaultRouter404Func 函数定义默认404处理
func DefaultRouter404Func(ctx Context) {
	const page404 string = "404 page not found\n"
	ctx.WriteHeader(404)
	ctx.WriteString(page404)
}

// RoutesInject 方法将路由配置注入到路由中。
func (config *RouterConfig) RoutesInject(r RouterMethod) {
	// handler
	r.AddHandler(config.Method, config.Path, config.Handler)

	// middleware
	r.AddMiddleware(config.Method, config.Path, config.Middleware...)

	// routes
	r = r.Group(config.Path)
	for _, i := range config.Routes {
		i.RoutesInject(r)
	}
}

// Group 返回一个组路由方法。
func (m *RouterMethodStd) Group(path string) RouterMethod {
	// 将路径前缀和路径参数分割出来
	args := strings.Split(path, " ")
	prefix := args[0]
	tags := path[len(prefix):]

	// 构建新的路由方法配置器
	return &RouterMethodStd{
		RouterCore:          m.RouterCore,
		ControllerParseFunc: m.ControllerParseFunc,
		prefix:              m.prefix + prefix,
		tags:                tags + m.tags,
	}
}

func (m *RouterMethodStd) registerHandlers(method, path string, hs ...interface{}) {
	handler := NewHandlerFuncs(hs)
	if len(handler) > 0 {
		m.RouterCore.RegisterHandler(method, m.prefix+path+m.tags, handler)
	}
}

// AddHandler 添加一个新路由。
//
// 方法和RegisterHandler方法的区别在于AddHandler方法不会继承Group的路径和参数信息，AddMiddleware相同。
func (m *RouterMethodStd) AddHandler(method, path string, hs ...interface{}) RouterMethod {
	m.registerHandlers(method, path, hs)
	return m
}

// AddMiddleware 给路由器添加一个中间件函数。
func (m *RouterMethodStd) AddMiddleware(method, path string, hs ...HandlerFunc) RouterMethod {
	if len(hs) > 0 {
		m.RegisterMiddleware(method, m.prefix+path+m.tags, hs)
	}
	return m
}

// AddController 方式使用内置的控制器解析函数解析控制器获得路由配置。
//
// 如果控制器实现了RoutesInjecter接口，调用控制器自身注入路由。
func (m *RouterMethodStd) AddController(cs ...Controller) RouterMethod {
	for _, c := range cs {
		if rj, ok := c.(RoutesInjecter); ok {
			rj.RoutesInject(m)
			continue
		}

		config, err := m.ControllerParseFunc(c)
		if err == nil {
			config.RoutesInject(m)
		} else {
			fmt.Println(err)
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

// Insert 方法实现middNode添加一个子节点。
func (t *middNode) Insert(key string, val HandlerFuncs) {
	t.recursiveInsertTree(key, key, val)
}

// Lookup Find if seachKey exist in current radix tree and return its value
func (t *middNode) Lookup(searchKey string) HandlerFuncs {
	searchKey = strings.Split(searchKey, " ")[0]
	if searchKey[len(searchKey)-1] == '*' {
		searchKey = searchKey[:len(searchKey)-1]
	}
	return t.recursiveLoopup(searchKey)
}

// InsertNode 新增Node
func (t *middNode) InsertNode(path, key string, value HandlerFuncs) {
	if len(path) == 0 {
		// 路径空就设置当前node的值
		t.key = key
		t.val = CombineHandlerFuncs(t.val, value)
	} else {
		// 否则新增node
		t.children = append(t.children, &middNode{path: path, key: key, val: value})
	}
}

// SplitNode 对指定路径为edgeKey的Node分叉，公共前缀路径为pathKey
func (t *middNode) SplitNode(pathKey, edgeKey string) *middNode {
	for i := range t.children {
		if t.children[i].path == edgeKey {
			newNode := &middNode{path: pathKey}
			newNode.children = append(newNode.children, &middNode{
				path:     strings.TrimPrefix(edgeKey, pathKey),
				key:      t.children[i].key,
				val:      t.children[i].val,
				children: t.children[i].children,
			})
			t.children[i] = newNode
			return newNode
		}
	}
	return nil
}

// 给currentNode递归添加，路径为containKey的Node。
func (t *middNode) recursiveInsertTree(containKey string, targetKey string, targetValue HandlerFuncs) {
	for i := range t.children {
		subStr, find := getSubsetPrefix(containKey, t.children[i].path)
		if find {
			if subStr == t.children[i].path {
				nextTargetKey := strings.TrimPrefix(containKey, t.children[i].path)
				t.children[i].recursiveInsertTree(nextTargetKey, targetKey, targetValue)
			} else {
				newNode := t.SplitNode(subStr, t.children[i].path)
				if newNode == nil {
					panic("Unexpect error on split node")
				}

				newNode.InsertNode(strings.TrimPrefix(containKey, subStr), targetKey, targetValue)
			}
			return
		}
	}
	t.InsertNode(containKey, targetKey, targetValue)
}

// 递归获得searchNode路径为searchKey的Node数据。
func (t *middNode) recursiveLoopup(searchKey string) HandlerFuncs {
	if len(searchKey) == 0 {
		return t.val
	}

	for _, edgeObj := range t.children {
		// 寻找相同前缀node
		if contrainPrefix(searchKey, edgeObj.path) {
			nextSearchKey := strings.TrimPrefix(searchKey, edgeObj.path)
			return append(t.val, edgeObj.recursiveLoopup(nextSearchKey)...)
		}
	}

	if len(t.key) == 0 || t.key[len(t.key)-1] == '/' {
		return t.val
	}

	return nil
}
