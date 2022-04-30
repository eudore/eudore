package middleware

import (
	"fmt"
	"strings"

	"github.com/eudore/eudore"
)

// NewRefererFunc 函数创建Referer header检查中间件。
//
// ""                         =>    其他值未匹配时使用的默认值。
//
// "origin"                   =>    请求Referer和Host同源情况下，检查host为referer前缀，origin检查在其他值检查之前。
//
// "*"                        =>    任意域名端口
//
// "www.eudore.cn/*"          =>    www.eudore.cn域名全部请求，不指明http或https时为同时包含http和https
//
// "www.eudore.cn:*/*"        =>    www.eudore.cn任意端口的全部请求，不包含没有指明端口的情况。
//
// "www.eudore.cn/api/*"      =>    www.eudore.cn域名全部/api/前缀的请求
//
// "https://www.eudore.cn/*"  =>    www.eudore.cn仅匹配https。
func NewRefererFunc(data map[string]bool) eudore.HandlerFunc {
	originvalue, origin := data["origin"]
	delete(data, "origin")

	tree := new(refererNode)
	for k, v := range data {
		if strings.HasPrefix(k, "http://") || strings.HasPrefix(k, "https://") || k == "" {
			tree.insert(k).data = fmt.Sprint(v)
		} else {
			tree.insert("http://" + k).data = fmt.Sprint(v)
			tree.insert("https://" + k).data = fmt.Sprint(v)
		}
	}

	return func(ctx eudore.Context) {
		referer := ctx.GetHeader(eudore.HeaderReferer)
		if origin && checkRefererOrigin(ctx, referer) {
			if !originvalue {
				ctx.WriteHeader(eudore.StatusForbidden)
				ctx.WriteString("invalid Referer header " + referer)
				ctx.End()
			}
			return
		}
		node := tree.matchNode(referer)
		if node != nil && node.data == "false" {
			ctx.WriteHeader(eudore.StatusForbidden)
			ctx.WriteString("invalid Referer header " + referer)
			ctx.End()
		}
	}
}

func checkRefererOrigin(ctx eudore.Context, referer string) bool {
	if referer == "" || len(referer) < 8 {
		return false
	}
	return strings.HasPrefix(referer[7:], ctx.Host()) || strings.HasPrefix(referer[8:], ctx.Host())
}

type refererNode struct {
	path     string
	wildcard *refererNode
	children []*refererNode
	data     string
}

func (node *refererNode) insert(path string) *refererNode {
	paths := strings.Split(path, "*")
	newpaths := make([]string, 1, len(paths)*2-1)
	newpaths[0] = paths[0]
	for _, path := range paths[1:] {
		newpaths = append(newpaths, "*")
		if path != "" {
			newpaths = append(newpaths, path)
		}
	}
	for _, p := range newpaths {
		node = node.insertNode(p)
	}
	return node
}

func (node *refererNode) insertNode(path string) *refererNode {
	if path == "*" {
		if node.wildcard == nil {
			node.wildcard = &refererNode{path: path}
		}
		return node.wildcard
	}
	if path == "" {
		return node
	}

	for i := range node.children {
		subStr, find := getSubsetPrefix(path, node.children[i].path)
		if find {
			if subStr != node.children[i].path {
				node.children[i].path = strings.TrimPrefix(node.children[i].path, subStr)
				node.children[i] = &refererNode{
					path:     subStr,
					children: []*refererNode{node.children[i]},
				}
			}
			return node.children[i].insertNode(strings.TrimPrefix(path, subStr))
		}
	}
	newnode := &refererNode{path: path}
	node.children = append(node.children, newnode)
	// 常量node按照首字母排序。
	for i := len(node.children) - 1; i > 0; i-- {
		if node.children[i].path[0] < node.children[i-1].path[0] {
			node.children[i], node.children[i-1] = node.children[i-1], node.children[i]
		}
	}

	return newnode
}

func (node *refererNode) matchNode(path string) *refererNode {
	if path == "" && node.data != "" {
		return node
	}
	for _, current := range node.children {
		if strings.HasPrefix(path, current.path) {
			if result := current.matchNode(path[len(current.path):]); result != nil {
				return result
			}
		}
	}
	if node.wildcard != nil {
		if node.wildcard.children != nil {
			pos := strings.IndexByte(path, '/')
			if pos == -1 {
				pos = len(path)
			}
			if result := node.wildcard.matchNode(path[pos:]); result != nil {
				return result
			}
		}
		if node.wildcard.data != "" {
			return node.wildcard
		}
	}
	return nil
}
