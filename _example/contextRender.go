package main

/*
Render会根据请求的Accept header觉得使用哪种方式写入返回数据，需要api请求时按照标准请求，不会默认返回json。

默认方式是返回字符串(fmt.Fprintf(ctx, "%#v", data))。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewCore()
	httptest.NewClient(app).Stop(0)
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.Render(map[string]interface{}{
			"name":    "eudore",
			"message": "hello eudore",
		})
	})
	app.Listen(":8088")
	app.Run()
}
