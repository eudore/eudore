package eudore_test

import (
	"errors"
	"io"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"testing"
)

type baseMethod struct{}

func (baseMethod) Any() {}

type mybGroupcontroller struct {
	eudore.ControllerBase
	baseMethod
}

type mybGroupController struct {
	eudore.ControllerBase
	baseMethod
}

type mysGroup struct {
	eudore.ControllerSingleton
	baseMethod
}
type mysGroupcontroller struct {
	eudore.ControllerSingleton
	baseMethod
}
type mysGroupController struct {
	eudore.ControllerSingleton
	baseMethod
}

func (ctl *mysGroupController) Init(ctx eudore.Context) error {
	if ctx.GetParam("*") == "init" {
		return errors.New("test error init")
	}
	return ctl.ControllerSingleton.Init(ctx)
}
func (ctl *mysGroupController) Release(ctx eudore.Context) error {
	if ctx.GetParam("*") == "release" {
		return errors.New("test error release")
	}
	return ctl.ControllerSingleton.Release(ctx)
}

func TestControllerGroup2(*testing.T) {
	app := eudore.NewCore()
	app.SetParam("controllergroup", "g1").AddController(new(mybGroupcontroller))
	app.AddController(new(mybGroupcontroller))
	app.AddController(new(mybGroupController))
	app.AddController(new(mysGroup))
	app.AddController(new(mysGroupcontroller))
	app.AddController(new(mysGroupController))
	app.Run()
}

type myexecConrtoller struct {
	eudore.ControllerSingleton
	baseMethod
}

func (ctl *myexecConrtoller) Init(ctx eudore.Context) error {
	if ctx.GetParam("*") == "init" {
		return errors.New("test error init")
	}
	return ctl.ControllerSingleton.Init(ctx)
}
func (ctl *myexecConrtoller) Release(ctx eudore.Context) error {
	if ctx.GetParam("*") == "release" {
		return errors.New("test error release")
	}
	return ctl.ControllerSingleton.Release(ctx)
}

func (ctl *myexecConrtoller) Error() error {
	return errors.New("test error")
}
func (ctl *myexecConrtoller) RenderError1() (interface{}, error) {
	return "hello", errors.New("test error")
}
func (ctl *myexecConrtoller) RenderError2() (interface{}, error) {
	return "hello", nil
}

func (ctl *myexecConrtoller) Context(eudore.Context) {
}
func (ctl *myexecConrtoller) Render() interface{} {
	return "hello"
}
func (ctl *myexecConrtoller) ContextRender(eudore.Context) interface{} {
	return "hello"
}
func (ctl *myexecConrtoller) ContextError1(eudore.Context) error {
	return errors.New("test error")
}
func (ctl *myexecConrtoller) ContextError2(eudore.Context) error {
	return nil
}
func (ctl *myexecConrtoller) ContextRenderError1(eudore.Context) (interface{}, error) {
	return "hello", errors.New("test error")
}
func (ctl *myexecConrtoller) ContextRenderError2(eudore.Context) (interface{}, error) {
	return "hello", nil
}

func (ctl *myexecConrtoller) MapString(map[string]interface{}) {
}
func (ctl *myexecConrtoller) MapStringRender(map[string]interface{}) interface{} {
	return "hello"
}
func (ctl *myexecConrtoller) MapStringError1(map[string]interface{}) error {
	return errors.New("test error")
}
func (ctl *myexecConrtoller) MapStringError2(map[string]interface{}) error {
	return nil
}
func (ctl *myexecConrtoller) MapStringRenderError1(map[string]interface{}) (interface{}, error) {
	return "hello", errors.New("test error")
}
func (ctl *myexecConrtoller) MapStringRenderError2(map[string]interface{}) (interface{}, error) {
	return "hello", nil
}
func TestControllerExtendExec2(*testing.T) {
	app := eudore.NewCore()
	app.AddController(new(eudore.ControllerBase))
	app.AddController(new(myexecConrtoller))
	app.SetParam("controllergroup", "name").SetParam("enable-route-extend", "0").AddController(new(mysGroupController))

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/init").Do()
	client.NewRequest("GET", "/release").Do()
	client.NewRequest("GET", "/name/init").Do()
	client.NewRequest("GET", "/name/release").Do()

	client.NewRequest("GET", "/error").Do()
	client.NewRequest("GET", "/render/error1").Do()
	client.NewRequest("GET", "/render/error2").Do()
	client.NewRequest("GET", "/context").Do()
	client.NewRequest("GET", "/context/error1").Do()
	client.NewRequest("GET", "/context/error2").Do()
	client.NewRequest("GET", "/context/render/error1").Do()
	client.NewRequest("GET", "/context/render/error2").Do()
	client.NewRequest("GET", "/map/string").Do()
	client.NewRequest("GET", "/map/string/render").Do()
	client.NewRequest("GET", "/map/string/error1").Do()
	client.NewRequest("GET", "/map/string/error2").Do()
	client.NewRequest("GET", "/map/string/render/error1").Do()
	client.NewRequest("GET", "/map/string/render/error2").Do()

	app.Renderer = func(eudore.Context, interface{}) error {
		return errors.New("test render error")
	}
	client.NewRequest("GET", "/map/string/render").Do()
	app.Binder = func(eudore.Context, io.Reader, interface{}) error {
		return errors.New("test binder error")
	}
	client.NewRequest("GET", "/error").Do()
	client.NewRequest("GET", "/render/error1").Do()
	client.NewRequest("GET", "/render/error2").Do()
	client.NewRequest("GET", "/render").Do()
	client.NewRequest("GET", "/context").Do()
	client.NewRequest("GET", "/context/render").Do()
	client.NewRequest("GET", "/context/error1").Do()
	client.NewRequest("GET", "/context/error2").Do()
	client.NewRequest("GET", "/context/render/error1").Do()
	client.NewRequest("GET", "/context/render/error2").Do()
	client.NewRequest("GET", "/map/string").Do()
	client.NewRequest("GET", "/map/string/render").Do()
	client.NewRequest("GET", "/map/string/error1").Do()
	client.NewRequest("GET", "/map/string/error2").Do()
	client.NewRequest("GET", "/map/string/render/error1").Do()
	client.NewRequest("GET", "/map/string/render/error2").Do()

	for client.Next() {
		app.Error(client.Error())
	}
	app.Run()
}
