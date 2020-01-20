package eudore

/*
基于基数树算法实现一个完整路由器
*/

import (
	"strings"
)

const (
	fullNodeKindConst    uint8 = 1 << iota // 常量
	fullNodeKindRegex                      // 参数正则或函数校验
	fullNodeKindParam                      // 参数
	fullNodeKindValid                      // 通配符正则或函数校验
	fullNodeKindWildcard                   // 通配符
	fullNodeKindAnyMethod
)

type (
	// RouterCoreFull is implemented based on the radix tree to implement all router related features.
	//
	// With path parameters, wildcard parameters, default parameters, parameter verification, wildcard verification, multi-parameter regular capture is not implemented.
	//
	// RouterFull基于基数树实现，实现全部路由器相关特性。
	//
	// 具有路径参数、通配符参数、默认参数、参数校验、通配符校验，未实现多参数正则捕捉。
	RouterCoreFull struct {
		middlewares *trieNode
		node404     fullNode
		nodefunc404 HandlerFuncs
		node405     fullNode
		nodefunc405 HandlerFuncs
		root        fullNode
		get         fullNode
		post        fullNode
		put         fullNode
		delete      fullNode
		options     fullNode
		head        fullNode
		patch       fullNode
	}
	fullNode struct {
		path string
		kind uint8
		pnum uint8
		name string
		// 保存各类子节点
		Cchildren []*fullNode
		Rchildren []*fullNode
		Pchildren []*fullNode
		Vchildren []*fullNode
		Wchildren *fullNode
		// 默认标签的名称和值
		tags []string
		vals []string
		// 校验函数
		check ValidateStringFunc
		// 正则捕获名称和函数
		// names		[]string
		// find		RouterFindFunc
		// 路由匹配的处理者
		handlers HandlerFuncs
	}
)

// NewRouterFull 函数创建一个Full路由器。
func NewRouterFull() Router {
	return NewRouterStd(NewRouterCoreFull())
}

// NewRouterCoreFull 函数创建一个Full路由器核心，使用radix匹配。
func NewRouterCoreFull() RouterCore {
	return &RouterCoreFull{
		nodefunc404: HandlerFuncs{HandlerRouter404},
		nodefunc405: HandlerFuncs{HandlerRouter405},
		node404: fullNode{
			tags:     []string{ParamRoute},
			vals:     []string{"404"},
			handlers: HandlerFuncs{HandlerRouter404},
		},
		node405: fullNode{
			Wchildren: &fullNode{
				tags:     []string{ParamRoute},
				vals:     []string{"405"},
				handlers: HandlerFuncs{HandlerRouter405},
			},
		},
		middlewares: newTrieNode(),
	}
}

// RegisterMiddleware method register the middleware into the middleware tree and append the handler if it exists.
//
// RegisterMiddleware 注册中间件到中间件树中，如果存在则追加处理者。
func (r *RouterCoreFull) RegisterMiddleware(path string, hs HandlerFuncs) {
	path = strings.Split(path, " ")[0]
	r.middlewares.Insert(path, hs)
	if path == "" {
		r.node404.handlers = append(r.middlewares.vals, r.nodefunc404...)
		r.node405.Wchildren.handlers = append(r.middlewares.vals, r.nodefunc405...)
	}
}

// RegisterHandler method register a new method request path to the router
//
// The router matches the handlers available to the current path from the middleware tree and adds them to the front of the handler.
//
// RegisterHandler 给路由器注册一个新的方法请求路径
//
// 路由器会从中间件树中匹配当前路径可使用的处理者，并添加到处理者前方。
func (r *RouterCoreFull) RegisterHandler(method string, path string, handler HandlerFuncs) {
	switch method {
	case "NotFound", "404":
		r.nodefunc404 = handler
		r.node404.handlers = HandlerFuncsCombine(r.middlewares.vals, handler)
	case "MethodNotAllowed", "405":
		r.nodefunc405 = handler
		r.node405.Wchildren.handlers = HandlerFuncsCombine(r.middlewares.vals, handler)
	case MethodAny:
		handler = HandlerFuncsCombine(r.middlewares.Lookup(path), handler)
		for _, method := range RouterAllMethod {
			r.insertRoute(method, path, true, handler)
		}
	default:
		r.insertRoute(method, path, false, HandlerFuncsCombine(r.middlewares.Lookup(path), handler))
	}
}

// Add a new route Node.
//
// If the method does not support it will not be added, request to change the path will respond 405
//
// 添加一个新的路由Node。
//
// 如果方法不支持则不会添加，请求改路径会响应405
func (r *RouterCoreFull) insertRoute(method, key string, isany bool, val HandlerFuncs) {
	var currentNode = r.getTree(method)
	if currentNode == &r.node405 {
		return
	}

	// 创建节点
	args := strings.Split(key, " ")
	for _, path := range getSplitPath(args[0]) {
		currentNode = currentNode.InsertNode(path, newFullNode(path))
	}

	if isany {
		if currentNode.kind&fullNodeKindAnyMethod != fullNodeKindAnyMethod && currentNode.handlers != nil {
			return
		}
		currentNode.kind |= fullNodeKindAnyMethod
	} else {
		currentNode.kind &^= fullNodeKindAnyMethod
	}

	currentNode.handlers = val
	currentNode.SetTags(args)
}

// Match a request, if the method does not allow direct return to node405, no match returns node404.
//
// Note: 404 does not support extra parameters, not implemented.
//
// 匹配一个请求，如果方法不不允许直接返回node405，未匹配返回node404。
func (r *RouterCoreFull) Match(method, path string, params Params) HandlerFuncs {
	if n := r.getTree(method).recursiveLoopup(path, params); n != nil {
		return n
	}

	// 处理404
	r.node404.AddTagsToParams(params)
	return r.node404.handlers
}

// Create a 405 response radixNode.
//
// 创建一个405响应的radixNode。
func newFullNode405(args string, h HandlerFunc) *fullNode {
	newNode := &fullNode{
		Wchildren: &fullNode{
			handlers: HandlerFuncs{h},
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
		} else {
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
func loadCheckFunc(path string) (string, ValidateStringFunc) {
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

	// 调用validate部分创建check函数
	return name, GetValidateStringFunc(fname)
}

// InsertNode add a child node to the node.
//
// InsertNode 给节点添加一个子节点。
func (r *fullNode) InsertNode(path string, nextNode *fullNode) *fullNode {
	if len(path) == 0 {
		return r
	}
	nextNode.path = path
	switch nextNode.kind {
	case fullNodeKindConst:
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
	for i := range r.Cchildren {
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
		r.tags[i+1], r.vals[i+1] = split2byte(str, '=')
	}
}

// AddTagsToParams give the current Node tag to Params
//
// AddTagsToParams 将当前Node的tags给予Params
func (r *fullNode) AddTagsToParams(p Params) {
	for i := range r.tags {
		p.Add(r.tags[i], r.vals[i])
	}
}

// Get the tree of the corresponding method.
//
// Support eudore.RouterAllMethod these methods, weak support will return 405 processing tree.
//
// 获取对应方法的树。
//
// 支持eudore.RouterAllMethod这些方法,弱不支持会返回405处理树。
func (r *RouterCoreFull) getTree(method string) *fullNode {
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

// Recursively add a constant Node with a path of containKey to the current node
//
// targetKey and targetValue are new Node data.
//
// 给当前节点递归添加一个路径为containKey的常量Node
//
// targetKey和targetValue为新Node数据。
func (r *fullNode) recursiveInsertTree(containKey string, targetNode *fullNode) *fullNode {
	for i := range r.Cchildren {
		subStr, find := getSubsetPrefix(containKey, r.Cchildren[i].path)
		if find {
			if subStr == r.Cchildren[i].path {
				nextTargetKey := strings.TrimPrefix(containKey, r.Cchildren[i].path)
				return r.Cchildren[i].recursiveInsertTree(nextTargetKey, targetNode)
			}
			newNode := r.SplitNode(subStr, r.Cchildren[i].path)
			if newNode == nil {
				panic("Unexpect error on split node")
			}

			return newNode.InsertNode(strings.TrimPrefix(containKey, subStr), targetNode)
		}
	}

	return r.InsertNode(containKey, targetNode)
}

func (r *fullNode) recursiveLoopup(searchKey string, params Params) HandlerFuncs {

	// constant match, return data
	// 常量匹配，返回数据
	if len(searchKey) == 0 && r.handlers != nil {
		r.AddTagsToParams(params)
		return r.handlers
	}

	if len(searchKey) > 0 {
		// Traverse constant Node match
		// 遍历常量Node匹配
		for _, edgeObj := range r.Cchildren {
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

		// parameter matching
		// Check if there is a parameter match
		// 参数匹配
		// 检测是否存在参数匹配
		if r.pnum != 0 {
			pos := strings.IndexByte(searchKey, '/')
			if pos == -1 {
				pos = len(searchKey)
			}
			currentKey, nextSearchKey := searchKey[:pos], searchKey[pos:]

			// check parameter matching
			// 校验参数匹配
			for _, edgeObj := range r.Rchildren {
				if edgeObj.check(currentKey) {
					if n := edgeObj.recursiveLoopup(nextSearchKey, params); n != nil {
						params.Add(edgeObj.name, currentKey)
						return n
					}
				}
			}

			// 参数匹配
			// 变量Node依次匹配是否满足
			for _, edgeObj := range r.Pchildren {
				if n := edgeObj.recursiveLoopup(nextSearchKey, params); n != nil {
					params.Add(edgeObj.name, currentKey)
					return n
				}
			}
		}
	}

	// wildcard verification match
	// If the current Node has a wildcard processing method that directly matches, the result is returned.
	// 通配符校验匹配
	// 若当前Node有通配符处理方法直接匹配，返回结果。
	for _, edgeObj := range r.Vchildren {
		if edgeObj.check(searchKey) {
			edgeObj.AddTagsToParams(params)
			params.Add(edgeObj.name, searchKey)
			return edgeObj.handlers
		}
	}

	// If the current Node has a wildcard processing method that directly matches, the result is returned.
	// 若当前Node有通配符处理方法直接匹配，返回结果。
	if r.Wchildren != nil {
		r.Wchildren.AddTagsToParams(params)
		params.Add(r.Wchildren.name, searchKey)
		return r.Wchildren.handlers
	}

	// can't match, return nil
	// 无法匹配，返回空
	return nil
}
