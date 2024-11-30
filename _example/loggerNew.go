package main

/*
// LoggerConfig 定义loggerStd配置信息。
type LoggerConfig struct {
	// 设置额外的LoggerHandler，和配置初始化创建的Handlers排序后处理LoggerEntry。
	Handlers []LoggerHandler `alias:"handlers" json:"-" xml:"-" yaml:"-"`
	// 设置日志输出级别。
	Level LoggerLevel `alias:"level" json:"level" xml:"level" yaml:"level"`
	// 是否记录调用者信息。
	Caller bool `alias:"caller" json:"caller" xml:"caller" yaml:"caller"`
	// 设置Entry输出格式，默认值为json，
	// 如果为json/text启用NewLoggerFormatterJSON/NewLoggerFormatterText。
	Formatter string `alias:"formater" json:"formater" xml:"formater" yaml:"formater"`
	// 设置日志时间输出格式，默认值为DefaultLoggerFormatterFormatTime或time.RFC3339。
	TimeFormat string `alias:"timeformat" json:"timeformat" xml:"timeformat" yaml:"timeformat"`
	// 设置Entry过滤规则；如果非空启用NewLoggerHookFilter。
	HookFilter [][]string `alias:"hoolfilter" json:"hoolfilter" xml:"hoolfilter" yaml:"hoolfilter"`
	// 是否处理Fatal级别日志，调用应用结束方法；如果为true启用NewLoggerHookMeta。
	HookFatal bool `alias:"hookfatal" json:"hookfatal" xml:"hookfatal" yaml:"hookfatal"`
	// 是否采集Meta信息，记录日志count、size；如果为true启用NewLoggerHookFatal。
	HookMeta bool `alias:"hookmeta" json:"hookmeta" xml:"hookmeta" yaml:"hookmeta"`
	// 是否输出日志到os.Stdout标准输出流；如果存在Env EnvEudoreDaemonEnable时会强制修改为false；
	// 如果为true启动NewLoggerWriterStdout。
	Stdout bool `alias:"stdout" json:"stdout" xml:"stdout" yaml:"stdout"`
	// 是否输出日志时使用彩色Level，默认在windows系统下禁用。
	StdColor bool `alias:"stdcolor" json:"stdcolor" xml:"stdcolor" yaml:"stdcolor"`
	// 设置日志文件输出路径；如果非空启用NewLoggerWriterFile，
	// 如果Path包含关键字yyyy/mm/dd/hh或MaxSize非0则改为启用NewLoggerWriterRotate。
	Path string `alias:"path" json:"path" xml:"path" yaml:"path" description:"Output file path."`
	// 设置日志文件滚动size，在文件名后缀之前添加索引值。
	MaxSize uint64 `alias:"maxsize" json:"maxsize" xml:"maxsize" yaml:"maxsize" description:"roatte file max size"`
	// 设置日志文件最多保留天数，如果非0使用hookFileRecycle。
	MaxAge int `alias:"maxage" json:"maxage" xml:"maxage" yaml:"maxage"`
	// 设置日志文件最多保留数量，如果非0使用hookFileRecycle。
	MaxCount int `alias:"maxcount" json:"maxcount" xml:"maxcount" yaml:"maxcount"`
	// 设置日志文件软链接名称，如果非空使用hookFileLink。
	Link string `alias:"link" json:"link" xml:"link" yaml:"link" description:"Output file link to path."`
}

*/

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyLogger, eudore.NewLogger(&eudore.LoggerConfig{
		Stdout:     true,
		StdColor:   true,
		Path:       "app.log",
		Level:      eudore.LoggerInfo,
		TimeFormat: "Mon Jan 2 15:04:05 -0700 MST 2006",
		Caller:     true,
	}))

	app.Debug("debug")
	app.Info("info")
	app.Warning("warning")
	app.Error("error")
	app.SetLevel(eudore.LoggerDebug)
	app.Debug("debug")
	app.WithField("depth", "disable").Info("info")

	// WithField方法参数为logger=true会返回一个logger深拷贝
	logout := app.WithField("caller", "mylogout").WithField("logger", true)
	logout.WithField("level", "debug").Debug("debug")
	logout.WithField("level", "info").Info("info")
	logout.WithField("level", "warning").Warning("warning")
	logout.WithField("context", app.Context).WithField("level", "error").Error("error")

	app.CancelFunc()
	app.Run()
}
