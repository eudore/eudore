package policy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	// "github.com/kr/pretty"
)

func stringSliceNotIn(strs []string, str string) bool {
	for _, i := range strs {
		if i == str {
			return false
		}
	}
	return true
}

// ControllerAction 定义生成action参考控制器
type ControllerAction struct{}

// ControllerParam 方法定义ControllerAction生成action参数。
func (ControllerAction) ControllerParam(pkg, name, method string) string {
	pos := strings.LastIndexByte(pkg, '/') + 1
	if pos != 0 {
		pkg = pkg[pos:]
	}
	if strings.HasSuffix(name, "Controller") {
		name = name[:len(name)-len("Controller")]
	}

	return fmt.Sprintf("action=%s:%s:%s", pkg, name, method)
}

// NewSignaturerJwt 函数创建一个Jwt Signaturer
func NewSignaturerJwt(secret []byte) Signaturer {
	return verifyFunc(func(b []byte) string {
		h := hmac.New(sha256.New, secret)
		h.Write(b)
		return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
	})
}

type verifyFunc func([]byte) string

func (fn verifyFunc) Signed(claims interface{}) string {
	payload, _ := json.Marshal(claims)
	var unsigned string = `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.` + base64.RawURLEncoding.EncodeToString(payload)
	return fmt.Sprintf("%s.%s", unsigned, fn([]byte(unsigned)))
}

func (fn verifyFunc) Parse(token string, dst interface{}) error {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return errors.New("Error: incorrect # of results from string parsing.")
	}

	if fn([]byte(parts[0]+"."+parts[1])) != parts[2] {
		return errors.New("Error：jwt validation error.")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return err
	}

	err = json.Unmarshal(payload, dst)
	if err != nil {
		return err
	}
	return nil
}

type starTree struct {
	// Index int
	Name string
	Path string
	// Const   string
	children []*starTree
	wildcard *starTree
}

func newStarTree(data []string) *starTree {
	tree := &starTree{}
	for i := range data {
		tree.Insert(data[i])
	}
	return tree
}

func (tree *starTree) Insert(path string) {
	for {
		pos := strings.Index(path, "**")
		if pos == -1 {
			break
		}
		path = strings.Replace(path, "**", "*", -1)
	}
	for i, s := range strings.Split(path, "*") {
		if i != 0 {
			tree = tree.insertNodeWildcard()
		}
		if s != "" {
			tree = tree.insertNode(s, &starTree{})
		}
	}
	tree.Name = path
}

func (tree *starTree) insertNode(path string, next *starTree) *starTree {
	if path == "" {
		return tree
	}
	for i := range tree.children {
		subStr, find := getSubsetPrefix(path, tree.children[i].Path)
		if find {
			if subStr != tree.children[i].Path {
				tree.children[i].Path = strings.TrimPrefix(tree.children[i].Path, subStr)
				tree.children[i] = &starTree{
					Path:     subStr,
					children: []*starTree{tree.children[i]},
				}
			}
			return tree.children[i].insertNode(strings.TrimPrefix(path, subStr), next)
		}
	}

	tree.children = append(tree.children, next)
	next.Path = path
	for i := len(tree.children) - 1; i > 0; i-- {
		if tree.children[i].Path[0] < tree.children[i-1].Path[0] {
			tree.children[i], tree.children[i-1] = tree.children[i-1], tree.children[i]
		}
	}
	return next
}

func (tree *starTree) insertNodeWildcard() *starTree {
	if tree.wildcard == nil {
		tree.wildcard = new(starTree)
	}
	return tree.wildcard
}

func (tree *starTree) Match(path string) string {
	if path != "" {
		for _, child := range tree.children {
			if child.Path[0] >= path[0] {
				if len(path) >= len(child.Path) && path[:len(child.Path)] == child.Path {
					if n := child.Match(path[len(child.Path):]); n != "" {
						return n
					}
				}
				break
			}
		}
	} else {
		if tree.Name != "" {
			return tree.Name
		}
	}

	if tree.wildcard == nil {
		return ""
	}
	tree = tree.wildcard
	if tree.children == nil {
		return tree.Name
	}

	for i := len(getConstPrifix(path)); i > -1; i-- {
		n := tree.Match(path[i:])
		if n != "" {
			return n
		}
	}
	return ""
}

func getConstPrifix(str string) string {
	for i, s := range str {
		if s == ':' || s == '/' {
			return str[:i]
		}
	}
	return str
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
