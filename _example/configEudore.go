package main

/*
ConfigEudore需要使用指定对象保存配置，默认为map[string]interface{}。

可以自己指定结构体来保存配置，例如example中Config对象指定的user.name就是展开的一层结构体或map后设置，详细查看eudore.Set函数的文档。

config的Get & Set方法使用eudore.Get & eudore.Set方法实现。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

type (
	eudoreConfig struct {
		Bool   bool        `alias:"bool"`
		Int    int         `alias:"int"`
		String string      `alias:"string"`
		User   user        `alias:"user"`
		Struct interface{} `alias:"struct"`
	}
	user struct {
		Name string `alias:"name"`
		Mail string `alias:"mail"`
	}
)

func main() {
	conf := &eudoreConfig{}
	app := eudore.NewCore()
	httptest.NewClient(app).Stop(0)
	// 设置config
	app.Config = eudore.NewConfigEudore(conf)
	eudore.Set(app.Config, "print", eudore.NewPrintFunc(app.App))

	// 设属性
	app.Set("int", 20)
	app.Set("string", "app set string")
	app.Set("bool", true)
	app.Set("user.name", "EudoreName")
	app.Set("struct", struct {
		Name string
		Age  int
	}{"eudore", 2020})
	app.Set("field", "not found")

	// 读取部分配置
	app.Debugf("%#v", app.GetInt("int"))
	app.Debugf("%#v", app.GetInt("string"))
	app.Debugf("%#v", app.GetString("string"))
	app.Debugf("%#v", app.GetBool("bool"))
	app.Debugf("%#v", app.Get("struct"))
	app.Debugf("%#v", app.Get("field"))

	// 输出全部配置信息
	app.Debugf("%#v", conf)
	app.Debugf("%#v", app.Get(""))
	app.Run()
}
