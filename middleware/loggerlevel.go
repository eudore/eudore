package middleware

import (
	"github.com/eudore/eudore"
)

// NewLoggerLevelFunc 函数创建一个设置一次请求日志级别的中间件。
//
// 通过一个函数处理请求，返回一个0-4,代表日志级别Debug-Fatal,默认处理函数使用debug参数转换成日志级别数字。
func NewLoggerLevelFunc(fn func(ctx eudore.Context) int) eudore.HandlerFunc {
	if fn == nil {
		fn = func(ctx eudore.Context) int {
			level := ctx.GetQuery("debug")
			if level != "" {
				return eudore.GetStringInt(level)
			}
			return -1
		}
	}
	return func(ctx eudore.Context) {
		l := fn(ctx)
		if -1 < l && l < 5 {
			log := ctx.Logger().WithFields(nil, nil)
			log.SetLevel(eudore.LoggerLevel(l))
			ctx.SetLogger(log)
		}
	}
}
