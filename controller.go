package eudore

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

/*
Controller 定义控制器接口。

默认AutoRoute控制器实现下列功能:
	路由映射控制器，将方法转换成一般的路由处理函数，使用函数扩展(ControllerAutoRoute)。
	控制器构造错误传递(NewControllerError)
	控制器前置和后置处理函数,Init和Release方法在控制器方法前后调用
	自定义控制器函数映射关系(实现func ControllerRoute() map[string]string)
	自定义控制器路由组和路由参数(实现func ControllerParam(pkg, name, method string) string)
	控制器路由组合，如果组合一个名称为xxxController控制器，会组合获得xxx控制器的路由方法
	控制器方法组合，如果组合一个名称非xxxController控制器，可以控制器属性ctl.xxx直接调用方法。
*/
type Controller interface {
	Inject(Controller, Router) error
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

// ControllerAutoRoute 实现路由映射控制器根据方法注册对应的路由器方法。
type ControllerAutoRoute struct{}

type controllerError struct {
	Error error
	Name  string
}

// NewControllerError 函数返回一个控制器错误，在控制器Inject时返回对应的错误。
func NewControllerError(ctl Controller, err error) Controller {
	return &controllerError{
		Error: err,
		Name:  getConrtrollerName(ctl),
	}
}

// Inject 方法在注入路由规则时返回控制器错误。
func (ctl *controllerError) Inject(Controller, Router) error {
	return ctl.Error
}

// String 方法返回controllerError的控制器名称。
func (ctl *controllerError) String() string {
	return ctl.Name
}

// Inject 方法实现控制器注入到路由器的方法，ControllerAutoRoute控制器调用ControllerInjectAutoRoute方法注入。
func (ctl *ControllerAutoRoute) Inject(controller Controller, router Router) error {
	return ControllerInjectAutoRoute(controller, router)
}

// ControllerInjectAutoRoute 函数基于控制器规则生成路由规则，使用方法转换成处理函数支持路由器。
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
// 如果控制器实现controllerRoute接口，会替换自动分析路由路径，路由路径为空会忽略该方法。
//
// 如果控制器实现interface{ControllerParam(string, string, string) string}接口，使用改接口方法来生成路由参数。
//
// 如果控制器嵌入了其他基础控制器(控制器名称为:ControllerXxx)，控制器路由分析会忽略嵌入的控制器的全部方法。
func ControllerInjectAutoRoute(controller Controller, router Router) error {
	pType := reflect.TypeOf(controller)
	pValue := reflect.ValueOf(controller)
	iType := reflect.TypeOf(controller).Elem()

	// 添加控制器组。
	cname := iType.Name()
	cpkg := iType.PkgPath()
	router = router.Group(getContrllerRouterGroup(controller, cname, router))

	// 获取路由参数函数
	pfn := defaultRouteParam
	v, ok := controller.(controllerParam)
	if ok {
		pfn = v.ControllerParam
	}

	// 路由器注册控制器方法
	methods, paths := getSortMapValue(getRoutesWithName(controller))
	for i, method := range methods {
		m, ok := pType.MethodByName(method)
		if !ok || (!checkAllowMethod(method) && paths[i] == "") {
			continue
		}

		h := pValue.Method(m.Index)
		SetHandlerAliasName(h, fmt.Sprintf("%s.%s.%s", cpkg, cname, method))
		router.AddHandler(getRouteMethod(method), paths[i]+" "+pfn(cpkg, cname, method), h)
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
	case controllerHasSuffix(name):
		buf := make([]rune, 0, len(name)*2)
		for _, i := range name[:len(name)-10] {
			if 64 < i && i < 91 {
				buf = append(buf, '/', i+0x20)
			} else {
				buf = append(buf, i)
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

// getRoutesWithName 函数获得一个控制器类型注入的全部名称和路由路径的映射。
func getRoutesWithName(controller Controller) map[string]string {
	iType := reflect.TypeOf(controller)
	names := getContrllerAllMethos(iType)
	routes := make(map[string]string, len(names))
	for _, name := range names {
		if name != "" && !strings.HasPrefix(name, "Controller") {
			routes[name] = getRouteByName(name)
		}
	}
	for _, name := range getContrllerIgnoreMethos(iType) {
		delete(routes, name)
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

func getContrllerIgnoreMethos(iType reflect.Type) []string {
	var allname []string
	if iType.Kind() == reflect.Ptr {
		iType = iType.Elem()
	}
	if iType.Kind() == reflect.Struct {
		for i := 0; i < iType.NumField(); i++ {
			// 判断嵌入属性
			if iType.Field(i).Anonymous {
				var ignore []string
				if controllerHasSuffix(getReflectTypeName(iType.Field(i).Type)) {
					ignore = getContrllerIgnoreMethos(iType.Field(i).Type)
				} else {
					ignore = getContrllerAllMethos(iType.Field(i).Type)
				}
				allname = append(allname, ignore...)
			}
		}
	}
	return allname
}

// controllerHasSuffix 函数判断控制器名称后缀是否为"Controller"或"controller"。
func controllerHasSuffix(name string) bool {
	return strings.HasSuffix(name, "Controller") || strings.HasSuffix(name, "controller")
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

func getReflectTypeName(iType reflect.Type) string {
	if iType.Kind() == reflect.Ptr {
		iType = iType.Elem()
	}
	return iType.Name()
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
	for _, i := range []string{"Any", "Get", "Post", "Put", "Delete", "Head", "Patch", "Options", "Connect", "Trace"} {
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
