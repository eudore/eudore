package eudore

/*
基于基数树算法实现一个标准功能的路由器。
*/

import (
	"strings"
)

const (
	radixNodeKindConst uint8 = 1 << iota
	radixNodeKindParam
	radixNodeKindWildcard
)

// routerCoreRadix basic function router based on radix tree implementation.
//
// There are three basic functions: path parameter, wildcard parameter, default parameter, and parameter verification.
// RouterRadix基于基数树实现的基本功能路由器。
//
// 具有路径参数、通配符参数、默认参数三项基本功能。
type routerCoreRadix struct {
	// exception handling method
	// 异常处理方法
	node404 radixNode
	node405 radixNode
	// various methods routing tree
	// 各种方法路由树
	get     radixNode
	post    radixNode
	put     radixNode
	delete  radixNode
	head    radixNode
	patch   radixNode
	options radixNode
	connect radixNode
	trace   radixNode
}

// radix节点的定义
type radixNode struct {
	// 基本信息
	kind uint8
	path string
	name string
	// 每次类型子节点
	Cchildren []*radixNode
	Pchildren []*radixNode
	Wchildren *radixNode
	// 当前节点的数据
	params      *Params
	handlers    HandlerFuncs
	anyin       bool
	anyhandlers HandlerFuncs
}

// NewRouterCoreRadix 函数创建一个Full路由器核心，使用radix匹配。
func NewRouterCoreRadix() RouterCore {
	return &routerCoreRadix{
		node404: radixNode{
			params:   &Params{Keys: []string{ParamRoute}, Vals: []string{"404"}},
			handlers: HandlerFuncs{HandlerRouter404},
		},
		node405: radixNode{
			Wchildren: &radixNode{
				params:   &Params{Keys: []string{ParamRoute}, Vals: []string{"405"}},
				handlers: HandlerFuncs{HandlerRouter405},
			},
		},
	}
}

// HandleFunc register a new method request path to the router.
//
// HandleFunc 给路由器注册一个新的方法请求路径。
//
// 如果方法是Any会注册全部方法，同时非Any方法路由和覆盖Any方法路由。
func (r *routerCoreRadix) HandleFunc(method string, path string, handler HandlerFuncs) {
	switch method {
	case "NotFound", "404":
		r.node404.handlers = handler
	case "MethodNotAllowed", "405":
		r.node405.Wchildren.handlers = handler
	case MethodAny:
		for _, method := range RouterAllMethod {
			r.insertRoute(method, path, true, handler)
		}
	default:
		r.insertRoute(method, path, false, handler)
	}
}

// Match a request, if the method does not allow direct return to node405, no match returns node404.
//
// Note: 404 does not support extra parameters, not implemented.
//
// 匹配一个请求，如果方法不不允许直接返回node405，未匹配返回node404。
func (r *routerCoreRadix) Match(method, path string, params *Params) HandlerFuncs {
	if n := r.getTree(method).lookNode(path, params); n != nil {
		return n
	}

	// 处理404
	params.Combine(r.node404.params)
	return r.node404.handlers
}

// Get the tree of the corresponding method.
//
// Support eudore.RouterAllMethod these methods, weak support will return 405 processing tree.
//
// 获取对应方法的树。
//
// 支持eudore.RouterAllMethod这些方法,弱不支持会返回405处理树。
func (r *routerCoreRadix) getTree(method string) *radixNode {
	switch method {
	case MethodGet:
		return &r.get
	case MethodPost:
		return &r.post
	case MethodDelete:
		return &r.delete
	case MethodPut:
		return &r.put
	case MethodHead:
		return &r.head
	case MethodPatch:
		return &r.patch
	case MethodOptions:
		return &r.options
	case MethodConnect:
		return &r.connect
	case MethodTrace:
		return &r.trace
	default:
		return &r.node405
	}
}

// insertRoute add a new routing node.
//
// If the method is not supported, it will not be added. Requesting the path will respond 405.
//
// Cut the path by node type. Each path is a type of node, then append to the tree in turn, and then set the data to the last node.
//
// Path cut see getSpiltPath function, currently not perfect, processing regularity may be abnormal.
//
// 添加一个新的路由节点。
//
// 如果方法不支持则不会添加，请求该路径会响应405.
//
// 将路径按节点类型切割，每段路径即为一种类型的节点，然后依次向树追加，然后给最后的节点设置数据。
//
// insertRoute 路径切割见getSpiltPath函数，当前未完善，处理正则可能异常。
func (r *routerCoreRadix) insertRoute(method, key string, isany bool, val HandlerFuncs) {
	var currentNode = r.getTree(method)
	if currentNode == &r.node405 {
		return
	}

	params := NewParamsRoute(key)
	if params.Get(ParamRegister) == "off" {
		currentNode.deleteRoute(params.Get(ParamRoute), isany, val)
		return
	}
	// 创建节点
	for _, path := range getSplitPath(params.Get(ParamRoute)) {
		currentNode = currentNode.insertNode(path, newRadixNode(path))
	}
	currentNode.params = params
	if isany {
		if currentNode.anyin || currentNode.handlers == nil {
			currentNode.handlers = val
			currentNode.anyin = true
		}
		currentNode.anyhandlers = val
	} else {
		currentNode.handlers = val
		currentNode.anyin = false
	}
}

// Create a Radix tree Node that will set different node types based on the current route.
//
// '*' prefix is a wildcard node, ':' prefix is a parameter node, and other non-constant nodes.
//
// 创建一个Radix树Node，会根据当前路由设置不同的节点类型和名称。
//
// '*'前缀为通配符节点，':'前缀为参数节点，其他未常量节点。
func newRadixNode(path string) *radixNode {
	newNode := &radixNode{path: path}
	switch path[0] {
	case '*':
		newNode.kind = radixNodeKindWildcard
		if len(path) == 1 {
			newNode.name = "*"
		} else {
			newNode.name = path[1:]
		}
	case ':':
		newNode.kind = radixNodeKindParam
		newNode.name = path[1:]
	default:
		newNode.kind = radixNodeKindConst
	}
	return newNode
}

// Add a child node to the current node path.
//
// If the new node type is a constant node, look for nodes with the same prefix path.
// If there is a node with a common prefix, add the new node directly to the child node that matches the prefix node;
// If only the two nodes only have a common prefix, then fork and then add the child nodes.
//
// If the new node type is a parameter node, it will detect if the current parameter exists, and there is a return node that is already present.
//
// If the new node type is a wildcard node, set it directly to the current node's wildcard processing node.
//
// 给当前节点路径下添加一个子节点。
//
// 如果新节点类型是常量节点，寻找是否存在相同前缀路径的结点，
// 如果存在路径为公共前缀的结点，直接添加新结点为匹配前缀结点的子节点；
// 如果只是两结点只是拥有公共前缀，则先分叉然后添加子节点。
//
// 如果新节点类型是参数结点，会检测当前参数是否存在，存在返回已处在的节点。
//
// 如果新节点类型是通配符结点，直接设置为当前节点的通配符处理节点。
func (r *radixNode) insertNode(path string, nextNode *radixNode) *radixNode {
	if len(path) == 0 {
		return r
	}
	nextNode.path = path
	switch nextNode.kind {
	case radixNodeKindConst:
		for i := range r.Cchildren {
			subStr, find := getSubsetPrefix(path, r.Cchildren[i].path)
			if find {
				if subStr != r.Cchildren[i].path {
					r.Cchildren[i].path = strings.TrimPrefix(r.Cchildren[i].path, subStr)
					r.Cchildren[i] = &radixNode{
						kind:      fullNodeKindConst,
						path:      subStr,
						Cchildren: []*radixNode{r.Cchildren[i]},
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
	case radixNodeKindParam:
		for _, i := range r.Pchildren {
			if i.path == path {
				return i
			}
		}
		r.Pchildren = append(r.Pchildren, nextNode)
	case radixNodeKindWildcard:
		if r.Wchildren != nil {
			return r.Wchildren
		}
		r.Wchildren = nextNode
		// default:
		// 	panic("Undefined radix node type from router radix.")
	}
	return nextNode
}

// lookNode 按照顺序匹配一个路径。
//
// 依次检查常量节点、参数节点、通配符节点，如果有一个匹配就直接返回。
func (r *radixNode) lookNode(searchKey string, params *Params) HandlerFuncs {
	// 如果路径为空，当前节点就是需要匹配的节点，直接返回。
	if len(searchKey) == 0 && r.handlers != nil {
		params.Combine(r.params)
		return r.handlers
	}

	if len(searchKey) > 0 {
		// 遍历常量Node匹配，寻找具有相同前缀的那个节点
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

		if len(r.Pchildren) > 0 {
			pos := strings.IndexByte(searchKey, '/')
			if pos == -1 {
				pos = len(searchKey)
			}
			nextSearchKey := searchKey[pos:]

			// Whether the variable Node matches in sequence is satisfied
			// 遍历参数节点是否后续匹配
			for _, child := range r.Pchildren {
				if n := child.lookNode(nextSearchKey, params); n != nil {
					params.Add(child.name, searchKey[:pos])
					return n
				}
			}
		}
	}

	// If the current Node has a wildcard processing method that directly matches, the result is returned.
	// 若当前节点有通配符处理方法直接匹配，返回结果。
	if r.Wchildren != nil {
		params.Combine(r.Wchildren.params)
		params.Add(r.Wchildren.name, searchKey)
		return r.Wchildren.handlers
	}

	// can't match, return nil
	// 无法匹配，返回空
	return nil
}

func (r *radixNode) deleteRoute(path string, isany bool, val HandlerFuncs) {
	nodes := r.findNode(path)
	if nodes == nil {
		return
	}
	last := nodes[len(nodes)-1]
	// clean handler
	if isany {
		if last.anyin {
			last.handlers = nil
		}
		last.anyin = false
		last.anyhandlers = nil
	} else {
		last.handlers = last.anyhandlers
	}
	if last.handlers == nil {
		last.params = nil
	}
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

func (r *radixNode) findNode(path string) []*radixNode {
	args := getSplitPath(path)
	nodes := make([]*radixNode, 1, len(args)*2)
	nodes[0] = r
	for _, i := range args {
		last := nodes[len(nodes)-1]
		switch i[0] {
		case '*':
			if last.Wchildren == nil || (last.Wchildren.name != i && last.Wchildren.name != i[1:]) {
				return nil
			}
			nodes = append(nodes, last.Wchildren)
		case ':':
			var islook bool
			for _, child := range last.Pchildren {
				if child.name == i[1:] {
					islook = true
					nodes = append(nodes, child)
					break
				}
			}
			if !islook {
				return nil
			}
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

func (r *radixNode) findNodeConst(path string) []*radixNode {
	if path == "" {
		return []*radixNode{r}
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

func (r *radixNode) IsZero() bool {
	return r.handlers == nil && len(r.Cchildren) == 0 && len(r.Pchildren) == 0 && r.Wchildren == nil
}

func (r *radixNode) IsMarge() bool {
	if r.kind == radixNodeKindConst && r.handlers == nil && len(r.Cchildren) == 1 && len(r.Pchildren) == 0 && r.Wchildren == nil {
		r.Cchildren[0].path = r.path + r.Cchildren[0].path
		*r = *r.Cchildren[0]
		return true
	}
	return false
}

func (r *radixNode) deleteNode(node *radixNode) {
	switch node.kind {
	case radixNodeKindConst:
		r.Cchildren = radixRemoveNode(r.Cchildren, node)
	case radixNodeKindParam:
		r.Pchildren = radixRemoveNode(r.Pchildren, node)
	case radixNodeKindWildcard:
		r.Wchildren = nil
	}
}

func radixRemoveNode(nodes []*radixNode, node *radixNode) []*radixNode {
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
/api/*|{^0/api\\S+$}	[/api/ *|^0 /api\S+$]
/api/*|^\\$\\d+$		[/api/ *|^\$\d+$]
*/
func getSplitPath(key string) []string {
	if len(key) < 2 {
		return []string{"/"}
	}
	var strs []string
	var length = -1
	var isall = 0
	var isconst = false
	for i := range key {
		// 块模式匹配
		if isall > 0 {
			switch key[i] {
			case '{':
				isall++
			case '}':
				isall--
			}
			if isall > 0 {
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
		case ':', '*':
			// 变量模式
			isconst = false
			length++
			strs = append(strs, "")
		case '{':
			isall++
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
