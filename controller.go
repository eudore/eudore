package eudore

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

/*
Controller defines the controller interface.

The default AutoRoute controller implements the following functions:
	The controller method maps the route method path.
	Controller construction error delivery (NewControllerError)
	Custom controller function mapping relationship (implement func ControllerRoute() map[string]string)
	Custom controller routing group and routing parameters (implement func ControllerParam(pkg, name, method string) string)
	Controller routing combination, if a controller named xxxController is combined, the routing method of the xxx controller will be combined
	Controller method combination, if you combine a controller with a name other than xxxController, you can directly call the method in the controller property ctl.xxx.

Controller 定义控制器接口。

默认AutoRoute控制器实现下列功能:
	控制器方法映射路由方法路径。
	控制器构造错误传递(NewControllerError)
	自定义控制器函数映射关系(实现func ControllerRoute() map[string]string)
	自定义控制器路由组和路由参数(实现func ControllerParam(pkg, name, method string) string)
	控制器路由组合，如果组合一个名称为xxxController控制器，会组合获得xxx控制器的路由方法
	控制器方法组合，如果组合一个名称非xxxController控制器，可以控制器属性ctl.xxx直接调用方法。
*/
type Controller interface {
	Inject(Controller, Router) error
}

type controllerName interface {
	ControllerName() string
}

type controllerGroup interface {
	ControllerGroup(string) string
}

// controllerRoute 定义获得路由和方法映射的接口。
type controllerRoute interface {
	ControllerRoute() map[string]string
}

// controllerParam 定义获得一个路由参数的接口，转入pkg、controllername、methodname获得需要添加的路由参数。
type controllerParam interface {
	ControllerParam(string, string, string) string
}

// The ControllerAutoRoute implements the routing mapping controller to register the corresponding router method according to the method.
//
// ControllerAutoRoute 实现路由映射控制器根据方法注册对应的路由器方法。
type ControllerAutoRoute struct{}

type controllerError struct {
	Error error
	Name  string
}

// NewControllerError function returns a controller error, and the corresponding error is returned when the controller Inject.
//
// NewControllerError 函数返回一个控制器错误，在控制器Inject时返回对应的错误。
func NewControllerError(ctl Controller, err error) Controller {
	return &controllerError{
		Error: err,
		Name:  getControllerPathName(ctl),
	}
}

// The Inject method returns a controller error when injecting routing rules.
//
// Inject 方法在注入路由规则时返回控制器错误。
func (ctl *controllerError) Inject(Controller, Router) error {
	return ctl.Error
}

// The ControllerName method returns the controller name of controllerError.
//
// ControllerName 方法返回controllerError的控制器名称。
func (ctl *controllerError) ControllerName() string {
	return ctl.Name
}

// Inject method implements the method of injecting the controller into the router,
// and the ControllerAutoRoute controller calls the ControllerInjectAutoRoute method to inject.
//
// Inject 方法实现控制器注入到路由器的方法，ControllerAutoRoute控制器调用ControllerInjectAutoRoute方法注入。
func (ctl *ControllerAutoRoute) Inject(controller Controller, router Router) error {
	return ControllerInjectAutoRoute(controller, router)
}

// ControllerInjectAutoRoute function generates routing rules based on the controller rules, and the usage method is converted into a processing function to support routers.
//
// Routing group: If the'ControllerGroup(string) string' method is implemented, the routing group is returned; if the routing parameter ParamControllerGroup is included, it is used; otherwise, the controller name is used to turn the path.
//
// Routing path: Convert the method with the request method as the prefix to the routing method and path, and then use the map[method]path returned by the'ControllerRoute() map[string]string' method to overwrite the routing path.
//
// Method conversion rules: The method prefix must be a valid request method (within RouterAllMethod), the remaining path is converted to a path, ByName is converted to variable matching/:name, and the last By of the method path is converted to /*;
// The return path of ControllerRoute is'-' and the method is ignored. The first character is'', which means it is a path append parameter.
//
// Routing parameters: If you implement the'ControllerParam(string, string, string) string' method to return routing parameters, otherwise use "controllername=%s.%s controllermethod=%s".
//
// Controller combination: If the controller combines other objects, only the methods of the object whose name suffix is ​​Controller are reserved, and other methods with embedded properties will be ignored.
//
// ControllerInjectAutoRoute 函数基于控制器规则生成路由规则，使用方法转换成处理函数支持路由器。
//
// 路由组: 如果实现'ControllerGroup(string) string'方法返回路由组；如果包含路由参数ParamControllerGroup则使用;否则使用控制器名称驼峰转路径。
//
// 路由路径: 将请求方法为前缀的方法转换成路由方法和路径，然后使用'ControllerRoute() map[string]string'方法返回的map[method]path覆盖路由路径。
//
// 方法转换规则: 方法前缀必须是有效的请求方法(RouterAllMethod之内)，剩余路径驼峰转路径，ByName转换成变量匹配/:name,方法路径最后一个By转换成/*;
// ControllerRoute返回路径为'-'则忽略方法，第一个字符为' '表示为路径追加参数。
//
// 路由参数: 如果实现'ControllerParam(string, string, string) string'方法返回路由参数，否则使用"controllername=%s.%s controllermethod=%s"。
//
// 控制器组合: 如果控制器组合了其他对象，仅保留名称后缀为Controller的对象的方法，其他嵌入属性的方法将被忽略。
func ControllerInjectAutoRoute(controller Controller, router Router) error {
	iType := reflect.TypeOf(controller)
	iValue := reflect.ValueOf(controller)

	// 添加控制器组。
	cname := getControllerName(reflect.Indirect(iValue))
	cpkg := reflect.Indirect(iValue).Type().PkgPath()
	router = router.Group(getContrllerRouterGroup(controller, cname, router))

	// 获取路由参数函数
	pfn := defaultRouteParam
	v, ok := controller.(controllerParam)
	if ok {
		pfn = v.ControllerParam
	}

	// 路由器注册控制器方法
	names, paths := getSortMapValue(getControllerRoutes(controller))
	for i, name := range names {
		m, ok := iType.MethodByName(name)
		if !ok || paths[i] == "-" {
			continue
		}

		h := iValue.Method(m.Index).Interface()
		SetHandlerAliasName(h, fmt.Sprintf("%s.%s.%s", cpkg, cname, name))
		method := getMethodByName(name)
		if method == "" {
			method = "ANY"
		}
		router.AddHandler(method, paths[i]+" "+pfn(cpkg, cname, name), h)
	}
	return nil
}

func getContrllerRouterGroup(controller Controller, name string, router Router) (group string) {
	ctl, ok := controller.(controllerGroup)
	switch {
	case ok:
		group = ctl.ControllerGroup(name)
	case router.Params().Get(ParamControllerGroup) != "":
		group = router.Params().Get(ParamControllerGroup)
		router.Params().Del(ParamControllerGroup)
	case strings.HasSuffix(name, "Controller"):
		buf := make([]rune, 0, len(name)*2)
		for _, b := range name[:len(name)-10] {
			if 64 < b && b < 91 {
				buf = append(buf, '/', b+0x20)
			} else {
				buf = append(buf, b)
			}
		}
		group = string(buf)
	}
	if group != "" && group[0] != '/' {
		return "/" + group
	}
	return group
}

// defaultRouteParam 函数定义默认的控制器参数，可以通过实现controllerParam来覆盖该函数。
func defaultRouteParam(pkg, name, method string) string {
	return fmt.Sprintf("controllername=%s.%s controllermethod=%s", pkg, name, method)
}

func getSortMapValue(data map[string]string) ([]string, []string) {
	keys := make([]string, 0, len(data))
	vals := make([]string, len(data))
	for key := range data {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		vi, vj := getRoutePath(data[keys[i]]), getRoutePath(data[keys[j]])
		if vi == vj {
			return keys[i] < keys[j]
		}
		return vi < vj
	})

	for i, key := range keys {
		vals[i] = data[key]
	}
	return keys, vals
}

// getControllerRoutes 函数获得一个控制器类型注入的全部名称和路由路径的映射。
func getControllerRoutes(controller Controller) map[string]string {
	routes := getContrllerAllowMethos(reflect.ValueOf(controller))
	for name := range routes {
		if getMethodByName(name) != "" {
			routes[name] = getRouteByName(name)
		} else {
			delete(routes, name)
		}
	}
	// 如果控制器实现ControllerRoute接口，加载额外路由。
	controllerRoute, isRoute := controller.(controllerRoute)
	if isRoute {
		for name, path := range controllerRoute.ControllerRoute() {
			if len(path) > 0 && path[0] == ' ' {
				// ControllerRoute获得的路径是空格开头，表示为路由参数。
				routes[name] += path
			} else {
				routes[name] = path
			}
		}
	}
	return routes
}

func getContrllerAllowMethos(iValue reflect.Value) map[string]string {
	names := make(map[string]string)
	for _, name := range getContrllerAllMethos(iValue) {
		names[name] = ""
	}

	iValue = reflect.Indirect(iValue)
	iType := iValue.Type()
	if iValue.Kind() == reflect.Struct {
		// 删除嵌入非控制器方法
		for i := 0; i < iValue.NumField(); i++ {
			if iType.Field(i).Anonymous {
				if !strings.HasSuffix(getControllerName(iValue.Field(i)), "Controller") {
					for _, name := range getContrllerAllMethos(iValue.Field(i)) {
						delete(names, name)
					}
				}
			}
		}
		// 追加嵌入控制器方法
		for i := 0; i < iType.NumField(); i++ {
			if iType.Field(i).Anonymous {
				if strings.HasSuffix(getControllerName(iValue.Field(i)), "Controller") {
					for _, name := range getContrllerAllMethos(iValue.Field(i)) {
						names[name] = ""
					}
				}
			}
		}
	}
	return names
}

// getContrllerAllMethos 函数获得一共类型包含指针类型的全部方法名称。
func getContrllerAllMethos(iValue reflect.Value) []string {
	iType := iValue.Type()
	if iType.Kind() != reflect.Ptr {
		iType = reflect.New(iType).Type()
	}
	names := make([]string, iType.NumMethod())
	for i := 0; i < iType.NumMethod(); i++ {
		names[i] = iType.Method(i).Name
	}
	return names
}

func getControllerName(iValue reflect.Value) string {
	if iValue.Kind() == reflect.Ptr && iValue.IsNil() {
		iValue = reflect.New(iValue.Type().Elem())
	}
	var name string
	if iValue.Type().Implements(typeControllerName) && iValue.CanSet() {
		name = iValue.MethodByName("ControllerName").Call(nil)[0].String()
	} else {
		name = reflect.Indirect(iValue).Type().Name()
	}
	pos := strings.IndexByte(name, '[')
	if pos != -1 {
		name = name[:pos]
	}
	return name
}

// getRouteByName 函数使用函数名称生成路由路径。
func getRouteByName(name string) string {
	names := splitTitleName(name)
	if getMethodByName(names[0]) != "" {
		names = names[1:]
	}
	name = ""
	for i := 0; i < len(names); i++ {
		if names[i] == "By" {
			i++
			if i == len(names) {
				name = name + "/*"
			} else {
				name = name + "/:" + names[i]
			}
		} else {
			name = name + "/" + names[i]
		}
	}
	return strings.ToLower(name)
}

func getMethodByName(name string) string {
	name = strings.ToUpper(getFirstUp(name))
	if name == "ANY" {
		return MethodAny
	}
	for _, method := range RouterAllMethod {
		if method == name {
			return name
		}
	}
	return ""
}

func getFirstUp(name string) string {
	for i, c := range name {
		if 0x40 < c && c < 0x5B && i != 0 {
			return name[:i]
		}
	}
	return name
}

// splitTitleName 方法基于路径首字符大写切割
func splitTitleName(str string) []string {
	var body []byte
	for i := range str {
		if i != 0 && byteIn(str[i], 0x40) && byteIn(str[i-1], 0x60) {
			body = append(body, ' ')
			body = append(body, str[i])
		} else if i != 0 && i != len(str)-1 && byteIn(str[i], 0x40) && byteIn(str[i-1], 0x40) && byteIn(str[i+1], 0x60) {
			body = append(body, ' ')
			body = append(body, str[i])
		} else if byteIn(str[i], 0x40) && i != 0 {
			body = append(body, str[i]+0x20)
		} else {
			body = append(body, str[i])
		}
	}
	return strings.Split(string(body), " ")
}

func byteIn(b byte, r byte) bool {
	return r < b && b < r+0x1B
}
