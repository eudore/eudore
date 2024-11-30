package main

/*
创建daemon.Signal管理信号，给信号10注册Reload函数。
内置信号：
- 2 interrupt
- 12 restart
- 15 kill

使用 kill -10 {pid} 或curl 127.0.0.1:8088/api/reload，通过信号或Api方式触发Reload方法。
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
	app := &AppReload{
		App: eudore.NewApp(),
		ConfigReload: &ConfigReload{
			Name: "eudore",
			Time: time.Now().String(),
		},
	}
	app.SetValue(eudore.ContextKeyLogger, eudore.NewLoggerInit())
	app.SetValue(eudore.ContextKeyConfig, eudore.NewConfig(app.ConfigReload))
	app.ParseOption(
		daemon.NewParseCommand(),
		NewParseLogger(app.App),
		daemon.NewParseSignal(), // last parse
	)
	if app.Parse() != nil {
		return
	}

	// 注册系统信号reload
	d, ok := app.Value(eudore.ContextKeyDaemonSignal).(*daemon.Signal)
	if ok {
		d.Register(syscall.Signal(0x0a), app.Reload)
	}

	app.AddController(NewUserReloadController(app))
	// 注册api触发reload
	app.AnyFunc("/api/reload", func(ctx eudore.Context) {
		app.Reload(ctx.Context())
	})

	go func() {
		select {
		case <-app.Done():
		case <-time.After(60 * time.Second):
			app.CancelFunc()
		}
	}()

	app.Listen(":8088")
	app.Run()
}

// Reload 方法加载配置并注册路由
func (app *AppReload) Reload(context.Context) error {
	app.Time = time.Now().Format(time.DateTime)
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

func NewParseLogger(app *eudore.App) eudore.ConfigParseFunc {
	return func(context.Context, eudore.Config) error {
		app.SetValue(eudore.ContextKeyLogger, eudore.NewLogger(&eudore.LoggerConfig{
			Stdout:   true,
			StdColor: true,
			Path:     "daemon.log",
		}))
		return nil
	}
}
