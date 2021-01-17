package middleware

import (
	"bytes"
	"strings"

	"github.com/eudore/eudore"
)

type rewriteNode struct {
	path     string
	wildcard *rewriteNode
	children []*rewriteNode
	Index    []int
	Data     []string
}

type matchResult struct {
	Params []string
	Index  []int
	Data   []string
}

// NewRewriteFunc 函数创建一个请求路径重写处理函数，当前默认使用树匹配实现。
//
// 匹配路径中使用'*'代表当前位置到下一个'/'匹配内容，如果'*'在结尾表示任意字符串
//
// 目标路径中$0到$9表示匹配路径中'*'出现位置匹配到的字符串，最多匹配10个'*'，否在解析错误。
//
// "/js/*"                     =>  "/public/js*"
//
// "/api/v1/*"                 =>  "/api/v3/$0"
//
// "/api/v1/users/*/orders/*"  =>  "/api/v3/user/$0/order/$1"
//
// "/d/*"                      =>  "/d/$0-$0"
//
// 若运行出现`panic: runtime error: index out of range`，请检测'$?'的值是否超出了'*'数量。
func NewRewriteFunc(data map[string]string) eudore.HandlerFunc {
	node := new(rewriteNode)
	for k, v := range data {
		node.insert(k, v)
	}
	return node.HandleHTTP
}

func (node *rewriteNode) HandleHTTP(ctx eudore.Context) {
	ctx.Request().URL.Path = node.rewrite(ctx.Path())
}

func (node *rewriteNode) insert(path, mapping string) {
	paths := strings.Split(path, "*")
	newpaths := make([]string, 1, len(paths)*2-1)
	newpaths[0] = paths[0]
	for _, path := range paths[1:] {
		newpaths = append(newpaths, "*")
		if path != "" {
			newpaths = append(newpaths, path)
		}
	}
	index, data := getRewritePathData(mapping, len(paths)+46)
	node.insertNode(newpaths, index, data)
}

func getRewritePathData(path string, num int) ([]int, []string) {
	paths := strings.Split(path, "$")
	Index := make([]int, 1, len(paths)*2-1)
	Data := make([]string, 1, len(paths)*2-1)
	Index[0] = -1
	Data[0] = paths[0]
	for _, path := range paths[1:] {
		Index = append(Index, num-int(path[0]))
		Data = append(Data, "")
		if path[1:] != "" {
			Index = append(Index, -1)
			Data = append(Data, path[1:])
		}
	}
	return Index, Data
}

func (node *rewriteNode) rewrite(path string) string {
	result := &matchResult{}
	if !node.matchNode(path, result) {
		return path
	}
	buffer := bytes.NewBuffer(nil)
	for i := range result.Index {
		if result.Index[i] == -1 {
			buffer.WriteString(result.Data[i])
		} else {
			buffer.WriteString(result.Params[result.Index[i]])
		}
	}
	return buffer.String()
}

func (node *rewriteNode) insertNode(path []string, index []int, data []string) {
	if len(path) == 0 {
		node.Index = index
		node.Data = data
		return
	}
	for i := range node.children {
		subStr, find := getSubsetPrefix(path[0], node.children[i].path)
		if find {
			if subStr != node.children[i].path {
				node.children[i].path = strings.TrimPrefix(node.children[i].path, subStr)
				node.children[i] = &rewriteNode{
					path:     subStr,
					children: []*rewriteNode{node.children[i]},
				}
			}
			path[0] = strings.TrimPrefix(path[0], subStr)
			if path[0] == "" {
				path = path[1:]
			}
			node.children[i].insertNode(path, index, data)
			return
		}
	}
	newnode := &rewriteNode{path: path[0]}
	if path[0] == "*" {
		node.wildcard = newnode
	} else {
		node.children = append(node.children, newnode)
		// 常量node按照首字母排序。
		for i := len(node.children) - 1; i > 0; i-- {
			if node.children[i].path[0] < node.children[i-1].path[0] {
				node.children[i], node.children[i-1] = node.children[i-1], node.children[i]
			}
		}
	}

	if len(path) == 1 {
		newnode.Index = index
		newnode.Data = data
	} else {
		newnode.insertNode(path[1:], index, data)
	}
}

func (node *rewriteNode) matchNode(path string, result *matchResult) bool {
	if path == "" && node.Index != nil {
		result.Index = node.Index
		result.Data = node.Data
		return true
	}
	for _, current := range node.children {
		if strings.HasPrefix(path, current.path) && current.matchNode(path[len(current.path):], result) {
			return true
		}
	}
	if node.wildcard != nil {
		if node.wildcard.children != nil {
			pos := strings.IndexByte(path, '/')
			if pos == -1 {
				pos = len(path)
			}
			if node.wildcard.matchNode(path[pos:], result) {
				result.Params = append(result.Params, path[:pos])
				return true
			}
		}
		if node.wildcard.Index != nil {
			result.Index = node.wildcard.Index
			result.Data = node.wildcard.Data
			result.Params = append(result.Params, path)
			return true
		}
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
