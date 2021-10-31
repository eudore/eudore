package main

/*
在Config默认解析函数ConfigAllParseFunc中包含eudore.NewConfigParseJSON("config")函数，用于解析json文件。
*/

import (
	"github.com/eudore/eudore"
	"os"
)

var filepath = "example.json"
var content = []byte(`{
	"default": true,
	"help": true,
	"mods.debug": {
		"debug": true
	}
}`)

func main() {
	// 创建临时文件
	tmpfile, _ := os.Create(filepath)
	defer os.Remove(tmpfile.Name())
	tmpfile.Write(content)

	app := eudore.NewApp()
	// 设置config路径
	app.Set("config", []string{filepath, "example", "/dev/null"})
	app.Set("help", true)
	app.Options(app.Parse())
	app.Debug(app.Get(""))
	app.CancelFunc()
	app.Run()
}
