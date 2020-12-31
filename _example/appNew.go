package main

/*
eudore.App对象的简单组装各类对象，实现Options、Listen和Run方法。

Options 方法加载app组件，option类型为context.Context、Logger、Config、Server、Router、Binder、Renderer、Validater时会设置app属性，并设置组件的print属性，如果类型为error将作为app结束错误返回给Run方法。

type App struct {
	context.Context    `alias:"context"`
	context.CancelFunc `alias:"cancelfunc"`
	Config             `alias:"config"`
	Logger             `alias:"logger"`
	Server             `alias:"server"`
	Router             `alias:"router"`
	Binder             `alias:"binder"`
	Renderer           `alias:"renderer"`
	Validater          `alias:"validater"`
	GetWarp            `alias:"getwarp"`
	HandlerFuncs       `alias:"handlerfuncs"`
	ContextPool        sync.Pool `alias:"contextpool"`
	CancelError        error     `alias:"cancelerror"`
	cancelMutex        sync.Mutex
}
*/

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewApp()
	app.Set("workdir", ".")
	app.Options(app.Parse())
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore")
	})
	app.AnyFunc("/data", func(ctx eudore.Context) interface{} {
		// 返回interface{}并直接Render
		return map[string]interface{}{
			"aa": 11,
			"bb": 22,
		}
	})
	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
