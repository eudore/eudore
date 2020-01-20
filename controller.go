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
		Inject(Controller, Router) error
	}
	// controllerRoute 定义获得路由和方法映射的接口。
	controllerRoute interface {
		ControllerRoute() map[string]string
	}
	controllerRouteParam interface {
		GetRouteParam(string, string, string) string
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
	//
	// 默认模板路由可以通过重写GetRouteParam方法，重新定义template参数。
	//
	// 如果Data不为空且未写入数据，会调用Render渲染数据。
	//
	// 如果渲染出html需要app.Renderer支持。
	ControllerView struct {
		ContextData
		Data map[string]interface{}
	}
)

var (
	typeController    = reflect.TypeOf((*Controller)(nil)).Elem()
	controllerNewFunc = make(map[reflect.Type]ControllerHandlerFunc)
)

// init 函数注册控制器默认处理的3*4种处理函数类型。
func init() {
	initControllerHandler()
	initControllerHandlerContext()
	initControllerHandlerMap()
}

func initControllerHandler() {
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
}
func initControllerHandlerContext() {
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
func initControllerHandlerMap() {
	// func(map[string]interface{})
	RegisterControllerHandlerFunc(func(map[string]interface{}) {}, func(ctx Context, controller Controller, index int) {
		req := make(map[string]interface{})
		err := ctx.Bind(&req)
		if err != nil {
			ctx.Fatalf("controller bind error: %v", err)
			return
		}
		reflect.ValueOf(controller).Method(index).Call([]reflect.Value{reflect.ValueOf(req)})
	})

	// func(map[string]interface{}) interface{}
	RegisterControllerHandlerFunc(func(map[string]interface{}) interface{} {
		return nil
	}, func(ctx Context, controller Controller, index int) {
		req := make(map[string]interface{})
		err := ctx.Bind(&req)
		if err != nil {
			ctx.Fatalf("controller bind error: %v", err)
			return
		}
		data := reflect.ValueOf(controller).Method(index).Call([]reflect.Value{reflect.ValueOf(req)})[0].Interface()
		if data != nil && ctx.Response().Size() == 0 {
			err := ctx.Render(data)
			if err != nil {
				ctx.Fatal(err)
			}
		}
	})

	// func(map[string]interface{}) error
	RegisterControllerHandlerFunc(func(map[string]interface{}) error {
		return nil
	}, func(ctx Context, controller Controller, index int) {
		req := make(map[string]interface{})
		err := ctx.Bind(&req)
		if err != nil {
			ctx.Fatalf("controller bind error: %v", err)
			return
		}
		ierr := reflect.ValueOf(controller).Method(index).Call([]reflect.Value{reflect.ValueOf(req)})[0].Interface()
		if ierr != nil {
			ctx.Fatal(ierr)
		}
	})

	// func(map[string]interface{}) (interface{}, error)
	RegisterControllerHandlerFunc(func(map[string]interface{}) (interface{}, error) {
		return nil, nil
	}, func(ctx Context, controller Controller, index int) {
		req := make(map[string]interface{})
		err := ctx.Bind(&req)
		if err != nil {
			ctx.Fatalf("controller bind error: %v", err)
			return
		}
		data, err := reflect.ValueOf(controller).Method(index).Interface().(func(interface{}) (interface{}, error))(req)
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
// 如果控制器实现interface{GetRouteParam(string, string, string) string}接口，使用改接口方法来生成路由参数。
//
// 如果控制器嵌入了其他基础控制器(控制器名称为:ControllerXxx)，控制器路由分析会忽略嵌入的控制器的全部方法。
//
// 如果控制器具有非空和导出的Chan、Func、Interface、Map、Ptr、Slice、Array类型的成员，会知道赋值给新控制器。
//
// 方法类型可以调用ListExtendControllerHandlerFunc()函数查看
//
// 注意：ControllerBaseInject执行的每次控制器会使用sync.Pool分配和回收。
func ControllerBaseInject(controller Controller, router Router) error {
	pType := reflect.TypeOf(controller)
	iType := reflect.TypeOf(controller).Elem()
	iValue := reflect.ValueOf(controller).Elem()

	// 获取控制器可导出非空属性
	var keys []int
	var vals []reflect.Value
	for i := 0; i < iValue.NumField(); i++ {
		field := iValue.Field(i)
		// go1.13 reflect.Value.IsZero
		if !checkValueIsZero(field) && field.CanSet() {
			keys = append(keys, i)
			vals = append(vals, iValue.Field(i))
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
	cpkg := iType.PkgPath()
	group := router.GetParam("controllergroup")
	if group != "" {
		router = router.SetParam("controllergroup", "").Group(group)
	} else if strings.HasSuffix(cname, "Controller") {
		router = router.Group("/" + strings.ToLower(strings.TrimSuffix(cname, "Controller")))
	} else if strings.HasSuffix(cname, "controller") {
		router = router.Group("/" + strings.ToLower(strings.TrimSuffix(cname, "controller")))
	}

	// 获取路由参数函数
	pfn := defaultRouteParam
	v, ok := controller.(controllerRouteParam)
	if ok {
		pfn = v.GetRouteParam
	}

	// 路由器注册控制器方法
	for method, path := range getRoutesWithName(controller) {
		m, ok := pType.MethodByName(method)
		if !ok || path == "" {
			continue
		}

		h := convertBaseHandler(controller, m.Index, pool)
		SetHandlerFuncName(h, fmt.Sprintf("%s.%s.%s", cpkg, cname, method))
		router.AddHandler(getRouteMethod(method), path+" "+pfn(cpkg, cname, method), h)
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

func defaultRouteParam(pkg, name, method string) string {
	return fmt.Sprintf("controllername=%s.%s controllermethod=%s", pkg, name, method)
}

// ControllerSingletonInject 方法实现注入单例控制器。
//
// 单例控制器的方法规则与ControllerBaseInject相同。
func ControllerSingletonInject(controller Controller, router Router) error {
	iType := reflect.TypeOf(controller)

	// 添加控制器组。
	cname := iType.Name()
	cpkg := iType.PkgPath()
	group := router.GetParam("controllergroup")
	if group != "" {
		router = router.SetParam("controllergroup", "").Group(group)
	} else if strings.HasSuffix(cname, "Controller") {
		router = router.Group("/" + strings.ToLower(strings.TrimSuffix(cname, "Controller")))
	} else if strings.HasSuffix(cname, "controller") {
		router = router.Group("/" + strings.ToLower(strings.TrimSuffix(cname, "controller")))
	}

	// 获取路由参数函数
	pfn := defaultRouteParam
	v, ok := controller.(controllerRouteParam)
	if ok {
		pfn = v.GetRouteParam
	}

	for method, path := range getRoutesWithName(controller) {
		m, ok := iType.MethodByName(method)
		if !ok || path == "" {
			continue
		}

		h := convertSingletonHandler(controller, m.Index)
		SetHandlerFuncName(h, fmt.Sprintf("%s.%s.%s", cpkg, cname, method))
		router.AddHandler(getRouteMethod(method), path+" "+pfn(cpkg, cname, method), h)
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

// getRoutesWithName 函数获得一个控制器类型注入的全部名称和路由路径的映射。
func getRoutesWithName(controller Controller) map[string]string {
	iType := reflect.TypeOf(controller)
	names := getContrllerAllowMethos(iType)
	routes := make(map[string]string, len(names))
	for _, name := range names {
		if name != "" {
			routes[name] = getRouteByName(name)
		}
	}

	// 如果控制器实现ControllerRoute接口，加载额外路由。
	controllerRoute, isRoute := controller.(controllerRoute)
	if isRoute {
		for name, path := range controllerRoute.ControllerRoute() {
			routes[name] = path
		}
	}
	return routes
}

// getContrllerAllowMethos 函数获得一个类型除去忽略方法意外的全部方法名称。
func getContrllerAllowMethos(iType reflect.Type) []string {
	if iType.Kind() == reflect.Ptr {
		iType = iType.Elem()
	}
	if strings.HasPrefix(iType.Name(), "Controller") || strings.HasPrefix(iType.Name(), "controller") {
		return nil
	}

	allname := getContrllerAllMethos(iType)
	ignore := getContrllerIgnoreMethos(iType)
	for i := 0; i < len(allname); i++ {
		for j := 0; j < len(ignore); j++ {
			if allname[i] == ignore[j] {
				allname[i] = ""
				break
			}
		}
	}
	return allname
}

// getContrllerIgnoreMethos 函数获得一个类型忽略的全部方法，如果类型名称或者类型嵌入类型名称前缀是Controller则忽略其全部方法。
func getContrllerIgnoreMethos(iType reflect.Type) []string {
	if iType.Kind() == reflect.Ptr {
		iType = iType.Elem()
	}

	var ms []string
	if strings.HasPrefix(iType.Name(), "Controller") || strings.HasPrefix(iType.Name(), "controller") {
		ms = getContrllerAllMethos(iType)
	}
	if iType.Kind() == reflect.Struct {
		for i := 0; i < iType.NumField(); i++ {
			// Controller为前缀的嵌入控制器。
			// 判断嵌入属性
			if iType.Field(i).Anonymous {
				ms = append(ms, getContrllerIgnoreMethos(iType.Field(i).Type)...)
			}
		}
	}
	return ms
}

// getContrllerAllMethos 函数获得一共类型包含指针类型的全部方法名称。
func getContrllerAllMethos(iType reflect.Type) []string {
	if iType.Kind() != reflect.Ptr {
		iType = reflect.New(iType).Type()
	}
	names := make([]string, iType.NumMethod())
	for i := 0; i < iType.NumMethod(); i++ {
		names[i] = iType.Method(i).Name
	}
	return names
}

// getRouteByName 函数使用函数名称生成路由路径。
func getRouteByName(name string) string {
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

// splitName 方法基于路径首字符大写切割
func splitName(name string) (strs []string) {
	var head int
	for i, c := range name {
		if 0x40 < c && c < 0x5B && i != 0 {
			strs = append(strs, name[head:i])
			head = i
		}
	}
	strs = append(strs, name[head:])
	return
}

// Init 实现控制器初始方法。
func (ctl *ControllerBase) Init(ctx Context) error {
	ctl.Context = ctx
	return nil
}

// Release 实现控制器释放方法。
func (ctl *ControllerBase) Release() error {
	return nil
}

// Inject 方法实现控制器注入到路由器的方法，ControllerBase控制器调用ControllerBaseInject方法注入。
func (ctl *ControllerBase) Inject(controller Controller, router Router) error {
	return ControllerBaseInject(controller, router)
}

// ControllerRoute 方法返回默认路由信息。
func (ctl *ControllerBase) ControllerRoute() map[string]string {
	return nil
}

// GetRouteParam 方法添加路由参数信息。
func (ctl *ControllerBase) GetRouteParam(pkg, name, method string) string {
	return defaultRouteParam(pkg, name, method)
}

// Init 实现控制器初始方法。
func (ctl *ControllerData) Init(ctx Context) error {
	ctl.ContextData.Context = ctx
	return nil
}

// Release 实现控制器释放方法。
func (ctl *ControllerData) Release() error {
	return nil
}

// Inject 方法实现控制器注入到路由器的方法，ControllerData控制器调用ControllerBaseInject方法注入。
func (ctl *ControllerData) Inject(controller Controller, router Router) error {
	return ControllerBaseInject(controller, router)
}

// ControllerRoute 方法返回默认路由信息。
func (ctl *ControllerData) ControllerRoute() map[string]string {
	return nil
}

// GetRouteParam 方法添加路由参数信息。
func (ctl *ControllerData) GetRouteParam(pkg, name, method string) string {
	return defaultRouteParam(pkg, name, method)
}

// Init 实现控制器初始方法,单例控制器初始化不执行任何内容。
func (ctl *ControllerSingleton) Init(ctx Context) error {
	return nil
}

// Release 实现控制器释放方法,单例控制器释放不执行任何内容。
func (ctl *ControllerSingleton) Release() error {
	return nil
}

// Inject 方法实现控制器注入到路由器的方法，ControllerSingleton控制器调用ControllerSingletonInject方法注入。
func (ctl *ControllerSingleton) Inject(controller Controller, router Router) error {
	return ControllerSingletonInject(controller, router)
}

// ControllerRoute 方法返回默认路由信息。
func (ctl *ControllerSingleton) ControllerRoute() map[string]string {
	return nil
}

// GetRouteParam 方法添加路由参数信息。
func (ctl *ControllerSingleton) GetRouteParam(pkg, name, method string) string {
	return defaultRouteParam(pkg, name, method)
}

// defaultGetViewTemplate 通过控制器名称和方法名称获得模板路径。
//
// 格式: views/controller/%s/%s.html
//
// MyUserController Index => views/controller/my/user/index.html
func defaultGetViewTemplate(cname string, method string) string {
	if strings.HasSuffix(cname, "Controller") {
		cname = strings.TrimSuffix(cname, "Controller")
	} else if strings.HasSuffix(cname, "controller") {
		cname = strings.TrimSuffix(cname, "controller")
	}
	names := splitName(cname)
	for i := range names {
		names[i] = strings.ToLower(names[i])
	}
	return fmt.Sprintf("views/controller/%s/%s.html", strings.Join(names, "/"), strings.ToLower(method))
}

// Init 实现控制器初始方法。
func (ctl *ControllerView) Init(ctx Context) error {
	ctl.Context = ctx
	ctl.Data = make(map[string]interface{})
	return nil
}

// Release 实现控制器释放方法。
func (ctl *ControllerView) Release() error {
	if ctl.Response().Size() == 0 && len(ctl.Data) != 0 {
		return ctl.Render(ctl.Data)
	}
	return nil
}

// Inject 方法实现控制器注入到路由器的方法，ControllerView控制器调用ControllerViewInject方法注入。
//
// ControllerView控制器在Release时，如果未写入数据会自动写入数据。
func (ctl *ControllerView) Inject(controller Controller, router Router) error {
	return ControllerBaseInject(controller, router)
}

// ControllerRoute 方法返回默认路由信息。
func (ctl *ControllerView) ControllerRoute() map[string]string {
	return nil
}

// GetRouteParam 方法返回路由的参数，View控制器会附加模板信息。
func (ctl *ControllerView) GetRouteParam(pkg, name, method string) string {
	return fmt.Sprintf("controllername=%s.%s controllermethod=%s template=%s", pkg, name, method, defaultGetViewTemplate(name, method))
}

// SetTemplate 方法设置模板文件路径。
func (ctl *ControllerView) SetTemplate(path string) {
	ctl.SetParam("template", path)
}
