package test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/eudore/eudore"
	// "github.com/kr/pretty"
)

type EudoreContext = eudore.Context
type Context struct {
	EudoreContext
	eudore.ParamsArray
}

func (ctx *Context) Params() eudore.Params {
	return &ctx.ParamsArray
}
func (ctx *Context) GetParam(key string) string {
	return ctx.ParamsArray.Get(key)
}
func (ctx *Context) Get(key string) string {
	return ctx.ParamsArray.Get(key)
}

func (ctx *Context) Add(key string, val string) {
	fmt.Println("Add", key, val)
	ctx.ParamsArray.Set(key, val)
}
func (ctx *Context) Set(key string, val string) {
	fmt.Println("Set", key, val)
	ctx.ParamsArray.Set(key, val)
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
}

func TestRadixRouter(*testing.T) {
	tree := eudore.NewRouterRadix()
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
	// fmt.Printf("%# v\n", pretty.Formatter(tree))
	ctx := &Context{}
	runCheck(tree.Match("GET", "/api/v1/node/11", ctx), ctx)
	runCheck(tree.Match("GET", "/api/v1/node/111", ctx), ctx)
	runCheck(tree.Match("GET", "/api/v1/node/111/111", ctx), ctx)
	runCheck(tree.Match("GET", "/api/v1/list", ctx), ctx)
	runCheck(tree.Match("GET", "/api/v1/list33", ctx), ctx)
	runCheck(tree.Match("GET", "/api/v1/get", ctx), ctx)
	runCheck(tree.Match("GET", "/api/v2/get", ctx), ctx)
	runCheck(tree.Match("GET", "/api/v3/111", ctx), ctx)
	runCheck(tree.Match("GET", "/note/get/eudore/2", ctx), ctx)
	runCheck(tree.Match("GET", "/note/get/eudore", ctx), ctx)
	runCheck(tree.Match("GET", "/note/set/eudore", ctx), ctx)
	runCheck(tree.Match("GET", "/node", ctx), ctx)
}

func TestRadixPath1(*testing.T) {
	tree := eudore.NewRouterRadix()
	tree.HandleFunc(newNodeData("/"))
	ctx := &Context{}
	tree.HandleFunc("404", "", eudore.HandlerFuncs{echoRoute("404")})
	tree.HandleFunc("405", "", eudore.HandlerFuncs{echoRoute("405")})
	runCheck(tree.Match("GET", "/", ctx), ctx)
	runCheck(tree.Match("GET", "/index", ctx), ctx)
	runCheck(tree.Match("11", "/", ctx), ctx)
}

func TestRadixPath2(*testing.T) {
	tree := eudore.NewRouterRadix()
	tree.HandleFunc(newNodeData("/authorizations"))
	tree.HandleFunc(newNodeData("/authorizations/:id"))
	ctx := &Context{}
	tree.HandleFunc("404", "", eudore.HandlerFuncs{echoRoute("404")})
	tree.HandleFunc("405", "", eudore.HandlerFuncs{echoRoute("405")})
	runCheck(tree.Match("GET", "/authorizations", ctx), ctx)
	runCheck(tree.Match("GET", "/authorizations/:id", ctx), ctx)
	runCheck(tree.Match("11", "/", ctx), ctx)
}

func TestFullRouter(*testing.T) {
	tree := eudore.NewRouterFull()
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
	runCheck(tree.Match("GET", "/2", ctx), ctx)
	runCheck(tree.Match("GET", "/22", ctx), ctx)
	runCheck(tree.Match("GET", "/abc", ctx), ctx)
	runCheck(tree.Match("GET", "/abc123", ctx), ctx)
	runCheck(tree.Match("GET", "/1/2", ctx), ctx)
	runCheck(tree.Match("GET", "/1/22", ctx), ctx)
	runCheck(tree.Match("GET", "/1/abc", ctx), ctx)
	runCheck(tree.Match("GET", "/1/abc123", ctx), ctx)
	runCheck(tree.Match("GET", "/2/abc123", ctx), ctx)
	runCheck(tree.Match("GET", "/api/23", ctx), ctx)
	runCheck(tree.Match("GET", "/api/3", ctx), ctx)
}
