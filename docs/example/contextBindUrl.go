package main

/*
默认只有Get和Head请求时，使用url参数绑定参数。

BinderURLWithBinder函数可以时其他方法也使用url参数绑定。

当body是使用url编码时，也可以直接绑定数据。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

type (
	putFileInfo struct {
		Name         string `json:"name" set:"name"`
		Type         string `json:"type" set:"type"`
		Size         int    `json:"size" set:"size"`
		LastModified int64  `json:"lastModified" set:"lastModified"`
	}
)

func main() {
	app := eudore.NewCore()
	// 附加非GET和HEAD方法下使用url参数绑定。
	app.Binder = eudore.BinderURLWithBinder(app.Binder)

	app.AnyFunc("/file/data/:path", func(ctx eudore.Context) {
		var info putFileInfo
		ctx.Bind(&info)
		ctx.RenderWith(&info, eudore.RenderIndentJSON)
	})

	client := httptest.NewClient(app)
	// get方法使用url参数绑定
	client.NewRequest("GET", "/file/data/2?name=eudore&type=2&size=722").WithHeaderValue("Content-Type", "application/x-www-form-urlencoded").Do().CheckStatus(200).Out()
	// put方法使用url参数绑定，需要BinderURLWithBinder函数支持。
	client.NewRequest("PUT", "/file/data/2?name=eudore&type=2&size=722").WithHeaderValue("Content-Type", "application/x-www-form-urlencoded").Do().CheckStatus(200).Out()
	// url body绑定
	client.NewRequest("PUT", "/file/data/2").WithBodyString("name=eudore&type=2&size=722").WithHeaderValue("Content-Type", "application/x-www-form-urlencoded").Do().CheckStatus(200).Out()
	for client.Next() {
		app.Error(client.Error())
	}

	app.Listen(":8088")
	app.Run()
}
