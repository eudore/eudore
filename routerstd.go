package eudore

/*
基于基数树算法实现一个完整路由器
*/

import (
	"strings"
)

const (
	stdNodeKindConst         uint8 = 1 << iota // 常量
	stdNodeKindParamValid                      // 参数校验
	stdNodeKindParam                           // 参数
	stdNodeKindWildcardValid                   // 通配符校验
	stdNodeKindWildcard                        // 通配符
	stdNodeKindAnyMethod
)

// routerCoreStd is implemented based on the radix tree to implement all router related features.
//
// With path parameters, wildcard parameters, default parameters, parameter verification, wildcard verification, multi-parameter regular capture is not implemented.
//
// RouterStd基于基数树实现，实现全部路由器相关特性。
//
// 具有路径参数、通配符参数、默认参数、参数校验、通配符校验，未实现多参数正则捕捉。
type routerCoreStd struct {
	node404 stdNode
	node405 stdNode
	get     stdNode
	post    stdNode
	put     stdNode
	delete  stdNode
	head    stdNode
	patch   stdNode
	options stdNode
	connect stdNode
	trace   stdNode
}
type stdNode struct {
	kind uint8
	pnum uint8
	path string
	name string
	// 保存各类子节点
	Cchildren  []*stdNode
	PVchildren []*stdNode
	Pchildren  []*stdNode
	WVchildren []*stdNode
	Wchildren  *stdNode
	// 默认标签的名称和值
	params      *Params
	handlers    HandlerFuncs
	anyin       bool
	anyhandlers HandlerFuncs
	check       func(string) bool
}

// NewRouterCoreStd 函数创建一个Std路由器核心，使用radix匹配。
func NewRouterCoreStd() RouterCore {
	return &routerCoreStd{
		node404: stdNode{
			params:   &Params{Keys: []string{ParamRoute}, Vals: []string{"404"}},
			handlers: HandlerFuncs{HandlerRouter404},
		},
		node405: stdNode{
			Wchildren: &stdNode{
				params:   &Params{Keys: []string{ParamRoute}, Vals: []string{"405"}},
				handlers: HandlerFuncs{HandlerRouter405},
			},
		},
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
func (r *routerCoreStd) Match(method, path string, params *Params) HandlerFuncs {
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
func (r *routerCoreStd) getTree(method string) *stdNode {
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

// Add a new route Node.
//
// If the method does not support it will not be added, request to change the path will respond 405
//
// 添加一个新的路由Node。
//
// 如果方法不支持则不会添加，请求改路径会响应405
func (r *routerCoreStd) insertRoute(method, key string, isany bool, val HandlerFuncs) {
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
		currentNode = currentNode.insertNode(path, newStdNode(path))
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

// 创建一个Radix树Node，会根据当前路由设置不同的节点类型和名称。
//
// '*'前缀为通配符节点，':'前缀为参数节点，其他未常量节点,如果通配符和参数结点后带有符号'|'则为校验结点。
func newStdNode(path string) *stdNode {
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
			if name, fn := loadCheckFunc(path); len(name) > 0 {
				// 无法获得校验函数抛出错误
				if fn == nil {
					panic("loadCheckFunc path is invalid, load func failure " + path)
				}
				newNode.kind, newNode.name, newNode.check = stdNodeKindWildcardValid, name, fn
			}
		}
	case ':':
		newNode.kind = stdNodeKindParam
		newNode.name = path[1:]
		// 如果路径后序具有'|'符号，则截取后端名称返回校验函数
		// 并升级成校验参数Node
		if name, fn := loadCheckFunc(path); len(name) > 0 {
			if fn == nil {
				panic("loadCheckFunc path is invalid, load func failure " + path)
			}
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
func loadCheckFunc(path string) (string, func(string) bool) {
	path = path[1:]
	// 截取参数名称和校验函数名称
	name, fname := split2byte(path, '|')
	if name == "" || fname == "" {
		return "", nil
	}
	// 如果是正则表达式开头，添加默认正则校验函数名称。
	if fname[0] == '^' && fname[len(fname)-1] == '$' {
		fname = "regexp:" + fname
	}

	// 调用validate部分创建check函数
	return name, GetValidateStringFunc(fname)
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
		r.Wchildren = nextNode
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

func (r *stdNode) lookNode(searchKey string, params *Params) HandlerFuncs {
	// constant match, return data
	// 常量匹配，返回数据
	if len(searchKey) == 0 && r.handlers != nil {
		params.Combine(r.params)
		return r.handlers
	}

	if len(searchKey) > 0 {
		// Traverse constant Node match
		// 遍历常量Node匹配
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
			for _, child := range r.PVchildren {
				if child.check(currentKey) {
					if n := child.lookNode(nextSearchKey, params); n != nil {
						params.Add(child.name, currentKey)
						return n
					}
				}
			}

			// 参数匹配
			// 变量Node依次匹配是否满足
			for _, child := range r.Pchildren {
				if n := child.lookNode(nextSearchKey, params); n != nil {
					params.Add(child.name, currentKey)
					return n
				}
			}
		}
	}

	// wildcard verification match
	// If the current Node has a wildcard processing method that directly matches, the result is returned.
	// 通配符校验匹配
	// 若当前Node有通配符处理方法直接匹配，返回结果。
	for _, child := range r.WVchildren {
		if child.check(searchKey) {
			params.Combine(child.params)
			params.Add(child.name, searchKey)
			return child.handlers
		}
	}

	// If the current Node has a wildcard processing method that directly matches, the result is returned.
	// 若当前Node有通配符处理方法直接匹配，返回结果。
	if r.Wchildren != nil {
		params.Combine(r.Wchildren.params)
		params.Add(r.Wchildren.name, searchKey)
		return r.Wchildren.handlers
	}

	// can't match, return nil
	// 无法匹配，返回空
	return nil
}

func (r *stdNode) deleteRoute(path string, isany bool, val HandlerFuncs) {
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

func (r *stdNode) IsZero() bool {
	return r.handlers == nil && len(r.Cchildren) == 0 && len(r.Pchildren) == 0 && len(r.PVchildren) == 0 && len(r.WVchildren) == 0 && r.Wchildren == nil
}

func (r *stdNode) IsMarge() bool {
	if r.kind == stdNodeKindConst && r.handlers == nil && len(r.Cchildren) == 1 && len(r.Pchildren) == 0 && len(r.PVchildren) == 0 && len(r.WVchildren) == 0 && r.Wchildren == nil {
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
