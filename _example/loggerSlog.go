package main

import (
	"context"
	"os"

	"github.com/eudore/eudore"
	"log/slog"
)

func main() {
	app := eudore.NewApp()

	log := slog.New(slog.NewTextHandler(os.Stderr, nil))
	app.SetValue(eudore.ContextKeyLogger, NewLoggerWithSlog(log))

	app.WithFields([]string{"animal", "number", "size"}, []interface{}{"walrus", 1, 10}).Info("A walrus appears")

	app.Debug("debug")
	app.Info("info")
	app.Warning("warning")
	app.Error("error")
	app.SetLevel(eudore.LoggerDebug)
	app.Debug("debug")
	app.WithField("depth", "disable").Info("info")

	// logrus WithFields方法每次都返回一个深拷贝，不需要使用为nil会返回一个logout深拷贝
	logout := app.WithField("caller", "mylogout") // .WithFields(nil)
	logout.WithField("level", "debug").Debug("debug")
	logout.WithField("level", "info").Info("info")
	logout.WithField("level", "warning").Warning("warning")
	logout.WithField("level", "error").Error("error")

	app.CancelFunc()
	app.Run()
}

func NewLoggerWithSlog(log *slog.Logger) eudore.Logger {
	return eudore.NewLogger(&eudore.LoggerConfig{
		// 禁用全部选项 防止创建默认Handler
		Stdout:    false,
		HookFatal: false,
		HookMeta:  false,
		Formatter: "disable",
		// 指定Handler
		Handlers: []eudore.LoggerHandler{
			&LoggerHandlerSlog{log},
		},
	})
}

type LoggerHandlerSlog struct {
	Logger *slog.Logger
}

// HandlerPriority 方法返回Handler处理顺序优先级，小值优先。
func (log *LoggerHandlerSlog) HandlerPriority() int {
	return 0
}

// HandlerEntry 方法将eudore.LoggerEntry 使用Slog输出。
func (log *LoggerHandlerSlog) HandlerEntry(entry *eudore.LoggerEntry) {
	if entry.Level == eudore.LoggerDiscard {
		return
	}

	ctx := context.Background()
	attrs := make([]slog.Attr, 0, len(entry.Keys))
	for i, key := range entry.Keys {
		if key == "context" {
			// 查找context.Context 通常链路追踪需要传递。
			c, ok := entry.Vals[i].(context.Context)
			if ok {
				ctx = c
				continue
			}
		}
		attrs = append(attrs, slog.Any(entry.Keys[i], entry.Vals[i]))
	}
	log.Logger.LogAttrs(ctx, levelMapping[entry.Level], entry.Message, attrs...)
}

var levelMapping = [...]slog.Level{
	slog.LevelDebug,
	slog.LevelInfo,
	slog.LevelWarn,
	slog.LevelError,
	slog.LevelError,
}
