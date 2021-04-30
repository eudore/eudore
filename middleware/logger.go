package middleware

import (
	"github.com/eudore/eudore"
	"time"
)

// NewLoggerFunc 函数创建一个请求日志记录中间件。
//
// app参数传入*eudore.App需要使用其Logger输出日志，paramsh获取Context.Params如果不为空则添加到输出日志条目中
//
// 状态码如果为40x、50x输出日志级别为Error。
func NewLoggerFunc(app *eudore.App, params ...string) eudore.HandlerFunc {
	keys := []string{"method", "path", "realip", "proto", "host", "status", "request-time", "size"}
	headerkeys := [...]string{eudore.HeaderXRequestID, eudore.HeaderXTraceID, eudore.HeaderLocation}
	headernames := [...]string{"x-request-id", "x-trace-id", "location"}
	return func(ctx eudore.Context) {
		now := time.Now()
		ctx.Next()
		status := ctx.Response().Status()
		// 连续WithField保证field顺序
		out := app.WithFields(keys, []interface{}{
			ctx.Method(), ctx.Path(), ctx.RealIP(), ctx.Request().Proto,
			ctx.Host(), status, time.Now().Sub(now).String(), ctx.Response().Size(),
		})

		for _, param := range params {
			val := ctx.GetParam(param)
			if val != "" {
				out = out.WithField(param, val)
			}
		}

		if xforward := ctx.GetHeader(eudore.HeaderXForwardedFor); len(xforward) > 0 {
			out = out.WithField("x-forward-for", xforward)
		}
		headers := ctx.Response().Header()
		for i, key := range headerkeys {
			val := headers.Get(key)
			if val != "" {
				out = out.WithField(headernames[i], val)
			}
		}

		if status < 500 {
			out.Info()
		} else {
			if err := ctx.Err(); err != nil {
				out = out.WithField("error", err.Error())
			}
			out.Error()
		}
	}
}
