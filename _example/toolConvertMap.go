package main

/*
ConvertMap和ConvertMapString的区别的转换成map[interface{}]interface{}和map[string]interface{}。

暂时未处理对象循环引用对象。
*/

import (
	"fmt"
	"github.com/eudore/eudore"
	"time"
)

type (
	configMap struct {
		Name       string
		Num        int
		Now        time.Time
		Fields     []mField
		ConfigAuth mConfigAuth
		Map        map[int]string
	}
	mField struct {
		Index int
		Name  string
	}
	mConfigAuth struct {
		Path   string
		Key    string
		Secret string
	}
)

func main() {
	data := &configMap{
		Name: "eudore",
		Num:  22,
		Now:  time.Now(),
		Fields: []mField{
			{1, "a"},
			{2, "b"},
			{3, "c"},
		},
		ConfigAuth: mConfigAuth{
			Path:   "/tmp",
			Key:    "876926",
			Secret: "uwvdjqbwi",
		},
		Map: map[int]string{
			1: "A",
			2: "B",
			3: "C",
		},
	}
	fmt.Printf("%#v\n\n", eudore.ConvertMap(data))
	target := eudore.ConvertMapString(data)
	fmt.Printf("%#v\n", target)
}
