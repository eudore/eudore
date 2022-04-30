package main

/*
日志field脱敏只针对指定field字段进行脱敏，对于message正则脱敏过于消耗性能。

本例子作为演示，实现eudore.LoggerStdData接口，捕捉到*eudore.LoggerStd对象写入，对部分数据修改实现指定字段脱敏，也可以实现自定义日志处理。
*/

import (
	"github.com/eudore/eudore"
	"strings"
)

func main() {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyLogger, NewLoggerSensitive(nil))
	app.WithField("username", "eudore").WithField("phone", "15057135056").Info()
	app.WithField("username", "eudore").WithField("email", "eudore@eudore.cn").Info()
	app.CancelFunc()
	app.Run()
}

type LoggerSensitive struct {
	eudore.LoggerStdData
}

func NewLoggerSensitive(data eudore.LoggerStdData) eudore.Logger {
	if data == nil {
		data = eudore.NewLoggerStdDataJSON(nil)
	}
	data = LoggerSensitive{data}
	return eudore.NewLoggerStd(data)
}

func (log LoggerSensitive) GetLogger() *eudore.LoggerStd {
	entry := log.LoggerStdData.GetLogger()
	_, ok := entry.LoggerStdData.(LoggerSensitive)
	if !ok {
		entry.LoggerStdData = LoggerSensitive{entry.LoggerStdData}
	}
	return entry
}

func (log LoggerSensitive) PutLogger(entry *eudore.LoggerStd) {
	// 脱敏处理
	for i, key := range entry.Keys {
		for _, name := range []string{"username", "password", "ipcard", "email", "phone"} {
			if key == name {
				val, ok := entry.Vals[i].(string)
				if ok {
					entry.Vals[i] = HandlerSensitive(key, val)
				}
				break
			}
		}
	}
	log.LoggerStdData.PutLogger(entry)
}

func HandlerSensitive(key, val string) string {
	length := len(val)
	switch {
	case key == "phone" && length == 11:
		return val[:3] + "****" + val[7:]
	case key == "email" && strings.Contains(val, "@"):
		index := strings.IndexByte(val, '@')
		return val[:3] + "***" + val[index:]
	case length < 1:
		return "***"
	case length == 2:
		return val[:1] + "*"
	case length == 3:
		return val[:1] + "*" + val[2:3]
	case length == 4:
		return val[:1] + "**" + val[3:4]
	default:
		return val[:2] + "***" + val[length-2:]
	}
}
