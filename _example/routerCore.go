package main

/*
从routerstd.go简化

RouterCore仅实现匹配算法，不包含上层封装方法。
*/

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const (
	stdNodeKindConst    uint16 = 1 << iota // 常量
	stdNodeKindParam                       // 参数
	stdNodeKindWildcard                    // 通配符

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

	HeaderAllow        = "Allow"
	HeaderXEudoreRoute = "X-Eudore-Route"
	ParamAllow         = "allow"
	ParamRoute         = "route"
)

var (
	// RouterAllMethod 定义路由器允许注册的全部方法，注册其他方法别忽略,前六种方法始终存在。
	RouterAllMethod = []string{MethodGet, MethodPost, MethodPut, MethodDelete, MethodHead, MethodPatch, MethodOptions, MethodConnect, MethodTrace}
	// RouterAnyMethod 定义Any方法的注册使用的方法。
	RouterAnyMethod        = []string{MethodGet, MethodPost, MethodPut, MethodDelete, MethodHead, MethodPatch}
	defaultRouterAnyMethod = append([]string{}, RouterAnyMethod...)
)

// routerCoreStd is implemented based on the radix tree to implement all router related features.
//
// With path parameters, wildcard parameters, default parameters, parameter verification, wildcard verification, multi-parameter regular capture is not implemented.
//
// RouterStd基于基数树实现，实现全部路由器相关特性。
//
// 具有路径参数、通配符参数、默认参数、参数校验、通配符校验，未实现多参数正则捕捉。
type routerCoreStd struct {
	root       *stdNode
	params404  Params
	params405  Params
	handler404 HandlerFuncs
	handler405 HandlerFuncs
}

type stdNode struct {
	isany uint16
	kind  uint16
	pnum  uint32
	path  string
	name  string
	route string

	// 默认标签的名称和值
	params    [7]Params
	handlers  [7]HandlerFuncs
	others    map[string]stdOtherHandler
	Wchildren *stdNode
	Cchildren []*stdNode
	Pchildren []*stdNode
}

type stdOtherHandler struct {
	any     bool
	params  Params
	handler HandlerFuncs
}

type RouterCore interface {
	HandleFunc(string, string, HandlerFuncs)
	Match(string, string, *Params) HandlerFuncs
}
type HandlerFunc func(http.ResponseWriter, *http.Request, Params)
type HandlerFuncs []HandlerFunc

// Params 定义用于保存一些键值数据。
type Params []string

func main() {
	router := NewRouterCoreStd()
	router.HandleFunc("ANY", "/api/* action=api", HandlerFuncs{HandlerPrintRoute})
	router.HandleFunc("ANY", "/api/users action=user:Get", HandlerFuncs{HandlerPrintRoute})
	router.HandleFunc("ANY", "/api/users/:id action=user:GetById", HandlerFuncs{HandlerPrintRoute})
	router.HandleFunc("PUT", "index action=index", HandlerFuncs{HandlerPrintRoute})

	srv := &http.Server{
		Addr: ":8088",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var p Params
			for _, fn := range router.Match(r.Method, r.URL.Path, &p) {
				fn(w, r, p)
			}
		}),
	}
	srv.ListenAndServe()
}

func HandlerPrintRoute(w http.ResponseWriter, r *http.Request, p Params) {
	w.Write([]byte(p.Get(ParamRoute) + " " + p.String()))
}

// HandlerRouter405 function defines the default 405 processing and returns Allow and X-Match-Route Header.
//
// HandlerRouter405 函数定义默认405处理,返回Allow和X-Match-Route Header。
func HandlerRouter405(w http.ResponseWriter, r *http.Request, p Params) {
	const page405 string = "405 method not allowed"
	w.Header().Set(HeaderAllow, p.Get(ParamAllow))
	w.Header().Set(HeaderXEudoreRoute, p.Get(ParamRoute))
	w.WriteHeader(405)
	w.Write([]byte(page405))
}

// HandlerRouter404 function defines the default 404 processing.
//
// HandlerRouter404 函数定义默认404处理。
func HandlerRouter404(w http.ResponseWriter, r *http.Request, p Params) {
	const page404 string = "404 page not found"
	w.WriteHeader(404)
	w.Write([]byte(page404))
}

// NewRouterCoreStd function creates a Std router core and uses radix to match. For the function description, please refer to the Router document.
//
// NewRouterCoreStd 函数创建一个Std路由器核心，使用radix匹配,功能说明见Router文档。
func NewRouterCoreStd() RouterCore {
	return &routerCoreStd{
		root:       &stdNode{},
		handler404: HandlerFuncs{HandlerRouter404},
		handler405: HandlerFuncs{HandlerRouter405},
	}
}

// HandleFunc method register a new method request path to the router
//
// The router matches the handlers available to the current path from the middleware tree and adds them to the front of the handler.
//
// HandleFunc 给路由器注册一个新的方法请求路径
//
// 路由器会从中间件树中匹配当前路径可使用的处理者，并添加到处理者前方。
func (r *routerCoreStd) HandleFunc(method string, path string, handler HandlerFuncs) {
	switch method {
	case "NotFound", "404":
		r.params404 = NewParamsRoute(path)[2:]
		r.handler404 = handler
	case "MethodNotAllowed", "405":
		r.params405 = NewParamsRoute(path)[2:]
		r.handler405 = handler
	case MethodAny, MethodGet, MethodPost, MethodPut, MethodDelete, MethodHead, MethodPatch:
		r.insertRoute(method, path, handler)
	default:
		for _, m := range RouterAllMethod {
			if method == m {
				r.insertRoute(method, path, handler)
				return
			}
		}
	}
}

// Match a request, if the method does not allow direct return to node405, no match returns node404.
//
// 匹配一个请求，如果方法不允许直接返回node405，未匹配返回node404。
func (r *routerCoreStd) Match(method, path string, params *Params) HandlerFuncs {
	node := r.root.lookNode(path, params)
	if node == nil {
		// 处理404
		*params = params.Set(ParamRoute, "404").Add(r.params404...)
		return r.handler404
	}
	// default method
	for i, m := range defaultRouterAnyMethod {
		if m == method {
			if node.handlers[i] != nil {
				*params = params.Set(ParamRoute, node.params[i][1]).Add(node.params[i][2:]...)
				return node.handlers[i]
			}
			break
		}
	}
	// other method
	handlers, ok := node.others[method]
	if ok {
		*params = params.Set(ParamRoute, handlers.params[1]).Add(handlers.params[2:]...)
		return handlers.handler
	}
	// 处理405
	pos := strings.IndexByte(node.route, ';')
	*params = params.Set(ParamRoute, node.route[pos+1:]).Add(ParamAllow, node.route[:pos]).Add(r.params405...)
	return r.handler405
}

// Add a new route Node.
//
// 添加一个新的路由Node。
func (r *routerCoreStd) insertRoute(method, path string, val HandlerFuncs) {
	var currentNode = r.root
	params := NewParamsRoute(path)
	for _, path := range getSplitPath(params.Get(ParamRoute)) {
		currentNode = currentNode.insertNode(path, r.newStdNode(path))
	}
	currentNode.setHandler(method, params, val)
	currentNode.setRoute()
}

// 创建一个Radix树Node，会根据当前路由设置不同的节点类型和名称。
//
// '*'前缀为通配符节点，':'前缀为参数节点，其他未常量节点,如果通配符和参数节点后带有符号'|'则为校验节点。
func (r *routerCoreStd) newStdNode(path string) *stdNode {
	newNode := &stdNode{path: path}
	switch path[0] {
	case '*':
		newNode.kind = stdNodeKindWildcard
		if len(path) == 1 {
			newNode.name = "*"
		} else {
			newNode.name = path[1:]
		}
	case ':':
		newNode.kind = stdNodeKindParam
		newNode.name = path[1:]
	// 常量Node
	default:
		newNode.kind = stdNodeKindConst
	}
	return newNode
}

func (r *stdNode) setHandler(method string, params Params, handler HandlerFuncs) {
	if method == MethodAny {
		r.setHandlerAny(params, handler)
		return
	}

	for i := uint(0); i < 6; i++ {
		if defaultRouterAnyMethod[i] == method {
			r.params[i] = params
			r.handlers[i] = handler
			r.isany &^= 1 << i
			return
		}
	}
	if r.others == nil {
		r.others = make(map[string]stdOtherHandler)
	}
	r.others[method] = stdOtherHandler{params: params, handler: handler}
}

func (r *stdNode) setHandlerAny(params Params, handler HandlerFuncs) {
	// 设置标准Any
	for i := uint(0); i < 6; i++ {
		if r.isany>>i&0x1 == 0x1 || r.handlers[i] == nil {
			r.params[i] = params
			r.handlers[i] = handler
			r.isany |= 1 << i
		}
	}
	// 设置others any
	for _, method := range RouterAnyMethod {
		i := getStringInIndex(method, defaultRouterAnyMethod)
		if i == -1 {
			if r.others == nil {
				r.others = make(map[string]stdOtherHandler)
			}
			r.others[method] = stdOtherHandler{any: true, params: params, handler: handler}
		}
	}
	r.isany |= 0x40
	r.params[6] = params
	r.handlers[6] = handler
}

func getStringInIndex(str string, strs []string) int {
	for i := range strs {
		if str == strs[i] {
			return i
		}
	}
	return -1
}

func (r *stdNode) setRoute() {
	var allow string
	var route string
	for i := uint(0); i < 6; i++ {
		if r.handlers[i] != nil {
			allow = allow + ", " + RouterAllMethod[i]
			route = r.params[i].Get(ParamRoute)
		}
	}
	for name, other := range r.others {
		allow = allow + ", " + name
		route = other.params.Get(ParamRoute)
	}
	if allow != "" {
		allow = allow[2:]
	}
	r.route = allow + ";" + route
}

// insertNode add a child node to the node.
//
// insertNode 给节点添加一个子节点。
func (r *stdNode) insertNode(path string, nextNode *stdNode) *stdNode {
	if len(path) == 0 {
		return r
	}
	nextNode.path = path
	switch nextNode.kind {
	case stdNodeKindConst:
		return r.insertNodeConst(path, nextNode)
	case stdNodeKindParam:
		for _, i := range r.Pchildren {
			if i.path == path {
				return i
			}
		}
		r.pnum++
		r.Pchildren = append(r.Pchildren, nextNode)
	case stdNodeKindWildcard:
		if r.Wchildren == nil {
			r.Wchildren = nextNode
		} else {
			r.Wchildren.path = nextNode.path
			r.Wchildren.name = nextNode.name
		}
		return r.Wchildren
		// default:
		// 	panic("Undefined radix node type from router std.")
	}
	return nextNode
}

// insertNodeConst 方法处理添加常量node。
func (r *stdNode) insertNodeConst(path string, nextNode *stdNode) *stdNode {
	// 变量添加常量node
	for i := range r.Cchildren {
		subStr, find := getSubsetPrefix(path, r.Cchildren[i].path)
		if find {
			if subStr != r.Cchildren[i].path {
				r.Cchildren[i].path = strings.TrimPrefix(r.Cchildren[i].path, subStr)
				r.Cchildren[i] = &stdNode{
					kind:      stdNodeKindConst,
					path:      subStr,
					Cchildren: []*stdNode{r.Cchildren[i]},
				}
			}
			return r.Cchildren[i].insertNode(strings.TrimPrefix(path, subStr), nextNode)
		}
	}
	r.Cchildren = append(r.Cchildren, nextNode)
	// 常量node按照首字母排序。
	for i := len(r.Cchildren) - 1; i > 0; i-- {
		if r.Cchildren[i].path[0] < r.Cchildren[i-1].path[0] {
			r.Cchildren[i], r.Cchildren[i-1] = r.Cchildren[i-1], r.Cchildren[i]
		}
	}
	return nextNode
}

func (r *stdNode) lookNode(searchKey string, params *Params) *stdNode {
	// constant match, return data
	// 常量匹配，返回数据
	if len(searchKey) == 0 && r.route != "" {
		return r
	}

	if len(searchKey) > 0 {
		// Traverse constant Node match
		// 遍历常量Node匹配，数据量少使用二分查找无效
		for _, child := range r.Cchildren {
			if child.path[0] >= searchKey[0] {
				if len(searchKey) >= len(child.path) && searchKey[:len(child.path)] == child.path {
					nextSearchKey := searchKey[len(child.path):]
					if n := child.lookNode(nextSearchKey, params); n != nil {
						return n
					}
				}
				break
			}
		}

		// parameter matching, Check if there is a parameter match
		// 参数匹配 检测是否存在参数匹配
		if r.pnum != 0 {
			pos := strings.IndexByte(searchKey, '/')
			if pos == -1 {
				pos = len(searchKey)
			}
			currentKey, nextSearchKey := searchKey[:pos], searchKey[pos:]

			// 参数匹配
			// 变量Node依次匹配是否满足
			for _, child := range r.Pchildren {
				if n := child.lookNode(nextSearchKey, params); n != nil {
					*params = params.Add(child.name, currentKey)
					return n
				}
			}
		}
	}

	// If the current Node has a wildcard processing method that directly matches, the result is returned.
	// 若当前Node有通配符处理方法直接匹配，返回结果。
	if r.Wchildren != nil {
		*params = params.Add(r.Wchildren.name, searchKey)
		return r.Wchildren
	}

	// can't match, return nil
	// 无法匹配，返回空
	return nil
}

/*
The string is cut according to the Node type.
将字符串按Node类型切割
String path cutting example:
字符串路径切割例子：
/				[/]
/api/note/		[/api/note/]
//api/*			[/api/ *]
//api/*name		[/api/ *name]
/api/get/		[/api/get/]
/api/get		[/api/get]
/api/:get		[/api/ :get]
/api/:get/*		[/api/ :get / *]
/api/:name/info/*		[/api/ :name /info/ *]
/api/:name|^\\d+$/info	[/api/ :name|^\d+$ /info]
/api/*|{^0/api\\S+$}	[/api/ *|{^0/api\S+$}]
/api/*|^\\$\\d+$		[/api/ *|^\$\d+$]
*/
func getSplitPath(key string) []string {
	if len(key) < 2 {
		return []string{"/"}
	}
	if key[0] != '/' {
		key = "/" + key
	}
	var strs []string
	var length = -1
	var isblock = 0
	var isconst = false
	for i := range key {
		// 块模式匹配
		if isblock > 0 {
			switch key[i] {
			case '{':
				isblock++
			case '}':
				isblock--
			}
			if isblock > 0 {
				strs[length] = strs[length] + key[i:i+1]
			}
			continue
		}
		switch key[i] {
		case '/':
			// 常量模式，非常量模式下创建新字符串
			if !isconst {
				length++
				strs = append(strs, "")
				isconst = true
			}
		case ':':
			// 变量模式
			isconst = false
			length++
			strs = append(strs, "")
		case '*':
			// 通配符模式
			isconst = false
			length++
			strs = append(strs, key[i:])
			return strs
		case '{':
			isblock++
			continue
		}
		strs[length] = strs[length] + key[i:i+1]
	}
	return strs
}

// Get the largest common prefix of the two strings,
// return the largest common prefix and have the largest common prefix.
//
// 获取两个字符串的最大公共前缀，返回最大公共前缀和是否拥有最大公共前缀。
func getSubsetPrefix(str1, str2 string) (string, bool) {
	findSubset := false
	for i := 0; i < len(str1) && i < len(str2); i++ {
		if str1[i] != str2[i] {
			retStr := str1[:i]
			return retStr, findSubset
		}
		findSubset = true
	}

	if len(str1) > len(str2) {
		return str2, findSubset
	} else if len(str1) == len(str2) {
		return str1, str1 == str2
	}

	return str1, findSubset
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

func split2byte(str string, b byte) (string, string) {
	pos := strings.IndexByte(str, b)
	if pos == -1 {
		return str, ""
	}
	return str[:pos], str[pos+1:]
}

// NewParamsRoute 方法根据一个路由路径创建Params，支持路由路径块模式。
func NewParamsRoute(path string) Params {
	route := getRoutePath(path)
	args := strings.Split(path[len(route):], " ")
	if args[0] == "" {
		args = args[1:]
	}
	params := make(Params, 0, len(args)*2+2)
	params = append(params, ParamRoute, route)
	for _, str := range args {
		k, v := split2byte(str, '=')
		if v != "" {
			params = append(params, k, v)
		}
	}
	return params
}

// Clone 方法深复制一个ParamArray对象。
func (p Params) Clone() Params {
	params := make(Params, len(p))
	copy(params, p)
	return params
}

// CombineWithRoute 方法将params数据合并到p，用于路由路径合并。
func (p Params) CombineWithRoute(params Params) Params {
	p[1] = p[1] + params[1]
	for i := 2; i < len(params); i += 2 {
		p = p.Set(params[i], params[i+1])
	}
	return p
}

// String 方法输出Params成字符串。
func (p Params) String() string {
	b := &bytes.Buffer{}
	for i := 0; i < len(p); i += 2 {
		if (p[i] != "" && p[i+1] != "") || i == 0 {
			if b.Len() != 0 {
				b.WriteString(" ")
			}
			fmt.Fprintf(b, "%s=%s", p[i], p[i+1])
		}
	}
	return b.String()
}

// MarshalJSON 方法设置Params json序列化显示的数据。
func (p Params) MarshalJSON() ([]byte, error) {
	data := make(map[string]string, len(p)/2)
	for i := 0; i < len(p); i += 2 {
		if p[i+1] != "" || i == 0 {
			data[p[i]] = p[i+1]
		}
	}
	return json.Marshal(data)
}

// Get 方法返回一个参数的值。
func (p Params) Get(key string) string {
	for i := 0; i < len(p); i += 2 {
		if p[i] == key {
			return p[i+1]
		}
	}
	return ""
}

// Add 方法添加一个参数。
func (p Params) Add(vals ...string) Params {
	return append(p, vals...)
}

// Set 方法设置一个参数的值。
func (p Params) Set(key, val string) Params {
	for i := 0; i < len(p); i += 2 {
		if p[i] == key {
			p[i+1] = val
			return p
		}
	}
	return append(p, key, val)
}

// Del 方法删除一个参数值
func (p Params) Del(key string) {
	for i := 0; i < len(p); i += 2 {
		if p[i] == key {
			p[i+1] = ""
		}
	}
}
