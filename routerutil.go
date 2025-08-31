package eudore

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"strings"
)

var routerLoggerKinds = [...]string{
	"handler",
	"controller",
	"middleware",
	"extend",
	"error",
	"metadata",
}

func getRouterLoggerKind(kind int, str string) int {
	str = strings.ReplaceAll(strings.ToLower(str), " ", "")
	for _, k := range strings.Split(str, "|") {
		switch k {
		case "all":
			kind = 0x3f
		case "~all":
			kind = 0
		default:
			pos := sliceIndex(routerLoggerKinds[:], strings.TrimPrefix(k, "~"))
			if pos != -1 {
				if strings.HasPrefix(k, "~") {
					kind &^= (1 << pos)
				} else {
					kind |= (1 << pos)
				}
			}
		}
	}
	return kind
}

func getRouterDepthWithFunc(start, size int, fn string) int {
	pc := make([]uintptr, size)
	n := runtime.Callers(start+1, pc)
	if n > 0 {
		index := start
		frames := runtime.CallersFrames(pc[:n])
		frame, more := frames.Next()
		for more {
			if strings.HasSuffix(frame.Function, fn) {
				return index
			}

			index++
			frame, more = frames.Next()
		}
	}
	return start
}

// The route information added by the addHandler method.
func addMetadataRouter(r *MetadataRouter, method, path string,
	handlers []HandlerFunc,
) {
	names := make([]string, len(handlers))
	for i := range handlers {
		names[i] = fmt.Sprint(handlers[i])
	}
	r.Methods = append(r.Methods, method)
	r.Paths = append(r.Paths, getRoutePath(path))
	r.Params = append(r.Params, NewParamsRoute(path))
	r.HandlerNames = append(r.HandlerNames, names)
}

// middlewareTree defines the middleware storage tree.
type middlewareTree struct {
	root  middlewareNode
	index int
}
type middlewareNode = radixNode[*middlewareDatas, middlewareDatas]

type middlewareData struct {
	index    int
	handlers HandlerFuncs
}
type middlewareDatas []middlewareData

func (data *middlewareDatas) Insert(i ...any) error {
	*data = append(*data, middlewareData{i[0].(int), i[1].(HandlerFuncs)})
	return nil
}

// The Insert method implements middlewareNode to add a child node.
func (t *middlewareTree) Insert(path string, val []HandlerFunc) {
	t.index++
	_ = t.root.insert(path, t.index, val)
}

// Lookup Find if seachKey exist in current trie tree and return its value.
func (t *middlewareTree) Lookup(path string) []HandlerFunc {
	var data []middlewareData
	for _, v := range t.root.lookPath(path) {
		data = append(data, *v...)
	}
	sort.Slice(data, func(i, j int) bool {
		return data[i].index < data[j].index
	})

	handlers := make([]HandlerFunc, 0, len(data))
	for i := range data {
		handlers = append(handlers, data[i].handlers...)
	}
	return handlers
}

// The clone method deeply copies this middleware node.
func (t *middlewareTree) clone() *middlewareTree {
	return &middlewareTree{
		root:  *middlewareClone(&t.root),
		index: t.index,
	}
}

func middlewareClone(node *middlewareNode) *middlewareNode {
	next := &middlewareNode{
		path: node.path,
	}
	if node.data != nil {
		next.data = new(middlewareDatas)
		*next.data = append(*next.data, *node.data...)
	}
	for i := range node.child {
		next.child = append(next.child, middlewareClone(node.child[i]))
	}
	return next
}

// NewRouterCoreHost function creates [RouterCore] to implement host matching.
//
// When HandleFunc, if the registration path param has [ParamRouteHost],
// the registration path is {host}{path}. host ignores prefix *.
//
// When HandleHTTP, if ctx.Host() has a matching pattern,
// the matching path is {pattern}{path}.
//
// * matches the char before the next . or : in pattern.
func NewRouterCoreHost(core RouterCore) RouterCore {
	if core == nil {
		core = NewRouterCoreMux()
	}
	return &routerCoreHost{RouterCore: core}
}

type routerCoreHost struct {
	RouterCore
	patterns routerHostNode
}

func (r *routerCoreHost) Mount(ctx context.Context) {
	anyMount(ctx, r.RouterCore)
}

func (r *routerCoreHost) Unmount(ctx context.Context) {
	anyUnmount(ctx, r.RouterCore)
}

func (r *routerCoreHost) HandleFunc(m string, path string, hs []HandlerFunc) {
	host := NewParamsRoute(path).Get(ParamRouteHost)
	if host == "" {
		r.RouterCore.HandleFunc(m, path, hs)
		return
	}

	for _, h := range strings.Split(host, ",") {
		pattern := strings.TrimPrefix(h, "*")
		r.patterns.insert(h, pattern)
		r.RouterCore.HandleFunc(m, fmt.Sprintf("{%s}%s", pattern, path), hs)
	}
}

// The Match method returns the HandleHTTP function to process the request,
// use the host value to match the request handler function.
//
// You can improve performance by replacing app.serveContext with r.HandleHTTP.
func (r *routerCoreHost) Match(string, string, *Params) []HandlerFunc {
	return []HandlerFunc{r.HandleHTTP}
}

func (r *routerCoreHost) HandleHTTP(ctx Context) {
	path := ctx.Path()
	// find pattern by host
	pattern := r.patterns.lookNode(ctx.Host())
	if pattern != "" {
		path = pattern + path
	}

	params := ctx.Params()
	ctx.SetHandlers(-1, r.RouterCore.Match(ctx.Method(), path, params))
	params.Set(ParamRoute, (*params)[1][len(pattern):])
	ctx.Next()
}

type routerHostNode struct {
	path     string
	data     string
	child    []*routerHostNode
	wildcard *routerHostNode
}

func (node *routerHostNode) insert(host, pattern string) {
	for i, route := range strings.Split(host, "*") {
		if i != 0 {
			node = node.insertNode(&routerHostNode{path: "*"})
		}
		node = node.insertNode(&routerHostNode{path: route})
	}
	node.data = pattern
}

func (node *routerHostNode) insertNode(next *routerHostNode) *routerHostNode {
	if next.path == "" {
		return node
	}

	if next.path == "*" {
		if node.wildcard == nil {
			node.wildcard = next
		}
		return node.wildcard
	}

	for i := range node.child {
		prefix, find := getSubsetPrefix(next.path, node.child[i].path)
		if find {
			if prefix != node.child[i].path {
				sub := node.child[i].path[len(prefix):]
				node.child[i].path = sub
				node.child[i] = &routerHostNode{
					path:  prefix,
					child: []*routerHostNode{node.child[i]},
				}
			}
			next.path = next.path[len(prefix):]
			return node.child[i].insertNode(next)
		}
	}
	node.child = append(node.child, next)
	for i := len(node.child) - 1; i > 0; i-- {
		if node.child[i].path[0] < node.child[i-1].path[0] {
			node.child[i], node.child[i-1] = node.child[i-1], node.child[i]
		}
	}
	return next
}

func (node *routerHostNode) lookNode(path string) string {
	if path == "" && node.data != "" {
		return node.data
	}

	if path != "" {
		char := path[0]
		for _, child := range node.child {
			if child.path[0] < char {
				continue
			}

			if child.path[0] == char && strings.HasPrefix(path, child.path) {
				data := child.lookNode(path[len(child.path):])
				if data != "" {
					return data
				}
			}
			break
		}
	}

	if node.wildcard != nil {
		if node.wildcard.child != nil {
			data := node.wildcard.lookNode(path[indexBytes(path):])
			if data != "" {
				return data
			}
		}
		if node.wildcard.data != "" {
			return node.wildcard.data
		}
	}
	return ""
}

var splitCharsURL = []byte{'.', ':'}

func indexBytes(path string) int {
	pos := len(path)
	for i := range splitCharsURL {
		p := strings.IndexByte(path[:pos], splitCharsURL[i])
		if p != -1 && p < pos {
			pos = p
		}
	}
	return pos
}
