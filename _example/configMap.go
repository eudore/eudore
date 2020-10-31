package main

/*
默认使用ConfigMap，配置使用map[string]interface{}存储。

如果访问help对象，返回的数据map["help"]，不会像ConfigEudore一样层层选择对象去访问数据。

*/

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewApp()
	app.Set("int", 20)
	app.Set("string", "app set string")
	app.Set("bool", true)
	app.Set("struct", struct {
		Name string
		Age  int
	}{"eudore", 2020})
	app.Debugf("%#v", app.GetInt("int"))
	app.Debugf("%#v", app.GetInt("string"))
	app.Debugf("%#v", app.GetString("string"))
	app.Debugf("%#v", app.GetBool("bool"))
	app.Debugf("%#v", app.Get("struct"))
	app.Debugf("%#v", app.Get("ptr"))
	app.Debugf("%#v", app.Get(""))

	app.Set("", map[string]interface{}{
		"nil data": nil,
	})
	app.Debugf("%#v", app.Get(""))

	app.CancelFunc()
	app.Run()
}
