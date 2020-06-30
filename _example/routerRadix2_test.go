package eudore_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/eudore/eudore"
)

type Context struct {
	eudore.Context
	httpParams eudore.Params
}

func (ctx *Context) Params() *eudore.Params {
	return &ctx.httpParams
}
func (ctx *Context) GetParam(key string) string {
	return ctx.httpParams.Get(key)
}

func newNodeData(path string) (string, string, eudore.HandlerFuncs) {
	return "GET", path, eudore.HandlerFuncs{echoRoute(path)}
}

func echoRoute(path string) eudore.HandlerFunc {
	path = strings.Split(path, " ")[0]
	return func(ctx eudore.Context) {
		if ctx.GetParam("route") != path {
			panic(fmt.Sprintf("route: '%s' '%s', params: %v", path, ctx.GetParam("route"), ctx.Params()))
		}
	}
}

func runCheck(hs eudore.HandlerFuncs, ctx *Context) {
	for _, h := range hs {
		h(ctx)
	}
	ctx.httpParams = eudore.Params{}
}

func TestRadixRouter2(*testing.T) {
	tree := eudore.NewRouterCoreRadix()
	tree.HandleFunc(newNodeData("/api/v1/node/"))
	tree.HandleFunc(newNodeData("/api/v1/:list/11"))
	tree.HandleFunc(newNodeData("/api/v1/:list/22"))
	tree.HandleFunc(newNodeData("/api/v1/:list/:name"))
	tree.HandleFunc(newNodeData("/api/v1/:list/*name"))
	tree.HandleFunc(newNodeData("/api/v1/:list version:v1"))
	tree.HandleFunc(newNodeData("/api/v1/* version:v1"))
	tree.HandleFunc(newNodeData("/api/v2/* version:v2"))
	tree.HandleFunc(newNodeData("/api/*"))
	tree.HandleFunc(newNodeData("/note/get/:name"))
	tree.HandleFunc(newNodeData("/note/:method/:name"))
	tree.HandleFunc(newNodeData("/*"))
	tree.HandleFunc(newNodeData("/"))
	ctx := &Context{}
	runCheck(tree.Match("GET", "/api/v1/node/11", ctx.Params()), ctx)
	runCheck(tree.Match("GET", "/api/v1/node/111", ctx.Params()), ctx)
	runCheck(tree.Match("GET", "/api/v1/node/111/111", ctx.Params()), ctx)
	runCheck(tree.Match("GET", "/api/v1/list", ctx.Params()), ctx)
	runCheck(tree.Match("GET", "/api/v1/list33", ctx.Params()), ctx)
	runCheck(tree.Match("GET", "/api/v1/get", ctx.Params()), ctx)
	runCheck(tree.Match("GET", "/api/v2/get", ctx.Params()), ctx)
	runCheck(tree.Match("GET", "/api/v3/111", ctx.Params()), ctx)
	runCheck(tree.Match("GET", "/note/get/eudore/2", ctx.Params()), ctx)
	runCheck(tree.Match("GET", "/note/get/eudore", ctx.Params()), ctx)
	runCheck(tree.Match("GET", "/note/set/eudore", ctx.Params()), ctx)
	runCheck(tree.Match("GET", "/node", ctx.Params()), ctx)
}

func TestRadixPath1(*testing.T) {
	tree := eudore.NewRouterCoreRadix()
	tree.HandleFunc(newNodeData("/"))
	ctx := &Context{}
	tree.HandleFunc("404", "", eudore.HandlerFuncs{echoRoute("404")})
	tree.HandleFunc("405", "", eudore.HandlerFuncs{echoRoute("405")})
	runCheck(tree.Match("GET", "/", ctx.Params()), ctx)
	runCheck(tree.Match("GET", "/index", ctx.Params()), ctx)
	runCheck(tree.Match("11", "/", ctx.Params()), ctx)
}

func TestRadixPath2(*testing.T) {
	tree := eudore.NewRouterCoreRadix()
	tree.HandleFunc(newNodeData("/authorizations"))
	tree.HandleFunc(newNodeData("/authorizations/:id"))
	ctx := &Context{}
	tree.HandleFunc("404", "", eudore.HandlerFuncs{echoRoute("404")})
	tree.HandleFunc("405", "", eudore.HandlerFuncs{echoRoute("405")})
	runCheck(tree.Match("GET", "/authorizations", ctx.Params()), ctx)
	runCheck(tree.Match("GET", "/authorizations/:id", ctx.Params()), ctx)
	runCheck(tree.Match("11", "/", ctx.Params()), ctx)
}

func TestFullRouter2(*testing.T) {
	tree := eudore.NewRouterCoreFull()
	tree.HandleFunc(newNodeData("/"))
	tree.HandleFunc(newNodeData("/*"))
	tree.HandleFunc(newNodeData("/:id|^[a-z]*$"))
	tree.HandleFunc(newNodeData("/:id|min:4"))
	tree.HandleFunc(newNodeData("/:id|isnum"))
	tree.HandleFunc(newNodeData("/:id"))
	tree.HandleFunc(newNodeData("/api/*id|^[0-9]$"))
	tree.HandleFunc(newNodeData("/1/*id|min:4"))
	tree.HandleFunc(newNodeData("/1/*id|isnum"))
	tree.HandleFunc(newNodeData("/1/*id"))
	tree.HandleFunc(newNodeData("/*id"))
	ctx := &Context{}
	runCheck(tree.Match("GET", "/2", ctx.Params()), ctx)
	runCheck(tree.Match("GET", "/22", ctx.Params()), ctx)
	runCheck(tree.Match("GET", "/abc", ctx.Params()), ctx)
	runCheck(tree.Match("GET", "/abc123", ctx.Params()), ctx)
	runCheck(tree.Match("GET", "/1/2", ctx.Params()), ctx)
	runCheck(tree.Match("GET", "/1/22", ctx.Params()), ctx)
	runCheck(tree.Match("GET", "/1/abc", ctx.Params()), ctx)
	runCheck(tree.Match("GET", "/1/abc123", ctx.Params()), ctx)
	runCheck(tree.Match("GET", "/2/abc123", ctx.Params()), ctx)
	runCheck(tree.Match("GET", "/api/23", ctx.Params()), ctx)
	runCheck(tree.Match("GET", "/api/3", ctx.Params()), ctx)
}
