package main

/*
std 是否输出到os.Stdout标准输出。
Path 输出到文件的路径，如果path为空，std会设置成true
Level 日志输出级别，可以使用SetLevel来修改。
TimeFormat 日志时间格式化格式

LoggerStd的配置，可以使用*LoggerStdConfig或者map类型。

type LoggerStdConfig struct {
	Std        bool        `alias:"std"`
	Path       string      `alias:"path"`
	Level      LoggerLevel `alias:"level"`
	TimeFormat string      `alias:"timeformat"`
}
*/

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewCore()
	// 创建日志
	app.Logger = eudore.NewLoggerStd(map[string]interface{}{
		"std":        false,
		"path":       "",
		"Level":      "1",
		"TimeFormat": "Mon Jan 2 15:04:05 -0700 MST 2006",
	})

	app.Debug("debug")
	app.Info("info")
	app.Warning("warning")
	app.Error("error")
	app.SetLevel(eudore.LogDebug)
	app.Debug("debug")
	app.Info("info")

	app.Logger.Sync()
}
