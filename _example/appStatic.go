package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewApp()
	// 添加静态文件处理
	app.GetFunc("/js/*", eudore.NewStaticHandler(""))
	// WriteFile 调用http.ServeFile实现，可以额外添加etag计算等逻辑，文件路径拼接需要注意清理。
	app.GetFunc("/css/*", func(ctx eudore.Context) {
		ctx.WriteFile("static" + ctx.Path())
	})
	app.GetFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore")
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/").Do().CheckStatus(200).Out()
	client.NewRequest("GET", "/js/index.js").Do().CheckStatus(200).Out()

	app.CancelFunc()
	app.Run()
}
