package test

import (
	"testing"

	"fmt"
	"github.com/eudore/eudore"
	// "github.com/kr/pretty"
)

type Params struct{}

func (*Params) GetParam(string) string {
	return ""
}

func (*Params) AddParam(key string, val string) {
	fmt.Println("Add", key, val)
}
func (*Params) SetParam(key string, val string) {
	fmt.Println("Set", key, val)
}

func newNodeData(path string) (string, string, eudore.HandlerFuncs) {
	return "GET", path, eudore.HandlerFuncs{echoRoute}
}

func echoRoute(ctx eudore.Context) {
	fmt.Println(ctx.GetParam("route"))
}

func TestRadixRouter(*testing.T) {
	tree := eudore.NewRouterRadix()
	tree.RegisterHandler(newNodeData("/api/v1/node/"))
	tree.RegisterHandler(newNodeData("/api/v1/:list/11"))
	tree.RegisterHandler(newNodeData("/api/v1/:list/22"))
	tree.RegisterHandler(newNodeData("/api/v1/:list/:name"))
	tree.RegisterHandler(newNodeData("/api/v1/:list/*name"))
	tree.RegisterHandler(newNodeData("/api/v1/:list version:v1"))
	tree.RegisterHandler(newNodeData("/api/v1/* version:v1"))
	tree.RegisterHandler(newNodeData("/api/v2/* version:v2"))
	tree.RegisterHandler(newNodeData("/api/*"))
	tree.RegisterHandler(newNodeData("/note/get/:name"))
	tree.RegisterHandler(newNodeData("/note/:method/:name"))
	tree.RegisterHandler(newNodeData("/*"))
	tree.RegisterHandler(newNodeData("/"))
	// fmt.Printf("%# v\n", pretty.Formatter(tree))
	// t.Log(getSpiltPath("/api/v1/:list/*"))
	params := &Params{}
	tree.Match("GET", "/api/v1/node/11", params)
	tree.Match("GET", "/api/v1/node/111", params)
	tree.Match("GET", "/api/v1/node/111/111", params)
	tree.Match("GET", "/api/v1/list", params)
	tree.Match("GET", "/api/v1/list33", params)
	tree.Match("GET", "/api/v1/get", params)
	tree.Match("GET", "/api/v2/get", params)
	tree.Match("GET", "/api/v3/111", params)
	tree.Match("GET", "/note/get/eudore/2", params)
	tree.Match("GET", "/note/get/eudore", params)
	tree.Match("GET", "/note/set/eudore", params)
	tree.Match("GET", "/node", params)
}

func TestRadixPath1(*testing.T) {
	tree := eudore.NewRouterRadix()
	tree.RegisterHandler(newNodeData("/"))
	params := &Params{}
	// tree.Set("404", eudore.HandlerFunc(echoRoute))
	tree.RegisterHandler("404", "", eudore.HandlerFuncs{echoRoute})
	tree.RegisterHandler("405", "", eudore.HandlerFuncs{echoRoute})
	tree.Match("GET", "/", params)
	tree.Match("GET", "/index", params)
	tree.Match("11", "/", params)
}

func TestRadixPath2(*testing.T) {
	tree := eudore.NewRouterRadix()
	tree.RegisterHandler(newNodeData("/authorizations"))
	tree.RegisterHandler(newNodeData("/authorizations/:id"))
	params := &Params{}
	// tree.Set("404", eudore.HandlerFunc(echoRoute))
	tree.RegisterHandler("404", "", eudore.HandlerFuncs{echoRoute})
	tree.RegisterHandler("405", "", eudore.HandlerFuncs{echoRoute})
	tree.Match("GET", "/authorizations", params)
	tree.Match("GET", "/authorizations/:id", params)
	tree.Match("11", "/", params)
}

func TestFullRouter(*testing.T) {
	tree := eudore.NewRouterFull()
	tree.RegisterHandler(newNodeData("/"))
	tree.RegisterHandler(newNodeData("/*"))
	tree.RegisterHandler(newNodeData("/:id|^[a-z]*$"))
	tree.RegisterHandler(newNodeData("/:id|min:4"))
	tree.RegisterHandler(newNodeData("/:id|isnum"))
	tree.RegisterHandler(newNodeData("/:id"))
	tree.RegisterHandler(newNodeData("/api/*id|^[0-9]$"))
	tree.RegisterHandler(newNodeData("/1/*id|min:4"))
	tree.RegisterHandler(newNodeData("/1/*id|isnum"))
	tree.RegisterHandler(newNodeData("/1/*id"))
	tree.RegisterHandler(newNodeData("/*id"))
	params := &Params{}
	tree.Match("GET", "/2", params)
	tree.Match("GET", "/22", params)
	tree.Match("GET", "/abc", params)
	tree.Match("GET", "/abc123", params)
	tree.Match("GET", "/1/2", params)
	tree.Match("GET", "/1/22", params)
	tree.Match("GET", "/1/abc", params)
	tree.Match("GET", "/1/abc123", params)
	tree.Match("GET", "/2/abc123", params)
	tree.Match("GET", "/api/23", params)
	tree.Match("GET", "/api/3", params)
}
