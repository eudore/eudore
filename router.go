package eudore

import (
	"fmt"
)

// 默认http请求方法
const (
	MethodAny		=	"ANY"
	MethodGet		=	"GET"
	MethodPost		=	"POST"
	MethodPut		=	"PUT"
	MethodDelete	=	"DELETE"
	MethodHead		=	"HEAD"
	MethodPatch		=	"PATCH"
	MethodOptions	=	"OPTIONS"
	MethodConnect	=	"CONNECT"
	MethodTrace		=	"TRACE"
)

type (
	// Router method
	// 路由默认直接注册的方法，其他方法可以使用RouterRegister接口直接注册。
	RouterMethod interface {
		// SubRoute(path string, router Router)
		Group(string) RouterMethod
		AddHandler(...Handler)
		Any(string, Handler)
		AnyFunc(string, HandlerFunc)
		Delete(string, Handler)
		DeleteFunc(string, HandlerFunc)
		Get(string, Handler)
		GetFunc(string, HandlerFunc)
		Head(string, Handler)
		HeadFunc(string, HandlerFunc)
		Options(string, Handler)
		OptionsFunc(string, HandlerFunc)
		Patch(string, Handler)
		PatchFunc(string, HandlerFunc)
		Post(string, Handler)
		PostFunc(string, HandlerFunc)
		Put(string, Handler)
		PutFunc(string, HandlerFunc)
	}
	// Router Core
	RouterCore interface {
		Middleware
		RegisterMiddleware(...Handler)
		RegisterHandler(method string, path string, handler Handler)
		Match(string, string, Params) Middleware
	}
	// router
	Router interface {
		Component
		RouterCore
		RouterMethod
	}


	// std router
	RouterStd struct {
		RouterCore
		RouterMethod
	}
	RouterMethodStd struct {
		RouterCore
		prefix 	string
	}
	RouterEmpty struct {
		Middleware
		RouterMethod
	}
	// router config
	// 存储路由配置，用于构造路由。
	RouterConfig struct {
		Type		string
		Path		string
		Method		string
		Handler		string
		Middleware  []string
		Router		[]*RouterConfig
	}
)

// check RouterStd has Router interface
var _ Router		=	&RouterStd{}
var _ RouterMethod	=	&RouterMethodStd{}


// new router
func NewRouter(name string, arg interface{}) (Router, error) {
	name = AddComponetPre(name, "router")
	c, err := NewComponent(name, arg)
	if err != nil {
		return nil, err
	}
	r, ok := c.(Router)
	if ok {
		return r, nil
	}
	return nil, fmt.Errorf("Component %s cannot be converted to Router type", name)
}

func NewRouterMust(name string, arg interface{}) Router {
	r, err := NewRouter(name, arg)
	if err != nil {
		panic(err)
	}
	return r
}

// Create a router component of the same type based on the router.
//
// 根据路由器创建一个类型相同的路由器组件。
func NewRouterClone(r Router) Router {
	return NewRouterMust(r.GetName(), nil)
}

// 未实现。
func SetRouterConfig(r Router, c *RouterConfig) error {
	// add Middleware
	for _, i := range c.Middleware {
		r.RegisterMiddleware(ConfigLoadMiddleware(i))
	}
	// add route
	if len(c.Type) == 0 {
		r.RegisterHandler(c.Method, c.Path, ConfigLoadHandleFunc(c.Handler))
		return nil
	} 
	r2, err := NewRouter(c.Type, c)
	if err != nil {
		return err
	}
	for _, i := range c.Router {
		SetRouterConfig(r2, i)
	}
	// r.SubRoute(c.Path, r2)
	return nil 
}

func (*RouterStd) GetName() string {
	return ComponentRouterStdName
}

func (*RouterStd) Version() string {
	return ComponentRouterStdVersion
}



func NewRouterEmpty(arg interface{}) (Router, error) {
	m, ok := arg.(Middleware)
	if !ok {
		h, ok := arg.(Handler)
		if !ok {
			h = HandlerFunc(HandleEmpty)
		}
		m = NewMiddlewareBase(h)
	}
	r := &RouterEmpty{Middleware:	m}
	r.RouterMethod = &RouterMethodStd{RouterCore: r}
	return r, nil
}

func (*RouterEmpty) RegisterMiddleware(...Handler) {
	// Do nothing because empty router does not process entries.
}
func (*RouterEmpty) RegisterHandler(method string, path string, handler Handler) {
	// Do nothing because empty router does not process entries.
}
func (*RouterEmpty) Match(string, string, Params) (Middleware ) {
	// Do nothing because empty router does not process entries.
	return nil
}

func (*RouterEmpty) GetName() string {
	return ComponentRouterEmptyName
}

func (*RouterEmpty) Version() string {
	return ComponentRouterEmptyVersion
}



func (m *RouterMethodStd) Register(mr RouterCore) {
	m.RouterCore = mr
}

func (m *RouterMethodStd) SubRoute(path string, router Router) {
	m.RegisterHandler(MethodAny, path, router)
}

func (m *RouterMethodStd) Group(prefix string) RouterMethod {
	return &RouterMethodStd{
		RouterCore:	m.RouterCore,
		prefix:		prefix,
	}
}

func (m *RouterMethodStd) AddHandler(hs ...Handler) {
	m.RegisterMiddleware(hs...)
}

// Router Register handler
func (m *RouterMethodStd) Any(path string, h Handler) {
	m.RegisterHandler(MethodAny, m.prefix + path, h)
}

func (m *RouterMethodStd) Get(path string, h Handler) {
	m.RegisterHandler(MethodGet, m.prefix + path, h)
}

func (m *RouterMethodStd) Post(path string, h Handler) {
	m.RegisterHandler(MethodPost, m.prefix + path, h)
}

func (m *RouterMethodStd) Put(path string, h Handler) {
	m.RegisterHandler(MethodPut, m.prefix + path, h)
}

func (m *RouterMethodStd) Delete(path string, h Handler) {
	m.RegisterHandler(MethodDelete, m.prefix + path, h)
}

func (m *RouterMethodStd) Head(path string, h Handler) {
	m.RegisterHandler(MethodHead, m.prefix + path, h)
}

func (m *RouterMethodStd) Patch(path string, h Handler) {
	m.RegisterHandler(MethodPatch, m.prefix + path, h)
}

func (m *RouterMethodStd) Options(path string, h Handler) {
	m.RegisterHandler(MethodOptions, m.prefix + path, h)
}


// RouterRegister handle func
func (m *RouterMethodStd) AnyFunc(path string, h HandlerFunc) {
	m.RegisterHandler(MethodAny, m.prefix + path, h)
}

func (m *RouterMethodStd) GetFunc(path string, h HandlerFunc) {
	m.RegisterHandler(MethodGet, m.prefix + path, h)
}

func (m *RouterMethodStd) PostFunc(path string, h HandlerFunc) {
	m.RegisterHandler(MethodPost, m.prefix + path, h)
}

func (m *RouterMethodStd) PutFunc(path string, h HandlerFunc) {
	m.RegisterHandler(MethodPut, m.prefix + path, h)
}

func (m *RouterMethodStd) DeleteFunc(path string, h HandlerFunc) {
	m.RegisterHandler(MethodDelete, m.prefix + path, h)
}

func (m *RouterMethodStd) HeadFunc(path string, h HandlerFunc) {
	m.RegisterHandler(MethodHead, m.prefix + path, h)
}

func (m *RouterMethodStd) PatchFunc(path string, h HandlerFunc) {
	m.RegisterHandler(MethodPatch, m.prefix + path, h)
}

func (m *RouterMethodStd) OptionsFunc(path string, h HandlerFunc) {
	m.RegisterHandler(MethodOptions, m.prefix + path, h)
}