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
	setConfig struct {
		Name       string         `set:"name" json:"name"`
		Num        int            `set:"num" json:"num"`
		Now        time.Time      `set:"now" json:"now"`
		Fields     []gField       `set:"fields" json:"fields"`
		ConfigAuth gConfigAuth    `set:"configauth" json:"configauth"`
		Map        map[int]string `set:"map" json:"map"`
	}
	gField struct {
		Index int    `set:"index" json:"index"`
		Name  string `set:"name" json:"name"`
	}
	gConfigAuth struct {
		Path   string `set:"path" json:"path"`
		Key    string `set:"key" json:"key"`
		Secret string `set:"secret" json:"secret"`
	}
)

func main() {
	data := new(setConfig)
	eudore.Set(data, "name", "name is eudore")
	eudore.Set(data, "num", 99)
	eudore.Set(data, "fields.2.index", 99)
	eudore.Set(data, "fields.3.name", "index3 name")
	eudore.Set(data, "configauth.key", "config key")
	eudore.Set(data, "map.9", "map9 hello")
	fmt.Printf("%#v\n\n", data)
	fmt.Printf("%#v \n", eudore.Get(data, "name"))
	fmt.Printf("%#v \n", eudore.Get(data, "fields.2"))
	fmt.Printf("%#v \n", eudore.Get(data, "configauth"))
	fmt.Printf("%#v \n", eudore.Get(data, "configauth.key"))
	fmt.Printf("%#v \n", eudore.Get(data, "map.0"))
}
