package eudore

/*
Application

定义基本的Application对象，实际使用可能需要组合Application对象生成新的实例对象，例如Core、Eudore。

文件：app.go core.go eudore.go
*/

import (
	"context"
	"sync"
)

type (
	// PoolGetFunc 定义sync.Pool对象使用的构造函数。
	PoolGetFunc func() interface{}
	// The App combines the main functional interfaces, and the instantiation operations such as startup require additional packaging.
	//
	// App 组合主要功能接口，启动等实例化操作需要额外封装。
	//
	// App初始化顺序请按照，Logger-Init、Config、Logger、...
	App struct {
		context.Context
		Config `set:"config"`
		Logger `set:"logger"`
		Server `set:"server"`
		Router `set:"router"`
		Binder
		Renderer
		ContextPool sync.Pool
	}
)

// NewApp 函数创建一个App对象。
func NewApp() *App {
	app := &App{
		Context:  context.Background(),
		Config:   NewConfigMap(nil),
		Logger:   NewLoggerInit(),
		Server:   NewServerStd(nil),
		Router:   NewRouterRadix(),
		Binder:   BinderDefault,
		Renderer: RenderDefault,
	}
	app.ContextPool.New = func() interface{} {
		return NewContextBase(app)
	}
	Set(app.Router, "print", NewLoggerPrintFunc(app.Logger))
	Set(app.Server, "print", NewLoggerPrintFunc(app.Logger))
	return app
}
