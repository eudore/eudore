package showmem

import (
	"fmt"
	"reflect"
	"github.com/eudore/eudore"

	//  "github.com/kr/pretty"
)

var objs map[string]interface{}

func init() {
	objs = make(map[string]interface{})
}

func RegisterObject(key string, val interface{}) {
	objs[key] = val
}

func DeleteObject(key string) {
	delete(objs, key)
}

func ListMem(ctx eudore.Context) {
	keys := make([]string, 0, len(objs))
	for k, _ := range objs {
		keys = append(keys, k)
	}
	ctx.WriteRender(keys)
}

func ShowMem(ctx eudore.Context) {
	val := objs[ctx.GetParam("name")]
	fields := make(map[string]string)

	fmt.Println(reflect.TypeOf(val))
	fmt.Println(reflect.TypeOf(val).Elem().NumField())
	var len int = reflect.TypeOf(val).Elem().NumField()
	pt := reflect.TypeOf(val).Elem()
	pv := reflect.Indirect(reflect.ValueOf(val))
	for i := 0; i< len; i++ {
		ctx.Debug(i, pt.Field(i).Name, pv.Field(i))
		fields[pt.Field(i).Name] = fmt.Sprint(pv.Field(i).Interface())
	}
	ctx.Response().Header().Add(eudore.HeaderContentType, "text/plain; charset=utf-8")
	eudore.RendererIndentJson.Render(ctx.Response(), fields)
	// ctx.WriteJson(fields)
	
	// fmt.Fprintf(ctx, "%# v", pretty.Formatter(val))
}
