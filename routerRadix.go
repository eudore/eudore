package eudore

/*
基于基数树算法实现一个标准功能的路由器。
*/

import (
	"strings"
)

const (
	radixNodeKindConst	uint8	=	iota
	radixNodeKindParam
	radixNodeKindWildcard
)

type (
	// Basic function router based on radix tree implementation.
	//
	// There are three basic functions: path parameter, wildcard parameter, default parameter, and parameter verification.
	// 基于基数树实现的基本功能路由器。
	//
	// 具有路径参数、通配符参数、默认参数三项基本功能。
	RouterRadix struct {
		RouterMethod
		// save middleware
		// 保存注册的中间件信息
		midds		*middNode
		// exception handling method
		// 异常处理方法
		node404		HandlerFuncs
		node405		*radixNode
		// various methods routing tree
		// 各种方法路由树
		root		radixNode
		get			radixNode
		post		radixNode
		put			radixNode
		delete		radixNode
		options		radixNode
		head		radixNode
		patch		radixNode
	}
	// radix节点的定义
	radixNode struct {
		// 基本信息
		kind		uint8
		path		string
		name		string
		// 每次类型子节点
		Cchildren	[]*radixNode
		Pchildren	[]*radixNode
		Wchildren	*radixNode
		// 当前节点的数据
		tags		[]string
		vals		[]string
		handlers		HandlerFuncs
	}
)


func NewRouterRadix(interface{}) (Router, error) {
	r := &RouterRadix{
		node404:	HandlerFuncs{DefaultRouter404Func},
		node405:	newRadixNode405("405", DefaultRouter405Func),
		midds:		&middNode{},
	}
	r.RouterMethod = &RouterMethodStd{RouterCore:	r}
	return r, nil
}

// Register the middleware into the middleware tree and append the handler if it exists.
//
// If the method is not empty, the path is empty and the modified path is '/'.
//
// 注册中间件到中间件树中，如果存在则追加处理者。
//
// 如果方法非空，路径为空，修改路径为'/'。
func (r *RouterRadix) RegisterMiddleware(method ,path string, hs HandlerFuncs) {
	if len(method) != 0 && len(path) == 0 {
		path = "/"
	}
	if method == MethodAny {
		if path == "/" {
			r.midds.Insert("", hs)
			return
		}
		for _, method = range RouterAllMethod {
			r.midds.Insert(method + path, hs)
		}
	}else {
		r.midds.Insert(method + path, hs)
	}	
}

// Register a new method request path to the router
//
// The router matches the handlers available to the current path from the middleware tree and adds them to the front of the handler.
//
// 给路由器注册一个新的方法请求路径
// 
// 路由器会从中间件树中匹配当前路径可使用的处理者，并添加到处理者前方。
func (r *RouterRadix) RegisterHandler(method string, path string, handler HandlerFuncs) {
	if method == MethodAny{
		for _, method := range RouterAllMethod {
			r.InsertRoute(method, path, CombineHandlers(r.midds.Lookup(method + path), handler))
		}
	}else {
		r.InsertRoute(method, path, CombineHandlers(r.midds.Lookup(method + path), handler))
	}
}

// Add a new routing node.
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
// 路径切割见getSpiltPath函数，当前未完善，处理正则可能异常。
func (t *RouterRadix) InsertRoute(method, key string, val HandlerFuncs) {
	var currentNode *radixNode = t.getTree(method)
	if currentNode == t.node405 {
		return
	}

	// 创建节点
	args := strings.Split(key, " ")
	for _, path := range getSpiltPath(args[0]) {
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
func (t *RouterRadix) Match(method, path string, params Params) HandlerFuncs {
	if n := t.getTree(method).recursiveLoopup(path, params); n != nil {
		return n
	}
	// t.node404.GetTags(params)
	return t.node404
}

// The router Set method can set the handlers for 404 and 405.
//
// 路由器Set方法可以设置404和405的处理者。
func (r *RouterRadix) Set(key string, i interface{}) error {
	h, ok := i.(HandlerFunc)
	if !ok {
		h, ok = i.(func(Context))
	}
	if !ok {
		return ErrRouterSetNoSupportType
	}
	args := strings.Split(key, " ")
	switch args[0] {
	case "404":
		r.node404 = HandlerFuncs{h}
	case "405":
		r.node405 = newRadixNode405(key, h)
	}
	return nil
}

// Returns the component name of the current router.
//
// 返回当前路由器的组件名称。
func (*RouterRadix) GetName() string {
	return ComponentRouterRadixName
}

// Returns the component version of the current router.
//
// 返回当前路由器的组件版本。
func (*RouterRadix) Version() string {
	return ComponentRouterRadixVersion
}

// Create a 405 response radixNode.
//
// 创建一个405响应的radixNode。
func newRadixNode405(args string, h HandlerFunc) *radixNode {
	newNode := &radixNode{
		Wchildren:	&radixNode{
			handlers:	HandlerFuncs{h},
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
		}else{
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
		for i, _ := range r.Cchildren {
			subStr, find := getSubsetPrefix(path, r.Cchildren[i].path)
			if find {
				if subStr == r.Cchildren[i].path {
					nextTargetKey := strings.TrimPrefix(path, r.Cchildren[i].path)
					return r.Cchildren[i].InsertNode(nextTargetKey, nextNode)	
				}else {
					newNode := r.SplitNode(subStr, r.Cchildren[i].path)
					if newNode == nil {
						panic("Unexpect error on split node")
					}
					return newNode.InsertNode(strings.TrimPrefix(path, subStr), nextNode)
				}
			}
		}
		r.Cchildren = append(r.Cchildren, nextNode)
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

// Bifurcate the child node whose path is edgeKey, and the fork common prefix path is pathKey.
//
// 对指定路径为edgeKey的子节点分叉，分叉公共前缀路径为pathKey。
func (r *radixNode) SplitNode(pathKey, edgeKey string) *radixNode {
	for i, _ := range r.Cchildren {
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
		r.tags[i + 1], r.vals[i + 1] = split2byte(str, ':')
	}
}

// Give the current Node tag to Params
//
// 将当前Node的tags给予Params
func (r *radixNode) AddTagsToParams(p Params) {
	for i, _ := range r.tags {
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
func (t *RouterRadix) getTree(method string) *radixNode {
	switch method {
	case MethodGet:
		return &t.get
	case MethodPost:
		return &t.post
	case MethodDelete:
		return &t.delete
	case MethodPut:
		return &t.put
	case MethodHead:
		return &t.head
	case MethodOptions:
		return &t.options
	case MethodPatch:
		return &t.patch
	default:
		return t.node405
	}
}

// 按照顺序匹配一个路径。
//
// 依次检查常量节点、参数节点、通配符节点，如果有一个匹配就直接返回。
func (searchNode *radixNode) recursiveLoopup(searchKey string, params Params) HandlerFuncs {
	// 如果路径为空，当前节点就是需要匹配的节点，直接返回。
	if len(searchKey) == 0 && searchNode.handlers != nil {
		searchNode.AddTagsToParams(params)
		return searchNode.handlers
	}

	// 遍历常量Node匹配，寻找具有相同前缀的那个节点
	for _, edgeObj := range searchNode.Cchildren {
		if contrainPrefix(searchKey, edgeObj.path) {
			nextSearchKey := searchKey[len(edgeObj.path):]
			if n := edgeObj.recursiveLoopup(nextSearchKey, params);n != nil {
				return n
			}
			// TODO: 待优化测试，只有一个相同前缀，当前应该直接退出遍历
			break
		}
	}

	if len(searchNode.Pchildren) > 0 && len(searchKey) > 0 {
		pos := strings.IndexByte(searchKey, '/')
		if pos == -1 {
			pos = len(searchKey) 
		}
		nextSearchKey := searchKey[pos:]

		// Whether the variable Node matches in sequence is satisfied
		// 遍历参数节点是否后续匹配
		for _, edgeObj := range searchNode.Pchildren {
			if n := edgeObj.recursiveLoopup(nextSearchKey, params);n != nil {
				params.AddParam(edgeObj.name, searchKey[:pos])
				return n
			}
		}
	}
	
	// If the current Node has a wildcard processing method that directly matches, the result is returned.
	// 若当前节点有通配符处理方法直接匹配，返回结果。
	if searchNode.Wchildren != nil {
		searchNode.Wchildren.AddTagsToParams(params)
		params.AddParam(searchNode.Wchildren.name, searchKey)
		return searchNode.Wchildren.handlers
	}

	// can't match, return nil
	// 无法匹配，返回空
	return nil
}

/*
The string is cut according to the Node type.
The current segmentation rule is not detailed enough, and the special characters are mis-segmented.
将字符串按Node类型切割，当前分割规则不够详细，特殊字符会误分割。

String path cutting example:
字符串路径切割例子：
/				[/]
/api/note/		[/api/note/]
//api/*			[/api/ *]
//api/*name		[/api/ *name]
/api/get/		[/api/get/]
/api/get		[/api/get]
/api/:get		[/api/ :get]
/api/:get/*			[/api/ :get / *]
/api/:name/info/*	[/api/ :name /info/ *]
*/
func getSpiltPath(key string) []string {
	if len(key) < 2 {
		return []string{"/"}
	}
	var strs []string
	var last bool = false
	for i, str := range strings.Split(key, "/") {
		// Filter the '/' in the path
		// 过滤路径中的'/'
		if len(str) == 0 {
			if i == 0 {
				strs = []string{"/"}
			}
			continue
		}
		// Supplemental separator
		// 补充分隔符
		if last {
			last = false
			strs = append(strs, "/")
		}else {
			lastappend(strs, '/')
		}
		// Handling special prefix paths
		// 处理特殊前缀路径
		if lastisbyte(strs, '/') && (str[0] == ':' || str[0] == '*') {
			strs = append(strs, str)
			last = true
			continue
		}
		// append constants
		// 追加常量
		num := len(strs) - 1
		strs[num] = strs[num] + str
	
	}
	return strs
}

// Modify the last string to end with the specified byte.
//
// 修改最后一个字符串结尾为指定byte。
func lastappend(strs []string, b byte) {
	num := len(strs) - 1
	laststr := strs[num]
	if laststr[len(laststr) - 1 ] != b {
		strs[num] = strs[num] + string(b)
	}
}

// Check if the end of the last string is the specified byte.
//
// 检测最后一个字符串的结尾是否为指定byte。
func lastisbyte(strs []string, b byte) bool {
	num := len(strs) - 1
	if num < 0 {
		return false
	}
	laststr := strs[num]
	return laststr[len(laststr) - 1 ] == b
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
	for i := 0; i < len(str2) ; i++ {
		if str1[i] != str2[i] {
			return false
		}
	}

	return true
}
