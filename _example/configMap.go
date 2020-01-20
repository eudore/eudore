package main

/*
默认使用ConfigMap，配置使用map[string]interface{}存储。

如果访问keys.help对象，返回的数据map["keys.help"]，不会像ConfigEudore一样层层选择对象去访问数据。

*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewCore()
	httptest.NewClient(app).Stop(0)
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
	app.Run()
}
