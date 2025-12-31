package eudore

// Implementing a full-featured [RouterCore] based on radix tree

import (
	"context"
	"fmt"
	"strings"
)

// routerCoreMux is implemented based on the radix tree to realize registration
// and matching of all routers.
type routerCoreMux struct {
	Root        *nodeMux
	AnyMethods  []string
	AllMethods  []string
	Params404   Params
	Params405   Params
	Handler404  []HandlerFunc
	Handler405  []HandlerFunc
	funcCreator FuncCreator
}

type nodeMux struct {
	path    string
	name    string
	route   string
	childc  []*nodeMux
	childpv []*nodeMux
	childp  []*nodeMux
	childwv []*nodeMux
	childw  *nodeMux
	check   func(string) bool
	// handlers
	handlers   []nodeMuxHandler
	anyHandler []HandlerFunc
	anyParams  Params
}

type nodeMuxHandler struct {
	method string
	params Params
	funcs  []HandlerFunc
}

// NewRouterCoreMux function creates the [RouterCore] implemented by radix.
//
// The [DefaultRouterAnyMethod] [DefaultRouterAllMethod] data will be copied
// when created.
func NewRouterCoreMux() RouterCore {
	return &routerCoreMux{
		Root:        &nodeMux{},
		AnyMethods:  append([]string{}, DefaultRouterAnyMethod...),
		AllMethods:  append([]string{}, DefaultRouterAllMethod...),
		Handler404:  []HandlerFunc{HandlerRouter404},
		Handler405:  []HandlerFunc{HandlerRouter405},
		funcCreator: DefaultFuncCreator,
	}
}

// The Mount method get [ContextKeyFuncCreator] from [context.Context] to
// create a verification function.
//
// You can make the validation constructor implement the CreateFunc method
// of [FuncCreator].
func (mux *routerCoreMux) Mount(ctx context.Context) {
	fc, ok := ctx.Value(ContextKeyFuncCreator).(FuncCreator)
	if ok {
		mux.funcCreator = fc
	}
}

// HandleFunc method register a new route to the router
//
// The router matches the handlers available to the current path from
// the middleware tree and adds them to the front of the handler.
func (mux *routerCoreMux) HandleFunc(method, path string, hs []HandlerFunc) {
	// Keep the handler as nil behavior,
	// ignore and do not recommend using nil hander.
	if hs == nil {
		return
	}

	switch method {
	case MethodNotFound:
		mux.Params404 = NewParamsRoute(path)[2:]
		mux.Handler404 = hs
	case MethodNotAllowed:
		mux.Params405 = NewParamsRoute(path)[2:]
		mux.Handler405 = hs
	default:
		if method == MethodAny || sliceIndex(mux.AllMethods, method) != -1 {
			mux.insertRoute(method, path, hs)
		}
	}
}

// Match a request, if the method does not allow direct return to node405,
// no match returns node404.
func (mux *routerCoreMux) Match(method, path string, p *Params) []HandlerFunc {
	node := mux.Root.lookNode(path, p)
	// 404
	if node == nil {
		*p = p.Add(mux.Params404...)
		return mux.Handler404
	}

	// no-any method
	p.Set(ParamRoute, node.route)
	for _, h := range node.handlers {
		if h.method == method {
			*p = p.Add(h.params...)
			return h.funcs
		}
	}

	// any method
	if node.anyHandler != nil {
		for _, m := range mux.AnyMethods {
			if m == method {
				*p = p.Add(node.anyParams...)
				return node.anyHandler
			}
		}
	}

	// 405
	allow := strings.Join(mux.getAllows(node), ", ")
	*p = p.Add(ParamAllow, allow).Add(mux.Params405...)
	return mux.Handler405
}

func (mux *routerCoreMux) getAllows(node *nodeMux) []string {
	if node.handlers == nil {
		return mux.AnyMethods
	}

	methods := make([]string, len(node.handlers))
	for i, h := range node.handlers {
		methods[i] = h.method
	}
	return methods
}

// Add a new route Node.
func (mux *routerCoreMux) insertRoute(method, path string, val []HandlerFunc) {
	node := mux.Root
	params := NewParamsRoute(path)
	paths := getSplitPath(params.Get(ParamRoute))
	// create a node
	for _, route := range paths {
		next := &nodeMux{path: route}
		switch route[0] {
		case ':', '*':
			next.name, next.check = mux.loadCheck(route)
		}
		node = node.insertNode(next)
		if route[0] == '*' {
			break
		}
	}
	node.route = strings.Join(paths, "")
	node.setHandler(method, params[2:], val)
}

// Load the checksum function by name.
func (mux *routerCoreMux) loadCheck(path string) (string, func(string) bool) {
	if len(path) == 1 {
		return path, nil
	}
	path = path[1:]
	// Cutting verification function name and parameter
	name, fname, _ := strings.Cut(path, "|")
	if name == "" || fname == "" {
		return path, nil
	}
	// If the prefix is '^', add the regular check function name
	if fname[0] == '^' && fname[len(fname)-1] == '$' {
		fname = "regexp:" + fname
	}

	// use [FuncCreator] to create a check function
	fn, err := mux.funcCreator.CreateFunc(FuncCreateString, fname)
	if err != nil {
		panic(fmt.Errorf(ErrRouterMuxLoadInvalidFunc, path, err))
	}
	return name, fn.(func(string) bool)
}

func (node *nodeMux) setHandler(method string, p Params, hs []HandlerFunc) {
	if method == MethodAny {
		node.anyHandler = hs
		node.anyParams = p
		return
	}

	for i, h := range node.handlers {
		if h.method == method {
			node.handlers[i].params = p
			node.handlers[i].funcs = hs
			return
		}
	}

	node.handlers = append(node.handlers, nodeMuxHandler{method, p, hs})
}

// insertNode add a child node to the node.
func (node *nodeMux) insertNode(next *nodeMux) *nodeMux {
	switch {
	case next.name == "": // const
		return node.insertNodeConst(next)
	case next.path[0] == ':': // param
		if next.check == nil { // param verification
			node.childp, next = nodeMuxSetNext(node.childp, next)
		} else {
			node.childpv, next = nodeMuxSetNext(node.childpv, next)
			// If node.childp is empty,
			// the variable phase will be skipped lookNode.
			if node.childp == nil {
				node.childp = make([]*nodeMux, 0)
			}
		}
	case next.path[0] == '*': // wildcard
		if next.check != nil { // wildcard verification
			node.childwv, next = nodeMuxSetNext(node.childwv, next)
			return next
		}

		// Copy next data and keep the original child info of node.childw.
		if node.childw != nil {
			node.childw.path = next.path
			node.childw.name = next.name
			return node.childw
		}
		node.childw = next
	}
	return next
}

func nodeMuxSetNext(nodes []*nodeMux, next *nodeMux) ([]*nodeMux, *nodeMux) {
	path := next.path
	// check if the current node exists.
	for _, node := range nodes {
		if node.path == path {
			return nodes, node
		}
	}
	return append(nodes, next), next
}

// The insertNodeConst method handles adding constant nodes.
func (node *nodeMux) insertNodeConst(next *nodeMux) *nodeMux {
	for i := range node.childc {
		prefix, find := getSubsetPrefix(next.path, node.childc[i].path)
		if find {
			// Split node path
			if len(prefix) != len(node.childc[i].path) {
				node.childc[i].path = node.childc[i].path[len(prefix):]
				node.childc[i] = &nodeMux{
					path:   prefix,
					childc: []*nodeMux{node.childc[i]},
				}
			}
			next.path = next.path[len(prefix):]

			if next.path == "" {
				return node.childc[i]
			}
			return node.childc[i].insertNodeConst(next)
		}
	}

	node.childc = append(node.childc, next)
	// Constant node is sorted by first char.
	for i := len(node.childc) - 1; i > 0; i-- {
		if node.childc[i].path[0] < node.childc[i-1].path[0] {
			node.childc[i], node.childc[i-1] = node.childc[i-1], node.childc[i]
		}
	}
	return next
}

//nolint:cyclop,gocyclo,gocognit,nestif
func (node *nodeMux) lookNode(path string, params *Params) *nodeMux {
	if path != "" {
		// constant Node match
		for _, child := range node.childc {
			if child.path[0] >= path[0] {
				if strings.HasPrefix(path, child.path) {
					if n := child.lookNode(path[len(child.path):], params); n != nil {
						return n
					}
				}
				break
			}
		}

		// parameter matching, Check if there is a parameter match
		if node.childp != nil {
			pos := strings.IndexByte(path, '/')
			if pos == -1 {
				pos = len(path)
			}
			current, next := path[:pos], path[pos:]

			// check parameter matching
			for _, child := range node.childpv {
				if child.check(current) {
					if n := child.lookNode(next, params); n != nil {
						*params = params.Add(child.name, current)
						return n
					}
				}
			}
			for _, child := range node.childp {
				if n := child.lookNode(next, params); n != nil {
					*params = params.Add(child.name, current)
					return n
				}
			}
		}
	} else if node.route != "" {
		// constant match, return data
		return node
	}

	// wildcard valid node
	for _, child := range node.childwv {
		if child.check(path) {
			*params = params.Add(child.name, path)
			return child
		}
	}
	// wildcard node
	if node.childw != nil {
		*params = params.Add(node.childw.name, path)
		return node.childw
	}
	// can't match, return nil
	return nil
}

/*
The string is cut according to the Node type, String path cutting example:

	/				[/]
	/api/note/		[/api/note/]
	//api/*			[//api/ *]
	//api/*name		[//api/ *name]
	/api/get/		[/api/get/]
	/api/get		[/api/get]
	/api/:get		[/api/ :get]
	/api/:get/*		[/api/ :get / *]
	{/api/**{}:get/*}		[/api/**{ :get / *}]
	/api/:name/info/*		[/api/ :name /info/ *]
	/api/:name|^\\d+$/info	[/api/ :name|^\d+$ /info]
	/api/*|{^0/api\\S+$}	[/api/ *|^0/api\S+$]
	/api/*|^\\$\\d+$		[/api/ *|^\$\d+$]
*/
func getSplitPath(path string) []string {
	var strs []string
	bytes := make([]byte, 0, 64)
	var isblock, isconst bool
	for _, b := range path {
		// block pattern
		if isblock {
			if b == '}' {
				if len(bytes) != 0 && bytes[len(bytes)-1] != '\\' {
					isblock = false
					continue
				}
				// escaping }
				bytes = bytes[:len(bytes)-1]
			}
			bytes = append(bytes, string(b)...)
			continue
		}
		switch b {
		case '/':
			// constant mode, creates a new string in non-constant mode
			if !isconst {
				isconst = true
				strs = append(strs, string(bytes))
				bytes = bytes[:0]
			}
		case ':', '*':
			// variable pattern or wildcard pattern
			isconst = false
			strs = append(strs, string(bytes))
			bytes = bytes[:0]
		case '{':
			isblock = true
			continue
		}
		bytes = append(bytes, string(b)...)
	}
	strs = append(strs, string(bytes))
	if strs[0] == "" {
		strs = strs[1:]
	}
	return strs
}

// Get the largest common prefix of the two strings,
// return the largest common prefix and have the largest common prefix.
func getSubsetPrefix(str2, str1 string) (string, bool) {
	if len(str2) < len(str1) {
		str1, str2 = str2, str1
	}

	for i := range str1 {
		if str1[i] != str2[i] {
			return str1[:i], i > 0
		}
	}
	return str1, true
}
