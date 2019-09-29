package show

import (
	"fmt"
	"github.com/eudore/eudore"
	"github.com/kr/pretty"
	// "reflect"
	"strings"
)

var objs = make(map[string]interface{})

// RegisterObject 函数注册显示的对象。
func RegisterObject(key string, val interface{}) {
	objs[key] = val
}

// DeleteObject 函数删除可以显示的对象。
func DeleteObject(key string) {
	delete(objs, key)
}

// RoutesInject 函数注入show使用的路由。
func RoutesInject(r eudore.RouterMethod) {
	r = r.Group("/show")
	r.GetFunc("/", List)
	r.GetFunc("/*key", Showkey)
}

// List 函数处理List请求。
func List(ctx eudore.Context) {
	keys := make([]string, 0, len(objs))
	for k := range objs {
		keys = append(keys, k)
	}
	ctx.Render(keys)
}

// Showkey 函数显示key的数据。
func Showkey(ctx eudore.Context) {
	// 获取对象
	key, path := getNameAndPath(ctx.GetParam("key"))
	val := objs[key]
	if val != nil {
		val = eudore.Get(val, path)
	}
	if val == nil {
		ctx.WriteString("not found key: " + ctx.GetParam("key"))
		return
	}

	ctx.SetHeader(eudore.HeaderContentType, "text/plain; charset=utf-8")
	fmt.Fprintf(ctx, "%# v", pretty.Formatter(val))
}

func getNameAndPath(key string) (string, string) {
	keys := strings.Split(key, "/")
	return keys[0], strings.Join(keys[1:], ".")
}
