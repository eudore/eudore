package main

/*
LoggerStdConfig 定义loggerStd配置信息。
Writer 设置日志输出流，如果为空会使用Std和Path创建一个LoggerWriter。
Std 是否输出日志到os.Stdout标准输出流。
Path 指定文件输出路径,如果为空强制指定Std为true。
MaxSize 指定文件切割大小，需要Path中存在index字符串,用于替换成切割文件索引。
Link 如果非空会作为软连接的目标路径。
Level 日志输出级别。
TimeFormat 日志输出时间格式化格式。
FileLine 是否输出调用日志输出的函数和文件位置

LoggerStd的配置，可以使用*LoggerStdConfig或者map类型。

type LoggerStdConfig struct {
	Writer     LoggerWriter `json:"-" alias:"writer"`
	Std        bool         `json:"std" alias:"std"`
	Path       string       `json:"path" alias:"path"`
	MaxSize    uint64       `json:"maxsize" alias:"maxsize"`
	Link       string       `json:"link" alias:"link"`
	Level      LoggerLevel  `json:"level" alias:"level"`
	TimeFormat string       `json:"timeformat" alias:"timeformat"`
	FileLine   bool         `json:"fileline" alias:"fileline"`
}
*/

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyLogger, eudore.NewLoggerStd(map[string]interface{}{
		"std":        false,
		"path":       "",
		"Level":      "1",
		"TimeFormat": "Mon Jan 2 15:04:05 -0700 MST 2006",
		"FileLine":   true,
	}))

	app.Debug("debug")
	app.Info("info")
	app.Warning("warning")
	app.Error("error")
	app.SetLevel(eudore.LogDebug)
	app.Debug("debug")
	app.WithField("depth", "disable").Info("info")

	// WithFields方法参数为nil会返回一个logout深拷贝
	logout := app.WithField("caller", "mylogout").WithFields(nil, nil)
	logout.WithField("level", "debug").Debug("debug")
	logout.WithField("level", "info").Info("info")
	logout.WithField("level", "warning").Warning("warning")
	logout.WithField("context", app.Context).WithField("context", app.Context).WithField("level", "error").Error("error")

	app.CancelFunc()
	app.Run()
}
