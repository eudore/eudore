# Controller

eudore控制器实现通过分析控制器得到路由方法路径和闭包执行控制器方法。

控制器接口需要实现初始化、释放和注入三个方法，初始化和释放完成控制器对象的初始化和释放过程，注入完成注入控制器路由到路由器中。

```golang
type Controller interface {
	Init(Context) error
	Release() error
	Inject(Controller, RouterMethod) error
}
```

# ControllerBase

ControllerBase是最基本的控制器实现，Init会初始化ctx，Inject调用ControllerBaseInject函数实现注入路由。

```golang
type ControllerBase struct {
	Context
}

// Init 实现控制器初始方法。
func (base *ControllerBase) Init(ctx Context) error {
	base.Context = ctx
	return nil
}

// Release 实现控制器释放方法。
func (base *ControllerBase) Release() error {
	return nil
}

// Inject 方法实现控制器注入到路由器的方法，ControllerBase控制器调用ControllerBaseInject方法注入。
func (base *ControllerBase) Inject(controller Controller, router RouterMethod) error {
	return ControllerBaseInject(controller, router)
}

// ControllerRoute 方法返回默认路由信息。
func (base *ControllerBase) ControllerRoute() map[string]string {
	return nil
}

```

## ControllerBaseInject

ControllerBaseInject函数最主要的路由注入功能。

实现主要分为四段：
- 获得控制器非空的可导出类型，用于初始化控制器复制这些属性，switch处理的这些类型的保存的地址，如果非空就保存下来
- 创建一个控制器池，reflect.New实现对象创建，同时将第一步获得的可导出非空属性，依次Field设置初始化控制器，然后返回interface{}
- 创建控制器路由组，如果控制器名称为XxxxController，那么路由组就是/xxxx。
- 获取控制器可以注册的方法名称和路由路径，然后读取路由方法，使用控制器和方法索引闭包一个请求上下文处理函数，设置改函数的名称，最后给路由注册函数，附加相关参数信息。

```golang
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
```

convertBaseHandler转换控制器方法成请求上下文处理函数。

通过控制器的方法索引获得到控制器方法处理函数，NewContrllerExecFunc会工具类型获得注册的控制器方法处理函数。

再返回闭包的请求上下文处理函数，处理sync.Pool和控制器执行流程。

```golang
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
```

getRoutesWithName方法用于获得全部控制器方法名称和路由路由，通过反射获得到路径和接口获得。

getContrllerMethosNum方法会加载这个类型禁止注册(基础控制器的方法)的全部方法名称，然后返回禁用的数量，那么就可以获得到允许的路由方法数量。

然后遍历控制器全部方法，然后使用方法获得转换的路由路径，同时checkControllerMethod检测该方法是否允许的。

最后检测是否实现了ControllerRoute接口，如果实现了就将自己定义的控制器方法路由信息覆盖。

```golang
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
```

getRouteName函数通过判断大小来分割路由路由，如果有ByXxx就是注册路由路径/:xxx

```golang
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

```

getContrllerMethosNum会进行预加载该类型禁用的方法名称，供后续checkControllerMethod方法使用数据。

会拿到控制器里面的满足嵌入属性、**基础控制器(对象命名Controller为前缀)**、实现控制器接口，这三个条件的对象的全部方法，就是需要禁用的基础控制器的全部方法。

如果不禁用，在注册应用控制器时，会将组合到的基础控制器的方法进行路由注册，但是这些方法是处理函数，不需要注册。

```golang
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
```
