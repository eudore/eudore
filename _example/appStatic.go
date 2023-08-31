package main

/*
默认HandlerExtender注册扩展具有NewHandlerEmbed和NewHandlerHTTPFileSystem函数。,
可以传递emebd.FS http.FileSystem类型处理者。

NewFileSystems函数可以将string io/fs.FS(embed.FS) net/http.FileSystem(http.Dir)转换成eudore.HandlerFunc

eudore.NewHandlerStatic(".")
eudore.NewHandlerEmbed(root)
http.Dir(".")
eudore.NewFileSystems(".", root)
*/

import (
	"embed"
	"net/http"

	"github.com/eudore/eudore"
)

//go:embed *.go
var root embed.FS

func main() {
	app := eudore.NewApp()
	// 添加静态文件处理
	app.GetFunc("/src/*", eudore.NewFileSystems(".", root))
	app.GetFunc("/js/*", eudore.NewHandlerStatic("."))
	app.GetFunc("/js/*", http.Dir("."))
	app.GetFunc("/css/*", eudore.NewHandlerEmbed(root))
	app.GetFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore")
	})

	app.Listen(":8088")
	app.Run()
}

// NewStaticHandlerWithCache 函数指定NewStaticHandler的缓存策略，默认为no-cache
func NewStaticHandlerWithCache(path, policy string) eudore.HandlerFunc {
	fn := eudore.NewHandlerStatic(path)
	return func(ctx eudore.Context) {
		ctx.SetHeader("Cache-Control", policy)
		fn(ctx)
	}
}
