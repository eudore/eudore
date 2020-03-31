package main

/*
bind根据请求中Content-Type Header来决定bind解析数据的方法，常用json和form两种。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"io"
	"mime/multipart"
	"os"
)

func main() {
	app := eudore.NewCore()
	app.PostFunc("/file/data/:path", postFile)

	client := httptest.NewClient(app)
	client.NewRequest("POST", "/file/data/content.text").WithBodyFormValue("name", "my name").WithBodyFormFile("file", "contextBindForm.go", "contextBindForm file content").Do()
	for client.Next() {
		app.Error(client.Error())
	}

	app.Run()
}

type (
	postFileRequest struct {
		Name string                `alias:"name" json:"name"`
		File *multipart.FileHeader `alias:"file"`
		// 如果上传多个文件，使用下面一行File定义，同时读取多个表单文件,表达多值一样。
		// File	[]*multipart.FileHeader	`alias:"file"`
	}
)

// 上传文件
func postFile(ctx eudore.Context) (err error) {
	// 读取表达文件
	var file postFileRequest
	ctx.Bind(&file)

	ctx.Debugf("name: %s", ctx.FormValue("name"))
	ctx.Debugf("%#v", file)

	// 读取文件
	upfile, err := file.File.Open()
	if err != nil {
		return err
	}
	defer upfile.Close()

	// 创建接入文件，没有检查目录存在
	newfile, err := os.Create("/tmp/eudore/" + ctx.GetParam("path"))
	if err != nil {
		return err
	}
	defer newfile.Close()

	// 文件写入
	_, err = io.Copy(newfile, upfile)
	return
}
