package main

import (
	"fmt"

	"github.com/eudore/eudore"
)

/*
type Context interface {
	// param query header cookie form
	Params() *Params
	GetParam(string) string
	SetParam(string, string)
	Querys() url.Values
	GetQuery(string) string
	GetHeader(string) string
	SetHeader(string, string)
	Cookies() []Cookie
	GetCookie(string) string
	SetCookie(*CookieSet)
	SetCookieValue(string, string, int)
	FormValue(string) string
	FormValues() map[string][]string
	FormFile(string) *multipart.FileHeader
	FormFiles() map[string][]*multipart.FileHeader
	...
}

// SetCookie 定义响应返回的set-cookie header的数据生成
type CookieSet = http.Cookie
// Cookie 定义请求读取的cookie header的键值对数据存储
type Cookie struct {
	Name  string
	Value string
}
*/

func main() {
	app := eudore.NewApp()
	// -------------------- Querys --------------------
	// Querys 读取http路径参数，不会读取Body。
	// 例如 http://localhost:8088/querys?name=eudore&year=2024
	app.AnyFunc("/querys", func(ctx eudore.Context) {
		vals, _ := ctx.Querys()
		ctx.Debugf("querys: %v", vals)
		ctx.Debugf("name: %s", ctx.GetQuery("name"))
	})

	// -------------------- Header --------------------
	// ctx.GetHeader(key) => ctx.Request().Header.Get(key)
	// ctx.SetHeader(key, val) = > ctx.Response().Header().Set(key, val)
	app.AnyFunc("/get", func(ctx eudore.Context) {
		// 获取一个请求header
		ua := ctx.GetHeader("User-Agent")
		ctx.SetHeader("Name", "eudore")
		ctx.Infof("user-agent: %s", ua)
		ctx.WriteString(ua)
	})

	// -------------------- Params --------------------
	// 路由器设置当前l路由组内置参数。
	app.Params().Set("router", "std")
	// 注册路径设置默认参数。
	app.AnyFunc("/params/* version=v1", func(ctx eudore.Context) {
		// 注册路由参数为默认值，请求上下文可修改当前参数。
		ctx.SetParam("name", "eudore")
		// 从参数获取路由匹配模式
		ctx.Debug("route:", ctx.GetParam("route"))
		ctx.Render(ctx.Params())
	})

	// -------------------- Cookie --------------------
	app.AnyFunc("/set", func(ctx eudore.Context) {
		ctx.SetCookie(&eudore.CookieSet{
			Name:     "set1",
			Value:    "val1",
			Path:     "/",
			HttpOnly: true,
		})
		ctx.SetCookieValue("set", "eudore", 0)
		ctx.SetCookieValue("name", "eudore", 600)
	})
	app.AnyFunc("/get", func(ctx eudore.Context) {
		ctx.Infof("cookie name value is: %s", ctx.GetCookie("name"))
		for _, i := range ctx.Cookies() {
			fmt.Fprintf(ctx, "%s: %s\n", i.Name, i.Value)
		}
	})

	// -------------------- Form Data --------------------
	/*
	   可以使用ctx.Request().Form的标准库方法，eudore与nethttp form解析逻辑不同。
	   如果body是http.NoBody时解析url参数，使用FormValue和GetQuery获取数据。
	   如果body是url格式，使用FormValue获取数据。
	   如果body是form格式，使用FormValue获取数据
	   对于form上传文件只能使用FormFile方法获得文件。
	*/
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.Debug("Content-Type", ctx.GetHeader(eudore.HeaderContentType))
		ctx.FormValue("haha")
		ctx.FormValue("name")
		ctx.FormFile("haha")
		ctx.FormFile("file")
		vals, _ := ctx.FormValues()
		for key, val := range vals {
			ctx.Debug("get from value:", key, val)
		}

		for key, file := range ctx.FormFiles() {
			ctx.Debug("get from file:", key, file[0].Filename)
		}
	})

	// -------------------- Redirect --------------------
	app.AnyFunc("/redirect/*", func(ctx eudore.Context) {
		ctx.Redirect(308, "hello")
	})
	app.GetFunc("/hello", func(ctx eudore.Context) {
		// 设置状态码和返回字符串
		ctx.WriteHeader(eudore.StatusOK)
		ctx.WriteString("hello eudore")
	})

	// -------------------- Request Info --------------------
	app.AnyFunc("/* version=v1", func(ctx eudore.Context) {
		ctx.WriteString("host: " + ctx.Host() + "\n")
		ctx.WriteString("method: " + ctx.Method() + "\n")
		ctx.WriteString("path: " + ctx.Path() + "\n")
		ctx.WriteString("params: " + ctx.Params().String() + "\n")
		ctx.WriteString("real ip: " + ctx.RealIP() + "\n")
		body, _ := ctx.Body()
		if len(body) > 0 {
			ctx.WriteString("body: " + string(body))
		}
	})

	app.Listen(":8088")
	app.Run()
}
