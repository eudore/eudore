package main

/*
Form相关方法定义。
type Context interface {
	FormValue(string) string
	FormValues() map[string][]string
	FormFile(string) *multipart.FileHeader
	FormFiles() map[string][]*multipart.FileHeader
	...
}
*/

import (
	"fmt"
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewCore()
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		for key, val := range ctx.FormValues() {
			fmt.Println(key, val)
		}

		for key, file := range ctx.FormFiles() {
			fmt.Println(key, file[0].Filename)
		}
	})
	app.Listen(":8088")
	app.Run()
}
