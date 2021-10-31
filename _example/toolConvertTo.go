package main

/*
ConvertTo和区别在于ConvertMap，ConvertTo会把数据转换到目标对象中，而ConvertMap会统一递归转换成map。
*/

import (
	"fmt"
	"github.com/eudore/eudore"
	"time"
)

type (
	configTo struct {
		Name       string
		Num        int
		Now        time.Time
		Fields     []tField
		ConfigAuth *tConfigAuth
		Map        map[int]string
	}
	tField struct {
		Index int
		Name  string
	}
	tConfigAuth struct {
		Path   string
		Key    string
		Secret string
	}
)

func main() {
	src := &configTo{
		Name: "eudore",
		Num:  22,
		Now:  time.Now(),
		Fields: []tField{
			{1, "a"},
			{2, "b"},
			{3, "c"},
		},
		ConfigAuth: &tConfigAuth{
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
	// 结构体转换空map
	var data map[string]interface{}
	err := eudore.ConvertTo(src, &data)
	fmt.Printf("%#v %v\n\n", data, err)

	// map转换结构体
	tar2 := &configTo{
		ConfigAuth: &tConfigAuth{},
	}
	err = eudore.ConvertTo(&data, tar2)
	fmt.Printf("%#v %v\n", tar2, err)
}
