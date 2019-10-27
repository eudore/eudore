package main

/*

 */

import (
	"fmt"
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewCore()
	app.AnyFunc("/get", func(ctx eudore.Context) {
		// 遍历请求header
		for k, v := range ctx.Request().Header {
			fmt.Fprintf(ctx, "%s: %s\n", k, v)
		}
		// 获取一个请求header
		ctx.Infof("user-agent: %s", ctx.GetHeader("User-Agent"))
	})
	app.Listen(":8088")
	app.Run()
}
