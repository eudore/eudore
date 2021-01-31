package main

import (
	"github.com/eudore/eudore"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.StandardLogger()
	log.Formatter = new(logrus.JSONFormatter)
	app := eudore.NewApp(NewLoggerLogrus(log))

	app.WithFields([]string{"animal", "number", "size"}, []interface{}{"walrus", 1, 10}).Info("A walrus appears")

	app.Debug("debug")
	app.Info("info")
	app.Warning("warning")
	app.Error("error")
	app.SetLevel(eudore.LogDebug)
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

func NewLoggerLogrus(log *logrus.Logger) eudore.Logger {
	return Logger{log}
}

type Logger struct {
	*logrus.Logger
}
type Entry struct {
	*logrus.Entry
}

func (log Logger) WithField(key string, value interface{}) eudore.Logger {
	return Entry{log.Logger.WithField(key, value)}
}

func (log Logger) WithFields(key []string, value []interface{}) eudore.Logger {
	filed := make(logrus.Fields, len(key))
	for i := range key {
		filed[key[i]] = value[i]
	}
	return Entry{log.Logger.WithFields(filed)}
}

func (log Logger) SetLevel(level eudore.LoggerLevel) {
	log.Logger.SetLevel(logrus.Level(5 - level))
}

func (log Logger) Sync() error {
	return nil
}

func (log Entry) WithField(key string, value interface{}) eudore.Logger {
	return Entry{log.Entry.WithField(key, value)}
}

func (log Entry) WithFields(key []string, value []interface{}) eudore.Logger {
	filed := make(logrus.Fields, len(key))
	for i := range key {
		filed[key[i]] = value[i]
	}
	return Entry{log.Logger.WithFields(filed)}
}

func (log Entry) SetLevel(level eudore.LoggerLevel) {
}

func (log Entry) Sync() error {
	return nil
}
