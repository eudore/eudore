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
		if requestID := ctx.GetHeader(eudore.HeaderXRequestID); len(requestID) > 0 {
			out = out.WithField("x-request-id", requestID)
		}
		if parentID := ctx.GetHeader(eudore.HeaderXParentID); len(parentID) > 0 {
			out = out.WithField("x-parent-id", parentID)
		}

		if 300 < status && status < 400 && status != 304 {
			out = out.WithField("location", ctx.Response().Header().Get(eudore.HeaderLocation))
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
