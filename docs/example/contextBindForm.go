package main

/*
bind根据请求中Content-Type Header来决定bind解析数据的方法，常用json和form两种。
*/

import (
	"github.com/eudore/eudore"
	"io"
	"mime/multipart"
	"os"
)

type (
	postFileRequest struct {
		File *multipart.FileHeader `set:"file"`
		// 如果上传多个文件，使用下面一行File定义，同时读取多个表单文件,表达多值一样。
		// File	[]*multipart.FileHeader	`set:"file"`
	}
)

func main() {
	app := eudore.NewCore()
	app.PostFunc("/file/data/:path", postFile)
	app.Listen(":8088")
	app.Run()
}

// 上传文件
func postFile(ctx eudore.Context) {
	// 读取表达文件
	var file postFileRequest
	ctx.Bind(&file)

	// 创建接入文件，没有检查目录存在
	newfile, err := os.Create("/tmp/eudore/" + ctx.GetParam("path"))
	if err != nil {
		ctx.Fatal(err)
		return
	}
	defer newfile.Close()

	// 读取文件
	upfile, err := file.File.Open()
	if err != nil {
		ctx.Fatal(err)
		return
	}
	defer upfile.Close()

	// 文件写入
	_, err = io.Copy(newfile, upfile)
	if err != nil {
		ctx.Fatal(err)
	}
}
