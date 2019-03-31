package eudore

/*
基于基数树算法实现一个完整功能的路由器。
未实现正则参数捕捉功能。
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
		// 保存中间件
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
	radixNode struct {
		kind		uint8
		path		string
		name		string
		Cchildren	[]*radixNode
		Pchildren	[]*radixNode
		Wchildren	*radixNode
		// data
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
// 注册中间件到中间件树中，如果存在则追加处理者。
func (r *RouterRadix) RegisterMiddleware(method ,path string, hs HandlerFuncs) {
	// Correct the data: If the method is not empty, the path is empty and the modified path is '/'.
	// 修正数据：如果方法非空，路径为空，修改路径为'/'。
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

// Add a new route Node.
//
// If the method does not support it will not be added, request to change the path will respond 405
//
// 添加一个新的路由Node。
//
// 如果方法不支持则不会添加，请求改路径会响应405
func (t *RouterRadix) InsertRoute(method, key string, val HandlerFuncs) {
	var currentNode, newNode *radixNode = t.getTree(method), nil
	// Unsupported request method, return 405 processing tree directly
	// 未支持的请求方法,直接返回405处理树
	if currentNode == t.node405 {
		return
	}
	args := strings.Split(key, " ")
	// Cut the path by Node and append it down one by one.
	// 将路径按Node切割，然后依次向下追加。
	for _, path := range getSpiltPath(args[0]) {
		// Create a new radixNode and set the Node type
		// 创建一个新的radixNode，并设置Node类型
		newNode = newRadixNode(path)
		// If it is a constant node, recursively add by radix tree rule
		// 如果是常量节点，按基数树规则递归添加
		if newNode.kind == radixNodeKindConst {
			currentNode = currentNode.recursiveInsertTree(path, newNode)
		}else {
			currentNode = currentNode.InsertNode(path, newNode)
		}		
	}

	// Tree fork ends Node is set to the current routing data.
	// 树分叉结尾Node设置为当前路由数据。
	newNode.handlers = val
	newNode.SetTags(args)
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
// 创建一个Radix树Node，会根据当前路由设置不同的节点类型。
//
// '*'前缀为通配符节点，':'前缀为参数节点，其他未常量节点。
func newRadixNode(path string) *radixNode {
	newNode := &radixNode{path: path}
	// Create a different Node with a more path prefix type
	// 更具路径前缀类型，创建不同的Node
	switch path[0] {
	// wildcard Node
	// 通配符Node
	case '*':
		newNode.kind = radixNodeKindWildcard
		if len(path) == 1 {
			// set the default name
			// 设置默认名称
			newNode.name = "*"
		}else{
			newNode.name = path[1:]
		}
	// parameter Node
	// 参数Node
	case ':':
		newNode.kind = radixNodeKindParam
		newNode.name = path[1:]
	// constant Node
	// 常量Node
	default:
		newNode.kind = radixNodeKindConst
	}
	return newNode
}

// Add a child node to the node.
//
// 给节点添加一个子节点。
func (r *radixNode) InsertNode(path string, newNode *radixNode) *radixNode {
	if len(path) == 0 {
		return r
	}
	newNode.path = path
	switch newNode.kind {
	case radixNodeKindParam:
		// The path exists to return the old Node
		// 路径存在返回旧Node
		for _, i := range r.Pchildren {
			if i.path == path {
				return i
			}
		}
		r.Pchildren = append(r.Pchildren, newNode)
	case radixNodeKindWildcard:
		// Set the wildcard Node data.
		// 设置通配符Node数据。
		r.Wchildren = newNode
	default:
		// Append the constant Node, the constant Node is inserted recursively, and the same prefix will not appear.
		// 追加常量Node，常量Node由递归插入，不会出现相同前缀。
		r.Cchildren = append(r.Cchildren, newNode)
	}
	return newNode
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
func (r *radixNode) GetTags(p Params) {
	for i, _ := range r.tags {
		p.AddParam(r.tags[i], r.vals[i])
	}
}

// Bifurcate the child node whose path is edgeKey, and the fork common prefix path is pathKey
//
// 对指定路径为edgeKey的子节点分叉，分叉公共前缀路径为pathKey
func (r *radixNode) SplitNode(pathKey, edgeKey string) *radixNode {
	for i, _ := range r.Cchildren {
		// Find the child node whose path is the edgeKey path, then fork
		// 找到路径为edgeKey路径的子节点，然后分叉
		if r.Cchildren[i].path == edgeKey {
			newNode := &radixNode{path: pathKey}

			r.Cchildren[i].path = strings.TrimPrefix(edgeKey, pathKey)
			newNode.Cchildren = append(newNode.Cchildren, r.Cchildren[i])

			r.Cchildren[i] = newNode
			// return the fork to the newly created Node
			// 返回分叉新创建的Node
			return newNode
		}
	}
	return nil
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

// Recursively add a constant Node with a path of containKey to the current node
//
// targetKey and targetValue are new Node data.
//
// 给当前节点递归添加一个路径为containKey的常量Node
//
// targetKey和targetValue为新Node数据。
func (currentNode *radixNode) recursiveInsertTree(containKey string, targetNode *radixNode) *radixNode {
	for i, _ := range currentNode.Cchildren {
		// Traverse to check if the path of the current child node and the insertion path of the new node have a common path
		// subStr is the public path of the two, find indicates whether there is
		// 遍历检查当前子节点的路径和新节点的插入路径是否有公共路径
		// subStr是两者的公共路径，find表示是否有
		subStr, find := getSubsetPrefix(containKey, currentNode.Cchildren[i].path)
		if find {
			// If the current child node's path is equal to the public maximum path, then the current child node adds a new node
			// 如果当前子节点的路径等于公共最大路径，则该当前子节点添加新节点
			if subStr == currentNode.Cchildren[i].path {
				// The path to the new node is the insertion path that first filters the back part of the public path.
				// 新节点的路径为插入路径先过滤公共路径的后面部分。
				nextTargetKey := strings.TrimPrefix(containKey, currentNode.Cchildren[i].path)
				// The current child node original has more than one child, so you need to add it recursively.
				// 当前子节点原版存在原本有多个child，所以需要递归添加
				return currentNode.Cchildren[i].recursiveInsertTree(nextTargetKey, targetNode)	
			}else {
				// If the public path is not equal to the path of the child node
				// will fork the path of the current child node
				// The child node and the new node after the fork have the public path, then add the new node
				// 如果公共路径不等于子节点的路径
				// 则将当前子节点的路径分叉
				// 分叉后的子节点和新节点就拥有了公共路径，然后添加新节点
				newNode := currentNode.SplitNode(subStr, currentNode.Cchildren[i].path)
				if newNode == nil {
					panic("Unexpect error on split node")
				}

				// Add a new node
				// After the fork, the tree must have only one child node without the same path, so add a new node directly.
				// 添加新的节点
				// 分叉后树一定只有一个没有相同路径的子节点，所以直接添加新节点。
				return newNode.InsertNode(strings.TrimPrefix(containKey, subStr), targetNode)
			}
		}
	}

	// All child nodes do not have the same prefix path, directly add a new node to a new child node
	// 所有子节点都没有相同前缀路径存在，直接添加新节点为一个新的子节点
	return currentNode.InsertNode(containKey, targetNode)
}


// 安装顺序匹配一个路径。
func (searchNode *radixNode) recursiveLoopup(searchKey string, params Params) HandlerFuncs {
	// constant match, return data
	// 常量匹配，返回数据
	if len(searchKey) == 0 && searchNode.handlers != nil {
		searchNode.GetTags(params)
		return searchNode.handlers
	}

	// Traverse constant Node match
	// 遍历常量Node匹配
	for _, edgeObj := range searchNode.Cchildren {
		// Find the same prefix node to further match
		// The current Node path must be the prefix of the searchKey.
		// 寻找相同前缀node进一步匹配
		// 当前Node的路径必须为searchKey的前缀。
		if contrainPrefix(searchKey, edgeObj.path) {
			// Remove the prefix path to get an unmatched path.
			// 除去前缀路径，获得未匹配路径。
			nextSearchKey := searchKey[len(edgeObj.path):]
			// Then the current Node recursively judges
			// 然后当前Node递归判断
			if n := edgeObj.recursiveLoopup(nextSearchKey, params);n != nil {
				return n
			}
		}
	}

	// parameter matching
	// Check if there is a parameter match
	// 参数匹配
	// 检测是否存在参数匹配
	if len(searchNode.Pchildren) > 0 && len(searchKey) > 0 {
		// Find the string cutting position
		// strings.IndexByte is a C implementation, does not have a more efficient method, and is more readable.
		// 寻找字符串切割位置
		// strings.IndexByte为c语言实现，未拥有更有效方法，且可读性更强。
		pos := strings.IndexByte(searchKey, '/')
		if pos == -1 {
			pos = len(searchKey) 
		}
		// The remaining string is cut to the beginning of the first '/'
		// 剩余字符串切割为第一个'/'开头
		nextSearchKey := searchKey[pos:]

		// Whether the variable Node matches in sequence is satisfied
		// 变量Node依次匹配是否满足
		for _, edgeObj := range searchNode.Pchildren {
			// All parameter nodes are recursively judged
			// 所有参数节点递归判断
			if n := edgeObj.recursiveLoopup(nextSearchKey, params);n != nil {
				// Add variable parameters
				// 添加变量参数
				params.AddParam(edgeObj.name, searchKey[:pos])
				return n
			}
		}
	}
	
	// wildcard matching
	// If the current Node has a wildcard processing method that directly matches, the result is returned.
	// 通配符匹配
	// 若当前Node有通配符处理方法直接匹配，返回结果。
	if searchNode.Wchildren != nil {
		searchNode.Wchildren.GetTags(params)
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
