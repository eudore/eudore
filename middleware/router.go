package middleware

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/eudore/eudore"
)

// NewRouterFunc 函数创建一个路由器中间件，将根据路由路径匹配执行对应的多个处理函数。
//
// 如果key为"router"，val类型为eudore.Router，则使用改路由器处理请求。
func NewRouterFunc(data map[string]interface{}) eudore.HandlerFunc {
	router, ok := data["router"].(eudore.Router)
	delete(data, "router")
	if !ok {
		router = eudore.NewRouterStd(nil)
		router.AddHandler("404", "", eudore.HandlerEmpty)
		router.AddHandler("405", "", eudore.HandlerEmpty)
	}

	for k, v := range data {
		pos := strings.IndexByte(k, '/')
		if pos > 1 {
			router.AddHandler(strings.TrimSpace(k[:pos]), k[pos:], v)
		} else {
			router.AddHandler(eudore.MethodAny, k, v)
		}
	}

	return func(ctx eudore.Context) {
		index, handler := ctx.GetHandler()
		hs := handler[index+1:]
		route := ctx.GetParam(eudore.ParamRoute)
		ctx.SetHandler(-1, eudore.NewHandlerFuncsCombine(router.Match(ctx.Method(), ctx.Path(), ctx.Params()), hs))
		if route != "" {
			ctx.SetParam(eudore.ParamRoute, route)
		}
		ctx.Next()
	}
}

// NewRouterRewriteFunc 函数创建一个根据Router中间件实现的请求路径重写中间件。
//
// RouterRewrite中间件使用参数和Rewrite中间件完全相同。
func NewRouterRewriteFunc(data map[string]string) eudore.HandlerFunc {
	mapping := make(map[string]interface{}, len(data))
	for k, v := range data {
		k = getRouterRewritePath(k)
		mapping[k] = newRouterRewriteFunc(v)
	}
	return NewRouterFunc(mapping)
}

// getRouterRewritePath 函数将Rewrite路由路径转换成默认eudore.Router使用路径。
func getRouterRewritePath(path string) string {
	str := ""
	num := 0
	length := len(path) - 1
	for i := range path {
		if path[i] == '*' {
			if i != length {
				str = fmt.Sprintf("%s:%d", str, num)
			} else {
				str = fmt.Sprintf("%s*%d", str, num)
			}
			num++
		} else {
			str = str + string(path[i])
		}
	}
	return str
}

// newRouterRewriteFunc 函数创建一个对应路径的路径重写处理函数。
func newRouterRewriteFunc(path string) eudore.HandlerFunc {
	paths := strings.Split(path, "$")
	Index := make([]string, 1, len(paths)*2-1)
	Data := make([]string, 1, len(paths)*2-1)
	Index[0] = ""
	Data[0] = paths[0]
	for _, path := range paths[1:] {
		Index = append(Index, path[0:1])
		Data = append(Data, "")
		if path[1:] != "" {
			Index = append(Index, "")
			Data = append(Data, path[1:])
		}
	}
	return func(ctx eudore.Context) {
		buffer := bytes.NewBuffer(nil)
		for i := range Index {
			if Index[i] == "" {
				buffer.WriteString(Data[i])
			} else {
				buffer.WriteString(ctx.GetParam(Index[i]))
			}
		}
		ctx.Request().URL.Path = buffer.String()
	}
}
