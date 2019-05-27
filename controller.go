package eudore

import (
	"fmt"
	"sync"
	"reflect"
	"strings"
)

type (
	ControllerParseFunc func(Controller) (*RouterConfig, error)
	Controller interface{
		Init(Context) error
		Release() error
	}
	ControllerRoute interface {
		ControllerRoute() map[string]string
	}
	ControllerBase struct{
		Context
	}
	ControllerData struct{
		ContextData
	}
)

func ControllerBaseParseFunc(controller Controller) (*RouterConfig, error) {	
	iType := reflect.TypeOf(controller)
	pool := sync.Pool{
		New: func() interface{} {
			return reflect.New(iType.Elem()).Interface()
		},
	}

	// 检查控制器是否实现ControllerRoute接口
	controllerRoute, isRoute := controller.(ControllerRoute)
	if !isRoute {
		return nil, fmt.Errorf("%s not suppert ControllerBaseParseFunc, not ControllerRoute", iType.Name())
	}

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

		configs = append(configs, &RouterConfig{
			Method:	strings.ToUpper(method),
			Path:	path,
			Handler:	convertHandler(pool, controller, m.Index),
		})
	}
	return &RouterConfig{Routes:	configs}, nil
}

func convertHandler(pool sync.Pool, controller Controller, index int) HandlerFunc {
	iType := reflect.TypeOf(controller)
	fType := iType.Method(index)
	// 获取函数参数信息
	var num  int = fType.Type.NumIn() - 1
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
	var head int = 0
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
	var head int = 0
	var isBy bool
	for i, c := range name {
		if 0x40 < c && c < 0x5B && i != 0 {
			if isBy {
				strs = append(strs, strings.ToLower(name[head:i]) )
			}
			if name[head:i] == "By" {
				isBy = true
			}
			head = i
		}
	}
	if isBy {
		strs = append(strs, strings.ToLower(name[head:len(name)]) )
	}
	return

}

func (c *ControllerBase) Init(ctx Context) error {
	c.Context = ctx
	return nil
}

func (c *ControllerBase) Release() error {
	return nil
}

func (c *ControllerData) Init(ctx Context) error {
	c.ContextData.Context = ctx
	return nil
}

func (c *ControllerData) Release() error {
	return nil
}