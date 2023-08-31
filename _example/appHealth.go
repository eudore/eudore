package main

/*
HandlerMetadata返回App全部All的Metadata() any方法数据，
Metadata前两个字段为Health和Name。

type Metadata struct {
    Health bool   `alias:"health" json:"health" xml:"health" yaml:"health"`
    Name   string `alias:"name" json:"name" xml:"name" yaml:"name"`
}
*/
import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyLogger, eudore.NewLogger(&eudore.LoggerConfig{
		Stdout:   true,
		StdColor: true,
		HookMeta: true,
	}))
	app.SetValue(eudore.ContextKeyHandlerExtender, eudore.NewHandlerExtender())
	app.SetValue(eudore.ContextKeyFuncCreator, eudore.NewFuncCreator())
	app.SetValue(eudore.ContextKeyRender, eudore.RenderJSON)
	app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))
	app.AddMiddleware(middleware.NewLoggerFunc(app))
	app.AddMiddleware(middleware.NewRecoverFunc())
	app.GetFunc("/health", eudore.HandlerMetadata)
	app.GetFunc("/panic", func(ctx eudore.Context) {
		panic(ctx)
	})

	app.Listen(":8087")
	app.Run()
}
