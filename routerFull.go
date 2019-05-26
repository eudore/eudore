package eudore

/*
基于基数树算法实现一个完整路由器
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
		Print		func(...interface{})	`set:"print"`
		middtree		*middNode
		node404		fullNode
		nodefunc404	HandlerFuncs
		node405		fullNode
		nodefunc405	HandlerFuncs
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
		path		string
		kind		uint8
		pnum		uint8
		name		string
		// 保存各类子节点
		Cchildren	[]*fullNode
		Rchildren	[]*fullNode
		Pchildren	[]*fullNode
		Vchildren	[]*fullNode
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

		Print:		func(...interface{}) {},
		nodefunc404:	HandlerFuncs{DefaultRouter404Func},
		nodefunc405:	HandlerFuncs{DefaultRouter405Func},
		node404:	fullNode{
			tags:	[]string{ParamRoute},
			vals:	[]string{"404"},
			handlers:	HandlerFuncs{DefaultRouter404Func},
		},
		node405:	fullNode{
			Wchildren:	&fullNode{
				tags:	[]string{ParamRoute},
				vals:	[]string{"405"},
				handlers:	HandlerFuncs{DefaultRouter405Func},
			},
		},
		middtree:		&middNode{},
	}
	r.RouterMethod = &RouterMethodStd{
		RouterCore:			r,
		ControllerParseFunc:	ControllerBaseParseFunc,
	}
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
	r.Print("RegisterMiddleware:", method, path, GetHandlerNames(hs))
	if method == MethodAny {
		if path == "/" {
			r.middtree.Insert("", hs)
			r.node404.handlers = append(r.middtree.val, r.nodefunc404...)
			r.node405.Wchildren.handlers = append(r.middtree.val, r.nodefunc405...)
			return
		}
		for _, method = range RouterAllMethod {
			r.middtree.Insert(method + path, hs)		
		}
	}else {
		r.middtree.Insert(method + path, hs)
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
	r.Print("RegisterHandler:", method, path, GetHandlerNames(handler))
	if method == MethodAny{
		for _, method := range RouterAllMethod {
			r.insertRoute(method, path, CombineHandlerFuncs(r.middtree.Lookup(method + path), handler))
		}
	}else {
		r.insertRoute(method, path, CombineHandlerFuncs(r.middtree.Lookup(method + path), handler))
	}
}

// Add a new route Node.
//
// If the method does not support it will not be added, request to change the path will respond 405
//
// 添加一个新的路由Node。
//
// 如果方法不支持则不会添加，请求改路径会响应405
func (t *RouterFull) insertRoute(method, key string, val HandlerFuncs) {
	var currentNode *fullNode = t.getTree(method)
	if currentNode == &t.node405 {
		return
	}

	// 创建节点
	args := strings.Split(key, " ")
	for _, path := range getSpiltPath(args[0]) {
		currentNode = currentNode.InsertNode(path, newFullNode(path))
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
func (t *RouterFull) Match(method, path string, params Params) HandlerFuncs {
	if n := t.getTree(method).recursiveLoopup(path, params); n != nil {
		return n
	}

	// 处理404
	t.node404.AddTagsToParams(params)
	return t.node404.handlers
}

// The router Set method can set the handlers for 404 and 405.
//
// 路由器Set方法可以设置404和405的处理者。
func (r *RouterFull) Set(key string, i interface{}) error {
	args := strings.Split(key, " ")
	switch args[0] {
	case "404":
		hs := NewHandlerFuncs(i)
		if hs == nil {
			return ErrRouterSetNoSupportType
		}
		r.node404.SetTags(args)
		r.node404.handlers = append(r.middtree.val, hs...)
		r.nodefunc404 = hs
	case "405":
		hs := NewHandlerFuncs(i)
		if hs == nil {
			return ErrRouterSetNoSupportType
		}
		r.node405.Wchildren.SetTags(args)
		r.node405.Wchildren.handlers = append(r.middtree.val, hs...)
		r.nodefunc405 = hs
	default:
		return ErrComponentNoSupportField
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

// 创建一个Radix树Node，会根据当前路由设置不同的节点类型和名称。
//
// '*'前缀为通配符节点，':'前缀为参数节点，其他未常量节点,如果通配符和参数结点后带有符号'|'则为校验结点。
func newFullNode(path string) *fullNode {
		newNode := &fullNode{path: path}
		switch path[0] {
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
func (r *fullNode) InsertNode(path string, nextNode *fullNode) *fullNode {
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
		r.Pchildren = append(r.Pchildren, nextNode)
	case fullNodeKindRegex:
		// parameter check node
		// 参数校验节点
		for _, i := range r.Rchildren {
			if i.path == path {
				return i
			}
		}
		r.pnum++
		r.Rchildren = append(r.Rchildren, nextNode)
	case fullNodeKindValid:
		// wildcard check Node
		// 通配符校验Node
		for _, i := range r.Vchildren {
			if i.path == path {
				return i
			}
		}
		r.Vchildren = append(r.Vchildren, nextNode)
	case fullNodeKindWildcard:
		// Set the wildcard Node data.
		// 设置通配符Node数据。
		r.Wchildren = nextNode
	default:
		panic("Undefined radix node type from router full.")
	}
	return nextNode
}

// Bifurcate the child node whose path is edgeKey, and the fork common prefix path is pathKey
//
// 对指定路径为edgeKey的子节点分叉，分叉公共前缀路径为pathKey
func (r *fullNode) SplitNode(pathKey, edgeKey string) *fullNode {
	for i, _ := range r.Cchildren {
		if r.Cchildren[i].path == edgeKey {
			newNode := &fullNode{path: pathKey}
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
func (r *fullNode) AddTagsToParams(p Params) {
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
		return &t.node405
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
		subStr, find := getSubsetPrefix(containKey, currentNode.Cchildren[i].path)
		if find {
			if subStr == currentNode.Cchildren[i].path {
				nextTargetKey := strings.TrimPrefix(containKey, currentNode.Cchildren[i].path)
				return currentNode.Cchildren[i].recursiveInsertTree(nextTargetKey, targetNode)	
			}else {
				newNode := currentNode.SplitNode(subStr, currentNode.Cchildren[i].path)
				if newNode == nil {
					panic("Unexpect error on split node")
				}

				return newNode.InsertNode(strings.TrimPrefix(containKey, subStr), targetNode)
			}
		}
	}

	return currentNode.InsertNode(containKey, targetNode)
}


func (searchNode *fullNode) recursiveLoopup(searchKey string, params Params) HandlerFuncs {

	// constant match, return data
	// 常量匹配，返回数据
	if len(searchKey) == 0 && searchNode.handlers != nil {
		searchNode.AddTagsToParams(params)
		return searchNode.handlers
	}

	// Traverse constant Node match
	// 遍历常量Node匹配
	for _, edgeObj := range searchNode.Cchildren {
		if contrainPrefix(searchKey, edgeObj.path) {
			nextSearchKey := searchKey[len(edgeObj.path):]
			if n := edgeObj.recursiveLoopup(nextSearchKey, params);n != nil {
				return n
			}
			break
		}
	}

	// parameter matching
	// Check if there is a parameter match
	// 参数匹配
	// 检测是否存在参数匹配
	if searchNode.pnum != 0 && len(searchKey) > 0 {
		pos := strings.IndexByte(searchKey, '/')
		if pos == -1 {
			pos = len(searchKey) 
		}
		currentKey, nextSearchKey := searchKey[:pos], searchKey[pos:]

		// check parameter matching
		// 校验参数匹配
		for _, edgeObj := range searchNode.Rchildren {
			if edgeObj.check(currentKey) {
				if n := edgeObj.recursiveLoopup(nextSearchKey, params);n != nil {
					params.AddParam(edgeObj.name, currentKey)
					return n
				}
			}
		}

		// 参数匹配
		// 变量Node依次匹配是否满足
		for _, edgeObj := range searchNode.Pchildren {
			if n := edgeObj.recursiveLoopup(nextSearchKey, params);n != nil {
				params.AddParam(edgeObj.name, currentKey)
				return n
			}
		}
	}
	
	// wildcard verification match
	// If the current Node has a wildcard processing method that directly matches, the result is returned.
	// 通配符校验匹配
	// 若当前Node有通配符处理方法直接匹配，返回结果。
	for _, edgeObj := range searchNode.Vchildren {
		if edgeObj.check(searchKey) {
			edgeObj.AddTagsToParams(params)
			params.AddParam(edgeObj.name, searchKey)
			return edgeObj.handlers
		}
	}

	// If the current Node has a wildcard processing method that directly matches, the result is returned.
	// 若当前Node有通配符处理方法直接匹配，返回结果。
	if searchNode.Wchildren != nil {
		searchNode.Wchildren.AddTagsToParams(params)
		params.AddParam(searchNode.Wchildren.name, searchKey)
		return searchNode.Wchildren.handlers
	}

	// can't match, return nil
	// 无法匹配，返回空
	return nil
}
