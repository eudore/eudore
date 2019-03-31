package eudore

/*
基于基数树算法实现一个标准路由器
*/

import (
	"strings"
)

const (
	fullNodeKindConst	uint8	=	iota	// 常量
	fullNodeKindRegex		// 参数正则或函数校验
	fullNodeKindParam		// 参数
	fullNodeKindValid		// 通配符正则或函数校验
	fullNodeKindWildcard	// 通配符
)

type (
	// Routing data check function
	//
	// 路由数据校验函数
	RouterCheckFunc func(string) bool
	// RouterFindFunc func() []string
	
	// Routing data validation function creation function
	//
	// Construct a check function by specifying a string
	//
	// 路由数据校验函数的创建函数
	// 
	// 通过指定字符串构造出一个校验函数
	RouterNewCheckFunc func(string) RouterCheckFunc

	// RouterFull is implemented based on the radix tree to implement all router related features.
	//
	// With path parameters, wildcard parameters, default parameters, parameter verification, wildcard verification, multi-parameter regular capture is not implemented.
	//
	// RouterFull基于基数树实现，实现全部路由器相关特性。
	//
	// 具有路径参数、通配符参数、默认参数、参数校验、通配符校验，未实现多参数正则捕捉。
	RouterFull struct {
		RouterMethod
		midds		*middNode
		node404		HandlerFuncs
		node405		*fullNode
		root		fullNode
		get			fullNode
		post		fullNode
		put			fullNode
		delete		fullNode
		options		fullNode
		head		fullNode
		patch		fullNode
	}
	fullNode struct {
		// 当前node的路径
		path		string
		// node类型
		kind		uint8
		// 
		pnum		uint8
		// 当前节点返回参数名称
		name		string
		// 常量Node
		Cchildren	[]*fullNode
		// 校验参数Node
		Rchildren	[]*fullNode
		// 参数Node
		Pchildren	[]*fullNode
		// 校验通配符Node
		Vchildren	[]*fullNode
		// 通配符Node
		Wchildren	*fullNode
		// 默认标签的名称和值
		tags		[]string
		vals		[]string
		// 校验函数
		check		RouterCheckFunc
		// 正则捕获名称和函数
		// names		[]string
		// find		RouterFindFunc
		// 路由匹配的处理者
		handlers	HandlerFuncs
	}
)


func NewRouterFull(interface{}) (Router, error) {
	r := &RouterFull{
		node404:	HandlerFuncs{DefaultRouter404Func},
		node405:	newFullNode405("405", DefaultRouter405Func),
		midds:		&middNode{},
	}
	r.RouterMethod = &RouterMethodStd{RouterCore:	r}
	return r, nil
}

// Register the middleware into the middleware tree and append the handler if it exists.
//
// 注册中间件到中间件树中，如果存在则追加处理者。
func (r *RouterFull) RegisterMiddleware(method ,path string, hs HandlerFuncs) {
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
			r.midds.recursiveInsertTree(method + path, method + path, hs)
		}
	}else {
		r.midds.recursiveInsertTree(method + path, method + path, hs)		
	}
}

// Register a new method request path to the router
//
// The router matches the handlers available to the current path from the middleware tree and adds them to the front of the handler.
//
// 给路由器注册一个新的方法请求路径
// 
// 路由器会从中间件树中匹配当前路径可使用的处理者，并添加到处理者前方。
func (r *RouterFull) RegisterHandler(method string, path string, handler HandlerFuncs) {
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
func (t *RouterFull) InsertRoute(method, key string, val HandlerFuncs) {
	var currentNode, newNode *fullNode = t.getTree(method), nil
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
		newNode = newFullNode(path)
		// If it is a constant node, recursively add by radix tree rule
		// 如果是常量节点，按基数树规则递归添加
		if newNode.kind == fullNodeKindConst {
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
func (t *RouterFull) Match(method, path string, params Params) HandlerFuncs {
	if n := t.getTree(method).recursiveLoopup(path, params); n != nil {
		return n
	}
	return t.node404
}

// The router Set method can set the handlers for 404 and 405.
//
// 路由器Set方法可以设置404和405的处理者。
func (r *RouterFull) Set(key string, i interface{}) error {
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
		r.node405 = newFullNode405(key, h)
	}
	return nil
}


// Returns the component name of the current router.
//
// 返回当前路由器的组件名称。
func (*RouterFull) GetName() string {
	return ComponentRouterFullName
}

// Returns the component version of the current router.
//
// 返回当前路由器的组件版本。
func (*RouterFull) Version() string {
	return ComponentRouterFullVersion
}

// Create a 405 response radixNode.
//
// 创建一个405响应的radixNode。
func newFullNode405(args string, h HandlerFunc) *fullNode {
	newNode := &fullNode{
		Wchildren:	&fullNode{
			handlers:	HandlerFuncs{h},
		},
	}
	newNode.Wchildren.SetTags(strings.Split(args, " "))
	return newNode
}

func newFullNode(path string) *fullNode {
		newNode := &fullNode{path: path}
	// Create a different Node with a more path prefix type
		// 更具路径前缀类型，创建不同的Node
		switch path[0] {
		// 通配符Node
		case '*':
			newNode.kind = fullNodeKindWildcard
			if len(path) == 1 {
				newNode.name = "*"
			}else{
				newNode.name = path[1:]
				// 如果路径后序具有'|'符号，则截取后端名称返回校验函数
				// 并升级成校验通配符Node
				if name, fn := loadCheckFunc(path); len(name) > 0 {
					// 无法获得校验函数抛出错误
					if fn == nil {
						panic("loadCheckFunc path is invalid, load func failure " + path)
					}
					newNode.kind, newNode.name, newNode.check = fullNodeKindValid, name, fn
				}
			}
		// 参数Node
		case ':':
			newNode.kind = fullNodeKindParam
			newNode.name = path[1:]
			// 如果路径后序具有'|'符号，则截取后端名称返回校验函数
			// 并升级成校验参数Node
			if name, fn := loadCheckFunc(path); len(name) > 0 {
				if fn == nil {
					panic("loadCheckFunc path is invalid, load func failure " + path)
				}
				newNode.kind, newNode.name, newNode.check = fullNodeKindRegex, name, fn
			}
		// 常量Node
		default:
			newNode.kind = fullNodeKindConst
		}
		return newNode
}

// Load the checksum function by name.
//
// 根据名称加载校验函数。
func loadCheckFunc(path string) (string, RouterCheckFunc) {
	// invalid path
	// 无效路径
	if len(path) == 0 || (path[0] != ':' && path[0] != '*') {
		return "", nil
	}
	path = path[1:]
	// Intercept parameter name and check function name
	// 截取参数名称和校验函数名称
	name, fname := split2byte(path, '|')
	if len(name) == 0 {
		return "", nil
	}
	// regular
	// If it is the beginning of a regular expression, add the default regular check function name.
	// 正则
	// 如果是正则表达式开头，添加默认正则校验函数名称。
	if fname[0] == '^' {
		fname = "regexp:" + fname
	}
	// Determine if there is ':'
	// 判断是否有':'
	f2name, arg := split2byte(fname, ':')
	// no ':' is a fixed function, return directly
	// 没有':'为固定函数，直接返回
	if len(arg) == 0 {
		return name, ConfigLoadRouterCheckFunc(fname)
	}
	// There is a ':' variable function to create a checksum function
	// 有':'为变量函数，创建校验函数
	fn := ConfigLoadRouterNewCheckFunc(f2name)(arg)
	// save the newly created checksum function
	// 保存新建的校验函数
	ConfigSaveRouterCheckFunc(fname, fn)
	return name, fn
}

// Add a child node to the node.
//
// 给节点添加一个子节点。
func (r *fullNode) InsertNode(path string, newNode *fullNode) *fullNode {
	if len(path) == 0 {
		return r
	}
	newNode.path = path
	switch newNode.kind {
	case fullNodeKindParam:
		// parameter node
		// 参数节点

		// The path exists to return the old Node
		// 路径存在返回旧Node
		for _, i := range r.Pchildren {
			if i.path == path {
				return i
			}
		}
		r.pnum++
		r.Pchildren = append(r.Pchildren, newNode)
	case fullNodeKindRegex:
		// parameter check node
		// 参数校验节点
		for _, i := range r.Rchildren {
			if i.path == path {
				return i
			}
		}
		r.pnum++
		r.Rchildren = append(r.Rchildren, newNode)
	case fullNodeKindValid:
		// wildcard check Node
		// 通配符校验Node
		for _, i := range r.Vchildren {
			if i.path == path {
				return i
			}
		}
		r.Vchildren = append(r.Vchildren, newNode)
	case fullNodeKindWildcard:
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
func (r *fullNode) SetTags(args []string) {
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
func (r *fullNode) GetTags(p Params) {
	for i, _ := range r.tags {
		p.AddParam(r.tags[i], r.vals[i])
	}
}

// Bifurcate the child node whose path is edgeKey, and the fork common prefix path is pathKey
//
// 对指定路径为edgeKey的子节点分叉，分叉公共前缀路径为pathKey
func (r *fullNode) SplitNode(pathKey, edgeKey string) *fullNode {
	for i, _ := range r.Cchildren {
		// Find the child node whose path is the edgeKey path, then fork
		// 找到路径为edgeKey路径的子节点，然后分叉
		if r.Cchildren[i].path == edgeKey {
			newNode := &fullNode{path: pathKey}

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
func (t *RouterFull) getTree(method string) *fullNode {
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
func (currentNode *fullNode) recursiveInsertTree(containKey string, targetNode *fullNode) *fullNode {
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


func (searchNode *fullNode) recursiveLoopup(searchKey string, params Params) HandlerFuncs {
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
	if searchNode.pnum != 0 && len(searchKey) > 0 {
		// Find the string cutting position
		// strings.IndexByte is a C implementation, does not have a more efficient method, and is more readable.
		// 寻找字符串切割位置
		// strings.IndexByte为c语言实现，未拥有更有效方法，且可读性更强。
		pos := strings.IndexByte(searchKey, '/')
		if pos == -1 {
			pos = len(searchKey) 
		}
		// currentKey is the current string to be processed
		// nextSearchKey starts with the remaining string starting with the first '/'
		// currentKey为当前需要处理的字符串
		// nextSearchKey为剩余字符串为第一个'/'开头
		currentKey, nextSearchKey := searchKey[:pos], searchKey[pos:]

		// check parameter matching
		// 校验参数匹配
		for _, edgeObj := range searchNode.Rchildren {
			// All parameter nodes are recursively judged
			// 所有参数节点递归判断
			if edgeObj.check(currentKey) {
				if n := edgeObj.recursiveLoopup(nextSearchKey, params);n != nil {
					// Add variable parameters
					// 添加变量参数
					params.AddParam(edgeObj.name, currentKey)
					return n
				}
			}
		}
		// 参数匹配
		// 变量Node依次匹配是否满足
		for _, edgeObj := range searchNode.Pchildren {
			// All parameter nodes are recursively judged
			// 所有参数节点递归判断
			if n := edgeObj.recursiveLoopup(nextSearchKey, params);n != nil {
				// Add variable parameters
				// 添加变量参数
				params.AddParam(edgeObj.name, currentKey)
				return n
			}
		}
	}
	// wildcard matching
	// 通配符匹配
	
	// wildcard verification match
	// If the current Node has a wildcard processing method that directly matches, the result is returned.
	// 通配符校验匹配
	// 若当前Node有通配符处理方法直接匹配，返回结果。
	for _, edgeObj := range searchNode.Vchildren {
		if edgeObj.check(searchKey) {
			edgeObj.GetTags(params)
			params.AddParam(edgeObj.name, searchKey)
			return edgeObj.handlers
		}
	}

	// If the current Node has a wildcard processing method that directly matches, the result is returned.
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
