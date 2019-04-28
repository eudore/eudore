package eudore

import (
	"fmt"
	"time"
	"sync"
	"reflect"
	"strings"
)

type (
	Controller interface{
		Init(Context) error
	}
	ControllerRoute interface {
		ControllerRoute(string) string
	}
	ControllerIgnore interface {
		ControllerIgnore() []string
	}
	ControllerBase struct{
		Context
	}
	ControllerSession struct{
		ControllerBase
		Session 	map[string]interface{}
	}
)

func controllerRegister(router RouterMethod, controller Controller) {
	iType := reflect.TypeOf(controller)
	pool := sync.Pool{
		New: func() interface{} {
			return reflect.New(iType.Elem()).Interface()
		},
	}

	// 获取Before和After函数数据
	var before, after []int
	for i := 0; i < iType.NumMethod(); i++ {
		name := iType.Method(i).Name
		var method = getFirstUp(name)
		if method == "After" {
			after = append(after, i)
		}
		if method == "Before" {
			before = append(before, i)
		}
	}
	controllerRoute, isRoute := controller.(ControllerRoute)
	for i := 0; i < iType.NumMethod(); i++ {
		name := iType.Method(i).Name

		var method = getFirstUp(name)
		if !checkAllowMethod(method) {
			continue
		}

		var path string
		if isRoute {
			path = controllerRoute.ControllerRoute(iType.Method(i).Name)
		}
		if len(path) == 0 || path == "-" {
			continue
		}

		tmp := append(append(before, i), after...)
		fmt.Println(method, path, tmp)
		fn := convertHandler(pool, controller, tmp)
		router.AddHandler(strings.ToUpper(method), path, fn)
	}
}

func convertHandler(pool sync.Pool, controller Controller, indexs []int) HandlerFunc {
	iType := reflect.TypeOf(controller)
	// 获取函数参数信息
	var numin []int = make([]int, len(indexs))
	var args []string
	var typeIn []reflect.Type
	for i, index := range indexs {
		fType := iType.Method(index)
		numin[i] = fType.Type.NumIn() - 1
		for j := 0 ; j < numin[i]; j++ {
			typeIn = append(typeIn, fType.Type.In(i + 1)) 
		}
		args = append(args, getFuncArgs(fType.Name)...)
	}
	return func(ctx Context) {
		// 初始化
		controller := pool.Get().(Controller)
		controller.Init(ctx)

		var pos int = 0
		for i, index := range indexs {
			// 构建一个函数参数并执行
			params := make([]reflect.Value, numin[i])
			for j, _ := range params {
				params[j] = reflect.New(typeIn[pos])
				setWithString(params[j].Kind(), params[j], ctx.GetParam(args[pos]))	
				pos++
				params[j] = params[j].Elem()
			}
			reflect.ValueOf(controller).Method(index).Call(params)
		}
		
		pool.Put(controller)
	}
}

func checkAllowMethod(method string) bool {
	for _, i := range []string{"Any", "Get", "post"} {
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
	return ""
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

func (c *ControllerSession) AfterReleaseSession() {
	c.App().Cache.Set("ss", c.Session, time.Second * 3600)
}
