package main

import (
	"github.com/eudore/eudore"
)

type MyController struct{
	eudore.ControllerData
}

func (c *MyController) Init(ctx eudore.Context) error {
	c.Context = ctx
	return nil
}

func (*MyController) Any() {}

func (c *MyController) Get() {
	c.Debug("---")
}

func (c *MyController) GetIdById(id int) {
	c.Debug("id ", id)
	c.WriteRender(id)
}

func (c *MyController) GetInfoByIdName(id int, name string) {
	c.WriteRender(id)
	c.WriteString(name)
}

func (*MyController) GetIndex() {}
func (*MyController) GetContent() {}
func (*MyController) ControllerRoute() map[string]string {
	m := map[string]string{
		"Any": "/*name",
		"Get": "/*",
		"GetIdById": "/:id",
		"GetInfoByIdName": "/info/:id/:name",
		"GetIndex": "/index",
		"GetContent": "/content",
	}
	return m
}


func main() {
	app := eudore.NewCore()
	eudore.Set(app.Router, "print", app.Logger.Debug)
	app.AddController(&MyController{})

	app.Listen(":8088")
	app.Run()
}
