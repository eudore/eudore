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
	eudore.ControllerAutoRoute
	baseMethod
}

type mybGroupController struct {
	eudore.ControllerAutoRoute
	baseMethod
}

type mysGroup struct {
	eudore.ControllerAutoRoute
	baseMethod
}
type mysGroupcontroller struct {
	eudore.ControllerAutoRoute
	baseMethod
}
type mysGroupController struct {
	eudore.ControllerAutoRoute
	baseMethod
}

func (ctl *mysGroupController) InfoBy() {}

func (ctl *mysGroupController) Error(*mybGroupcontroller) {}

func TestControllerGroup2(*testing.T) {
	app := eudore.NewApp()
	app.Params().Set("controllergroup", "/g1")
	app.AddController(new(mybGroupcontroller))
	app.AddController(new(mybGroupcontroller))
	app.AddController(new(mybGroupController))
	app.AddController(new(mysGroup))
	app.AddController(new(mysGroupcontroller))
	app.AddController(new(mysGroupController))
	app.CancelFunc()
	app.Run()
}

type myexecConrtoller struct {
	eudore.ControllerAutoRoute
	baseMethod
}

func (ctl *myexecConrtoller) Error() error {
	return errors.New("test error myexecConrtoller.Error")
}
func (ctl *myexecConrtoller) RenderError1() (interface{}, error) {
	return "hello", errors.New("test error RenderError1")
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
	app := eudore.NewApp()
	app.AddController(new(eudore.ControllerAutoRoute))
	app.AddController(new(myexecConrtoller))
	app.Group(" controllergroup=name").AddController(new(mysGroupController))

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/init").Do()
	client.NewRequest("GET", "/release").Do()
	client.NewRequest("GET", "/name/info/init").Do()
	client.NewRequest("GET", "/name/info/release").Do()
	client.NewRequest("GET", "/name/info").Do()

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

	app.CancelFunc()
	app.Run()
}
