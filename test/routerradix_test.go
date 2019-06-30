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
	tree := &eudore.RouterRadix{}
	tree.InsertRoute(newNodeData("/api/v1/node/"))
	tree.InsertRoute(newNodeData("/api/v1/:list/11"))
	tree.InsertRoute(newNodeData("/api/v1/:list/22"))
	tree.InsertRoute(newNodeData("/api/v1/:list/:name"))
	tree.InsertRoute(newNodeData("/api/v1/:list/*name"))
	tree.InsertRoute(newNodeData("/api/v1/:list version:v1"))
	tree.InsertRoute(newNodeData("/api/v1/* version:v1"))
	tree.InsertRoute(newNodeData("/api/v2/* version:v2"))
	tree.InsertRoute(newNodeData("/api/*"))
	tree.InsertRoute(newNodeData("/note/get/:name"))
	tree.InsertRoute(newNodeData("/note/:method/:name"))
	tree.InsertRoute(newNodeData("/*"))
	tree.InsertRoute(newNodeData("/"))
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
	tree := &eudore.RouterRadix{}
	tree.InsertRoute(newNodeData("/"))
	params := &Params{}
	// tree.Set("404", eudore.HandlerFunc(echoRoute))
	tree.Set("404", echoRoute)
	tree.Set("405", echoRoute)
	tree.Match("GET", "/", params)
	tree.Match("GET", "/index", params)
	tree.Match("11", "/", params)
}

func TestRadixPath2(*testing.T) {
	tree := &eudore.RouterRadix{}
	tree.InsertRoute(newNodeData("/authorizations"))
	tree.InsertRoute(newNodeData("/authorizations/:id"))
	params := &Params{}
	// tree.Set("404", eudore.HandlerFunc(echoRoute))
	tree.Set("404", echoRoute)
	tree.Set("405", echoRoute)
	tree.Match("GET", "/authorizations", params)
	tree.Match("GET", "/authorizations/:id", params)
	tree.Match("11", "/", params)
}

func TestFullRouter(*testing.T) {
	tree := &eudore.RouterFull{}
	tree.InsertRoute(newNodeData("/"))
	tree.InsertRoute(newNodeData("/*"))
	tree.InsertRoute(newNodeData("/:id|^[a-z]*$"))
	tree.InsertRoute(newNodeData("/:id|min:4"))
	tree.InsertRoute(newNodeData("/:id|isnum"))
	tree.InsertRoute(newNodeData("/:id"))
	tree.InsertRoute(newNodeData("/api/*id|^[0-9]$"))
	tree.InsertRoute(newNodeData("/1/*id|min:4"))
	tree.InsertRoute(newNodeData("/1/*id|isnum"))
	tree.InsertRoute(newNodeData("/1/*id"))
	tree.InsertRoute(newNodeData("/*id"))
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
