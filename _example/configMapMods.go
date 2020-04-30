package main

/*
enable获得到的数组为需要加载的模式，额外会加载为当前操作系统的名称的模式，如果是docker环境则加载docker模式。

然后会依次将mods.xxx的数据加载到配置中。

实现参考eudore.ConfigParseMods
*/

import (
	"github.com/eudore/eudore"
	"os"
)

var configmapfilepath = "example.json"

func main() {
	content := []byte(`{
	"keys.default": true,
	"keys.help": true,
	"mods.debug": {
		"keys.debug": true
	}
}
`)
	tmpfile, _ := os.Create(configmapfilepath)
	defer os.Remove(tmpfile.Name())
	tmpfile.Write(content)
	// 模拟docker系统模式
	os.Create("/.dockerenv")
	defer os.Remove("/.dockerenv")

	app := eudore.NewApp()
	app.Config.Set("keys.config", configmapfilepath)
	app.Config.Set("enable", []string{"debug"})
	app.Options(app.Parse())

	app.CancelFunc()
	app.Run()
}
