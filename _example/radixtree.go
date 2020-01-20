package main

/*
radixrouter请参考框架源码实现。

test新增team
1、
test
2、分叉成
te
--st
3、新增
te
--st
--am

结果就是
test和team
*/

import (
	"fmt"
	"strings"
)

type (
	radixTree struct {
		root radixNode
	}
	radixNode struct {
		path     string
		children []*radixNode
		key      string
		val      interface{}
	}
)

func main() {
	tree := newRadixTree()
	tree.Insert("test", 1)
	tree.Insert("test22", 1)
	tree.Insert("team", 3)
	tree.Insert("apple", 4)
	tree.Insert("append", 12)
	tree.Insert("app", 5)
	tree.Insert("append", 6)
	tree.Insert("interface", 7)
	fmt.Println(tree.Lookup("append"))
}

func newRadixTree() *radixTree {
	return &radixTree{radixNode{}}
}

// 新增Node
func (r *radixNode) InsertNode(path, key string, value interface{}) {
	if len(path) == 0 {
		// 路径空就设置当前node的值
		r.key = key
		r.val = value
	} else {
		// 否则新增node
		r.children = append(r.children, &radixNode{path: path, key: key, val: value})
	}
}

// 对指定路径为edgeKey的Node分叉，公共前缀路径为pathKey
func (r *radixNode) SplitNode(pathKey, edgeKey string) *radixNode {
	for i := range r.children {
		// 找到路径为edgeKey路径的Node，然后分叉
		if r.children[i].path == edgeKey {
			// 创建新的分叉Node，路径为公共前缀路径pathKey
			newNode := &radixNode{path: pathKey}
			// 将原来edgeKey的数据移动到新的分叉Node之下
			// 直接新增Node，原Node数据仅改变路径为截取后的后段路径
			newNode.children = append(newNode.children, &radixNode{
				// 截取路径
				path: strings.TrimPrefix(edgeKey, pathKey),
				// 复制数据
				key:      r.children[i].key,
				val:      r.children[i].val,
				children: r.children[i].children,
			})
			// 设置radixNode的child[i]的Node为分叉Node
			// 原理路径Node的数据移到到了分叉Node的child里面，原Node对象GC释放。
			r.children[i] = newNode
			// 返回分叉新创建的Node
			return newNode
		}
	}
	return nil
}

func (t *radixTree) Insert(key string, val interface{}) {
	t.recursiveInsertTree(&t.root, key, key, val)
}

// 给currentNode递归添加，路径为containKey的Node
//
// targetKey和targetValue为新Node数据。
func (t *radixTree) recursiveInsertTree(currentNode *radixNode, containKey string, targetKey string, targetValue interface{}) {
	for i := range currentNode.children {
		// 检查当前遍历的Node和插入路径是否有公共路径
		// subStr是两者的公共路径，find表示是否有
		subStr, find := getSubsetPrefix(containKey, currentNode.children[i].path)
		if find {
			// 如果child路径等于公共最大路径，则该node添加child
			// child的路径为插入路径先过滤公共路径的后面部分。
			if subStr == currentNode.children[i].path {
				nextTargetKey := strings.TrimPrefix(containKey, currentNode.children[i].path)
				// 当前node新增子Node可能原本有多个child，所以需要递归添加
				t.recursiveInsertTree(currentNode.children[i], nextTargetKey, targetKey, targetValue)
			} else {
				// 如果公共路径不等于当前node的路径
				// 则将currentNode.children[i]路径分叉
				// 分叉后的就拥有了公共路径，然后添加新Node
				newNode := currentNode.SplitNode(subStr, currentNode.children[i].path)
				if newNode == nil {
					panic("Unexpect error on split node")
				}
				// 添加新的node
				// 分叉后树一定只有一个没有相同路径的child，所以直接添加node
				newNode.InsertNode(strings.TrimPrefix(containKey, subStr), targetKey, targetValue)
			}
			return
		}
	}
	// 没有相同前缀路径存在，直接添加为child
	currentNode.InsertNode(containKey, targetKey, targetValue)
}

//Lookup: Find if seachKey exist in current radix tree and return its value
func (t *radixTree) Lookup(searchKey string) (interface{}, bool) {
	return t.recursiveLoopup(&t.root, searchKey)
}

// 递归获得searchNode路径为searchKey的Node数据。
func (t *radixTree) recursiveLoopup(searchNode *radixNode, searchKey string) (interface{}, bool) {
	// 匹配node，返回数据
	if len(searchKey) == 0 {
		return searchNode.val, true
	}

	for _, edgeObj := range searchNode.children {
		// 寻找相同前缀node
		if contrainPrefix(searchKey, edgeObj.path) {
			// 截取为匹配的路径
			nextSearchKey := strings.TrimPrefix(searchKey, edgeObj.path)
			// 然后当前Node递归判断
			return t.recursiveLoopup(edgeObj, nextSearchKey)
		}
	}

	return nil, false
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
