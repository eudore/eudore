//go:build go1.16
// +build go1.16

package main

/*
允许指定文件目录路径，如果路径文件存在使用文件系统，否则使用embed文件系统。
开发时指定文件目录从文件系统加载资源。
部署时不指定文件目录从embed加载，或者允许使用文件系统覆盖embed文件。
*/

import (
	"embed"
	"github.com/eudore/eudore"
)

//go:embed *.go
var f embed.FS

func main() {
	app := eudore.NewApp()
	// 方式1：使用embed.FS扩展，路由参数dir指定多个允许存在的目录位置,分隔符';'。
	app.GetFunc("/static/* dir=.", f)
	// 方式2：使用NewHandlerEmbedFunc函数显示指定embed.FS和存在路径。
	app.GetFunc("/static/*", eudore.NewHandlerEmbedFunc(f, "."))

	app.Listen(":8088")
	app.Run()
}
