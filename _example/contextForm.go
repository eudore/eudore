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
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewCore()
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.FormValue("haha")
		ctx.FormValue("name")
		ctx.FormFile("haha")
		ctx.FormFile("file")
		for key, val := range ctx.FormValues() {
			fmt.Println(key, val)
		}

		for key, file := range ctx.FormFiles() {
			fmt.Println(key, file[0].Filename)
		}
	})

	client := httptest.NewClient(app)
	client.NewRequest("POST", "/").WithBodyFormValue("name", "my name", "message", "msg").WithBodyFormFile("file", "contextBindForm.go", "contextBindForm file content").Do()
	client.NewRequest("POST", "/").WithBodyJSONValue("name", "my name", "message", "msg").Do()

	for client.Next() {
		app.Error(client.Error())
	}

	app.Run()
}
