package eudore

import (
	"strings"
)

const (
	kindConst	uint8	=	iota
	kindParam
	kindWildcard
	kindRegex
)

type (
	RouterRadix struct {
		RouterMethod
		link 		Middleware
		root		radixNode
		get 		radixNode
		post		radixNode
		put			radixNode
		delete		radixNode
		options 	radixNode
		head		radixNode
		patch 		radixNode
	}
	radixNode struct {
		path		string
		kind		uint8
		name		string
		Cchildren	[]*radixNode
		Pchildren	[]*radixNode
		Wchildren	*radixNode
		// data
		key			string
		val			Middleware
	}
)


func NewRouterRadix(interface{}) (Router, error) {
	r := &RouterRadix{}
	r.RouterMethod = &RouterMethodStd{RouterCore:	r}
	return r, nil
}


func (r *RouterRadix) Handle(ctx Context) {
	h := r.Match(ctx.Method(), ctx.Path(), ctx)
	ctx.SetHandler(r.link)
	ctx.Next()
	ctx.SetHandler(h)
	ctx.Next()
}

func (r *RouterRadix) GetNext() Middleware {
	return r.link
}

func (r *RouterRadix) SetNext(m Middleware)  {
	if r.link == nil {
		r.link = m
	}else {
		GetMiddlewareEnd(r.link).SetNext(m)
	}
}

func (r *RouterRadix) RegisterMiddleware(hs ...Handler) {
	r.SetNext(NewMiddlewareLink(hs...))
}

func (r *RouterRadix) RegisterHandler(method string, path string, handler Handler) {
	if method == MethodAny{
		for _, method := range r.AllRouterMethod() {
			r.InsertRoute(method, path, NewMiddleware(handler))		
		}
	}else {
		r.InsertRoute(method, path, NewMiddleware(handler))	
	}
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



func NewRadixNode(path, key string, val Middleware) *radixNode {
		newNode := &radixNode{path: path, key: key, val: val}
		switch path[0] {
		case '*':
			newNode.kind = kindWildcard
			if len(path) == 1 {
				newNode.name = "*"
			}else{
				newNode.name = path[1:]
			}
		case ':':
			newNode.kind = kindParam
			newNode.name = path[1:]
		case '#':
			newNode.kind = kindRegex
		default:
			newNode.kind = kindConst
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
	case kindParam:
		// 路径存在返回旧Node
		for _, i := range r.Pchildren {
			if i.path == path {
				return i
			}
		}
		r.Pchildren = append(r.Pchildren, newNode)
	case kindRegex:
		// 正则Node，未实现
	case kindWildcard:
		// 设置通配符Node数据。
		r.Wchildren = newNode
	default:
		// 追加常量Node，常量Node由递归插入，不会出现相同前缀。
		r.Cchildren = append(r.Cchildren, newNode)
	}
	return newNode
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
	case "GET":
		return &t.get
	case "POST":
		return &t.post
	case "DELETE":
		return &t.delete
	case "PUT":
		return &t.put
	case "HEAD":
		return &t.head
	case "OPTIONS":
		return &t.options
	case "PATCH":
		return &t.patch
	default:
		return &t.root
	}
}

// 添加一个新的Node
func (t *RouterRadix) InsertRoute(method, key string, val Middleware) {
	var currentNode, newNode *radixNode = t.getTree(method), nil
	// 兼容带默认变量的路径
	key = strings.Split(key, " ")[0]
	// 将路径按Node切割，然后依次向下追加。
	for _, path := range getSpiltPath(key) {
		// 创建一个新的radixNode，并设置Node类型
		newNode = NewRadixNode(path, "", nil)
		if newNode.kind == kindConst {
			currentNode = t.recursiveInsertTree(currentNode, path, newNode)
		}else {
			currentNode = currentNode.InsertNode(path, newNode)
		}		
	}
	// 追加一个Node设置数据。
	newNode.key = key
	newNode.val = val
}



// 给currentNode递归添加，路径为containKey的Node
//
// targetKey和targetValue为新Node数据。
func (t *RouterRadix) recursiveInsertTree(currentNode *radixNode, containKey string, targetNode *radixNode) *radixNode {
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
				return t.recursiveInsertTree(currentNode.Cchildren[i], nextTargetKey, targetNode)	
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

func (t *RouterRadix) Match(method, path string, params Params) Middleware {
	return t.recursiveLoopup(t.getTree(method), path, params)
}

func (t *RouterRadix) recursiveLoopup(searchNode *radixNode, searchKey string, params Params) Middleware {
	// 常量Node匹配，返回数据
	if len(searchKey) == 0 && searchNode.val != nil {
		return searchNode.val
	}
	// 遍历常量Node匹配
	for _, edgeObj := range searchNode.Cchildren {
		// 寻找相同前缀node进一步匹配
		if subStr, find := getSubsetPrefix(searchKey, edgeObj.path); find && subStr == edgeObj.path{
			// 除去前缀路径，获得未匹配路径。
			nextSearchKey := searchKey[len(subStr):]
			// 然后当前Node递归判断
			n := t.recursiveLoopup(edgeObj, nextSearchKey, params)
			if n != nil {
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
			n := t.recursiveLoopup(edgeObj, nextSearchKey, params)
			if n != nil {
				// 添加变量参数
				params.AddParam(edgeObj.name, searchKey[:pos])
				return n
			}
		}
	}
	// 通配符
	// 若当前Node有通配符处理方法直接匹配，返回结果
	if searchNode.Wchildren != nil {
		params.AddParam(searchNode.Wchildren.name, searchKey)
		return searchNode.Wchildren.val
	}
	// 无法匹配
	return nil
}

/*
将字符串按Node类型切割

字符串路径切割例子：
/ 				[/ *]
/api/note/ 		[/api/note/ *]
//api/* 		[/api/ *]
//api/*name 	[/api/ *name]
/api/get/ 		[/api/get/ *]
/api/get 		[/api/get]
/api/:get 		[/api/ :get]
/api/:get/* 		[/api/ :get / *]
/api/:name/info/* 	[/api/ :name /info/ *]
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
		if str[0] == ':' || str[0] == '*' {
			strs = append(strs, str)
			last = true
			continue
		}
		// 追加常量
		num := len(strs) - 1
		strs[num] = strs[num] + str //+ "/" 
	
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
