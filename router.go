package eudore

import (
	"fmt"
	"sort"
	"strings"
)

// router-std component, route parameter type flag
//
// router-std组件，路由参数类型标志
const (
	CONST = 2 << iota
	//PARAM value store in Atts if the route have parameters
	PARAM
	//SUB value store in Atts if the route is a sub router
	SUB
	//WC value store in Atts if the route have wildcard
	WC
	//REGEX value store in Atts if the route contains regex
	REGEX
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
		SubRoute(path string, router Router)
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
		Match(Params) Middleware
	}
	// router
	Router interface {
		Component
		RouterCore
		RouterMethod
	}


	// std router
	RouterStd struct {
		RouterCore		`json:"-" yaml:"-"`
		RouterMethod	`json:"-" yaml:"-"`
	}
	RouterStdCore struct {
		Middleware		`json:"-" yaml:"-"`
		Routes			map[string][]*routeStd	
		head			Middleware		`json:"-" yaml:"-"`
	}
	RouterStdMethod struct {
		RouterCore					`json:"-" yaml:"-"`
	}
	routeStd struct {
		Path		string		`description:"route path."`
		Size		int
		Atts		[]int
		Tags		[]string
		keys		[]string
		vals		[]string
		isAny 		bool
		Sub			Router		`json:"-" yaml:"-"`
		Handler		Middleware	`json:"-" yaml:"-"`
	}
	MiddlewareRouter struct {
		RouterCore
		Next Middleware
	}
	// router config
	// 存储路由配置，用于构造路由。
	RouterConfig struct {
		Type		string				`json:",omitempty"`
		Path		string
		Method		string				`json:",omitempty"`
		Handler		string				`json:",omitempty"`
		Middleware  []string			`json:",omitempty"`
		Router		[]*RouterConfig		`json:",omitempty"`
	}
)

// check RouterStd has Router interface
var _ Router		=	&RouterStd{}
var _ RouterMethod	=	&RouterStdMethod{}


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

// Returns a subroute that the router matches.
//
// If it does not match to null.
//
// 返回路由器匹配的一个子路由。
//
// 如果未匹配到空。
func GetSubRouter(r Router, path string) Router {
	return GetSubRouterMethod(r, MethodAny, path)
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
	r.SubRoute(c.Path, r2)
	return nil 
}

// Returns a sub-router under a method.
//
// 返回一个方法下的子路由器。
func GetSubRouterMethod(r Router, method, path string) Router {
	p := make(Params)
	p.Set(ParamRouteMethod, method)
	p.Set(ParamRoutePath, path)
	if r2 := getSubRouter(r , p); r2 != r {
		return r2
	}
	return nil
}

func getSubRouter(r Router, params Params) Router {
	if len(params.Get(ParamRoutePath)) == 0 {
		return r
	}
	r2 := r.Match(params)
	if r3, ok := r2.(Router); ok {
		return getSubRouter(r3, params)
	}
	return r
}


// Match a handler and directly use it with the Context object.
//
// Then set the tail handler appended by the SetNext method to be the follower.
//
// 匹配出一个处理者，并直接给Context对象并使用。
//
// 然后设置SetNext方法追加的尾处理者为后续处理者。
func (m *MiddlewareRouter) Handle(ctx Context) {
	ctx.SetHandler(m.Match(ctx.Params()))
	ctx.Next()
	ctx.SetHandler(m.Next)
	ctx.Next()
}

// The return processing middleware is nil.
//
// The router is stateless and cannot return directly to the next handler.
//
// When the router processes it, it will match the next handler and directly use it for the Context object.
//
// 返回处理中间件为空。
//
// 路由器是无状态的，无法直接返回下一个处理者。
//
// 在路由器处理时会匹配出下一个处理者，并直接给Context对象并使用。
func (m *MiddlewareRouter) GetNext() Middleware {
	return nil
}

// Set the post-processing chain after the route is processed.
//
// 设置路由处理完后的后序处理链。
func (m *MiddlewareRouter) SetNext(nm Middleware) {
	// 请求尾处理
	if nm == nil {
		m.Next = nil
		return
	}
	// 尾追加处理中间件
	link := m.Next
	n := link.GetNext();
	for n != nil {
		link = n
		n = link.GetNext();
	}
	link.SetNext(nm)
}

// Create a basic route handler with component name: "router-std".
//
// 创建一个基础路由处理器，组件名称：“router-std”。
func NewRouterStd(interface{}) (Router, error) {
	rc := NewRouterStdCore()
	return &RouterStd{
		RouterCore:		rc,
		RouterMethod:	&RouterStdMethod{
			RouterCore:		rc,
		},
	}, nil
}

func NewRouterStdCore() RouterCore {
	rc := &RouterStdCore{
		Routes:		make(map[string][]*routeStd),

	}
	rc.Middleware = &MiddlewareRouter{
		RouterCore:	rc,
		Next:	nil,
	}
	return rc
}

// Use a router to process a Context object.
//
// Set the handler for the Context object and turn it on.
//
// 使用路由器处理一个Context对象。
//
// 给Context对象设置好处理者，然后开启处理。
func (r *RouterStdCore) Handle(ctx Context) {
	ctx.SetHandler(r.GetNext())
	ctx.Next()
}

// Returns the first handler of the router.
//
// If no pre-match handler is registered, it will return directly to the router's processor.
//
// 返回路由器的第一个处理者。
//
// 如果没有注册匹配前处理者，会直接返回路由器的处理器。
func (r *RouterStdCore) GetNext() Middleware {
	if r.head == nil {
		return r.Middleware
	}
	return r.head
}

// 根据输出的参数匹配返回一个处理中间件。
//
// 需要ParamRouteMethod和ParamRoutePath参数，子路由会截取ParamRoutePath的值。
//
// 同时会给params添加路由的相关参数。
func (r *RouterStdCore) Match(params Params) Middleware {
	// check register method
	rs, ok := r.Routes[params.Get(ParamRouteMethod)]
	if !ok {
		return nil
	}
	// each method router
	path := params.Get(ParamRoutePath)
	for _, r2 := range rs {
		if r2.match(path) {
			// 增加路由参数
			r2.addArgs(params, path)
			if r2.Sub != nil {
				// 修改匹配路由路径
				params.Set(ParamRoutePath, path[len(r2.Path):])
				return r2.Sub
			}
			return r2.Handler
		}
	}
	return nil
	// return r.notFound, "404"
}



// Returns all routing methods supported by RouterStd.
//
// 返回RouterStd支持的全部路由方法。
func (*RouterStdCore) AllRouterMethod() []string {
	// return []string{MethodAny, MethodGet, MethodPost, MethodPut, MethodDelete}
	return []string{MethodAny, MethodGet, MethodPost, MethodPut, MethodDelete, MethodHead, MethodPatch, MethodOptions}
}

// RouterStd adds routing matching pre-processing middleware.
//
// SetNext(Middleware) can be added to the middleware after the matching process, but it is not recommended.
//
// The execution order is RegisterMiddleware (hs ... Handler) registered in order,
// Match (Params) Middleware return handler, SetNext (Middleware) registered in order
//
// RouterStd增加路由匹配前处理中间件。
//
// SetNext(Middleware) 可加入匹配处理完后中间件，但不推荐使用。
//
// 执行顺序为RegisterMiddleware(hs ...Handler)按顺序注册、Match(Params) Middleware返回的处理者、SetNext(Middleware)按顺序注册
func (r *RouterStdCore) RegisterMiddleware(hs ...Handler) {
	ml := NewMiddlewareLink(hs...)
	r.getend(ml).SetNext(r.Middleware)
	if r.head == nil {
		r.head = ml
	}else {
		r.getend(r.head).SetNext(ml)
	}
}

// return the last processing middleware of the parameter link
//
// 返回参数link的最后一个处理中间件
func (r *RouterStdCore) getend(link Middleware) Middleware {
	if link == nil {
		link = r.head
	}
	next := link.GetNext()
	for next != r.Middleware && next != nil {
		link = next
		next = link.GetNext()
	}
	return link
}

// Register a handler for a method path.
//
// The handler is converted to the Middleware type for the Handler type, 
//
// and the handler is registered as the child route for the Router type.
//
// 给一个方法路径注册一个处理者。
//
// handler为Handler类型会转换成Middleware类型，handler为Router类型会注册为子路由。
func (r *RouterStdCore) RegisterHandler(method string, path string, handler Handler) {
	route := newRouteStd(path, handler)
	route.isAny = method == MethodAny
	// Any方法注册全部方法路由
	if method == MethodAny {
		for _, i := range r.AllRouterMethod() {
			r.addroute(i, route)
		}
	}else {
		r.addroute(method, route)
	}
}

// Used to add a route to the specified method.
//
// Overwrite the original route if a route exists.
//
// If the new route is the Any method and the old route is not the Any method,
//
// The route is not added, and the Any method is prohibited from overwriting individual methods.
//
// 用于对指定方法添加路由。
//
// 如果路由存在，则覆盖原路由。
//
// 如果新路由是Any方法且旧路由非Any方法，
//
// 则不会添加路由，禁止Any方法覆盖单独方法。
func (r *RouterStdCore) addroute(method string, route *routeStd) {
	for i, rs := range r.Routes[method] {
		if rs.Path == route.Path {
			// 如果新路由是Any方法 旧路由非Any方法
			// 禁止Any方法覆盖单独方法
			if !rs.isAny && route.isAny {
				return
			}
			// 覆盖旧路径
			r.Routes[method][i] = route
			return
		}
	}
	// 追加新的路由
	routes := append(r.Routes[method] , route)
	// 对路由顺序排序
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].Path > routes[j].Path
	})
	r.Routes[method] = routes
}


func (r *RouterStd) GetName() string {
	return ComponentRouterStdName
}

func (r *RouterStd) Version() string {
	return ComponentRouterStdVersion
}








func newRouteStd(path string, h Handler) *routeStd {
	r := &routeStd{}
	// set handle
	switch m := h.(type) {
	case Router:
		r.Sub = m 
	case Middleware:
		r.Handler = m
	default:
		r.Handler = NewMiddlewareBase(h)
	}
	// handle args
	args := strings.Split(path, " ")
	path = args[0]
	r.keys = make([]string, len(args))
	r.vals = make([]string, len(args))
	for i, k := range args {
		r.keys[i], r.vals[i] = split2(k, ":")
	}
	r.keys[0], r.vals[0] = ParamRoutes, path
	// 修正路由规则
	if strings.HasSuffix(path, "/") {
		path = path + "*"
	}
	r.Path = path
	// set tags
	ss := strings.Split(path, "/")
	var atts = make([]int, len(ss))
	var tags = make([]string, len(ss))
	for i, s := range ss {
		if len(s) > 0 {
			switch s[0] {
			case ':':
				atts[i] = PARAM
				tags[i] = s[1:]
			// case '#':
				// r.[i] = REGEX
				// r.Atts |= REGEX
			case '*':
				atts[i] = WC
				if len(s) > 1 {
					tags[i] = s[1:]
				}else {
					tags[i] = "*"
				}
				break
			default:
				atts[i] = CONST
				tags[i] = s
			}
		}
		r.Size++
	}
	// set data
	r.Atts = make([]int, r.Size)
	r.Tags = make([]string, r.Size)
	copy(r.Atts, atts)
	copy(r.Tags, tags)
	return r
}

func (r *routeStd) match(path string) bool {
	ss := strings.Split(path, "/")
	if len(ss) < r.Size {
		return false
	}
	for i, v := range r.Atts {
		switch v {
		case CONST:
			if r.Tags[i] != ss[i] {
				return false
			}
		// case PARAM:
		case WC:
			return true
		}
	}
	return true
}

func (r *routeStd) addArgs(params Params, path string) {
	// default param
	for i, k := range r.keys {
		params.Add(k, r.vals[i])
	}
	// route param
	ss := strings.Split(path, "/")
	for i, v := range r.Atts {
		switch v {
		case PARAM:
			params.Add(r.Tags[i], ss[i])
		case WC:
			params.Add(r.Tags[i], strings.Join(ss[i:], "/"))
		}
	}
}





















func (m *RouterStdMethod) Register(mr RouterCore) {
	m.RouterCore = mr
}

func (m *RouterStdMethod) SubRoute(path string, router Router) {
	m.RegisterHandler(MethodAny, path, router)
}

func (m *RouterStdMethod) AddHandler(hs ...Handler) {
	m.RegisterMiddleware(hs...)
}

// Router Register handler
func (m *RouterStdMethod) Any(path string, h Handler) {
	m.RegisterHandler(MethodAny, path, h)
}

func (m *RouterStdMethod) Get(path string, h Handler) {
	m.RegisterHandler(MethodGet, path, h)
}

func (m *RouterStdMethod) Post(path string, h Handler) {
	m.RegisterHandler(MethodPost, path, h)
}

func (m *RouterStdMethod) Put(path string, h Handler) {
	m.RegisterHandler(MethodPut, path, h)
}

func (m *RouterStdMethod) Delete(path string, h Handler) {
	m.RegisterHandler(MethodDelete, path, h)
}

func (m *RouterStdMethod) Head(path string, h Handler) {
	m.RegisterHandler(MethodHead, path, h)
}

func (m *RouterStdMethod) Patch(path string, h Handler) {
	m.RegisterHandler(MethodPatch, path, h)
}

func (m *RouterStdMethod) Options(path string, h Handler) {
	m.RegisterHandler(MethodOptions, path, h)
}


// RouterRegister handle func
func (m *RouterStdMethod) AnyFunc(path string, h HandlerFunc) {
	m.RegisterHandler(MethodAny, path, h)
}

func (m *RouterStdMethod) GetFunc(path string, h HandlerFunc) {
	m.RegisterHandler(MethodGet, path, h)
}

func (m *RouterStdMethod) PostFunc(path string, h HandlerFunc) {
	m.RegisterHandler(MethodPost, path, h)
}

func (m *RouterStdMethod) PutFunc(path string, h HandlerFunc) {
	m.RegisterHandler(MethodPut, path, h)
}

func (m *RouterStdMethod) DeleteFunc(path string, h HandlerFunc) {
	m.RegisterHandler(MethodDelete, path, h)
}

func (m *RouterStdMethod) HeadFunc(path string, h HandlerFunc) {
	m.RegisterHandler(MethodHead, path, h)
}

func (m *RouterStdMethod) PatchFunc(path string, h HandlerFunc) {
	m.RegisterHandler(MethodPatch, path, h)
}

func (m *RouterStdMethod) OptionsFunc(path string, h HandlerFunc) {
	m.RegisterHandler(MethodOptions, path, h)
}