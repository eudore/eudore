package eudore

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
	// 路由数据校验函数
	RouterCheckFunc func(string) bool
	RouterFindFunc func() []string
	// 路由数据校验函数的创建函数
	// 
	// 通过指定字符串构造出一个校验函数
	RouterNewCheckFunc func(string) RouterCheckFunc
	// RouterFull基于基数树实现，实现全部路由器相关特性。
	RouterFull struct {
		RouterMethod
		// link		Middleware
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
		names		[]string
		find		RouterFindFunc
		// 路由匹配的处理者
		// handle		Middleware
		handlers	HandlerFuncs
	}
)


func NewRouterFull(interface{}) (Router, error) {
	r := &RouterFull{
		node404:	HandlerFuncs{RouterDefault404Func},
		node405:	newFullNode405("405", RouterDefault405Func),
	}
	r.RouterMethod = &RouterMethodStd{RouterCore:	r}
	return r, nil
}


func (r *RouterFull) Handle(ctx Context) {
	h := r.Match(ctx.Method(), ctx.Path(), ctx)
	ctx.SetHandler(h)
	ctx.Next()
}

func (r *RouterFull) RegisterHandler(method string, path string, handler HandlerFuncs) {
	if method == MethodAny{
		for _, method := range r.AllRouterMethod() {
			r.InsertRoute(method, path, handler)		
		}
	}else {
		r.InsertRoute(method, path, handler)
	}
}

// 添加一个新的Node
func (t *RouterFull) InsertRoute(method, key string, val HandlerFuncs) {
	var currentNode, newNode *fullNode = t.getTree(method), nil
	// 未支持的请求方法
	if currentNode == t.node405 {
		return
	}
	args := strings.Split(key, " ")
	// 将路径按Node切割，然后依次向下追加。
	for _, path := range getSpiltPath(args[0]) {
		// 创建一个新的fullNode，并设置Node类型
		newNode = NewFullNode(path)
		if newNode.kind == fullNodeKindConst {
			currentNode = currentNode.recursiveInsertTree(path, newNode)
		}else {
			currentNode = currentNode.InsertNode(path, newNode)
		}		
	}
	// 树分叉结尾Node添加这次Route数据。
	newNode.handlers = val
	newNode.SetTags(args)
}

// 匹配一个请求，如果方法不不允许直接返回node405，未匹配返回node404。
func (t *RouterFull) Match(method, path string, params Params) HandlerFuncs {
	if n := t.getTree(method).recursiveLoopup(path, params); n != nil {
		return n
	}
	return t.node404
}

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


func (*RouterFull) GetName() string {
	return ComponentRouterFullName
}

func (*RouterFull) Version() string {
	return ComponentRouterFullVersion
}


func (*RouterFull) AllRouterMethod() []string {
	return []string{MethodAny, MethodGet, MethodPost, MethodPut, MethodDelete, MethodHead, MethodPatch, MethodOptions}
}


func newFullNode405(args string, h HandlerFunc) *fullNode {
	newNode := &fullNode{
		Wchildren:	&fullNode{
			handlers:	HandlerFuncs{h},
		},
	}
	newNode.Wchildren.SetTags(strings.Split(args, " "))
	return newNode
}

func NewFullNode(path string) *fullNode {
		newNode := &fullNode{path: path}
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

// 更具名称加载校验函数。
func loadCheckFunc(path string) (string, RouterCheckFunc) {
	// 无效路径
	if len(path) == 0 || (path[0] != ':' && path[0] != '*') {
		return "", nil
	}
	path = path[1:]
	// 截取参数名称和校验函数名称
	name, fname := split2byte(path, '|')
	if len(name) == 0 {
		return "", nil
	}
	// 正则
	// 如果是正则表达式开头，添加默认正则校验函数名称。
	if fname[0] == '^' {
		fname = "regexp:" + fname
	}
	// 判断是否有':'
	f2name, arg := split2byte(fname, ':')
	// 没有':'为固定函数，直接返回
	if len(arg) == 0 {
		return name, ConfigLoadRouterCheckFunc(fname)
	}
	// 有':'为变量函数，创建校验函数
	fn := ConfigLoadRouterNewCheckFunc(f2name)(arg)
	// 保存新建的校验函数
	ConfigSaveRouterCheckFunc(fname, fn)
	return name, fn
}

// 新增Node
func (r *fullNode) InsertNode(path string, newNode *fullNode) *fullNode {
	if len(path) == 0 {
		return r
	}
	newNode.path = path
	switch newNode.kind {
	case fullNodeKindParam:
		// 路径存在返回旧Node
		for _, i := range r.Pchildren {
			if i.path == path {
				return i
			}
		}
		r.pnum++
		r.Pchildren = append(r.Pchildren, newNode)
	case fullNodeKindRegex:
		for _, i := range r.Rchildren {
			if i.path == path {
				return i
			}
		}
		r.pnum++
		r.Rchildren = append(r.Rchildren, newNode)
	case fullNodeKindValid:
		// 通配符校验Node
		for _, i := range r.Vchildren {
			if i.path == path {
				return i
			}
		}
		r.Vchildren = append(r.Vchildren, newNode)
	case fullNodeKindWildcard:
		// 设置通配符Node数据。
		r.Wchildren = newNode
	default:
		// 追加常量Node，常量Node由递归插入，不会出现相同前缀。
		r.Cchildren = append(r.Cchildren, newNode)
	}
	return newNode
}


// 给当前Node设置tags
func (r *fullNode) SetTags(args []string) {
	if len(args) == 0 {
		return
	}
	r.tags = make([]string, len(args))
	r.vals = make([]string, len(args))
	// 第一个参数名称默认为route
	r.tags[0] = "route"
	r.vals[0] = args[0]
	for i, str := range args[1:] {
		r.tags[i + 1], r.vals[i + 1] = split2byte(str, ':')
	}
}

// 将当前Node的tags给予Params
func (r *fullNode) GetTags(p Params) {
	for i, _ := range r.tags {
		p.AddParam(r.tags[i], r.vals[i])
	}
}

// 对指定路径为edgeKey的Node分叉，公共前缀路径为pathKey
func (r *fullNode) SplitNode(pathKey, edgeKey string) *fullNode {
	for i, _ := range r.Cchildren {
		// 找到路径为edgeKey路径的Node，然后分叉
		if r.Cchildren[i].path == edgeKey {
			newNode := &fullNode{path: pathKey}

			r.Cchildren[i].path = strings.TrimPrefix(edgeKey, pathKey)
			newNode.Cchildren = append(newNode.Cchildren, r.Cchildren[i])

			r.Cchildren[i] = newNode
			// 返回分叉新创建的Node
			return newNode
		}
	}
	return nil
}

// 获取对应方法的Tree
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

// 给currentNode递归添加，路径为containKey的Node
//
// targetKey和targetValue为新Node数据。
func (currentNode *fullNode) recursiveInsertTree(containKey string, targetNode *fullNode) *fullNode {
	for i, _ := range currentNode.Cchildren {
		// 检查当前遍历的Node和插入路径是否有公共路径
		// subStr是两者的公共路径，find表示是否有
		subStr, find := getSubsetPrefix(containKey, currentNode.Cchildren[i].path)
		if find {
			// 如果child路径等于公共最大路径，则该node添加child
			// child的路径为插入路径先过滤公共路径的后面部分。
			if subStr == currentNode.Cchildren[i].path {
				nextTargetKey := strings.TrimPrefix(containKey, currentNode.Cchildren[i].path)
				// 当前node新增子Node可能原本有多个child，所以需要递归添加
				return currentNode.Cchildren[i].recursiveInsertTree(nextTargetKey, targetNode)	
			}else {
				// 如果公共路径不等于当前node的路径
				// 则将currentNode.children[i]路径分叉
				// 分叉后的就拥有了公共路径，然后添加新Node
				newNode := currentNode.SplitNode(subStr, currentNode.Cchildren[i].path)
				if newNode == nil {
					panic("Unexpect error on split node")
				}
				// 添加新的node
				// 分叉后树一定只有一个没有相同路径的child，所以直接添加node
				
				return newNode.InsertNode(strings.TrimPrefix(containKey, subStr), targetNode)
			}
		}
	}
	// 没有相同前缀路径存在，直接添加为child
	return currentNode.InsertNode(containKey, targetNode)
}


func (searchNode *fullNode) recursiveLoopup(searchKey string, params Params) HandlerFuncs {
	// 常量Node匹配，返回数据
	if len(searchKey) == 0 && searchNode.handlers != nil {
		searchNode.GetTags(params)
		return searchNode.handlers
	}
	// 遍历常量Node匹配
	for _, edgeObj := range searchNode.Cchildren {
		// 寻找相同前缀node进一步匹配
		// 当前Node的路径必须为searchKey的前缀。
		if subStr, find := getSubsetPrefix(searchKey, edgeObj.path); find && subStr == edgeObj.path{
			// 除去前缀路径，获得未匹配路径。
			nextSearchKey := searchKey[len(subStr):]
			// 然后当前Node递归判断
			if n := edgeObj.recursiveLoopup(nextSearchKey, params);n != nil {
				return n
			}
		}
	}
	// 参数
	if searchNode.pnum != 0 && len(searchKey) > 0 {
		// 寻找字符串切割位置
		// strings.IndexByte为c语言实现，未写过更有效方法，且可读性更强。
		pos := strings.IndexByte(searchKey, '/')
		if pos == -1 {
			pos = len(searchKey) 
		}
		// currentKey为当前需要处理的字符串
		// nextSearchKey为剩余字符串为第一个'/'开头
		currentKey, nextSearchKey := searchKey[:pos], searchKey[pos:]
		// 校验参数匹配
		for _, edgeObj := range searchNode.Rchildren {
			if edgeObj.check(currentKey) {
				if n := edgeObj.recursiveLoopup(nextSearchKey, params);n != nil {
					// 添加变量参数
					params.AddParam(edgeObj.name, currentKey)
					return n
				}
			}
		}
		// 参数匹配
		// 变量Node依次匹配是否满足
		for _, edgeObj := range searchNode.Pchildren {
			// 递归判断
			if n := edgeObj.recursiveLoopup(nextSearchKey, params);n != nil {
				// 添加变量参数
				params.AddParam(edgeObj.name, currentKey)
				return n
			}
		}
	}
	// 通配符
	// 通配符校验匹配
	// 若当前Node有通配符处理方法直接匹配，返回结果。
	for _, edgeObj := range searchNode.Vchildren {
		if edgeObj.check(searchKey) {
			edgeObj.GetTags(params)
			params.AddParam(edgeObj.name, searchKey)
			return edgeObj.handlers
		}
	}
	// 返回通配符处理。
	if searchNode.Wchildren != nil {
		searchNode.Wchildren.GetTags(params)
		params.AddParam(searchNode.Wchildren.name, searchKey)
		return searchNode.Wchildren.handlers
	}
	// 无法匹配
	return nil
}