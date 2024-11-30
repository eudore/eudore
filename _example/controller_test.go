package eudore_test

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/eudore/eudore"
)

func TestControllerInject(*testing.T) {
	r := NewRouter(nil)
	r.AddController(new(autoController))
	r.AddController(new(typeController))
	r.AddController(NewErrController(-10))
	r.AddController(NewErrController(10))
	r.AddController(new(MethodController))
	r.Group(" controllergroup=route").AddController(new(RouteController))
	r.AddController(&typeNameController[int]{})
}

type errController struct {
	ControllerAutoRoute
}

func NewErrController(i int) Controller {
	ctl := new(errController)
	if i < 0 {
		// controller创建错误处理。
		return NewControllerError(ctl, errors.New("int must grate 0"))
	}
	return ctl
}

func (ctl *errController) ControllerGroup(string, string) string {
	return "err"
}

func (ctl *errController) Any(ctx Context) {
	ctx.Info("errController Any")
}

// Get 方法注册不存在的扩展函数，触发注册error。
func (*errController) Get(ControllerAutoRoute) interface{} {
	return "get errController"
}

type autoController struct {
	ControllerAutoRoute
}

// Any 方法注册 Any /*路径。
func (*autoController) Any(ctx Context) {
	ctx.Info("autoController Any")
}

// Get 方法注册 Get /*路径。
func (*autoController) GetBy(ctx Context) interface{} {
	ctx.Debug("get")
	return "get autoController"
}

// GetInfoById 方法注册GET /info/:id/name 路由路径。
func (*autoController) GetInfoByIDName(ctx Context) interface{} {
	return ctx.GetParam("id")
}

// String 方法返回控制器名称，响应Router.AddController输出的名称。
func (*autoController) String() string {
	return "hello.autoController"
}

// Help 方法定义一个控制器本身的方法。
func (*autoController) Help(ctx Context) {}

func (*autoController) GetData(ctx Context) {}

func (*autoController) Inject(controller Controller, router Router) error {
	router = router.Group("")
	params := router.Params()
	*params = params.Set(ParamControllerGroup, "/auto")
	return ControllerInjectAutoRoute(controller, router)
}

// ControllerRoute 方法返回控制器路由推导修改信息。
func (*autoController) ControllerRoute() map[string]string {
	return map[string]string{
		// 修改Path方法的路由注册,路径为空就是忽略改方法
		"GetData": "-",
		"Help":    "/help",
		// 给Any方法自动生成的路径添加参数iaany
		"Any": " isany=1",
	}
}

// ControllerParam 方法添加路由参数信息。
func (*autoController) ControllerParam(pkg, name, method string) string {
	return fmt.Sprintf("source=ControllerParam cpkg=%s cname=%s cmethod=%s", pkg, name, method)
}

type typeController struct {
	ControllerAutoType[request14]
}

type request14 struct {
	Name string
}

func (*typeController) AnyH1(ctx Context, data *request14)               {}
func (*typeController) AnyH2(ctx Context, data *request14) error         { return nil }
func (*typeController) AnyH3(ctx Context, data *request14) (any, error)  { return nil, nil }
func (*typeController) AnyH4(ctx Context, data []request14)              {}
func (*typeController) AnyH5(ctx Context, data []request14) error        { return nil }
func (*typeController) AnyH6(ctx Context, data []request14) (any, error) { return nil, nil }

// Controllerwebsite 是基础方法的控制器
type Controllerwebsite struct {
	ControllerAutoRoute
}
type MethodController struct {
	// Controllerwebsite 因名称后缀不为Controller所以不会注册Hello方法为路由。
	Controllerwebsite
}
type tableController struct {
	ControllerAutoRoute
}

// RouteController 从tableController嵌入两个方法注册成路由。
type RouteController struct {
	tableController
	*baseMethod
}

type baseMethod struct{}

func (baseMethod) Any() {}

// Hello 方法返回heelo
func (ctl *Controllerwebsite) Hello() string {
	return "hello eudore"
}

// Get 方法不会被组合
func (ctl *Controllerwebsite) Get(ctx Context) {}

// Any 方法处理控制器全部请求。
func (ctl *MethodController) Any(ctx Context) {
	ctx.Debug("MethodController Any", ctl.Hello())
}

func (ctl *MethodController) ControllerGroup(pkg, name string) string {
	return ""
}

func (ctl tableController) ControllerName() string {
	return "tableController[Context]"
}

func (ctl *tableController) Hello() interface{} {
	return "hello eudore"
}

func (ctl *tableController) Any(ctx Context) {
	ctx.Debug("tableController Any", ctl.Hello())
}

type typeNameController[T any] struct {
	ControllerAutoRoute
}

func (ctl *typeNameController[T]) Any(ctx Context) {
	var t T
	ctx.Debugf("typeNameController is %T", t)
}
