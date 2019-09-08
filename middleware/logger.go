package middleware

import (
	"github.com/eudore/eudore"
	"time"
)

// NewLoggerFunc 函数创建一个请求日志记录中间件。
func NewLoggerFunc(app *eudore.App, params ...string) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		now := time.Now()
		f := eudore.Fields{
			"method": ctx.Method(),
			"path":   ctx.Path(),
			"remote": ctx.RealIP(),
			"proto":  ctx.Request().Proto(),
			"host":   ctx.Host(),
		}
		ctx.Next()
		status := ctx.Response().Status()
		f["status"] = status
		f["time"] = time.Now().Sub(now).String()
		f["size"] = ctx.Response().Size()

		for _, param := range params {
			val := ctx.GetParam(param)
			if val != "" {
				f[param] = val
			}
		}

		if requestID := ctx.GetHeader(eudore.HeaderXRequestID); len(requestID) > 0 {
			f["x-request-id"] = requestID
		}
		if parentID := ctx.GetHeader(eudore.HeaderXParentID); len(parentID) > 0 {
			f["x-parent-id"] = parentID
		}

		if 300 < status && status < 400 && status != 304 {
			f["location"] = ctx.Response().Header().Get(eudore.HeaderLocation)
		}
		if status < 400 {
			app.Logger.WithFields(f).Info()
		} else {
			app.Logger.WithFields(f).Error()
		}
	}
}
