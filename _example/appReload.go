package main

/*
添加配置解析NewParseDaemon函数，注册daemon.Signal管理信号。
然后给信号10注册Reload函数。
*/

import (
	"context"
	"fmt"
	"syscall"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/daemon"
)

type AppReload struct {
	*eudore.App
	*ConfigReload
}

type ConfigReload struct {
	Workdir string `json:"workdir" alias:"workdir"`
	Command string `json:"command" alias:"command"`
	Pidfile string `json:"pidfile" alias:"pidfile"`
	Name    string `alias:"name" json:"name"`
	Time    string `alias:"time" json:"time"`
}

func main() {
	conf := &ConfigReload{
		Name: "eudore",
		Time: time.Now().String(),
	}
	app := &AppReload{
		App:          eudore.NewApp(),
		ConfigReload: conf,
	}
	// 使用读写路由核心，允许并发增删路由规则。
	app.SetValue(eudore.ContextKeyRouter, eudore.NewRouter(eudore.NewRouterCoreLock(nil)))
	app.SetValue(eudore.ContextKeyConfig, eudore.NewConfig(conf))
	// 添加daemon对象
	app.ParseOption(daemon.NewParseDaemon(app.App))
	app.Parse()
	if app.Err() != nil {
		return
	}

	// 注册系统信号reload
	d, ok := app.Value(eudore.ContextKeyDaemonSignal).(*daemon.Signal)
	if ok {
		d.Register(syscall.Signal(0x0a), app.Reload)
	}
	// 注册api触发reload
	app.AnyFunc("/reload", app.Reload)

	app.AddController(NewUserReloadController(app))
	app.Listen(":8086")
	app.Run()
}

// Reload 方法加载配置并注册路由
func (app *AppReload) Reload(context.Context) error {
	app.Time = time.Now().String()
	app.Debug("app reload config")
	return nil
}

type UserReloadController struct {
	Name   string
	Config eudore.Config
	eudore.ControllerAutoRoute
}

func NewUserReloadController(app *AppReload) eudore.Controller {
	return &UserReloadController{Name: app.ConfigReload.Name, Config: app.App.Config}
}

func (ctl UserReloadController) Any(ctx eudore.Context) {
	// 使用属性或Get获取数据，Get方法带锁。
	ctx.WriteString(fmt.Sprintf("name is %s at %v", ctl.Name, ctl.Config.Get("time")))
}
