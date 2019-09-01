package eudore

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type (
	// ControllerHandlerFunc 定义控制器执行函数。
	ControllerHandlerFunc func(Context, Controller, int)
	// Controller 定义控制器必要的接口。
	Controller interface {
		Init(Context) error
		Release() error
		Inject(Controller, RouterMethod) error
	}
	// ControllerRoute 定义获得路由和方法映射的接口。
	ControllerRoute interface {
		ControllerRoute() map[string]string
	}
	// ControllerBase 实现基本控制器。
	ControllerBase struct {
		Context
	}
	// ControllerData 实现基于ContextData的控制器,基于ControllerBase扩展了额外的控制器方法。
	ControllerData struct {
		ContextData
	}
	// ControllerSingleton 实现单例控制器。
	ControllerSingleton struct{}
	// ControllerView 基于ControllerBase额外增加了控制器自动渲染数据。
	ControllerView struct {
		Context
		Data    map[string]interface{}
		TplName string
	}
)

var (
	typeController    = reflect.TypeOf((*Controller)(nil)).Elem()
	controllerMethods = make(map[reflect.Type]map[string]struct{})
	controllerNewFunc = make(map[reflect.Type]ControllerHandlerFunc)
)

func init() {
	// func()
	RegisterControllerHandlerFunc(func() {}, func(ctx Context, controller Controller, index int) {
		reflect.ValueOf(controller).Method(index).Call(nil)
	})

	// func() interface{}
	RegisterControllerHandlerFunc(func() interface{} {
		return nil
	}, func(ctx Context, controller Controller, index int) {
		data := reflect.ValueOf(controller).Method(index).Call(nil)[0].Interface()
		if data != nil && ctx.Response().Size() == 0 {
			err := ctx.Render(data)
			if err != nil {
				ctx.Fatal(err)
			}
		}
	})

	// func() error
	RegisterControllerHandlerFunc(func() error {
		return nil
	}, func(ctx Context, controller Controller, index int) {
		ierr := reflect.ValueOf(controller).Method(index).Call(nil)[0].Interface()
		if ierr != nil {
			ctx.Fatal(ierr)
		}
	})

	// func() (interface{}, error)
	RegisterControllerHandlerFunc(func() (interface{}, error) {
		return nil, nil
	}, func(ctx Context, controller Controller, index int) {
		data, err := reflect.ValueOf(controller).Method(index).Interface().(func() (interface{}, error))()
		if err == nil && data != nil && ctx.Response().Size() == 0 {
			err = ctx.Render(data)
		}
		if err != nil {
			ctx.Fatal(err)
		}
	})

	// func(Context)
	RegisterControllerHandlerFunc(func(Context) {}, func(ctx Context, controller Controller, index int) {
		reflect.ValueOf(controller).Method(index).Call([]reflect.Value{reflect.ValueOf(ctx)})
	})

	// func() interface{}
	RegisterControllerHandlerFunc(func(Context) interface{} {
		return nil
	}, func(ctx Context, controller Controller, index int) {
		data := reflect.ValueOf(controller).Method(index).Call([]reflect.Value{reflect.ValueOf(ctx)})[0].Interface()
		if data != nil && ctx.Response().Size() == 0 {
			err := ctx.Render(data)
			if err != nil {
				ctx.Fatal(err)
			}
		}
	})

	// func() error
	RegisterControllerHandlerFunc(func(Context) error {
		return nil
	}, func(ctx Context, controller Controller, index int) {
		ierr := reflect.ValueOf(controller).Method(index).Call([]reflect.Value{reflect.ValueOf(ctx)})[0].Interface()
		if ierr != nil {
			ctx.Fatal(ierr)
		}
	})

	// func() (interface{}, error)
	RegisterControllerHandlerFunc(func(Context) (interface{}, error) {
		return nil, nil
	}, func(ctx Context, controller Controller, index int) {
		data, err := reflect.ValueOf(controller).Method(index).Interface().(func(Context) (interface{}, error))(ctx)
		if err == nil && data != nil && ctx.Response().Size() == 0 {
			err = ctx.Render(data)
		}
		if err != nil {
			ctx.Fatal(err)
		}
	})
}

// NewContrllerExecFunc 使用控制器返回对应类型的控制器执行函数。
func NewContrllerExecFunc(controller Controller, index int) ControllerHandlerFunc {
	fType := reflect.ValueOf(controller).Method(index).Type()
	fn, ok := controllerNewFunc[fType]
	if ok {
		return fn
	}
	panic(fmt.Errorf(ErrFormatNewContrllerExecFuncTypeNotFunc, fType.String()))
}

// RegisterControllerHandlerFunc 函数注册一个函数类型的控制器执行函数。
//
// 可以使用ListExtendControllerHandlerFunc()函数查看已经注册的函数类型。
func RegisterControllerHandlerFunc(fn interface{}, val ControllerHandlerFunc) {
	iType := reflect.TypeOf(fn)
	if iType.Kind() != reflect.Func {
		panic(ErrRegisterControllerHandlerFuncParamNotFunc)
	}
	controllerNewFunc[iType] = val
}

// ListExtendControllerHandlerFunc 函数列出全部控制器执行函数的类型。
func ListExtendControllerHandlerFunc() []string {
	strs := make([]string, 0, len(controllerNewFunc))
	for i := range controllerNewFunc {
		strs = append(strs, i.String())
	}
	return strs
}

// ControllerBaseInject 定义基本的控制器实现函数。
//
// 如果控制器名称为XxxxController，控制器会注册到路由组/Xxxx下，注册的方法会附加请求上下文参数'controller'，指定控制器包名称。
//
// 请求方法为函数首字母大写单词，如果方法不是Get、Post、Put、Delete、Patch、Options方法，则使用Any方法。
//
// 请求路径为名称每首字母大写单词组成，忽略第一个表示请求方法的单词，如果前一个单位为'By'表示是变量。
//
// Hello() ANY /hello
//
// Get() GET /*
//
// GetId() GET /id
//
// GetById() GET /:id
//
// 如果控制器实现ControllerRoute接口，会替换自动分析路由路径，路由路径为空会忽略该方法。
//
// 如果控制器嵌入了其他基础控制器(控制器名称为:ControllerXxx)，控制器路由分析会忽略嵌入的控制器的全部方法。
//
// 如果控制器具有非空和导出的Chan、Func、Interface、Map、Ptr、Slice、Array类型的成员，会知道赋值给新控制器。
//
// 方法类型可以调用ListExtendControllerHandlerFunc()函数查看
//
// 注意：ControllerBaseInject执行的每次控制器会使用sync.Pool分配和回收。
func ControllerBaseInject(controller Controller, router RouterMethod) error {
	pType := reflect.TypeOf(controller)
	iType := reflect.TypeOf(controller).Elem()
	iValue := reflect.ValueOf(controller).Elem()

	// 获取控制器可导出非空属性
	var keys []int
	var vals []reflect.Value
	for i := 0; i < iValue.NumField(); i++ {
		field := iValue.Field(i)
		switch field.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice, reflect.Array:
			if !field.IsNil() && field.CanSet() {
				keys = append(keys, i)
				vals = append(vals, iValue.Field(i))
			}
		}
	}

	// 设置控制器Pool
	pool := &sync.Pool{
		New: func() interface{} {
			iValue := reflect.New(iType)
			for i, key := range keys {
				iValue.Elem().Field(key).Set(vals[i])
			}
			return iValue.Interface()
		},
	}

	// 添加控制器组。
	cname := iType.Name()
	cpkg := iType.PkgPath() + "." + cname
	if strings.HasSuffix(cname, "Controller") {
		router = router.Group("/" + strings.ToLower(strings.TrimSuffix(cname, "Controller")))
	}

	// 路由器注册控制器方法
	for method, path := range getRoutesWithName(controller) {
		m, ok := pType.MethodByName(method)
		if !ok || path == "" {
			continue
		}

		h := convertBaseHandler(controller, m.Index, pool)
		SetHandlerFuncName(h, fmt.Sprintf("%s.%s", cpkg, method))
		router.AddHandler(getRouteMethod(method), path+fmt.Sprintf(" controllername=%s controllermethod=%s", cpkg, method), h)
	}
	return nil
}

// convertHandler 实现返回一个HandlerFunc对象，用于执行一个控制器方法。
//
// 方法使用ControllerHandlerFunc执行。
func convertBaseHandler(controller Controller, index int, pool *sync.Pool) HandlerFunc {
	fn := NewContrllerExecFunc(controller, index)
	return func(ctx Context) {
		// 初始化
		controller := pool.Get().(Controller)
		err := controller.Init(ctx)
		if err != nil {
			ctx.Fatal(err)
			return
		}

		fn(ctx, controller, index)
		err = controller.Release()
		if err != nil {
			ctx.Fatal(err)
		}

		pool.Put(controller)
	}
}

func ControllerViewInject(controller Controller, router RouterMethod) error {
	pType := reflect.TypeOf(controller)
	iType := reflect.TypeOf(controller).Elem()
	pool := &sync.Pool{
		New: func() interface{} {
			return reflect.New(iType).Interface()
		},
	}

	// 添加控制器组。
	cname := iType.Name()
	cpkg := iType.PkgPath() + "." + cname
	if strings.HasSuffix(cname, "Controller") {
		router = router.Group("/" + strings.ToLower(strings.TrimSuffix(cname, "Controller")))
	}

	for method, path := range getRoutesWithName(controller) {
		m, ok := pType.MethodByName(method)
		if !ok || path == "" {
			continue
		}

		h := convertBaseHandler(controller, m.Index, pool)
		SetHandlerFuncName(h, fmt.Sprintf("%s.%s", cpkg, method))

		tpl := fmt.Sprintf("controller/%s/%s.tpl", cname, strings.ToLower(method))
		router.AddHandler(getRouteMethod(method), path+fmt.Sprintf(" controllername=%s controllermethod=%s defaultTpl=%s", cpkg, method, tpl), h)
	}
	return nil
}

// ControllerSingletonInject 方法实现注入单例控制器。
//
// 单例控制器的方法规则与ControllerBaseInject相同。
func ControllerSingletonInject(controller Controller, router RouterMethod) error {
	iType := reflect.TypeOf(controller)

	// 添加控制器组。
	cname := iType.Elem().Name()
	cpkg := iType.Elem().PkgPath() + "." + cname
	params := fmt.Sprintf(" controller=%s", cpkg)
	if strings.HasSuffix(cname, "Controller") {
		router = router.Group("/" + strings.ToLower(strings.TrimSuffix(cname, "Controller")))
	}

	for name, path := range getRoutesWithName(controller) {
		m, ok := iType.MethodByName(name)
		if !ok || path == "" {
			continue
		}

		h := convertSingletonHandler(controller, m.Index)
		SetHandlerFuncName(h, fmt.Sprintf("%s.%s", cpkg, name))
		router.AddHandler(getRouteMethod(name), path+params, h)
	}
	return nil
}

// convertSingletonHandler 方法返回一个单例控制器方法处理函数。
func convertSingletonHandler(controller Controller, index int) HandlerFunc {
	fn := NewContrllerExecFunc(controller, index)
	return func(ctx Context) {
		err := controller.Init(ctx)
		if err != nil {
			ctx.Fatal(err)
			return
		}

		fn(ctx, controller, index)

		err = controller.Release()
		if err != nil {
			ctx.Fatal(err)
		}
	}
}

func getRoutesWithName(controller Controller) map[string]string {
	iType := reflect.TypeOf(controller)
	routes := make(map[string]string, iType.NumMethod()-getContrllerMethosNum(iType))
	for i := 0; i < iType.NumMethod(); i++ {
		name := iType.Method(i).Name
		if !checkControllerMethod(iType, name) {
			routes[name] = getRouteName(name)
		}
	}

	// 如果控制器实现ControllerRoute接口，加载额外路由。
	controllerRoute, isRoute := controller.(ControllerRoute)
	if isRoute {
		for name, path := range controllerRoute.ControllerRoute() {
			routes[name] = path
		}
	}
	return routes
}

func getRouteName(name string) string {
	names := splitName(name)
	if checkAllowMethod(names[0]) {
		names = names[1:]
	}
	if len(names) == 0 {
		return "/*"
	}
	name = ""
	for i := 0; i < len(names); i++ {
		if names[i] == "By" {
			name = name + "/:" + names[i+1]
			i++
		} else {
			name = name + "/" + names[i]
		}
	}
	return strings.ToLower(name)
}

func getRouteMethod(name string) string {
	method := getFirstUp(name)
	if checkAllowMethod(method) {
		return strings.ToUpper(method)
	}
	return "ANY"
}

func getFirstUp(name string) string {
	for i, c := range name {
		if 0x40 < c && c < 0x5B && i != 0 {
			return name[:i]
		}
	}
	return name
}

func checkAllowMethod(method string) bool {
	for _, i := range []string{"Any", "Get", "Post", "Put", "Delete", "Patch", "Options"} {
		if i == method {
			return true
		}
	}
	return false
}

func getContrllerMethosNum(iType reflect.Type) int {
	iType = iType.Elem()
	methods := make(map[string]struct{})
	for i := 0; i < iType.NumField(); i++ {
		// Controller为前缀的嵌入控制器。
		// 判断嵌入属性
		if !iType.Field(i).Anonymous {
			continue
		}
		// 判断名称前缀
		if !strings.HasPrefix(iType.Field(i).Name, "Controller") {
			continue
		}

		// 转换成指针类型，获得指针接实者的方法。
		sType := iType.Field(i).Type
		switch sType.Kind() {
		case reflect.Ptr, reflect.Interface:
			break
		default:
			sType = reflect.New(sType).Type()
		}
		// 判断实现控制器接口。
		if !sType.Implements(typeController) {
			continue
		}

		for i := 0; i < sType.NumMethod(); i++ {
			methods[sType.Method(i).Name] = struct{}{}
		}
	}
	controllerMethods[iType] = methods
	return len(methods)
}

func checkControllerMethod(iType reflect.Type, method string) bool {
	_, ok := controllerMethods[iType.Elem()][method]
	return ok
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

// Init 实现控制器初始方法。
func (c *ControllerBase) Init(ctx Context) error {
	c.Context = ctx
	return nil
}

// Release 实现控制器释放方法。
func (c *ControllerBase) Release() error {
	return nil
}

// Inject 方法实现控制器注入到路由器的方法，ControllerBase控制器调用ControllerBaseInject方法注入。
func (c *ControllerBase) Inject(controller Controller, router RouterMethod) error {
	return ControllerBaseInject(controller, router)
}

// ControllerRoute 方法返回默认路由详细。
func (c *ControllerBase) ControllerRoute() map[string]string {
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

// Inject 方法实现控制器注入到路由器的方法，ControllerData控制器调用ControllerBaseInject方法注入。
func (c *ControllerData) Inject(controller Controller, router RouterMethod) error {
	return ControllerBaseInject(controller, router)
}

// ControllerRoute 方法返回默认路由详细。
func (c *ControllerData) ControllerRoute() map[string]string {
	return nil
}

// Init 实现控制器初始方法,单例控制器初始化不执行任何内容。
func (c *ControllerSingleton) Init(ctx Context) error {
	return nil
}

// Release 实现控制器释放方法,单例控制器释放不执行任何内容。
func (c *ControllerSingleton) Release() error {
	return nil
}

// Inject 方法实现控制器注入到路由器的方法，ControllerSingleton控制器调用ControllerSingletonInject方法注入。
func (c *ControllerSingleton) Inject(controller Controller, router RouterMethod) error {
	return ControllerSingletonInject(controller, router)
}

// ControllerRoute 方法返回默认路由详细。
func (c *ControllerSingleton) ControllerRoute() map[string]string {
	return nil
}

// Init 实现控制器初始方法。
func (c *ControllerView) Init(ctx Context) error {
	c.Context = ctx
	c.Data = make(map[string]interface{})
	c.TplName = ""
	return nil
}

// Release 实现控制器释放方法。
func (c *ControllerView) Release() error {
	if c.Response().Size() == 0 && len(c.Data) != 0 {
		if c.TplName == "" {
			c.TplName = c.GetParam("defaultTpl")
		}
	}
	return nil
}

// Inject 方法实现控制器注入到路由器的方法，ControllerView控制器调用ControllerViewInject方法注入。
//
// ControllerView控制器在Release时，如果未写入数据会自动写入数据。
func (c *ControllerView) Inject(controller Controller, router RouterMethod) error {
	return ControllerViewInject(controller, router)
}

// ControllerRoute 方法返回默认路由详细。
func (c *ControllerView) ControllerRoute() map[string]string {
	return nil
}
