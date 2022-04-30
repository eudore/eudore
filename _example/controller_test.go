package eudore_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/eudore/eudore"
)

func TestControllerError(t *testing.T) {
	app := eudore.NewApp()
	app.AddController(NewErrController(-10))
	app.AddController(NewErrController(10))

	app.CancelFunc()
	app.Run()
}

type errController struct {
	eudore.ControllerAutoRoute
}

func NewErrController(i int) eudore.Controller {
	ctl := new(errController)
	if i < 0 {
		// controller创建错误处理。
		return eudore.NewControllerError(ctl, errors.New("int must grate 0"))
	}
	return ctl
}

func (ctl *errController) ControllerGroup(string) string {
	return "err"
}

func (ctl *errController) Any(ctx eudore.Context) {
	ctx.Info("errController Any")
}

// Get 方法注册不存在的扩展函数，触发注册error。
func (*errController) Get(eudore.ControllerAutoRoute) interface{} {
	return "get errController"
}

func TestControllerAutoRoute(t *testing.T) {
	app := eudore.NewApp()
	app.AddController(new(autoController))

	app.CancelFunc()
	app.Run()
}

type autoController struct {
	eudore.ControllerAutoRoute
}

// Any 方法注册 Any /*路径。
func (*autoController) Any(ctx eudore.Context) {
	ctx.Info("autoController Any")
}

// Get 方法注册 Get /*路径。
func (*autoController) GetBy(ctx eudore.Context) interface{} {
	ctx.Debug("get")
	return "get autoController"
}

// GetInfoById 方法注册GET /info/:id/name 路由路径。
func (*autoController) GetInfoByIDName(ctx eudore.Context) interface{} {
	return ctx.GetParam("id")
}

// String 方法返回控制器名称，响应Router.AddController输出的名称。
func (*autoController) String() string {
	return "hello.autoController"
}

// Help 方法定义一个控制器本身的方法。
func (*autoController) Help(ctx eudore.Context) {}

func (*autoController) GetData(ctx eudore.Context) {}

func (*autoController) Inject(controller eudore.Controller, router eudore.Router) error {
	router = router.Group("")
	params := router.Params()
	*params = params.Set(eudore.ParamControllerGroup, "/auto")
	return eudore.ControllerInjectAutoRoute(controller, router)
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

func TestControllerCompose(t *testing.T) {
	app := eudore.NewApp()
	app.AddController(new(MethodController))
	app.AddController(new(RouteController))

	app.CancelFunc()
	app.Run()
}

// Controllerwebsite 是基础方法的控制器
type Controllerwebsite struct {
	eudore.ControllerAutoRoute
}
type MethodController struct {
	// Controllerwebsite 因名称后缀不为Controller所以不会注册Hello方法为路由。
	Controllerwebsite
}
type tableController struct {
	eudore.ControllerAutoRoute
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
func (ctl *Controllerwebsite) Get(ctx eudore.Context) {}

// Any 方法处理控制器全部请求。
func (ctl *MethodController) Any(ctx eudore.Context) {
	ctx.Debug("MethodController Any", ctl.Hello())
}

func (ctl tableController) ControllerName() string {
	return "tableController[Context]"
}

func (ctl *tableController) Hello() interface{} {
	return "hello eudore"
}

func (ctl *tableController) Any(ctx eudore.Context) {
	ctx.Debug("tableController Any", ctl.Hello())
}
