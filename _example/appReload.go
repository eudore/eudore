package main

import (
	"fmt"
	"github.com/eudore/eudore"
	"time"
)

type AppReload struct {
	*eudore.App
	*ConfigReload
}

type ConfigReload struct {
	Name string `alias:"name" json:"name"`
	Time string `alias:"time" json:"time"`
}

func main() {
	app := NewAppReload()
	app.Init()
	// 访问reload 触发重新加载
	app.AnyFunc("/reload", app.Init)
	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

func NewAppReload() *AppReload {
	conf := &ConfigReload{Name: "eudore"}
	return &AppReload{
		App: eudore.NewApp(
			// 使用读写路由核心，允许并发增删路由规则。
			eudore.NewRouterStd(eudore.NewRouterCoreLock(eudore.NewRouterCoreStd())),
			eudore.NewConfigEudore(conf),
		),
		ConfigReload: conf,
	}
}

// Init 方法加载配置并注册路由
func (app *AppReload) Init() error {
	app.Options(app.Parse())
	app.Time = time.Now().String()
	return app.AddController(NewUserReloadController(app))
}

type UserReloadController struct {
	Name   string
	Config eudore.Config
	eudore.ControllerAutoRoute
}

func NewUserReloadController(app *AppReload) eudore.Controller {
	return &UserReloadController{Name: app.ConfigReload.Name, Config: app.App.Config}
}

func (ctl UserReloadController) Any(ctx eudore.Context) interface{} {
	// 使用属性或Get获取数据，Get方法带锁。
	return fmt.Sprintf("name is %s at %v", ctl.Name, ctl.Config.Get("time"))
}
