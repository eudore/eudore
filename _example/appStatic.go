package main

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewApp()
	// 添加静态文件处理
	app.GetFunc("/js/*", NewStaticHandlerWithCache("", "public"))
	// WriteFile 调用http.ServeFile实现，可以额外添加etag计算等逻辑，文件路径拼接需要注意清理。
	app.GetFunc("/css/*", func(ctx eudore.Context) {
		ctx.WriteFile("static" + ctx.Path())
	})
	app.GetFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore")
	})

	app.Listen(":8088")
	app.Run()
}

// NewStaticHandlerWithCache 函数指定NewStaticHandler的缓存策略，默认为no-cache
func NewStaticHandlerWithCache(path, policy string) eudore.HandlerFunc {
	fn := eudore.NewStaticHandler("", path)
	return func(ctx eudore.Context) {
		ctx.SetHeader("Cache-Control", policy)
		fn(ctx)
	}
}
