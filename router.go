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
		AddHandler(...HandlerFunc)
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
		// Middleware
		// RegisterMiddleware(...Handler)
		RegisterHandler(method string, path string, handler HandlerFuncs)
		Match(string, string, Params) HandlerFuncs
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
		prefix		string
		handlers	HandlerFuncs
	}
	RouterEmpty struct {
		// Middleware
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
/*func SetRouterConfig(r Router, c *RouterConfig) error {
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
*/
func RouterDefault405Func(ctx Context) {
	const page405 string = "405 method not allowed"
	ctx.Response().Header().Add("Allow", "HEAD, GET, POST, PUT, DELETE, PATCH")
	ctx.WriteHeader(405)
	ctx.WriteString(page405)
}

func RouterDefault404Func(ctx Context) {
	const page404 string = "404 page not found"
	ctx.WriteHeader(404)
	ctx.WriteString(page404)
}

func (*RouterStd) GetName() string {
	return ComponentRouterStdName
}

func (*RouterStd) Version() string {
	return ComponentRouterStdVersion
}



func NewRouterEmpty(arg interface{}) (Router, error) {
	// m, ok := arg.(Middleware)
	// if !ok {
	// 	h, ok := arg.(Handler)
	// 	if !ok {
	// 		h = HandlerFunc(HandleEmpty)
	// 	}
	// 	m = NewMiddlewareBase(h)
	// }
	// r := &RouterEmpty{Middleware:	m}
	r := &RouterEmpty{}
	r.RouterMethod = &RouterMethodStd{RouterCore: r}
	return r, nil
}

/*func (*RouterEmpty) RegisterMiddleware(...Handler) {
	// Do nothing because empty router does not process entries.
}*/
func (*RouterEmpty) RegisterHandler(method string, path string, handler HandlerFuncs) {
	// Do nothing because empty router does not process entries.
}
func (*RouterEmpty) Match(string, string, Params) (HandlerFuncs) {
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

func (m *RouterMethodStd) Group(prefix string) RouterMethod {
	return &RouterMethodStd{
		RouterCore:	m.RouterCore,
		prefix:		m.prefix + prefix,
		handlers:	m.handlers,
	}
}

func (m *RouterMethodStd) AddHandler(hs ...HandlerFunc) {
	m.handlers =m.combineHandlers(hs)
}

func (m *RouterMethodStd) RegisterHandlers(method ,path string, hs ...HandlerFunc) {
	m.RouterCore.RegisterHandler(method, m.prefix + path, m.combineHandlers(hs))
}

func (m *RouterMethodStd) combineHandlers(handlers HandlerFuncs) HandlerFuncs {
	const abortIndex int8 = 63
	finalSize := len(m.handlers) + len(handlers)
	if finalSize >= int(abortIndex) {
		panic("too many handlers")
	}
	mergedHandlers := make(HandlerFuncs, finalSize)
	copy(mergedHandlers, m.handlers)
	copy(mergedHandlers[len(m.handlers):], handlers)
	return mergedHandlers
}

// Router Register handler
func (m *RouterMethodStd) Any(path string, h Handler) {
	m.RegisterHandlers(MethodAny, path, h.Handle)
}

func (m *RouterMethodStd) Get(path string, h Handler) {
	m.RegisterHandlers(MethodGet, path, h.Handle)
}

func (m *RouterMethodStd) Post(path string, h Handler) {
	m.RegisterHandlers(MethodPost, path, h.Handle)
}

func (m *RouterMethodStd) Put(path string, h Handler) {
	m.RegisterHandlers(MethodPut, path, h.Handle)
}

func (m *RouterMethodStd) Delete(path string, h Handler) {
	m.RegisterHandlers(MethodDelete, path, h.Handle)
}

func (m *RouterMethodStd) Head(path string, h Handler) {
	m.RegisterHandlers(MethodHead, path, h.Handle)
}

func (m *RouterMethodStd) Patch(path string, h Handler) {
	m.RegisterHandlers(MethodPatch, path, h.Handle)
}

func (m *RouterMethodStd) Options(path string, h Handler) {
	m.RegisterHandlers(MethodOptions, path, h.Handle)
}


// RouterRegister handle func
func (m *RouterMethodStd) AnyFunc(path string, h HandlerFunc) {
	m.RegisterHandlers(MethodAny, path, h)
}

func (m *RouterMethodStd) GetFunc(path string, h HandlerFunc) {
	m.RegisterHandlers(MethodGet, path, h)
}

func (m *RouterMethodStd) PostFunc(path string, h HandlerFunc) {
	m.RegisterHandlers(MethodPost, path, h)
}

func (m *RouterMethodStd) PutFunc(path string, h HandlerFunc) {
	m.RegisterHandlers(MethodPut, path, h)
}

func (m *RouterMethodStd) DeleteFunc(path string, h HandlerFunc) {
	m.RegisterHandlers(MethodDelete, path, h)
}

func (m *RouterMethodStd) HeadFunc(path string, h HandlerFunc) {
	m.RegisterHandlers(MethodHead, path, h)
}

func (m *RouterMethodStd) PatchFunc(path string, h HandlerFunc) {
	m.RegisterHandlers(MethodPatch, path, h)
}

func (m *RouterMethodStd) OptionsFunc(path string, h HandlerFunc) {
	m.RegisterHandlers(MethodOptions, path, h)
}