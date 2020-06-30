package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

// App 定义一个简单app
type App struct {
	*eudore.App
	Config
}

type Config struct {
	Name string
}

func main() {
	app := NewApp()
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("Hello, 世界")
	})
	app.Info("hello")

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/").Do().CheckStatus(200).OutBody()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

// NewApp 方法创建一个自定义app
func NewApp() *App {
	return &App{
		App: eudore.NewApp(),
		Config: Config{
			Name: "eudore",
		},
	}
}
