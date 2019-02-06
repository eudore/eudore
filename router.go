package eudore

import (
	"fmt"
	"sort"
	"strings"
	"net/http"
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
		AllRouterMethod() []string
		Any(string, Handler) Handler
		AnyFunc(string, HandlerFunc) Handler
		Delete(string, Handler) Handler
		DeleteFunc(string, HandlerFunc) Handler
		Get(string, Handler) Handler
		GetFunc(string, HandlerFunc) Handler
		Head(string, Handler) Handler
		HeadFunc(string, HandlerFunc) Handler
		Options(string, Handler) Handler
		OptionsFunc(string, HandlerFunc) Handler
		Patch(string, Handler) Handler
		PatchFunc(string, HandlerFunc) Handler
		Post(string, Handler) Handler
		PostFunc(string, HandlerFunc) Handler
		Put(string, Handler) Handler
		PutFunc(string, HandlerFunc) Handler
	}
	// Router Register
	RouterRegister interface {
		RegisterFunc(method string, path string, handle HandlerFunc) Handler
		RegisterHandler(method string, path string, handler Handler) Handler
		RegisterSubRoute(path string, router Router) Handler
		RegisterHandlers(...Handler) []Handler
	}
	// router
	Router interface {
		Component
		Handler
		RouterMethod
		RouterRegister
		// method path
		Match(string, string, Params) ([]Handler, string)
		GetSubRouter(string) Router
		NotFoundFunc(Handler)
	}
	// router config
	// 存储路由配置，用于构造路由。
	RouterConfig struct {
		Type		string				`json:",omitempty"`
		Path		string
		Method		string				`json:",omitempty"`
		Handler		string				`json:",omitempty"`
		Router		[]*RouterConfig		`json:",omitempty"`
	}

	RouterMethodStd struct {
		RouterRegister					`json:"-" yaml:"-"`
	}
	// std router
	RouterStd struct {
		RouterMethod
		Routes			map[string][]*routeStd
		handlers		[]Handler		`json:"-" yaml:"-"`
		notFound		Handler			`json:"-" yaml:"-"`
	}
	routeStd struct {
		Path		string		`description:"route path."`
		Size		int
		Atts		[]int
		Tags		[]string
		keys		[]string
		vals		[]string
		Sub			Router		`json:"-" yaml:"-"`
		handler		Handler		`json:"-" yaml:"-"`
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


func setRouter(r Router, c *RouterConfig) error {
	for _, i := range c.Router {
		if len(i.Handler) != 0 {
			r.RegisterFunc(i.Method, i.Path, ConfigLoadHandleFunc(i.Handler))
		}
		r2, err := NewRouter(i.Type, i)
		if err != nil {
			return err
		}
		r.RegisterSubRoute(i.Path, r2)
	}
	return nil 
}

func NewRouterMust(name string, arg interface{}) Router {
	r, err := NewRouter(name, arg)
	if err != nil {
		panic(err)
	}
	return r
}

func NewRouterClone(r Router) Router {
	return NewRouterMust(r.GetName(), nil)
}

func GetSubRouter(r Router, path string) Router {
	return r.GetSubRouter(path)
}

/*func GetSubRouter(r Router, path string) Router {
	matchcontext.request.Method = MethodAny
	matchcontext.request.URL.Path = path
	h, _ := r.Match(matchcontext)
	if r2, ok := h.(SubRouter);ok {
		r = r2.SubRoute()
		// if r3 := GetSubRouter(r, matchcontext.request.URL.Path); r3 != nil {
		// 	return r3
		// }
		return r
	}
	return nil
}*/

// func GetThisRouter(r Router, ctx Context) Router {
// 	ctx.URL().Path = ctx.Path()
// 	for {
// 		h, _ := r.Match(ctx)
// 		if r2, ok := h.(SubRouter);ok {
// 			r = r2.SubRoute()
// 		}else {
// 			break
// 		}
// 	}
// 	return r
// }











// Create a basic route handler with component name: "router-std"
//
// 创建一个基础z路由处理器，组件名称：“router-std”
func NewRouterStd(interface{}) (Router, error) {
	r := &RouterStd{
		Routes:		make(map[string][]*routeStd),
	}
	r.notFound = HandlerFunc(r.DefaultNotfound)
	r.RouterMethod = &RouterMethodStd{
		RouterRegister:		r,
	}
	return r, nil
}

// default handle func
/*func (r *RouterStd) DefaultHandle(ctx Context) {
	h, _ := r.Match(ctx)
	if h == nil {
		h = r.notFound
	}
	h.Handle(ctx)	
}*/

// default not found func
func (r *RouterStd) DefaultNotfound(ctx Context) {
	ctx.WriteHeader(http.StatusNotFound)
	ctx.Write([]byte("404 page not found"))
}


func (r *RouterStd) Handle(ctx Context) {
	// hs, _ := r.Match(ctx)
	// for _
	// if h != nil {
	// 	h.Handle(ctx)
	// }
}

func (r *RouterStd) Match(method, path string, params Params) ([]Handler, string) {
	// check register method
	rs, ok := r.Routes[method]
	if !ok {
		return nil, "405"
	}
	
	// each method router
	for _, r2 := range rs {
		if r2.match(path) {
			if r2.Sub != nil && path != r2.Path {
				h, path := r2.Sub.Match(method, path[len(r2.Path):], params)
				if len(h) > 0 {
					return append(r.handlers, h... ), r2.Path + " " + path
				}
			}
			r2.addArgs(params, path)
			return append(r.handlers, r2.handler ), r2.Path
		}
	}
	return nil, "404"
	// return r.notFound, "404"
}

func (r *RouterStd) GetSubRouter(path string) Router {
	rs, ok := r.Routes[MethodAny]
	if !ok {
		return nil
	}
	for _, r2 := range rs {
		if r2.match(path) {
			if r2.Sub != nil  {
				if path == r2.Path {
					return r2.Sub
				}
				if strings.HasPrefix(path, r2.Path) {
					return r2.Sub.GetSubRouter(path[len(r2.Path):])
				}
			}else {
				return r
			}
		}
	}
	return nil
}
/*
func (r *RouterStd) matchRoute(method , path string) (Handler, string) {
	rs, ok := r.Routes[ctx.Method()]
	if !ok {
		return nil, "405"
	}
	for _, r2 := range rs {
		if r2.matchpath(ctx) {
			return r2
		}
	}
	return nil
}*/


func (r *RouterStd) RegisterHandlers(h ...Handler) []Handler {
	if len(h) > 0 {
		r.handlers = append(r.handlers, h...)
	}
	return r.handlers
}



func (r *RouterStd) NotFoundFunc(h Handler) {
	r.notFound = h
}


func (r *RouterStd) RegisterHandler(method string, path string, handler Handler) Handler {
	route := newStdRoute(path, handler)
	r.add(method, route)	
	if method == MethodAny {
		for _, i := range r.AllRouterMethod() {
			r.add(i, route)
		}
	}
	return route
}

func (r *RouterStd) RegisterFunc(method string, path string, handle HandlerFunc) Handler {
	return r.RegisterHandler(method, path, HandlerFunc(handle))
}

func (r *RouterStd) RegisterSubRoute(path string, router Router) Handler {
	r2 := r.RegisterHandler(MethodAny, path, router)
	if route, ok := r2.(*routeStd); ok {
		route.Sub = router
	}
	return r2
}

func (r *RouterStd) add(method string, route *routeStd) {
	routes := append(r.Routes[method] , route)
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








type Token struct {
	raw    []int
	Tokens []string
	Size   int
}



func newStdRoute(path string, h Handler) *routeStd {
	r := &routeStd{handler:	h}
	args := strings.Split(path, " ")
	path = args[0]
	r.keys = make([]string, len(args) - 1)
	r.vals = make([]string, len(args) - 1)
	for i, k := range args[1:] {
		r.keys[i], r.vals[i] = split2(k, ":")
	}
	r.Path = path
	if strings.HasSuffix(path, "/") {
		path = path + "*"
	}
	ss := strings.Split(path, "/")
	var atts = make([]int, len(ss))
	var tags = make([]string, len(ss))
	for i, s := range ss {
		if len(s) > 0 {
			switch s[0] {
			case ':':
				atts[i] = PARAM
				tags[i] = s[1:]
			case '#':
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

func (r *routeStd) Handle(ctx Context) {
	r.handler.Handle(ctx)
}

// func (r *routeStd) matchpath(path string) bool {
// 	return r.Path == path || (r.Dir && strings.HasPrefix(path, r.Path))
// }


func (r *routeStd) SubRoute() Router {
	return r.Sub
}




// func (r *RouterStd) Match(ctx Context) (Handler, string) {
// 	// check register method
// 	rs, ok := r.Routes[ctx.Method()]
// 	if !ok {
// 		return nil
// 	}
// 	// each method router
// 	for _, r2 := range rs {
// 		if r2.Match(ctx) {
// 			if r2.Sub != nil {
// 				// clean last router url
// 				ctx.URL().Path = ctx.URL().Path[len(r2.Path):]
// 			}
// 			return r2
// 		}
// 	}
// 	return nil
// }























func (m *RouterMethodStd) AllRouterMethod() []string {
	// return []string{MethodAny, MethodGet, MethodPost, MethodPut, MethodDelete}
	return []string{MethodAny, MethodGet, MethodPost, MethodPut, MethodDelete, MethodHead, MethodPatch, MethodOptions}
}

func (m *RouterMethodStd) Register(mr RouterRegister) {
	m.RouterRegister = mr
}

// RouterRegister handler
func (m *RouterMethodStd) Any(path string, h Handler) Handler {
	return m.RouterRegister.RegisterHandler(MethodAny, path, h)
}

func (m *RouterMethodStd) Get(path string, h Handler) Handler {
	return m.RouterRegister.RegisterHandler(MethodGet, path, h)
}

func (m *RouterMethodStd) Post(path string, h Handler) Handler {
	return m.RouterRegister.RegisterHandler(MethodPost, path, h)
}

func (m *RouterMethodStd) Put(path string, h Handler) Handler {
	return m.RouterRegister.RegisterHandler(MethodPut, path, h)
}

func (m *RouterMethodStd) Delete(path string, h Handler) Handler {
	return m.RouterRegister.RegisterHandler(MethodDelete, path, h)
}

func (m *RouterMethodStd) Head(path string, h Handler) Handler {
	return m.RouterRegister.RegisterHandler(MethodHead, path, h)
}

func (m *RouterMethodStd) Patch(path string, h Handler) Handler {
	return m.RouterRegister.RegisterHandler(MethodPatch, path, h)
}

func (m *RouterMethodStd) Options(path string, h Handler) Handler {
	return m.RouterRegister.RegisterHandler(MethodOptions, path, h)
}


// RouterRegister handle func
func (m *RouterMethodStd) AnyFunc(path string, h HandlerFunc) Handler {
	return m.RouterRegister.RegisterFunc(MethodAny, path, h)	
}

func (m *RouterMethodStd) GetFunc(path string, h HandlerFunc) Handler {
	return m.RouterRegister.RegisterFunc(MethodGet, path, h)
}

func (m *RouterMethodStd) PostFunc(path string, h HandlerFunc) Handler {
	return m.RouterRegister.RegisterFunc(MethodPost, path, h)
}

func (m *RouterMethodStd) PutFunc(path string, h HandlerFunc) Handler {
	return m.RouterRegister.RegisterFunc(MethodPut, path, h)
}

func (m *RouterMethodStd) DeleteFunc(path string, h HandlerFunc) Handler {
	return m.RouterRegister.RegisterFunc(MethodDelete, path, h)
}

func (m *RouterMethodStd) HeadFunc(path string, h HandlerFunc) Handler {
	return m.RouterRegister.RegisterFunc(MethodHead, path, h)
}

func (m *RouterMethodStd) PatchFunc(path string, h HandlerFunc) Handler {
	return m.RouterRegister.RegisterFunc(MethodPatch, path, h)
}

func (m *RouterMethodStd) OptionsFunc(path string, h HandlerFunc) Handler {
	return m.RouterRegister.RegisterFunc(MethodOptions, path, h)
}

