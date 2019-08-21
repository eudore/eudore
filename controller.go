package eudore

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type (
	// ControllerParseFunc 函数定义解析控制器获得路由配置的方法。
	ControllerParseFunc func(Controller) (*RouterConfig, error)
	// Controller 定义控制器必要的接口。
	Controller interface {
		Init(Context) error
		Release() error
	}
	// ControllerRoute 定义获得路由和方法映射的接口。
	ControllerRoute interface {
		ControllerRoute() map[string]string
	}
	// ControllerBase 实现基本控制器。
	ControllerBase struct {
		Context
	}
	// ControllerData 实现基于ContextData的控制器。
	ControllerData struct {
		ContextData
	}
)

// ControllerBaseParseFunc 定义基本的控制器实现函数，控制器需要实现ControllerRoute接口，提供路由信息。
//
// 控制器函数是首字母大小词为路由方法，如果不是Get、Post、Put、Delete、Patch、Options方法外，则使用Any方法。
//
// GetId() 定义一个基本的控制器方法。
//
// GetIdByIdName(id int, name string)，By后面是使用的Context Param，会使用ctx.GetParam("id")来初始化id的值，name相同。
func ControllerBaseParseFunc(controller Controller) (*RouterConfig, error) {
	iType := reflect.TypeOf(controller)
	pool := &sync.Pool{
		New: func() interface{} {
			return reflect.New(iType.Elem()).Interface()
		},
	}

	// 检查控制器是否实现ControllerRoute接口
	controllerRoute, isRoute := controller.(ControllerRoute)
	if !isRoute {
		return nil, fmt.Errorf("%s not suppert ControllerBaseParseFunc, not ControllerRoute", iType.Name())
	}

	fnname := iType.Elem().PkgPath() + "." + iType.Elem().Name() + "."
	// 生成路由器配置信息。
	var configs = make([]*RouterConfig, 0, len(controllerRoute.ControllerRoute()))
	for name, path := range controllerRoute.ControllerRoute() {
		var method = getFirstUp(name)
		if !checkAllowMethod(method) {
			method = "ANY"
		}

		m, ok := iType.MethodByName(name)
		if !ok {
			continue
		}

		h := convertHandler(pool, controller, m.Index)
		SetHandlerFuncName(h, fnname+name)

		configs = append(configs, &RouterConfig{
			Method:  strings.ToUpper(method),
			Path:    path,
			Handler: HandlerFuncs{h},
		})
	}
	return &RouterConfig{Routes: configs}, nil
}

// convertHandler 实现返回一个HandlerFunc对象，用于执行一个控制器方法。
func convertHandler(pool *sync.Pool, controller Controller, index int) HandlerFunc {
	iType := reflect.TypeOf(controller)
	fType := iType.Method(index)
	// 获取函数参数信息
	var num int = fType.Type.NumIn() - 1
	var args []string = getFuncArgs(fType.Name)
	var typeIn []reflect.Type = make([]reflect.Type, num)
	for i := 0; i < num; i++ {
		typeIn[i] = fType.Type.In(i + 1)
	}
	return func(ctx Context) {
		// 初始化
		controller := pool.Get().(Controller)
		controller.Init(ctx)

		// 函数初始化参数并 调用
		params := make([]reflect.Value, num)
		for i := 0; i < num; i++ {
			params[i] = reflect.New(typeIn[i])
			setWithString(params[i].Kind(), params[i], ctx.GetParam(args[i]))
			params[i] = params[i].Elem()
		}
		reflect.ValueOf(controller).Method(index).Call(params)

		controller.Release()
		pool.Put(controller)
	}
}

func checkAllowMethod(method string) bool {
	for _, i := range []string{"Any", "Get", "Post", "Put", "Delete", "Patch", "Options"} {
		if i == method {
			return true
		}
	}
	return false
}

func splitName(name string) (strs []string) {
	var head int
	for i, c := range name {
		if 0x40 < c && c < 0x5B && i != 0 {
			strs = append(strs, name[head:i])
			head = i
		}
	}
	strs = append(strs, name[head:len(name)])
	return
}

func getFirstUp(name string) string {
	for i, c := range name {
		if 0x40 < c && c < 0x5B && i != 0 {
			return name[:i]
		}
	}
	return name
}

func getFuncArgs(name string) (strs []string) {
	var head int
	var isBy bool
	for i, c := range name {
		if 0x40 < c && c < 0x5B && i != 0 {
			if isBy {
				strs = append(strs, strings.ToLower(name[head:i]))
			}
			if name[head:i] == "By" {
				isBy = true
			}
			head = i
		}
	}
	if isBy {
		strs = append(strs, strings.ToLower(name[head:len(name)]))
	}
	return
}

// Init 实现控制器初始方法。
func (c *ControllerBase) Init(ctx Context) error {
	c.Context = ctx
	return nil
}

// Release 实现控制器释放方法。
func (c *ControllerBase) Release() error {
	return nil
}

// Init 实现控制器初始方法。
func (c *ControllerData) Init(ctx Context) error {
	c.ContextData.Context = ctx
	return nil
}

// Release 实现控制器释放方法。
func (c *ControllerData) Release() error {
	return nil
}
