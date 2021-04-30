package main

/*
Context.WriteFile处理基于http.ServeFile封装。

NewStaticHandler方法返回文件路径未 参数 + ctx.GetParam("path")或ctx.Path()

func NewStaticHandler(dir string) HandlerFunc {
	if dir == "" {
		dir = "."
	}
	return func(ctx Context) {
		path := ctx.GetParam("path")
		if path == "" {
			path = ctx.Path()
		}
		ctx.WriteFile(filepath.Join(dir, filepath.Clean("/"+path)))
	}
}
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
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

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/").Do().CheckStatus(200)
	client.NewRequest("GET", "/js/index.js").Do().CheckStatus(404)

	// app.CancelFunc()
	app.Run()
}

// NewStaticHandlerWithCache 函数指定NewStaticHandler的缓存策略，默认为no-cache
func NewStaticHandlerWithCache(path, policy string) eudore.HandlerFunc {
	fn := eudore.NewStaticHandler(path)
	return func(ctx eudore.Context) {
		ctx.SetHeader("Cache-Control", policy)
		fn(ctx)
	}
}
