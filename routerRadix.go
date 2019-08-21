package eudore

/*
基于基数树算法实现一个标准功能的路由器。
*/

import (
	"strings"
)

const (
	radixNodeKindConst uint8 = iota
	radixNodeKindParam
	radixNodeKindWildcard
)

type (
	// RouterRadix basic function router based on radix tree implementation.
	//
	// There are three basic functions: path parameter, wildcard parameter, default parameter, and parameter verification.
	// RouterRadix基于基数树实现的基本功能路由器。
	//
	// 具有路径参数、通配符参数、默认参数三项基本功能。
	RouterRadix struct {
		RouterMethod
		// save middleware
		// 保存注册的中间件信息
		Print    func(...interface{}) `set:"print"`
		middtree *middNode
		// exception handling method
		// 异常处理方法
		node404     radixNode
		nodefunc404 HandlerFuncs
		node405     radixNode
		nodefunc405 HandlerFuncs
		// various methods routing tree
		// 各种方法路由树
		root    radixNode
		get     radixNode
		post    radixNode
		put     radixNode
		delete  radixNode
		options radixNode
		head    radixNode
		patch   radixNode
	}
	// radix节点的定义
	radixNode struct {
		// 基本信息
		kind uint8
		path string
		name string
		// 每次类型子节点
		Cchildren []*radixNode
		Pchildren []*radixNode
		Wchildren *radixNode
		// 当前节点的数据
		tags     []string
		vals     []string
		handlers HandlerFuncs
	}
)

// NewRouterRadix 创建一个Radix路由器。
func NewRouterRadix() Router {
	r := &RouterRadix{
		Print:       func(...interface{}) {},
		nodefunc404: HandlerFuncs{DefaultRouter404Func},
		nodefunc405: HandlerFuncs{DefaultRouter405Func},
		node404: radixNode{
			tags:     []string{ParamRoute},
			vals:     []string{"404"},
			handlers: HandlerFuncs{DefaultRouter404Func},
		},
		node405: radixNode{
			Wchildren: &radixNode{
				tags:     []string{ParamRoute},
				vals:     []string{"405"},
				handlers: HandlerFuncs{DefaultRouter405Func},
			},
		},
		middtree: &middNode{},
	}
	r.RouterMethod = &RouterMethodStd{
		RouterCore:          r,
		ControllerParseFunc: ControllerBaseParseFunc,
	}
	return r
}

// RegisterMiddleware register the middleware into the middleware tree and append the handler if it exists.
//
// If the method is not empty, the path is empty and the modified path is '/'.
//
// RegisterMiddleware注册中间件到中间件树中，如果存在则追加处理者。
//
// 如果方法非空，路径为空，修改路径为'/'。
func (r *RouterRadix) RegisterMiddleware(method, path string, hs HandlerFuncs) {
	if len(method) != 0 && len(path) == 0 {
		path = "/"
	}
	r.Print("RegisterMiddleware:", method, path, GetHandlerFuncsName(hs))
	if method == MethodAny {
		if path == "/" {
			r.middtree.Insert("", hs)
			r.node404.handlers = append(r.middtree.val, r.nodefunc404...)
			r.node405.Wchildren.handlers = append(r.middtree.val, r.nodefunc405...)
			return
		}
		for _, method = range RouterAllMethod {
			r.middtree.Insert(method+path, hs)
		}
	} else {
		r.middtree.Insert(method+path, hs)
	}
}

// RegisterHandler register a new method request path to the router
//
// The router matches the handlers available to the current path from the middleware tree and adds them to the front of the handler.
//
// RegisterHandler 给路由器注册一个新的方法请求路径
//
// 路由器会从中间件树中匹配当前路径可使用的处理者，并添加到处理者前方。
func (r *RouterRadix) RegisterHandler(method string, path string, handler HandlerFuncs) {
	r.Print("RegisterHandler:", method, path, GetHandlerFuncsName(handler))
	switch method {
	case "NotFound", "404":
		r.nodefunc404 = handler
		r.node404.handlers = CombineHandlerFuncs(handler, r.middtree.val)
	case "MethodNotAllowed", "405":
		r.nodefunc405 = handler
		r.node405.Wchildren.handlers = CombineHandlerFuncs(handler, r.middtree.val)
	case MethodAny:
		for _, method := range RouterAllMethod {
			r.insertRoute(method, path, CombineHandlerFuncs(r.middtree.Lookup(method+path), handler))
		}
	default:
		r.insertRoute(method, path, CombineHandlerFuncs(r.middtree.Lookup(method+path), handler))
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
func (r *RouterRadix) insertRoute(method, key string, val HandlerFuncs) {
	var currentNode *radixNode = r.getTree(method)
	if currentNode == &r.node405 {
		return
	}

	// 创建节点
	args := strings.Split(key, " ")
	for _, path := range getSplitPath(args[0]) {
		currentNode = currentNode.InsertNode(path, newRadixNode(path))
	}

	currentNode.handlers = val
	currentNode.SetTags(args)
}

// Match a request, if the method does not allow direct return to node405, no match returns node404.
//
// Note: 404 does not support extra parameters, not implemented.
//
// 匹配一个请求，如果方法不不允许直接返回node405，未匹配返回node404。
//
// 注意：404不支持额外参数，未实现。
func (r *RouterRadix) Match(method, path string, params Params) HandlerFuncs {
	if n := r.getTree(method).recursiveLoopup(path, params); n != nil {
		return n
	}

	// 处理404
	r.node404.AddTagsToParams(params)
	return r.node404.handlers
}

// Set 方法允许设置Print属性，设置日志输出信息。
func (r *RouterRadix) Set(key string, value interface{}) error {
	switch val := value.(type) {
	case func(...interface{}):
		r.Print = val
	default:
		return ErrSeterNotSupportField
	}
	return nil
}

// Create a 405 response radixNode.
//
// 创建一个405响应的radixNode。
func newRadixNode405(args string, h HandlerFunc) *radixNode {
	newNode := &radixNode{
		Wchildren: &radixNode{
			handlers: HandlerFuncs{h},
		},
	}
	newNode.Wchildren.SetTags(strings.Split(args, " "))
	return newNode
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
func (r *radixNode) InsertNode(path string, nextNode *radixNode) *radixNode {
	if len(path) == 0 {
		return r
	}
	nextNode.path = path
	switch nextNode.kind {
	case radixNodeKindConst:
		for i := range r.Cchildren {
			subStr, find := getSubsetPrefix(path, r.Cchildren[i].path)
			if find {
				if subStr == r.Cchildren[i].path {
					nextTargetKey := strings.TrimPrefix(path, r.Cchildren[i].path)
					return r.Cchildren[i].InsertNode(nextTargetKey, nextNode)
				}
				newNode := r.SplitNode(subStr, r.Cchildren[i].path)
				if newNode == nil {
					panic("Unexpect error on split node")
				}
				return newNode.InsertNode(strings.TrimPrefix(path, subStr), nextNode)
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
		r.Wchildren = nextNode
	default:
		panic("Undefined radix node type")
	}
	return nextNode
}

// SplitNode bifurcate the child node whose path is edgeKey, and the fork common prefix path is pathKey.
//
// SplitNode 对指定路径为edgeKey的子节点分叉，分叉公共前缀路径为pathKey。
func (r *radixNode) SplitNode(pathKey, edgeKey string) *radixNode {
	for i := range r.Cchildren {
		if r.Cchildren[i].path == edgeKey {
			newNode := &radixNode{path: pathKey}
			newNode.Cchildren = append(newNode.Cchildren, r.Cchildren[i])

			r.Cchildren[i].path = strings.TrimPrefix(edgeKey, pathKey)
			r.Cchildren[i] = newNode
			return newNode
		}
	}
	return nil
}

// Set the tags for the current Node
//
// 给当前Node设置tags
func (r *radixNode) SetTags(args []string) {
	if len(args) == 0 {
		return
	}
	r.tags = make([]string, len(args))
	r.vals = make([]string, len(args))
	// The first parameter name defaults to route
	// 第一个参数名称默认为route
	r.tags[0] = ParamRoute
	r.vals[0] = args[0]
	for i, str := range args[1:] {
		r.tags[i+1], r.vals[i+1] = split2byte(str, '=')
	}
}

// AddTagsToParams give the current Node tag to Params
//
// AddTagsToParams 将当前Node的tags给予Params
func (r *radixNode) AddTagsToParams(p Params) {
	for i := range r.tags {
		p.AddParam(r.tags[i], r.vals[i])
	}
}

// Get the tree of the corresponding method.
//
// Support eudore.RouterAllMethod these methods, weak support will return 405 processing tree.
//
// 获取对应方法的树。
//
// 支持eudore.RouterAllMethod这些方法,弱不支持会返回405处理树。
func (r *RouterRadix) getTree(method string) *radixNode {
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
	case MethodOptions:
		return &r.options
	case MethodPatch:
		return &r.patch
	default:
		return &r.node405
	}
}

// 按照顺序匹配一个路径。
//
// 依次检查常量节点、参数节点、通配符节点，如果有一个匹配就直接返回。
func (r *radixNode) recursiveLoopup(searchKey string, params Params) HandlerFuncs {
	// 如果路径为空，当前节点就是需要匹配的节点，直接返回。
	if len(searchKey) == 0 {
		r.AddTagsToParams(params)
		return r.handlers
	}

	// 遍历常量Node匹配，寻找具有相同前缀的那个节点
	for _, edgeObj := range r.Cchildren {
		// fmt.Println(edgeObj.path,searchKey )
		if edgeObj.path[0] >= searchKey[0] {
			if len(searchKey) >= len(edgeObj.path) && searchKey[:len(edgeObj.path)] == edgeObj.path {
				nextSearchKey := searchKey[len(edgeObj.path):]
				if n := edgeObj.recursiveLoopup(nextSearchKey, params); n != nil {
					return n
				}
			}
			break
		}
	}

	if len(r.Pchildren) > 0 && len(searchKey) > 0 {
		pos := strings.IndexByte(searchKey, '/')
		if pos == -1 {
			pos = len(searchKey)
		}
		nextSearchKey := searchKey[pos:]

		// Whether the variable Node matches in sequence is satisfied
		// 遍历参数节点是否后续匹配
		for _, edgeObj := range r.Pchildren {
			if n := edgeObj.recursiveLoopup(nextSearchKey, params); n != nil {
				params.AddParam(edgeObj.name, searchKey[:pos])
				return n
			}
		}
	}

	// If the current Node has a wildcard processing method that directly matches, the result is returned.
	// 若当前节点有通配符处理方法直接匹配，返回结果。
	if r.Wchildren != nil {
		r.Wchildren.AddTagsToParams(params)
		params.AddParam(r.Wchildren.name, searchKey)
		return r.Wchildren.handlers
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
/api/*|^0/api\\S+$		[/api/ *|^0 /api\S+$]
/api/*|^\\$\\d+$		[/api/ *|^\$\d+$]
*/
func getSplitPath(key string) []string {
	if len(key) < 2 {
		return []string{"/"}
	}
	var strs []string
	var length int = -1
	var ismatch bool = false
	var isconst bool = false
	for i := range key {
		if ismatch {
			strs[length] = strs[length] + key[i:i+1]
			if key[i] == '$' && key[i-1] != '\\' && (i == len(key)-1 || key[i+1] == '/') {
				ismatch = false
			}
			continue
		}
		// fmt.Println(last, key[i:i+1])
		switch key[i] {
		case '/':
			if !isconst {
				length++
				strs = append(strs, "")
				isconst = true

			}
		case ':', '*':
			isconst = false
			if key[i-1] == '/' {
				length++
				strs = append(strs, "")
			}
		case '^':
			ismatch = true
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

// Check if the string str2 is the prefix of str1.
//
// 检测字符串str2是否为str1的前缀。
func contrainPrefix(str1, str2 string) bool {
	if len(str1) < len(str2) {
		return false
	}
	for i := 0; i < len(str2); i++ {
		if str1[i] != str2[i] {
			return false
		}
	}

	return true
}
