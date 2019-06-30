# Controller

eudore控制器实现通过分析控制器得到路由方法路径和闭包执行控制器方法。

控制器解析函数用路由器注册使用。

控制器接口需要实现初始化和释放两个方法，完成控制器对象的初始化和释放过程。

```golang
type ControllerParseFunc func(Controller) (*RouterConfig, error)
type Controller interface{
	Init(Context) error
	Release() error
}
```

默认实现了一个控制器解析函数`ControllerBaseParseFunc`, 需要控制器实现`ControllerRoute`接口，获得控制器的执行方法和路径，控制器的注册方法通过方法名称截取到方法，截取非法名称就是使用ANY方法。

然后使用函数分析控制器对象，然后闭包执行过程获得HandlerFunc函数对象，然后构建成一个RouterConfig配置。

```golang
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

```

`convertHandler`分析对应的方法的全部入参信息，返回`HandlerFunc`函数对象。

`HandlerFunc`函数先从池获得控制器对象，然后使用ctx调用初始化，再创建入参并初始化，然后调用对应的执行函数，最后释放控制器对象。

```golang
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

```

# 路由器控制器解析函数

如果路由器方法使用的是RouterMethodStd，可以使用Set方法或者构造新路由器来设置ControllerParseFunc属性。

暂时没有新的控制器执行方法设计，未使用改功能。

```golang
// 默认路由器方法注册实现
type RouterMethodStd struct {
	RouterCore
	ControllerParseFunc
	prefix		string
	tags		string
}

func (m *RouterMethodStd) AddController(cs ...Controller) RouterMethod {
	for _, c := range cs {
		// controllerRegister(m, c)
		config, err := m.ControllerParseFunc(c)
		if err == nil {
			config.Inject(m)
		}else {
			fmt.Println(err)
		}
	}
	return m
}
```