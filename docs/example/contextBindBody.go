package main

/*
bind根据请求中Content-Type Header来决定bind解析数据的方法，常用json和form两种。

例如存在Request Header Content-Type: application/json，Bind就会使用Json解析。

如果请求方法是Get或Head，使用Uri参数绑定。
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
	// 上传文件信息
	app.PutFunc("/file/data/:path", func(ctx eudore.Context) {
		var info putFileInfo
		ctx.Bind(&info)
		ctx.RenderWith(&info, eudore.RenderIndentJSON)
	})

	client := httptest.NewClient(app)
	client.NewRequest("PUT", "/file/data/2").WithHeaderValue("Content-Type", "application/json").WithBodyString(`{"name": "eudore","type": "file", "size":720,"lastModified":1257894000}`).Do().CheckStatus(200).Out()
	for client.Next() {
		app.Error(client.Error())
	}

	app.Listen(":8088")
	app.Run()
}
