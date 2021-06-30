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

type radixNode struct {
	path     string
	children []*radixNode
	val      interface{}
}

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
	fmt.Println(tree.Lookup("app"))
	fmt.Println(tree.Lookup("append"))
	tree.Delete("app")
	fmt.Println(tree.Lookup("app"))
	fmt.Println(tree.Lookup("append"))
}

func newRadixTree() *radixNode {
	return &radixNode{}
}

func (node *radixNode) Insert(path string, val interface{}) {
	if path == "" {
		node.val = val
		return
	}
	for i := range node.children {
		// 检查当前遍历的Node和插入路径是否有公共路径,subStr是两者的公共路径，find表示是否有
		subStr, find := getSubsetPrefix(path, node.children[i].path)
		if find {
			// 如果child路径大于公共最大路径，则进行node分裂
			if subStr != node.children[i].path {
				node.children[i].path = strings.TrimPrefix(node.children[i].path, subStr)
				node.children[i] = &radixNode{
					path:     subStr,
					children: []*radixNode{node.children[i]},
				}
			}
			node.children[i].Insert(strings.TrimPrefix(path, subStr), val)
			return
		}
	}
	// 没有相同前缀路径存在，直接添加为child
	node.children = append(node.children, &radixNode{path: path, val: val})
}

func (node *radixNode) Delete(path string) bool {
	if len(path) == 0 {
		node.val = nil
		return true
	}
	for i, child := range node.children {
		if contrainPrefix(path, child.path) && child.Delete(strings.TrimPrefix(path, child.path)) {
			if len(child.children) == 0 {
				for ; i < len(node.children)-1; i++ {
					node.children[i] = node.children[i+1]
				}
				node.children = node.children[:len(node.children)-1]
			}
			return true
		}
	}
	return false
}

// Lookup 递归获得searchNode路径为searchKey的Node数据。
func (node *radixNode) Lookup(path string) interface{} {
	if len(path) == 0 {
		return node.val
	}
	for _, child := range node.children {
		// 寻找相同前缀node,截取为匹配的路径,然后当前Node递归判断
		if contrainPrefix(path, child.path) {
			return child.Lookup(strings.TrimPrefix(path, child.path))
		}
	}
	return nil
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
