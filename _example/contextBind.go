package main

/*
bind根据请求中Content-Type Header来决定bind解析数据的方法，常用json和form两种。

例如存在Request Header Content-Type: application/json，Bind就会使用Json解析。

如果请求方法是Get或Head，使用Uri参数绑定。
*/

import (
	"encoding/xml"
	"mime/multipart"

	"github.com/eudore/eudore"
)

type fileInfo struct {
	// 可以删除xml字段
	XMLName xml.Name `xml:"Person"`
	Name    string   `json:"name" alias:"name"`
	// 绑定Form file
	File         *multipart.FileHeader `alias:"file"`
	Type         string                `json:"type" alias:"type"`
	Size         int                   `json:"size" alias:"size"`
	LastModified int64                 `json:"lastModified" alias:"lastModified"`
}

type userRequest struct {
	Username string `validate:"regexp:^[a-zA-Z]*$"`
	Name     string `validate:"nozero"`
	Age      int    `validate:"min:21,max:40"`
	Password string `validate:"len:>7"`
}

func main() {
	app := eudore.NewApp()
	// 上传文件信息
	app.AnyFunc("/file/:path", func(ctx eudore.Context) {
		body, err := ctx.Body()
		ctx.Debugf("file body: %s", body)
		var file fileInfo
		err = ctx.Bind(&file)
		if err != nil {
			ctx.Fatal(err)
			return
		}
		ctx.Debugf("file data: %#v", file)
		if file.File != nil {
			ctx.Debugf("file name: %#v size: %d", file.File.Filename, file.File.Size)
		}
	})
	app.AnyFunc("/body", func(ctx eudore.Context) {
		body, err := ctx.Body()
		ctx.Debugf("body: %s", body)
		var data map[string]interface{}
		err = ctx.Bind(&data)
		if err != nil {
			ctx.Fatal(err)
			return
		}
		ctx.Debugf("body data: %#v", data)
	})

	// 没有Content-Type时使用URL参数绑定
	app.NewRequest("PUT", "/file/uri?name=example.go&size=245")
	app.NewRequest("PUT", "/file/json",
		eudore.NewClientBodyJSON(&fileInfo{
			Name:         "eudore",
			Type:         "file",
			Size:         720,
			LastModified: 1257894000,
		}),
	)
	app.NewRequest("PUT", "/file/xml",
		eudore.NewClientBodyXML(&fileInfo{
			Name: "example.go",
			Size: 245,
		}),
	)
	body := eudore.NewClientBodyForm(nil)
	body.AddValue("name", "contextBindForm.go")
	body.AddFile("file", "contextBindForm.go", []byte("contextBindForm file content"))
	app.NewRequest("POST", "/file/contextBindForm.go", body)

	app.NewRequest("PUT", "/body",
		eudore.NewClientBodyJSON(&fileInfo{
			Name:         "eudore",
			Type:         "file",
			Size:         720,
			LastModified: 1257894000,
		}),
	)

	app.Listen(":8088")
	app.Run()
}
