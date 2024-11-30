package main

import (
	"fmt"
	"errors"

	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewApp()
	app.AddController(
		NewAutoController(app, 0),
		NewAutoController(app, 1),
	)

	app.Set("name", "eudore")
	app.Listen(":8088")
	app.Run()
}

type autoController struct {
	eudore.ControllerAutoRoute
	Config eudore.Config
	// Database *sql.DB
}

func NewAutoController(app *eudore.App, id int) eudore.Controller {
	if id == 0 {
		// 创建控制器时返回error。
		return eudore.NewControllerError(&autoController{}, errors.New("int must grate 0"))
	}
	return &autoController{
		Config: app.Config,
	}
}

// Any 方法注册 Any /*路径。
func (ctl *autoController) Any(ctx eudore.Context) {
	ctx.Info("autoController Any name:", ctl.Config.Get("name"))
}

// Get 方法注册 Get /*路径。
func (ctl *autoController) GetBy(ctx eudore.Context) interface{} {
	ctx.Debug("get", ctl.Config.Get("name"))
	return "get autoController"
}

// GetInfoById 方法注册GET /info/:id 路由路径。
func (*autoController) GetInfoById(ctx eudore.Context) interface{} {
	return ctx.GetParam("id")
}

// Help 方法定义一个控制器本身的方法。
func (*autoController) Help(ctx eudore.Context) {}

func (*autoController) GetData(ctx eudore.Context) {}

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
