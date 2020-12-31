package main

/*
App是一个自定义的程序主体，可以额外组合需要的App对象和方法。

例如定义一个Config结构体对象，可以使用app.Config.Name直接获得配置属性，也可以使用app.App.Get("name") app.App.ConfigGet("name")使用路径访问熟悉。

或者组合一个*sql.DB，直接使用App的数据库连接池，避免使用全局对象。

其他一些组合根据实际情况组合。
*/

import (
	"database/sql"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

// App 定义一个简单app
type App struct {
	*eudore.App
	*Config
	*sql.DB
}

type Config struct {
	Name string `alias:"name" json:"name"`
}

func main() {
	app := NewApp()
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("Hello, 世界")
	})
	app.GetFunc("/user",NewUserHandlr(app))
	app.Info("hello")

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/").Do().CheckStatus(200).OutBody()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

// NewApp 方法创建一个自定义app
func NewApp() *App {
 	conf := &Config{Name: "eudore"}
	return &App{
		App: eudore.NewApp(eudore.NewConfigEudore(conf)),
		Config: conf,
	}
}

// NewUserHandlr 方法闭包传递app对象，然后使用数据库进行操作。
func NewUserHandlr(app *App) eudore.HandlerFunc {
return func(ctx eudore.Context) {
	// app.QueryRow()
   // ...
}
}
