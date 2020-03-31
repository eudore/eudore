package main

/*
Cookie相关方法定义。
type (
	Context interface {
		Cookies() []Cookie
		GetCookie(name string) string
		SetCookie(cookie *SetCookie)
		SetCookieValue(string, string, int)
		...
	}

	// SetCookie 定义响应返回的set-cookie header的数据生成
	SetCookie = http.Cookie
	// Cookie 定义请求读取的cookie header的键值对数据存储
	Cookie struct {
		Name  string
		Value string
	}
)
*/

import (
	"fmt"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewCore()
	httptest.NewClient(app).Stop(0)
	app.AnyFunc("/set", func(ctx eudore.Context) {
		ctx.SetCookie(&eudore.SetCookie{
			Name:     "set1",
			Value:    "val1",
			Path:     "/",
			HttpOnly: true,
		})
		ctx.SetCookieValue("name", "eudore", 600)
	})
	app.AnyFunc("/get", func(ctx eudore.Context) {
		ctx.Infof("cookie name value is: %s", ctx.GetCookie("name"))
		for _, i := range ctx.Cookies() {
			fmt.Fprintf(ctx, "%s: %s\n", i.Name, i.Value)
		}
	})

	client := httptest.NewClient(app)
	client.WithHeaderValue("Cookie", "age=22; name=eudore; =00; tag=\007hs; aa=\"bb\"; ")
	client.NewRequest("PUT", "/get").Do().CheckStatus(200).Out()
	client.NewRequest("PUT", "/set").Do().CheckStatus(200).Out()
	for client.Next() {
		app.Error(client.Error())
	}

	app.Run()
}
