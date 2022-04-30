package main

/*
NewLookFunc 函数创建一个访问对象数据处理函数。

获取请求路由参数"*"为object访问路径，返回object指定属性的数据，允许使用下列参数：
	d=10 depth递归显时最大层数
	all=false 是否显时非导出属性
	format=html/json/text 设置数据显示格式
	godoc=https://golang.org 设置html格式链接的godoc服务地址
	width=60 设置html格式缩进宽度
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	config := make(map[interface{}]interface{})
	app := eudore.NewApp()
	app.Logger = app.WithField("key", "look").WithField("logger", true)
	app.Set("conf", config)

	var i interface{}

	config[true] = 1
	config[1] = 11
	config[uint(1)] = 11
	config[1.0] = 11.0
	config[complex(1, 1)] = complex(1, 1)
	config[struct{}{}] = "2"
	config[i] = 0

	app.AnyFunc("/eudore/debug/look/* godoc=/eudore/debug/pprof/godoc", middleware.NewLookFunc(app))

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/eudore/debug/look/?d=3").Do()
	client.NewRequest("GET", "/eudore/debug/look/?all=1").Do()
	client.NewRequest("GET", "/eudore/debug/look/?format=text").Do()
	client.NewRequest("GET", "/eudore/debug/look/?format=json").Do()
	client.NewRequest("GET", "/eudore/debug/look/?format=t2").Do()
	client.NewRequest("GET", "/eudore/debug/look/Config/Keys/2").Do()
	client.NewRequest("GET", "/eudore/debug/look/?d=3").WithHeaderValue(eudore.HeaderAccept, eudore.MimeApplicationJSON).Do()
	client.NewRequest("GET", "/eudore/debug/look/?d=3").WithHeaderValue(eudore.HeaderAccept, eudore.MimeTextHTML).Do()
	client.NewRequest("GET", "/eudore/debug/look/?d=3").WithHeaderValue(eudore.HeaderAccept, eudore.MimeText).Do()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
