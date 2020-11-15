package main

/*
enable获得到的数组为需要加载的模式，额外会加载为当前操作系统的名称的模式，如果是docker环境则加载docker模式。

然后会依次将mods.xxx的数据加载到配置中。

实现参考eudore.ConfigParseMods
*/

import (
	"encoding/json"
	"fmt"
	"github.com/eudore/eudore"
	"os"
)

type conf struct {
	Keys      map[string]interface{} `alias:"keys" json:"keys"`
	Config    string                 `alias:"config" json:"config"`
	Component *componentConfig       `alias:"component" json:"component"`
	Enable    []string               `alias:"enable" json:"enable"`
	Mods      map[string]*conf       `alias:"mods" json:"mods"`
}
type componentConfig struct {
	Logger *eudore.LoggerStdConfig `json:"logger" alias:"logger"`
	Server *eudore.ServerStdConfig `json:"server" alias:"server"`
}

var configfilepath = "example.json"

func main() {
	content := []byte(`{
	"keys": {
		"default": true,
		"help": true
	},
	"mods": {
		"debug": {
			"keys": {
				"debug": true
			},
			"component":{
				"server":{
					"readtimeout": "12s",
					"writetimeout": 3000000000
				}
			}
		}
	}
}
`)
	tmpfile, _ := os.Create(configfilepath)
	defer os.Remove(tmpfile.Name())
	tmpfile.Write(content)

	app := eudore.NewApp(eudore.NewConfigEudore(new(conf)))
	app.Config.Set("config", configfilepath)
	app.Config.Set("enable", []string{"debug"})
	app.Options(app.Parse())
	app.Info(app.Get("component.server.readtimeout"))

	indent, err := json.MarshalIndent(app.Config, "", "\t")
	fmt.Println(string(indent), err)
	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
