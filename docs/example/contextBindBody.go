package main

/*
bind根据请求中Content-Type Header来决定bind解析数据的方法，常用json和form两种。

例如存在Request Header Content-Type: application/json，Bind就会使用Json解析。

如果请求方法是Get或Head，使用Uri参数绑定。
*/

import (
	"fmt"
	"github.com/eudore/eudore"
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
	app.PutFunc("/file/data/:path", putFile)

	// 启动server
	app.Listen(":8088")
	app.Run()
}

// 上传文件信息
func putFile(ctx eudore.Context) {
	var info putFileInfo
	ctx.Bind(&info)
	fmt.Println(info)
}
