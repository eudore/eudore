package main

/*
Core是对eudore.App对象的简单封装，实现不到百行。
*/

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewCore()
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("host: " + ctx.Host())
		ctx.WriteString("\nmethod: " + ctx.Method())
		ctx.WriteString("\npath: " + ctx.Path())
		ctx.WriteString("\nreal ip: " + ctx.RealIP())
		ctx.WriteString("\nreferer: " + ctx.Referer())
		ctx.WriteString("\ncontext type: " + ctx.ContentType())
		body := ctx.Body()
		if len(body) > 0 {
			ctx.WriteString("\nbody: " + string(body))
		}
	})
	app.Listen(":8088")
	app.Run()
}
