package main

/*
在Config默认解析函数ConfigAllParseFunc中包含eudore.ConfigParseJSON函数，用于解析json文件。
*/

import (
	"github.com/eudore/eudore"
	"os"
)

var filepath = "example.json"

func main() {
	// 创建一个测试配置文件
	content := []byte(`{
	"default": true,
	"help": true,
	"mods.debug": {
		"debug": true
	}
}
`)
	tmpfile, _ := os.Create(filepath)
	defer os.Remove(tmpfile.Name())
	tmpfile.Write(content)

	app := eudore.NewApp()
	// 设置config路径
	app.Set("config", filepath)
	app.Set("help", true)
	app.Options(app.Parse())
	// app.CancelFunc()
	app.Run()
}
