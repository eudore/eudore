//go:build go1.16

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

//go:embed handler*.go
var f embed.FS

func main() {
	app := eudore.NewApp()
	// 方式1：使用embed.FS扩展，NewHandlerFileEmbed()是embed.FS对象显示调用。
	app.GetFunc("/static/*", f)
	app.GetFunc("/static/*", eudore.NewHandlerFileEmbed(f))
	// 方式2：使用NewHandlerFileSystems函数指定多个embed.FS和资源路径。
	// 参数允许为string iofs.FS http.FileSystem类型
	app.GetFunc("/static/* autoindex=true", eudore.NewHandlerFileSystems(f, "."))
	// 存在参数autoindex=true时，返回目录信息。

	app.Listen(":8088")
	app.Run()
}
