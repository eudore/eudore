package main

/*
Config需要使用指定对象保存配置，默认为map[string]interface{}。

可以自己指定结构体来保存配置，例如example中Config对象指定的user.name就是展开的一层结构体或map后设置，详细查看eudore.Set函数的文档。

config的Get & Set方法使用eudore.GetAnyByPath & eudore.SetAnyByPath方法实现。
*/

import (
	"os"

	"github.com/eudore/eudore"
)

type eudoreConfig struct {
	Config string`alias:"config"`
	Bool   bool        `alias:"bool"`
	Int    int         `alias:"int"`
	String string      `alias:"string"`
	User   user        `alias:"user" flag:"u"`
	Any    interface{} `alias:"any"`
}
type user struct {
	Name string `alias:"name"`
	Mail string `alias:"mail"`
}

func main() {
	os.Args = append(os.Args, "--user.name=eudoreName", "--config=config.json")
	os.Setenv("ENV_USER_MAIL", "post@eudore.cn")

	conf := &eudoreConfig{}
	app := eudore.NewApp()
	// 如果NewConfig参数为nil，使用map[string]any保存。
	app.SetValue(eudore.ContextKeyConfig, eudore.NewConfig(conf))
	app.Parse()

	// 设置属性
	app.Set("int", 20)
	app.Set("string", "app set string")
	app.Set("bool", true)
	app.Set("any", struct {
		Name string
		Age  int
	}{"eudore", 2020})

	// 读取部分配置
	app.Debugf("%#v", app.GetInt("int"))
	app.Debugf("%#v", app.GetString("string"))
	app.Debugf("%#v", app.GetBool("bool"))
	app.Debugf("%#v", app.Get("user"))
	app.Debugf("%#v", app.Get("any"))

	// 输出全部配置信息
	app.Debugf("%#v", conf)
	app.Debugf("%#v", app.Get(""))

	app.CancelFunc()
	app.Run()
}
