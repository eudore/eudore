
package test

import (
	"strings"
	"testing"
	
	"fmt"
	// "github.com/kr/pretty"
)

func TestStaticRouter(t *testing.T) {
	tree := NewRadixTree()
	tree.Insert("/api/v1/node/", "")
	tree.Insert("/api/v1/:list/11", "")
	tree.Insert("/api/v1/:list/22", "")
	tree.Insert("/api/v1/:list/:name", "")
	tree.Insert("/api/v1/:list/*name", "")
	tree.Insert("/api/v1/:list", "")
	tree.Insert("/api/v1/*", "")
	tree.Insert("/api/v2/*", "")
	// fmt.Printf("%# v\n", pretty.Formatter(tree))
	tree.Insert("/api/*", "")
	tree.Insert("/note/get/:name","")
	tree.Insert("/note/:method/:name","")
	tree.Insert("/*", "")
	tree.Insert("/", "")
	// t.Log(getSpiltPath("/api/v1/:list/*"))
	t.Log(tree.Lookup("/api/v1/node/11"))
	t.Log(tree.Lookup("/api/v1/node/111"))
	t.Log(tree.Lookup("/api/v1/node/111/111"))
	t.Log(tree.Lookup("/api/v1/list"))
	t.Log(tree.Lookup("/api/v1/list33"))
	t.Log(tree.Lookup("/api/v1/get"))
	t.Log(tree.Lookup("/api/v2/get"))
	t.Log(tree.Lookup("/api/v3/111"))
	t.Log(tree.Lookup("/note/get/eudore/2"))
	t.Log(tree.Lookup("/note/get/eudore"))
	t.Log(tree.Lookup("/note/set/eudore"))
	t.Log(tree.Lookup("/node"))
}


func TestAAA(t *testing.T) {
	tree := NewRadixTree()
	tree.Insert("/a/b", "")
	tree.Insert("/a/:b", "")
	tree.Insert("/", "")
	tree.Insert("/a/*", "")
	t.Log(tree.Lookup("/a/b/s"))
	t.Log(tree.Lookup("/a/b"))
	t.Log(tree.Lookup("/a/bs"))
	t.Log(tree.Lookup("/abs"))
}
func BenchmarkLooup(b *testing.B) {
	tree := NewRadixTree()
	tree.Insert("/", "/*")
	tree.Insert("/*", "/*")
	tree.Insert("/api/v1/:list/11", "")
	tree.Insert("/api/v1/:list/22", "")
	tree.Insert("/api/v1/:list/:name", "")
	tree.Insert("/api/v1/:list/*name", "")
	tree.Insert("/api/v1/:list", "")
	tree.Insert("/api/v1/:li8u", "")
	tree.Insert("/api/v1/*", "")
	tree.Insert("/api/v2/*", "")
	tree.Insert("/api/", "api")
	tree.Insert("/", "*")
	for i := 0; i < b.N; i++ {
		tree.Lookup("/api/v1/node/11")
		tree.Lookup("/api/v1/node/111")
		tree.Lookup("/api/v1/node/111/111")
	}
}



const (
	kindConst	uint8	=	iota
	kindParam
	kindWildcard
	kindRegex
)
type (
	Middleware = interface{}
	radixTree struct {
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

func NewRadixTree() *radixTree{
	return &radixTree{}
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
		for _, i := range r.Pchildren {
			if i.path == path {
				return i
			}
		}
		r.Pchildren = append(r.Pchildren, newNode)
	case kindRegex:
	case kindWildcard:
			r.Wchildren = newNode
	default:
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

func (t *radixTree) getTree(method string) *radixNode {
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
func (t *radixTree) Insert(key string, val Middleware) {
	val = key
	var head = &t.root
	var newNode *radixNode
	for _, path := range getSpiltPath(key) {
		newNode = NewRadixNode(path, "", nil)
		if newNode.kind == kindConst {
			head = t.recursiveInsertTree(head, path, newNode)
		}else {
			head = head.InsertNode(path, newNode)
		}		
	}
	newNode.key = key
	newNode.val = val
}

func (t *radixTree) InsertRoute(method, key string, val Middleware) {
	val = key
	var head = t.getTree(method)
	var newNode *radixNode
	for _, path := range getSpiltPath(key) {
		newNode = NewRadixNode(path, "", nil)
		if newNode.kind == kindConst {
			head = t.recursiveInsertTree(head, path, newNode)
		}else {
			head = head.InsertNode(path, newNode)
		}		
	}
	newNode.key = key
	newNode.val = val
}



// 给currentNode递归添加，路径为containKey的Node
//
// targetKey和targetValue为新Node数据。
func (t *radixTree) recursiveInsertTree(currentNode *radixNode, containKey string, targetNode *radixNode) *radixNode {
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

func (r *radixTree) LookupRoute(method, path string) (string, Middleware) {
	return path, r.recursiveLoopup(r.getTree(method), path)
}

func (r *radixTree) Lookup(path string) (string, Middleware) {
	return path, r.recursiveLoopup(&r.root, path)
}

func (t *radixTree) recursiveLoopup(searchNode *radixNode, searchKey string) Middleware {
	// 常量
	if len(searchKey) == 0 && searchNode.val != nil {
		return searchNode.val
	}
	for _, edgeObj := range searchNode.Cchildren {
		// 寻找相同前缀node
		if subStr, find := getSubsetPrefix(searchKey, edgeObj.path); find && subStr == edgeObj.path {
			// 截取为匹配的路径
			// nextSearchKey := strings.TrimPrefix(searchKey, edgeObj.path)
			// if subStr == searchKey {
			// 	return edgeObj.val
			// }
			nextSearchKey := searchKey[len(subStr):]
			fmt.Println(searchKey, nextSearchKey)
			// 然后当前Node递归判断
			n := t.recursiveLoopup(edgeObj, nextSearchKey)
			// fmt.Println(n)
			if n != nil {
				return n
			}
		}
	}
	// 参数
	if len(searchNode.Pchildren) > 0 && len(searchKey) > 0 {
		// pos := 0
		// length := len(searchKey) -1
		// for ; searchKey[pos] != '/' && pos < length ; pos++{
		// }
		// var nextSearchKey string
		// if pos < length {
		// 	nextSearchKey = searchKey[pos:]	
		// }
		
		pos := strings.IndexByte(searchKey, '/')
		if pos == -1 {
			pos = len(searchKey) 
		}
		nextSearchKey := searchKey[pos:]

		// fmt.Println("param: ",searchKey, nextSearchKey)
		for _, edgeObj := range searchNode.Pchildren {
			n := t.recursiveLoopup(edgeObj, nextSearchKey)
			if n != nil {
				fmt.Println("add param p:", edgeObj.name, searchKey[:pos])
				// if pos < len {
				// 	fmt.Println("add param p:", edgeObj.name, searchKey[:pos])
				// }else{
				// 	fmt.Println("add param p:", edgeObj.name, searchKey)
				// }
				return n
			}
		}
	}
	// 通配符
	if searchNode.Wchildren != nil {
		fmt.Println("add param w:", searchNode.Wchildren.name, searchKey)
		return searchNode.Wchildren.val
	}
	return nil
}



func getSpiltPath(key string) []string {
	if len(key) < 2 {
		return []string{"/"}
	}
	// if key[len(key) - 1] == '/' {
	// 	key += "*"
	// }
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

// 判断字符串str1的前缀是否是str2
func contrainPrefix(str1, str2 string) bool {
	if sub, find := getSubsetPrefix(str1, str2); find {
		return sub == str2
	}
	return false
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

