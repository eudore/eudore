package eudore

/*
基于基数树算法实现一个完整路由器
*/

import (
	"context"
	"fmt"
	"strings"
)

const (
	stdNodeKindConst         uint16 = 1 << iota // 常量
	stdNodeKindParamValid                       // 参数校验
	stdNodeKindParam                            // 参数
	stdNodeKindWildcardValid                    // 通配符校验
	stdNodeKindWildcard                         // 通配符
)

// routerCoreStd is implemented based on the radix tree to implement all router related features.
//
// RouterStd基于基数树实现，实现全部路由器相关特性。
type routerCoreStd struct {
	root        *stdNode
	params404   Params
	params405   Params
	handler404  HandlerFuncs
	handler405  HandlerFuncs
	FuncCreator FuncCreator
}

type stdNode struct {
	isany uint16
	kind  uint16
	pnum  uint32
	check func(string) bool
	path  string
	name  string
	route string

	// 默认标签的名称和值
	params     [7]Params
	handlers   [7]HandlerFuncs
	others     map[string]stdOtherHandler
	Wchildren  *stdNode
	Cchildren  []*stdNode
	Pchildren  []*stdNode
	PVchildren []*stdNode
	WVchildren []*stdNode
}

type stdOtherHandler struct {
	any     bool
	params  Params
	handler HandlerFuncs
}

// NewRouterCoreStd function creates a Std router core and uses radix to match.
// For the function description, please refer to the Router document.
//
// NewRouterCoreStd 函数创建一个Std路由器核心，使用radix匹配,功能说明见Router文档。
func NewRouterCoreStd() RouterCore {
	return &routerCoreStd{
		root:        &stdNode{},
		handler404:  HandlerFuncs{HandlerRouter404},
		handler405:  HandlerFuncs{HandlerRouter405},
		FuncCreator: DefaultFuncCreator,
	}
}

// Mount 方法获取ContextKeyFuncCreator，用于创建校验函数。
func (r *routerCoreStd) Mount(ctx context.Context) {
	fc, ok := ctx.Value(ContextKeyFuncCreator).(FuncCreator)
	if ok {
		r.FuncCreator = fc
	}
}

// HandleFunc method register a new method request path to the router
//
// The router matches the handlers available to the current path from
// the middleware tree and adds them to the front of the handler.
//
// HandleFunc 给路由器注册一个新的方法请求路径。
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
		for _, m := range DefaultRouterAllMethod {
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
	for i, m := range DefaultRouterCoreMethod {
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
	currentNode := r.root

	params := NewParamsRoute(path)
	if params.Get(ParamRegister) == "off" || val == nil {
		currentNode.deleteRoute(method, params.Get(ParamRoute))
		return
	}

	// 创建节点
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
			// 如果路径后序具有'|'符号，则截取后端名称返回校验函数
			// 并升级成校验通配符Node
			if name, fn := r.loadCheckFunc(path); len(name) > 0 {
				newNode.kind, newNode.name, newNode.check = stdNodeKindWildcardValid, name, fn
			}
		}
	case ':':
		newNode.kind = stdNodeKindParam
		newNode.name = path[1:]
		// 如果路径后序具有'|'符号，则截取后端名称返回校验函数
		// 并升级成校验参数Node
		if name, fn := r.loadCheckFunc(path); len(name) > 0 {
			newNode.kind, newNode.name, newNode.check = stdNodeKindParamValid, name, fn
		}
	// 常量Node
	default:
		newNode.kind = stdNodeKindConst
	}
	return newNode
}

// Load the checksum function by name.
//
// 根据名称加载校验函数。
func (r *routerCoreStd) loadCheckFunc(path string) (string, func(string) bool) {
	path = path[1:]
	// 截取参数名称和校验函数名称
	name, fname, _ := strings.Cut(path, "|")
	if name == "" || fname == "" {
		return "", nil
	}
	// 如果是正则表达式开头，添加默认正则校验函数名称
	if fname[0] == '^' && fname[len(fname)-1] == '$' {
		fname = "regexp:" + fname
	}

	// 调用FuncCreator创建check函数
	fn, err := r.FuncCreator.CreateFunc(FuncCreateString, fname)
	if err == nil {
		return name, fn.(func(string) bool)
	}
	// 无法获得校验函数抛出错误，由RouterStd recover。
	panic(fmt.Errorf(ErrFormarRouterStdLoadInvalidFunc, path, err))
}

func (r *stdNode) setHandler(method string, params Params, handler HandlerFuncs) {
	if method == MethodAny {
		r.setHandlerAny(params, handler)
		return
	}

	for i := uint(0); i < 6; i++ {
		if DefaultRouterCoreMethod[i] == method {
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
	for _, method := range DefaultRouterAnyMethod {
		i := getStringInIndex(method, DefaultRouterCoreMethod)
		if i == -1 {
			// 设置others any
			if r.others == nil {
				r.others = make(map[string]stdOtherHandler)
			}
			r.others[method] = stdOtherHandler{any: true, params: params, handler: handler}
		} else if r.isany>>i&0x1 == 0x1 || r.handlers[i] == nil {
			// 设置标准Any
			r.params[i] = params
			r.handlers[i] = handler
			r.isany |= 1 << i
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
			allow = allow + ", " + DefaultRouterAllMethod[i]
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
// 给当前节点路径下添加一个子节点。
//
// 如果新节点类型是常量节点，寻找是否存在相同前缀路径的节点，
// 如果存在路径为公共前缀的节点，直接添加新节点为匹配前缀节点的子节点；
// 如果只是两节点只是拥有公共前缀，则先分叉然后添加子节点。
//
// 如果新节点类型是参数/校验参数/校验通配符/节点，会检测当前参数是否存在，存在返回已处在的节点。
//
// 如果新节点类型是通配符节点，直接设置为当前节点的通配符处理节点。
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
	case stdNodeKindParamValid:
		for _, i := range r.PVchildren {
			if i.path == path {
				return i
			}
		}
		r.pnum++
		r.PVchildren = append(r.PVchildren, nextNode)
	case stdNodeKindWildcardValid:
		for _, i := range r.WVchildren {
			if i.path == path {
				return i
			}
		}
		r.WVchildren = append(r.WVchildren, nextNode)
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
	// 遍历添加常量node
	for i := range r.Cchildren {
		subStr, find := getSubsetPrefix(path, r.Cchildren[i].path)
		if find {
			// 如果常量node路径比公有前缀长，需要拆分node路径
			// 公有路径作为一个父node和拥有剩余路径的子node，使父node路径一定为新增node的前缀
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

//nolint:cyclop,gocyclo
func (r *stdNode) lookNode(searchKey string, params *Params) *stdNode {
	// constant match, return data
	if len(searchKey) == 0 && r.route != "" {
		return r
	}

	if len(searchKey) > 0 {
		// Traverse constant Node match
		for _, child := range r.Cchildren {
			if child.path[0] >= searchKey[0] {
				length := len(child.path)
				if len(searchKey) >= length && searchKey[:length] == child.path {
					if n := child.lookNode(searchKey[length:], params); n != nil {
						return n
					}
				}
				break
			}
		}

		// parameter matching, Check if there is a parameter match
		if r.pnum != 0 {
			pos := strings.IndexByte(searchKey, '/')
			if pos == -1 {
				pos = len(searchKey)
			}
			currentKey, nextSearchKey := searchKey[:pos], searchKey[pos:]

			// check parameter matching
			for _, child := range r.PVchildren {
				if child.check(currentKey) {
					if n := child.lookNode(nextSearchKey, params); n != nil {
						*params = params.Add(child.name, currentKey)
						return n
					}
				}
			}
			for _, child := range r.Pchildren {
				if n := child.lookNode(nextSearchKey, params); n != nil {
					*params = params.Add(child.name, currentKey)
					return n
				}
			}
		}
	}
	// If the current Node has a wildcard processing method that directly matches
	for _, child := range r.WVchildren {
		if child.check(searchKey) {
			*params = params.Add(child.name, searchKey)
			return child
		}
	}
	// If the current Node has a wildcard processing method that directly matches
	if r.Wchildren != nil {
		*params = params.Add(r.Wchildren.name, searchKey)
		return r.Wchildren
	}
	// can't match, return nil
	return nil
}

func (r *stdNode) deleteRoute(method, path string) {
	nodes := r.findNode(path)
	if nodes == nil {
		return
	}
	// clean hndler
	nodes[len(nodes)-1].delHandler(method)
	nodes[len(nodes)-1].setRoute()
	// clean node
	for i := len(nodes) - 2; i > -1; i-- {
		if nodes[i+1].IsZero() {
			nodes[i].deleteNode(nodes[i+1])
			nodes[i].IsMarge()
		} else if !nodes[i].IsMarge() {
			return
		}
	}
}

func (r *stdNode) delHandler(method string) {
	if method == MethodAny {
		for i := uint(0); i < 6; i++ {
			if r.isany>>i&0x1 == 0x1 {
				r.params[i] = nil
				r.handlers[i] = nil
				r.isany &^= 1 << i
			}
		}
		for k, v := range r.others {
			if v.any {
				delete(r.others, k)
			}
		}
		r.params[6] = nil
		r.handlers[6] = nil
		return
	}

	for i := uint(0); i < 6; i++ {
		if DefaultRouterAllMethod[i] == method {
			if r.isany>>i&0x1 == 0x0 {
				r.params[i] = nil
				r.handlers[i] = nil
				if r.handlers[6] != nil {
					r.params[i] = r.params[6]
					r.handlers[i] = r.handlers[6]
					r.isany |= 1 << i
				}
			}
			return
		}
	}
	if r.handlers[6] != nil {
		r.others[method] = stdOtherHandler{any: true, params: r.params[6], handler: r.handlers[6]}
	} else {
		delete(r.others, method)
	}
}

func (r *stdNode) findNode(path string) []*stdNode {
	args := getSplitPath(path)
	nodes := make([]*stdNode, 1, len(args)*2)
	nodes[0] = r
	for _, i := range args {
		last := nodes[len(nodes)-1]
		switch i[0] {
		case '*':
			child := last.findNodeWildcard(i)
			if child == nil {
				return nil
			}
			nodes = append(nodes, child)
		case ':':
			child := last.findNodeParam(i)
			if child == nil {
				return nil
			}
			nodes = append(nodes, child)
		default:
			childs := last.findNodeConst(i)
			if childs == nil {
				return nil
			}
			for i := len(childs) - 1; i > -1; i-- {
				nodes = append(nodes, childs[i])
			}
		}
	}
	return nodes
}

func (r *stdNode) findNodeWildcard(path string) *stdNode {
	if r.Wchildren != nil && r.Wchildren.path == path {
		return r.Wchildren
	}
	for _, child := range r.WVchildren {
		if child.path == path {
			return child
		}
	}
	return nil
}

func (r *stdNode) findNodeParam(path string) *stdNode {
	for _, child := range r.Pchildren {
		if child.path == path {
			return child
		}
	}
	for _, child := range r.PVchildren {
		if child.path == path {
			return child
		}
	}
	return nil
}

func (r *stdNode) findNodeConst(path string) []*stdNode {
	if path == "" {
		return []*stdNode{r}
	}
	for _, child := range r.Cchildren {
		if child.path[0] >= path[0] {
			if len(path) >= len(child.path) && path[:len(child.path)] == child.path {
				if n := child.findNodeConst(path[len(child.path):]); n != nil {
					return append(n, r)
				}
			}
			break
		}
	}
	return nil
}

func (r *stdNode) IsEmpty() bool {
	for i := range r.handlers {
		if r.handlers[i] != nil {
			return false
		}
	}
	// r.params = nil
	return len(r.others) == 0
}

func (r *stdNode) IsZero() bool {
	return r.IsEmpty() && len(r.Cchildren) == 0 && len(r.Pchildren) == 0 && len(r.PVchildren) == 0 &&
		len(r.WVchildren) == 0 && r.Wchildren == nil
}

func (r *stdNode) IsMarge() bool {
	if r.kind == stdNodeKindConst && r.IsEmpty() && len(r.Cchildren) == 1 && len(r.Pchildren) == 0 &&
		len(r.PVchildren) == 0 && len(r.WVchildren) == 0 && r.Wchildren == nil {
		r.Cchildren[0].path = r.path + r.Cchildren[0].path
		*r = *r.Cchildren[0]
		return true
	}
	return false
}

func (r *stdNode) deleteNode(node *stdNode) {
	switch node.kind {
	case stdNodeKindConst:
		r.Cchildren = stdRemoveNode(r.Cchildren, node)
	case stdNodeKindParam:
		r.PVchildren = stdRemoveNode(r.PVchildren, node)
		r.pnum--
	case stdNodeKindParamValid:
		r.Pchildren = stdRemoveNode(r.Pchildren, node)
		r.pnum--
	case stdNodeKindWildcardValid:
		r.WVchildren = stdRemoveNode(r.WVchildren, node)
	case stdNodeKindWildcard:
		r.Wchildren = nil
	}
}

func stdRemoveNode(nodes []*stdNode, node *stdNode) []*stdNode {
	if len(nodes) == 1 {
		return nil
	}
	for i, child := range nodes {
		if child == node {
			for ; i < len(nodes)-1; i++ {
				nodes[i] = nodes[i+1]
			}
			nodes = nodes[:len(nodes)-1]
		}
	}
	return nodes
}

/*
The string is cut according to the Node type, String path cutting example:
将字符串按Node类型切割，字符串路径切割例子：

	/				[/]
	/api/note/		[/api/note/]
	//api/*			[//api/ *]
	//api/*name		[//api/ *name]
	/api/get/		[/api/get/]
	/api/get		[/api/get]
	/api/:get		[/api/ :get]
	/api/:get/*		[/api/ :get / *]
	/api/:name/info/*		[/api/ :name /info/ *]
	/api/:name|^\\d+$/info	[/api/ :name|^\d+$ /info]
	/api/*|{^0/api\\S+$}	[/api/ *|^0/api\S+$]
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
	length := -1
	isblock := 0
	isconst := false
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
				strs[length] += key[i : i+1]
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
		case ':', '*':
			// 变量模式，通配符模式
			isconst = false
			length++
			strs = append(strs, "")
		case '{':
			isblock++
			continue
		}
		strs[length] += key[i : i+1]
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
