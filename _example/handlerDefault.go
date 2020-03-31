package main

/*
eudore默认支持多路由扩展和控制器函数扩展。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewCore()
	// eudore Render 返回的数据
	app.GetFunc("/*", func(eudore.Context) interface{} {
		return "hello eudore"
	})
	// 检查返回的err
	app.GetFunc("/err", func(eudore.Context) error {
		return nil
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/file/").Do().Out()
	for client.Next() {
		app.Error(client.Error())
	}
	client.Stop(0)

	app.Run()
}
