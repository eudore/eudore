package test

import (
	// "fmt"
	// "reflect"
	"testing"
	// "github.com/kr/pretty"
	"github.com/eudore/eudore"
)

type BaseController struct {
	eudore.ControllerSession
}

func (c *BaseController) Init(ctx eudore.Context) error {
	c.Context = ctx
	return nil
}
func (*BaseController) Any() {}
func (c *BaseController) Get() {
	c.Debug("---")
}
func (c *BaseController) GetIdById(id int) {
	c.Debug("id", id)
	c.WriteRender(id)
}

func (c *BaseController) GetInfoByIdName(id int, name string) {
	c.WriteRender(id)
	c.WriteString(name)
}

func (*BaseController) GetIndex()   {}
func (*BaseController) GetContent() {}
func (*BaseController) ControllerRoute() map[string]string {
	m := map[string]string{
		"Any":             "/*name",
		"Get":             "/*",
		"GetIdById":       "/:id",
		"GetInfoByIdName": "/info/:id/:name",
		"GetIndex":        "/index",
		"GetContent":      "/content",
	}
	return m
}

func h1(ctx eudore.Context) {
	ctx.WriteString("ControllerSession")
}

func TestMvc1(*testing.T) {
	app := eudore.NewCore()
	eudore.Set(app.Router, "debug", app.Logger.Debug)
	app.AddController(&BaseController{})

	config := &eudore.RouterConfig{
		Routes: []*eudore.RouterConfig{
			&eudore.RouterConfig{
				Method:  "GET",
				Path:    "/11",
				Handler: h1,
			},
			&eudore.RouterConfig{
				Method:  "GET",
				Path:    "/12",
				Handler: h1,
			},
		},
	}
	config.Inject(app.Router.Group(""))
	app.Listen(":8085")
	app.Run()
}
