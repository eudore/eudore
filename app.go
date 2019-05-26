/*
Application

定义基本的Application对象，实际使用可能需要组合Application对象生成新的实例对象，例如Core、Eudore。

文件：app.go core.go eudore.go
*/
package eudore


import (
	"fmt"
)

type (
	// sync.Pool对象使用的构造函数。
	PoolGetFunc func() interface{}
	// The App combines the main functional interfaces, and the instantiation operations such as startup require additional packaging.
	//
	// App组合主要功能接口，启动等实例化操作需要额外封装。
	App struct {
		Config			`set:"config"`
		Logger			`set:"logger"`
		Server			`set:"server"`
		Router			`set:"router"`
		Cache
		Session
		Client
		View
		Binder
		Renderer
	}
)

func NewApp() *App {
	return &App{
		Binder:	BinderDefault,
	}
}

// Load components in bulk, using null values.
//
// 批量加载组件，使用的参数都是空值。
func (app *App) RegisterComponents(names []string, args []interface{}) error {
	var err error
	errs := NewErrors()
	for i, name := range names {
		_, err = app.RegisterComponent(name, args[i])
		errs.HandleError(err)
	}
	return errs.GetError()
}

// Load a component and assign it to app.
//
// 加载一个组件，并赋值给app。
func (app *App) RegisterComponent(name string,  arg interface{}) (Component, error) {
	c, err := NewComponent(name, arg)
	if err != nil {
		app.Error(err)
		return nil, err
	}
	switch c.(type) {
	case Config:
		app.Config = c.(Config)
	case Logger:
		li, ok := app.Logger.(LoggerInitHandler)
		app.Logger = c.(Logger)
		if ok {
			li.NextHandler(app.Logger)
		}
	case Server:
		app.Server = c.(Server)
	case Router:
		app.Router = c.(Router)
	case Cache:
		app.Cache = c.(Cache)
	case Session:
		app.Session = c.(Session)
	case View:
		app.View = c.(View)
	default:
		err := fmt.Errorf("app undefined component: %s", name)
		app.Error(err)
		return nil, err
	}
	return c, nil
}


func (app *App) GetAllComponent() ([]string, []Component) {
	var names []string = []string{"config", "logger", "server", "router", "cache", "session", "view"}
	return names, []Component{
		app.Config,
		app.Logger,
		app.Server,
		app.Router,
		app.Cache,
		app.Session,
		// app.Client,
		app.View,
	}
}


func (app *App) InitComponent() error {
	names, components := app.GetAllComponent()
	for i, name := range names {
		if components[i] == nil {
			_, err := app.RegisterComponent(name, nil)
			if err != nil {
				return err
			}
		}
	}
	return nil
}