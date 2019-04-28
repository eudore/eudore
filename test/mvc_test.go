package test

import (
	// "fmt"
	// "reflect"
	"testing"
	// "github.com/kr/pretty"
	"github.com/eudore/eudore"
)

type BaseController struct{
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

func (*BaseController) GetIndex() {}
func (*BaseController) GetContent() {}
func (*BaseController) ControllerRoute(name string) string {
	m := map[string]string{
		"Any": "/*name",
		"Get": "/*",
		"GetIdById": "/:id",
		"GetInfoByIdName": "/info/:id/:name",
		"GetIndex": "/index",
		"GetContent": "/content",
	}
	return m[name]
}

func TestMvc1(*testing.T) {
	app := eudore.NewCore()
	app.AddController(&BaseController{})
	app.Listen(":8085")
	app.Run()
}
