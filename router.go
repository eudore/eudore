/*
Router

Router对象用于定义请求的路由

文件：router.go routerRadix.go routerFull.go
*/
package eudore

import (
	"fmt"
	"strings"
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
	// The route is directly registered by default. Other methods can be directly registered using the RouterRegister interface.
	//
	// 路由默认直接注册的方法，其他方法可以使用RouterRegister接口直接注册。
	RouterMethod interface {
		Group(string) RouterMethod
		AddMiddleware(...HandlerFunc) RouterMethod
		Any(string, ...Handler)
		AnyFunc(string, ...HandlerFunc)
		Delete(string, ...Handler)
		DeleteFunc(string, ...HandlerFunc)
		Get(string, ...Handler)
		GetFunc(string, ...HandlerFunc)
		Head(string, ...Handler)
		HeadFunc(string, ...HandlerFunc)
		Options(string, ...Handler)
		OptionsFunc(string, ...HandlerFunc)
		Patch(string, ...Handler)
		PatchFunc(string, ...HandlerFunc)
		Post(string, ...Handler)
		PostFunc(string, ...HandlerFunc)
		Put(string, ...Handler)
		PutFunc(string, ...HandlerFunc)
	}
	// The router core interface, performs routing, middleware registration, and matches a request and returns to the handler.
	//
	// 路由器核心接口，执行路由、中间件的注册和匹配一个请求并返回处理者。
	RouterCore interface {
		RegisterMiddleware(string, string, HandlerFuncs)
		RegisterHandler(string, string, HandlerFuncs)
		Match(string, string, Params) HandlerFuncs
	}
	// Router interface, you need to set the component, router method, router core three interfaces.
	//
	// 路由器接口，需要设置组件、路由器方法、路由器核心三个接口。
	Router interface {
		Component
		RouterCore
		RouterMethod
	}

	// 默认路由器方法注册实现
	RouterMethodStd struct {
		RouterCore
		prefix		string
		tags		string
	}
	// 存储中间件信息的基数树。
	//
	// 用于内存存储路由器中间件注册信息，并根据注册路由返回对应的中间件。
	middTree struct {
		root		middNode
	}
	middNode struct {
		path		string
		children	[]*middNode
		key			string
		val			HandlerFuncs
	}
	// Storage router configuration for constructing routers.
	//
	// 存储路由器配置，用于构造路由器。
	RouterConfig struct {
		// Type		string
		Path		string
		Method		string
		Handler		string
		Middleware  []string
		Routes		[]*RouterConfig
	}
)


// check RouterStd has Router interface
var (
	_ Router		=	&RouterRadix{}
	_ Router		=	&RouterFull{}
	_ RouterMethod	=	&RouterMethodStd{}
	RouterAllMethod	=	[]string{MethodGet, MethodPost, MethodPut, MethodDelete, MethodHead, MethodPatch, MethodOptions}
)


// new router
func NewRouter(name string, arg interface{}) (Router, error) {
	name = ComponentPrefix(name, "router")
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
func SetRouterConfig(r RouterMethod, c *RouterConfig) {
	// add Middleware
	if len(c.Method) == 0 {
		c.Method = MethodAny
	}
	if len(c.Path) > 0 && len(c.Handler) > 0 {

	}
	if len(c.Path) > 0 {
		r = r.Group(c.Path)
	}
	for _, i := range c.Middleware {
		r.AddMiddleware(ConfigLoadHandleFunc(i))
	}
	for _, i := range c.Routes {
		SetRouterConfig(r, i)
	}
}

func DefaultRouter405Func(ctx Context) {
	const page405 string = "405 method not allowed"
	ctx.Response().Header().Add("Allow", "HEAD, GET, POST, PUT, DELETE, PATCH")
	ctx.WriteHeader(405)
	ctx.WriteString(page405)
}

func DefaultRouter404Func(ctx Context) {
	const page404 string = "404 page not found"
	ctx.WriteHeader(404)
	ctx.WriteString(page404)
}




func (m *RouterMethodStd) Register(mr RouterCore) {
	m.RouterCore = mr
}

func (m *RouterMethodStd) Group(path string) RouterMethod {
	// 将路径前缀和路径参数分割出来
	args := strings.Split(path, " ")
	prefix := args[0]
	tags := path[len(prefix):]
	// 如果路径是'/*'或'/'结尾，则移除后缀。
	// '/*'为路由结尾，不可为路由前缀
	// '/'不可为路由前缀，会导致出现'//'
	if len(prefix) > 0 && prefix[len(prefix) - 1] == '*' {
		prefix = prefix[:len(prefix) - 1]
	}
	if len(prefix) > 0 && prefix[len(prefix) - 1] == '/' {
		prefix = prefix[:len(prefix) - 1]
	}
	// 构建新的路由方法配置器
	return &RouterMethodStd{
		RouterCore:	m.RouterCore,
		prefix:		m.prefix + prefix,
		tags:		tags + m.tags,
	}
}

func (m *RouterMethodStd) AddMiddleware(hs ...HandlerFunc) RouterMethod {
	m.RegisterMiddleware(MethodAny, m.prefix + "/", hs)
	return m
}

func (m *RouterMethodStd) registerHandlers(method ,path string, hs HandlerFuncs) {
	m.RouterCore.RegisterHandler(method, m.prefix + path + m.tags, hs)
}

// Router Register handler
func (m *RouterMethodStd) Any(path string, h ...Handler) {
	m.registerHandlers(MethodAny, path, handlesToFunc(h))
}

func (m *RouterMethodStd) Get(path string, h ...Handler) {
	m.registerHandlers(MethodGet, path, handlesToFunc(h))
}

func (m *RouterMethodStd) Post(path string, h ...Handler) {
	m.registerHandlers(MethodPost, path, handlesToFunc(h))
}

func (m *RouterMethodStd) Put(path string, h ...Handler) {
	m.registerHandlers(MethodPut, path, handlesToFunc(h))
}

func (m *RouterMethodStd) Delete(path string, h ...Handler) {
	m.registerHandlers(MethodDelete, path, handlesToFunc(h))
}

func (m *RouterMethodStd) Head(path string, h ...Handler) {
	m.registerHandlers(MethodHead, path, handlesToFunc(h))
}

func (m *RouterMethodStd) Patch(path string, h ...Handler) {
	m.registerHandlers(MethodPatch, path, handlesToFunc(h))
}

func (m *RouterMethodStd) Options(path string, h ...Handler) {
	m.registerHandlers(MethodOptions, path, handlesToFunc(h))
}

func handlesToFunc(hs []Handler) HandlerFuncs {
	h := make(HandlerFuncs, len(hs))
	for i, _ := range hs {
		h[i] = hs[i].Handle
	}
	return h
}


// RouterRegister handle func
func (m *RouterMethodStd) AnyFunc(path string, h ...HandlerFunc) {
	m.registerHandlers(MethodAny, path, h)
}

func (m *RouterMethodStd) GetFunc(path string, h ...HandlerFunc) {
	m.registerHandlers(MethodGet, path, h)
}

func (m *RouterMethodStd) PostFunc(path string, h ...HandlerFunc) {
	m.registerHandlers(MethodPost, path, h)
}

func (m *RouterMethodStd) PutFunc(path string, h ...HandlerFunc) {
	m.registerHandlers(MethodPut, path, h)
}

func (m *RouterMethodStd) DeleteFunc(path string, h ...HandlerFunc) {
	m.registerHandlers(MethodDelete, path, h)
}

func (m *RouterMethodStd) HeadFunc(path string, h ...HandlerFunc) {
	m.registerHandlers(MethodHead, path, h)
}

func (m *RouterMethodStd) PatchFunc(path string, h ...HandlerFunc) {
	m.registerHandlers(MethodPatch, path, h)
}

func (m *RouterMethodStd) OptionsFunc(path string, h ...HandlerFunc) {
	m.registerHandlers(MethodOptions, path, h)
}




func (t *middNode) Insert(key string, val HandlerFuncs) {
	// fmt.Println("Insert:", key)
	t.recursiveInsertTree(key, key ,val)
}

//Lookup: Find if seachKey exist in current radix tree and return its value
func (t *middNode) Lookup(searchKey string) HandlerFuncs {
	searchKey = strings.Split(searchKey, " ")[0]
	if searchKey[len(searchKey) - 1] == '*' {
		searchKey = searchKey[:len(searchKey) - 1]
	}
	// fmt.Println("Lookup:", searchKey)
	return t.recursiveLoopup(searchKey)
}

// 新增Node
func (r *middNode) InsertNode(path, key string, value HandlerFuncs) {
	if len(path) == 0 {
		// 路径空就设置当前node的值
		r.key = key
		r.val = CombineHandlers(r.val, value)
	}else {
		// 否则新增node
		r.children = append(r.children, &middNode{path: path, key: key, val: value})
	}
}

// 对指定路径为edgeKey的Node分叉，公共前缀路径为pathKey
func (r *middNode) SplitNode(pathKey, edgeKey string) *middNode {
	for i, _ := range r.children {
		// 找到路径为edgeKey路径的Node，然后分叉
		if r.children[i].path == edgeKey {
			// 创建新的分叉Node，路径为公共前缀路径pathKey
			newNode := &middNode{path: pathKey}
			// 将原来edgeKey的数据移动到新的分叉Node之下
			// 直接新增Node，原Node数据仅改变路径为截取后的后段路径
			newNode.children = append(newNode.children, &middNode{
				// 截取路径
				path:	strings.TrimPrefix(edgeKey, pathKey),
				// 复制数据
				key:	r.children[i].key,
				val:	r.children[i].val,
				children:	r.children[i].children,
			})
			// 设置middNode的child[i]的Node为分叉Node
			// 原理路径Node的数据移到到了分叉Node的child里面，原Node对象GC释放。
			r.children[i] = newNode
			// 返回分叉新创建的Node
			return newNode
		}
	}
	return nil
}


// 给currentNode递归添加，路径为containKey的Node
//
// targetKey和targetValue为新Node数据。
func (currentNode *middNode) recursiveInsertTree(containKey string, targetKey string, targetValue HandlerFuncs) {
	// fmt.Println("recursiveInsertTree", containKey, targetKey)
	for i, _ := range currentNode.children {
		// 检查当前遍历的Node和插入路径是否有公共路径
		// subStr是两者的公共路径，find表示是否有
		subStr, find := getSubsetPrefix(containKey, currentNode.children[i].path)
		if find {
			// 如果child路径等于公共最大路径，则该node添加child
			// child的路径为插入路径先过滤公共路径的后面部分。
			if subStr == currentNode.children[i].path {
				nextTargetKey := strings.TrimPrefix(containKey, currentNode.children[i].path)
				// 当前node新增子Node可能原本有多个child，所以需要递归添加
				currentNode.children[i].recursiveInsertTree(nextTargetKey, targetKey, targetValue)
			}else {
				// 如果公共路径不等于当前node的路径
				// 则将currentNode.children[i]路径分叉
				// 分叉后的就拥有了公共路径，然后添加新Node
				newNode := currentNode.SplitNode(subStr, currentNode.children[i].path)
				if newNode == nil {
					panic("Unexpect error on split node")
				}
				// 添加新的node
				// 分叉后树一定只有一个没有相同路径的child，所以直接添加node
				newNode.InsertNode(strings.TrimPrefix(containKey, subStr), targetKey, targetValue)
			}
			return
		}
	}
	// 没有相同前缀路径存在，直接添加为child
	currentNode.InsertNode(containKey, targetKey, targetValue)
}



// 递归获得searchNode路径为searchKey的Node数据。
func (searchNode *middNode) recursiveLoopup(searchKey string) (HandlerFuncs) {
	// fmt.Println("recursiveLoopup:", searchKey)
	// 匹配node，返回数据
	// fmt.Println(1)
	if len(searchKey) == 0  {
		return searchNode.val
	}

	// fmt.Println(2)
	for _, edgeObj := range searchNode.children {
		// 寻找相同前缀node
		if contrainPrefix(searchKey, edgeObj.path) {
			// 截取为匹配的路径
			nextSearchKey := strings.TrimPrefix(searchKey, edgeObj.path)
			// 然后当前Node递归判断
			return append(searchNode.val, edgeObj.recursiveLoopup(nextSearchKey)...)
		}
	}
	// fmt.Println(3)
	if len(searchNode.key) == 0 || searchNode.key[len(searchNode.key)-1] =='/'  {
		return searchNode.val
	}

	// fmt.Println(4)
	return nil
}
