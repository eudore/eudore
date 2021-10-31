package main

/*
eudore.NewConfigEudore.Get/Set方法会根据key中'.'做为分隔符，基于路径层次遍历访问对象的子属性。

app.Get("component.logger.path")
c.Component.Logger.Path
*/

import (
	"github.com/eudore/eudore"
	"time"
)

type conf struct {
	Keys      map[string]interface{} `alias:"keys" json:"keys"`
	Config    string                 `alias:"config" json:"config"`
	Component *componentConfig       `alias:"component" json:"component"`
}
type componentConfig struct {
	Logger *eudore.LoggerStdConfig `json:"logger" alias:"logger"`
	Server *eudore.ServerStdConfig `json:"server" alias:"server"`
}

func main() {
	c := &conf{
		Keys: map[string]interface{}{
			"default": true,
			"help":    true,
		},
		Component: &componentConfig{
			Logger: &eudore.LoggerStdConfig{
				Std:     true,
				Path:    "app.log",
				MaxSize: 50 << 20,
			},
			Server: &eudore.ServerStdConfig{
				ReadTimeout:  eudore.TimeDuration(time.Second * 12),
				WriteTimeout: eudore.TimeDuration(time.Second * 3),
			},
		},
	}

	app := eudore.NewApp(eudore.NewConfigEudore(c))
	app.Info(app.Get("component.server.readtimeout"), c.Component.Server.ReadTimeout)
	app.Info(app.Get("component.logger.path"), c.Component.Logger.Path)

	app.Listen(":8088")
	app.CancelFunc()
	app.Run()
}
