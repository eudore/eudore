package eudore

import (
	"fmt"
	"time"
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
	ControllerSession struct{
		ControllerBase
		Session 	map[string]interface{}
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
			continue
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

func (c *ControllerBase) GetQueryBool(key string) bool {
	return GetStringBool(c.GetQuery(key))
}

func (c *ControllerBase) GetQueryInt(key string) int {
	return GetStringInt(c.GetQuery(key))
}

func (c *ControllerBase) GetQueryUint64(key string) uint64 {
	return GetStringUint64(c.GetQuery(key))
}

func (c *ControllerBase) GetQueryFloat32(key string) float32 {
	return GetStringFloat32(c.GetQuery(key))
}

func (c *ControllerBase) GetQueryFloat64(key string) float64 {
	return GetStringFloat64(c.GetQuery(key))
}

func (c *ControllerBase) GetQueryString(key string) string {
	return c.GetQuery(key)
}



func (c *ControllerBase) GetQueryDefaultBool(key string, b bool) bool {
	return GetStringDefaultBool(c.GetQuery(key), b)
}

func (c *ControllerBase) GetQueryDefaultInt(key string, n int) int {
	return GetStringDefaultInt(c.GetQuery(key), n)
}

func (c *ControllerBase) GetQueryDefaultUint64(key string, n uint64) uint64 {
	return GetStringDefaultUint64(c.GetQuery(key), n)
}

func (c *ControllerBase) GetQueryDefaultFloat32(key string, f float32) float32 {
	return GetStringDefaultFloat32(c.GetQuery(key), f)
}

func (c *ControllerBase) GetQueryDefaultFloat64(key string, f float64) float64 {
	return GetStringDefaultFloat64(c.GetQuery(key), f)
}

func (c *ControllerBase) GetQueryDefaultString(key , str string) string {
	return GetStringDefault(c.GetQuery(key), str)
}



func (c *ControllerSession) Init(ctx Context) error {
	var ok bool
	c.Session, ok = ctx.App().Cache.Get("ss").(map[string]interface{})
	if !ok {
		c.Session = make(map[string]interface{})
	}
	return c.ControllerBase.Init(ctx)
}


func (c *ControllerSession) SetSession(key string, val interface{}) {
	c.Session[key] = val
}

func (c *ControllerSession) DelSession(key string) {
	delete(c.Session, key)
}

func (c *ControllerSession) GetSession(key string) interface{} {
	return c.Session[key]
}

func (c *ControllerSession) GetSessionBool(key string) bool {
	return GetBool(c.GetSession(key))
}

func (c *ControllerSession) GetSessionInt(key string) int {
	return GetInt(c.GetSession(key))
}

func (c *ControllerSession) GetSessionUint64(key string) uint64 {
	return GetUint64(c.GetSession(key))
}

func (c *ControllerSession) GetSessionFloat32(key string) float32 {
	return GetFloat32(c.GetSession(key))
}

func (c *ControllerSession) GetSessionFloat64(key string) float64 {
	return GetFloat64(c.GetSession(key))
}

func (c *ControllerSession) GetSessionString(key string) string {
	return GetString(key)
}

func (c *ControllerSession) Release() error {
	c.App().Cache.Set("ss", c.Session, time.Second * 3600)
	return c.ControllerBase.Release()
}
