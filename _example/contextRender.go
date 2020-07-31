package main

/*
Render会根据请求的Accept header觉得使用哪种方式写入返回数据，需要api请求时按照标准请求，不会默认返回json。

默认方式是返回字符串(fmt.Fprintf(ctx, "%#v", data))。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

type renderData struct {
	Name    string `xml:"name" json:"name"`
	Message string `xml:"message" json:"message"`
}

func main() {
	app := eudore.NewApp()
	app.AnyFunc("/*path", func(ctx eudore.Context) interface{} {
		return renderData{
			Name:    "eudore",
			Message: "hello eudore",
		}
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/1").WithHeaderValue("Accept", "application/json").Do().Out()
	client.NewRequest("GET", "/1").WithHeaderValue("Accept", "application/xml").Do().Out()
	client.NewRequest("GET", "/1").WithHeaderValue("Accept", eudore.MimeTextHTML).Do().Out()
	client.NewRequest("GET", "/1").WithHeaderValue("Accept", eudore.MimeTextPlain).Do().Out()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
