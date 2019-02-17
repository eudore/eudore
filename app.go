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
		Config
		Server
		Logger
		Router
		Cache
		Binder
		Renderer
		View
		// pools存储各种Context、、构造函数，用于sync.pool Get一个新对象。
		Pools map[string]PoolGetFunc
	}
)

func NewApp() *App {
	return &App{
		Pools:	make(map[string]PoolGetFunc),
	}
}

// Set the program sync.Pool create function
//
// 设置程序sync.Pool创建函数
func (app *App) RegisterPoolFunc(name string, fn PoolGetFunc) {
	app.Pools[name] = fn
}


// Load components in bulk, using null values.
//
// 批量加载组件，使用的参数都是空值。
func (app *App) RegisterComponents(names []string, args []interface{}) error {
	errs := NewErrors()
	for i, name := range names {
		errs.HandleError(app.RegisterComponent(name, args[i]))
	}
	return errs.GetError()
}

// Load a component and assign it to app.
//
// 加载一个组件，并赋值给app。
func (app *App) RegisterComponent(name string,  arg interface{}) error {
	c, err := NewComponent(name, arg)
	if err != nil {
		app.Error(err)
		return err
	}
	switch c.(type) {
	case Router:
		app.Router = c.(Router)
	case Logger:
		li, ok := app.Logger.(LoggerInitHandler)
		app.Logger = c.(Logger)
		if ok {
			li.NextHandler(app.Logger)
		}
	case Server:
		app.Server = c.(Server)
	case Cache:
		app.Cache = c.(Cache)
	case Config:
		app.Config = c.(Config)
	default:
		err := fmt.Errorf("undefined component: %s", name)
		app.Error(err)
		return err
	}
	return nil
}
