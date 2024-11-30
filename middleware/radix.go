package middleware

import (
	"bytes"
	"net/http"
	"net/textproto"
	"strings"

	"github.com/eudore/eudore"
)

var (
	stateTrue   = []byte{0}
	stateFalse  = []byte{}
	schemeHTTP  = "http://"
	schemeHTTPS = "https://"
)

// NewCORSFunc function creates middleware to implement handle CORS.
//
// pattens is the allowed origins, and headers is the headers added
// after successful cross-domain verification.
//
// If pattens is empty, any origin is allowed.
//
// If [eudore.HeaderAccessControlAllowMethods] or
// [eudore.HeaderAccessControlAllowHeaders] is empty, set it to *.
//
// When cors registration is not a global middleware,
// you need to register Options /* or 404 method for the last time,
// otherwise the Options request matches the default 404 and is not processed.
//
//	app.AddMiddleware("global", middleware.NewCORSFunc([]string{"example.com", "127.0.0.1:*"}, map[string]string{
//		"Access-Control-Allow-Credentials": "true",
//		"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, HEAD",
//		"Access-Control-Allow-Headers":     "Content-Type,X-Request-Id,X-CustomHeader",
//		"Access-Control-Expose-Headers":    "X-Request-Id",
//		"Access-Control-Max-Age":           "1000",
//	}))
//
// * matches the next character . or / or :, last * matches to the end.
func NewCORSFunc(patterns []string, headers map[string]string) Middleware {
	corsHeaders := make(http.Header, len(headers))
	corsHeaders[eudore.HeaderAccessControlAllowMethods] = []string{"*"}
	corsHeaders[eudore.HeaderAccessControlAllowHeaders] = []string{"*"}
	for k, v := range headers {
		corsHeaders[textproto.CanonicalMIMEHeaderKey(k)] = []string{v}
	}

	node := new(radixNode[byte])
	if patterns == nil {
		patterns = []string{"*"}
	}
	for _, pattern := range patterns {
		node.insert(pattern, stateTrue)
	}
	return func(ctx eudore.Context) {
		origin := ctx.GetHeader(eudore.HeaderOrigin)
		host := trimScheme(origin)
		// Check is no same-origin request
		// Origin header exists for cors and upgrade.
		if host == "" || host == ctx.Host() {
			return
		}

		h := ctx.Response().Header()
		headerVary(h, eudore.HeaderOrigin)
		if node.lookNode(host) == nil {
			writePage(ctx, eudore.StatusForbidden, DefaultPageCORS, host)
			ctx.End()
			return
		}

		h.Add(eudore.HeaderAccessControlAllowOrigin, origin)
		if ctx.Method() == eudore.MethodOptions {
			headerCopy(h, corsHeaders)
			ctx.WriteHeader(eudore.StatusNoContent)
			ctx.End()
		}
	}
}

// can inline with cost 49.
func trimScheme(host string) string {
	switch {
	case strings.HasPrefix(host, schemeHTTP):
		return host[7:]
	case strings.HasPrefix(host, schemeHTTPS):
		return host[8:]
	}
	return host
}

// The NewRefererCheckFunc function creates middleware to implement
// [eudore.HeaderReferer] check.
//
// The map key specifies the url pattern,
// and the map value sets whether it is allowed.
//
// There are three special values: empty string, *, origin, which respectively
// describe whether [eudore.HeaderReferer] is valid when there is
// no Header, any value, or the Host has the same origin.
//
// * matches the next character . or / or :, last * matches to the end.
func NewRefererCheckFunc(data map[string]bool) Middleware {
	originvalue, origin := data["origin"]
	delete(data, "origin")

	node := new(radixNode[byte])
	for k, v := range data {
		var state []byte
		if v {
			state = stateTrue
		} else {
			state = stateFalse
		}

		if strings.HasPrefix(k, schemeHTTP) ||
			strings.HasPrefix(k, schemeHTTPS) ||
			k == "" || k == "*" {
			node.insert(k, state)
		} else {
			node.insert(schemeHTTP+k, state)
			node.insert(schemeHTTPS+k, state)
		}
	}

	return func(ctx eudore.Context) {
		referer := ctx.GetHeader(eudore.HeaderReferer)
		if origin && refererCheckOrigin(ctx, referer) {
			if originvalue {
				return
			}
		} else if len(node.lookNode(referer)) == 1 {
			return
		}
		writePage(ctx, eudore.StatusForbidden, DefaultPageReferer, referer)
		ctx.End()
	}
}

func refererCheckOrigin(ctx eudore.Context, referer string) bool {
	if len(referer) < 8 {
		return false
	}

	pos := strings.Index(referer[:8], "://")
	if pos != -1 {
		referer = referer[pos+3:]
	}
	return strings.HasPrefix(referer, ctx.Host()+"/")
}

// NewRewriteFunc function creates middleware to implement path rewrite.
//
// pattern uses * as a placeholder; the new path uses $0-$9 to reference
// matching values, with a maximum of 10 values.
//
// * matches the next character /, last * matches to the end.
func NewRewriteFunc(data map[string]string) Middleware {
	node := new(radixNode[string])
	for k, v := range data {
		node.insert(k, splitRewritePattern(v, strings.Count(k, "*")))
	}
	return func(ctx eudore.Context) {
		params := []string{}
		pattern := node.lookNodeParams(ctx.Path(), &params)
		if pattern != nil {
			buf := bytes.NewBuffer(nil)
			for _, p := range pattern {
				if len(p) == 1 && int(p[0]) < len(params) {
					buf.WriteString(params[p[0]])
				} else {
					buf.WriteString(p)
				}
			}

			path := buf.String()
			r := ctx.Request()
			r.URL.Path = path
			r.RequestURI = r.URL.String()
		}
	}
}

type radixNode[T comparable] struct {
	path     string
	data     []T
	child    []*radixNode[T]
	wildcard *radixNode[T]
}

func (node *radixNode[T]) insert(path string, data []T) {
	for i, route := range strings.Split(path, "*") {
		if i != 0 {
			node = node.insertNode(&radixNode[T]{path: "*"})
		}
		node = node.insertNode(&radixNode[T]{path: route})
	}
	node.data = data
}

func (node *radixNode[T]) insertNode(next *radixNode[T]) *radixNode[T] {
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
				node.child[i].path = node.child[i].path[len(prefix):]
				node.child[i] = &radixNode[T]{
					path:  prefix,
					child: []*radixNode[T]{node.child[i]},
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

func (node *radixNode[T]) lookNode(path string) []T {
	if path == "" && node.data != nil {
		return node.data
	}

	if path != "" {
		for _, child := range node.child {
			if child.path[0] >= path[0] {
				length := len(child.path)
				if len(path) >= length && path[:length] == child.path {
					data := child.lookNode(path[length:])
					if data != nil {
						return data
					}
				}
				break
			}
		}
	}

	if node.wildcard != nil {
		if node.wildcard.child != nil {
			// diff split char
			pos := indexBytes(path)
			data := node.wildcard.lookNode(path[pos:])
			if data != nil {
				return data
			}
		}
		if node.wildcard.data != nil {
			return node.wildcard.data
		}
	}
	return nil
}

func (node *radixNode[T]) lookNodeParams(path string, params *[]string) []T {
	if path == "" && node.data != nil {
		return node.data
	}

	if path != "" {
		for _, child := range node.child {
			if child.path[0] >= path[0] {
				length := len(child.path)
				if len(path) >= length && path[:length] == child.path {
					data := child.lookNodeParams(path[length:], params)
					if data != nil {
						return data
					}
				}
				break
			}
		}
	}

	if node.wildcard != nil {
		if node.wildcard.child != nil {
			pos := strings.IndexByte(path, '/')
			if pos == -1 {
				pos = len(path)
			}
			data := node.wildcard.lookNodeParams(path[pos:], params)
			if data != nil {
				*params = append(*params, path[:pos])
				return data
			}
		}
		if node.wildcard.data != nil {
			*params = append(*params, path)
			return node.wildcard.data
		}
	}
	return nil
}

var splitCharsURL = []byte{'.', '/', ':'}

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

func splitRewritePattern(pattern string, count int) []string {
	var strs []string
	var bytes []byte
	var isvar bool
	for _, b := range pattern {
		if b == '$' {
			strs = append(strs, string(bytes))
			bytes = bytes[:0]
			isvar = true
			continue
		}

		if isvar {
			isvar = false
			index := byte(b) - 0x30
			if index < byte(count) {
				strs = append(strs, string([]byte{index}))
			} else {
				bytes = append(bytes, strs[len(strs)-1]...)
				bytes = append(bytes, '$')
				bytes = append(bytes, string(b)...)
				strs = strs[:len(strs)-1]
			}
		} else {
			bytes = append(bytes, string(b)...)
		}
	}
	if isvar {
		bytes = append(bytes, '$')
	}
	strs = append(strs, string(bytes))
	if strs[0] == "" {
		strs = strs[1:]
	}
	return strs
}

// from routermux.go.
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
