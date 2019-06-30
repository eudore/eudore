package show

import (
	"github.com/eudore/eudore"
	"reflect"
	"strings"

	"fmt"
	"github.com/kr/pretty"
)

var objs map[string]interface{} = make(map[string]interface{})

func RegisterObject(key string, val interface{}) {
	objs[key] = val
}

func DeleteObject(key string) {
	delete(objs, key)
}

func Inject(r eudore.RouterMethod) {
	r = r.Group("/show")
	r.GetFunc("/", List)
	r.GetFunc("/*key", Showkey)
}

func List(ctx eudore.Context) {
	keys := make([]string, 0, len(objs))
	for k, _ := range objs {
		keys = append(keys, k)
	}
	ctx.WriteRender(keys)
}

func getVal(key1, key2 string) interface{} {
	val, ok := objs[key1]
	if ok {
		if key2 != "" {
			return eudore.Get(val, strings.Replace(key2[1:], "/", ".", -1))
		}
		return val
	}

	index := strings.LastIndexByte(key1, '/')
	if index == -1 {
		return nil
	}

	key1 += key2
	return getVal(key1[0:index], key1[index:len(key1)])
}

func Showkey(ctx eudore.Context) {
	val := getVal(ctx.GetParam("key"), "")
	if val == nil {
		ctx.WriteString("not found key: " + ctx.GetParam("key"))
		return
	}

	var length int = reflect.TypeOf(val).Elem().NumField()
	fields := make(map[string]interface{}, length)
	pt := reflect.TypeOf(val).Elem()
	pv := reflect.ValueOf(val).Elem()
	for i := 0; i < length; i++ {
		if pv.Field(i).CanInterface() {
			fields[pt.Field(i).Name] = pv.Field(i).Interface()
		}
	}
	ctx.SetHeader(eudore.HeaderContentType, "text/plain; charset=utf-8")
	// ctx.WriteRender(fields)
	// ctx.WriteJson(fields)
	fmt.Fprintf(ctx, "%# v", pretty.Formatter(val))
}
