package eudore

import (
	"strings"
)

const (
	radixNodeKindConst	uint8	=	iota
	radixNodeKindParam
	radixNodeKindWildcard
)

type (
	// Basic function router based on radix tree implementation
	//
	// 基于基数树实现的基本功能路由器
	RouterRadix struct {
		RouterMethod
		node404		HandlerFuncs
		node405		*radixNode
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
		node404:	HandlerFuncs{RouterDefault404Func},
		node405:	newRadixNode405("405", RouterDefault405Func),
	}
	r.RouterMethod = &RouterMethodStd{RouterCore:	r}
	return r, nil
}


func (r *RouterRadix) Handle(ctx Context) {
	h := r.Match(ctx.Method(), ctx.Path(), ctx)
	ctx.SetHandler(h)
	ctx.Next()
}

func (r *RouterRadix) RegisterHandler(method string, path string, handler HandlerFuncs) {
	if method == MethodAny{
		for _, method := range r.AllRouterMethod() {
			r.InsertRoute(method, path, handler)
		}
	}else {
		r.InsertRoute(method, path, handler)
	}
}

// 添加一个新的Node
func (t *RouterRadix) InsertRoute(method, key string, val HandlerFuncs) {
	var currentNode, newNode *radixNode = t.getTree(method), nil
	// 未支持的请求方法
	if currentNode == t.node405 {
		return
	}
	args := strings.Split(key, " ")
	// 将路径按Node切割，然后依次向下追加。
	for _, path := range getSpiltPath(args[0]) {
		// 创建一个新的radixNode，并设置Node类型
		newNode = NewRadixNode(path)
		if newNode.kind == radixNodeKindConst {
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
func (t *RouterRadix) Match(method, path string, params Params) HandlerFuncs {
	if n := t.getTree(method).recursiveLoopup(path, params); n != nil {
		return n
	}
	// t.node404.GetTags(params)
	return t.node404
}

func (r *RouterRadix) Set(key string, i interface{}) error {
	h, ok := i.(HandlerFunc)
	if !ok {
		h, ok = i.(func(Context))
	}
	if !ok {
		return ErrRouterSetNoSupportType
	}
	// m := NewMiddleware(h)
	args := strings.Split(key, " ")
	switch args[0] {
	case "404":
		r.node404 = HandlerFuncs{h}
	case "405":
		r.node405 = newRadixNode405(key, h)
	}
	return nil
}

func (*RouterRadix) GetName() string {
	return ComponentRouterRadixName
}

func (*RouterRadix) Version() string {
	return ComponentRouterRadixVersion
}


func (*RouterRadix) AllRouterMethod() []string {
	return []string{MethodAny, MethodGet, MethodPost, MethodPut, MethodDelete, MethodHead, MethodPatch, MethodOptions}
}


func newRadixNode405(args string, h HandlerFunc) *radixNode {
	newNode := &radixNode{
		Wchildren:	&radixNode{
			handlers:	HandlerFuncs{h},
		},
	}
	newNode.Wchildren.SetTags(strings.Split(args, " "))
	return newNode
}

func NewRadixNode(path string) *radixNode {
		newNode := &radixNode{path: path}
		// 更具路径前缀类型，创建不同的Node
		switch path[0] {
		// 通配符Node
		case '*':
			newNode.kind = radixNodeKindWildcard
			if len(path) == 1 {
				// 设置默认名称
				newNode.name = "*"
			}else{
				newNode.name = path[1:]
			}
		// 参数Node
		case ':':
			newNode.kind = radixNodeKindParam
			newNode.name = path[1:]
		// 常量Node
		default:
			newNode.kind = radixNodeKindConst
		}
		return newNode
}

// 新增Node
func (r *radixNode) InsertNode(path string, newNode *radixNode) *radixNode {
	if len(path) == 0 {
		return r
	}
	newNode.path = path
	switch newNode.kind {
	case radixNodeKindParam:
		// 路径存在返回旧Node
		for _, i := range r.Pchildren {
			if i.path == path {
				return i
			}
		}
		r.Pchildren = append(r.Pchildren, newNode)
	case radixNodeKindWildcard:
		// 设置通配符Node数据。
		r.Wchildren = newNode
	default:
		// 追加常量Node，常量Node由递归插入，不会出现相同前缀。
		r.Cchildren = append(r.Cchildren, newNode)
	}
	return newNode
}

// 给当前Node设置tags
func (r *radixNode) SetTags(args []string) {
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
func (r *radixNode) GetTags(p Params) {
	for i, _ := range r.tags {
		p.AddParam(r.tags[i], r.vals[i])
	}
}

// 对指定路径为edgeKey的Node分叉，公共前缀路径为pathKey
func (r *radixNode) SplitNode(pathKey, edgeKey string) *radixNode {
	for i, _ := range r.Cchildren {
		// 找到路径为edgeKey路径的Node，然后分叉
		if r.Cchildren[i].path == edgeKey {
			newNode := &radixNode{path: pathKey}

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


// 给currentNode递归添加，路径为containKey的Node
//
// targetKey和targetValue为新Node数据。
func (currentNode *radixNode) recursiveInsertTree(containKey string, targetNode *radixNode) *radixNode {
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



func (searchNode *radixNode) recursiveLoopup(searchKey string, params Params) HandlerFuncs {
	// 常量Node匹配，返回数据
	if len(searchKey) == 0 && searchNode.handlers != nil {
		searchNode.GetTags(params)
		return searchNode.handlers
	}
	// 遍历常量Node匹配
	for _, edgeObj := range searchNode.Cchildren {
		// 寻找相同前缀node进一步匹配
		// 当前Node的路径必须为searchKey的前缀。
		if IsPrefix(searchKey, edgeObj.path) {
			// 除去前缀路径，获得未匹配路径。
			nextSearchKey := searchKey[len(edgeObj.path):]
			// 然后当前Node递归判断
			if n := edgeObj.recursiveLoopup(nextSearchKey, params);n != nil {
				return n
			}
		}
	}
	// 参数
	if len(searchNode.Pchildren) > 0 && len(searchKey) > 0 {
		// 寻找字符串切割位置
		// strings.IndexByte为c语言实现，未写过更有效方法，且可读性更强。
		pos := strings.IndexByte(searchKey, '/')
		if pos == -1 {
			pos = len(searchKey) 
		}
		// 剩余字符串为第一个'/'开头
		nextSearchKey := searchKey[pos:]

		// 变量Node依次匹配是否满足
		for _, edgeObj := range searchNode.Pchildren {
			// 递归判断
			if n := edgeObj.recursiveLoopup(nextSearchKey, params);n != nil {
				// 添加变量参数
				params.AddParam(edgeObj.name, searchKey[:pos])
				return n
			}
		}
	}
	// 通配符
	// 若当前Node有通配符处理方法直接匹配，返回结果。
	if searchNode.Wchildren != nil {
		searchNode.Wchildren.GetTags(params)
		params.AddParam(searchNode.Wchildren.name, searchKey)
		return searchNode.Wchildren.handlers
	}
	// 无法匹配
	return nil
}

/*
将字符串按Node类型切割，当前分割规则不够详细。

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
		// 过滤路径中的/
		if len(str) == 0 {
			if i == 0 {
				strs = []string{"/"}
			}
			continue
		}
		// 补充分隔符
		if last {
			last = false
			strs = append(strs, "/")
		}else {
			lastappend(strs, '/')
		}
		// 处理特殊前缀路径
		if lastisbyte(strs, '/') && (str[0] == ':' || str[0] == '*') {
			strs = append(strs, str)
			last = true
			continue
		}
		// 追加常量
		num := len(strs) - 1
		strs[num] = strs[num] + str
	
	}
	return strs
}

// 修改最后一个字符串结尾为b
func lastappend(strs []string, b byte) {
	num := len(strs) - 1
	laststr := strs[num]
	if laststr[len(laststr) - 1 ] != b {
		strs[num] = strs[num] + string(b)
	}
}

// 检测最后一个字符串的结尾是否为指导byte
func lastisbyte(strs []string, b byte) bool {
	num := len(strs) - 1
	if num < 0 {
		return false
	}
	laststr := strs[num]
	return laststr[len(laststr) - 1 ] == b
}


// 获取两个字符串的最大公共前缀，返回最大公共前缀和是否拥有最大公共前缀
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
		//fix "" not a subset of ""
		return str1, str1 == str2
	}

	return str1, findSubset
}

// 获取两个字符串的最大公共前缀，返回最大公共前缀和是否拥有最大公共前缀
func IsPrefix(str1, str2 string) bool {
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
