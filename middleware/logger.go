package middleware

import (
	"github.com/eudore/eudore"
	"time"
)

// NewLoggerFunc 函数创建一个请求日志记录中间件。
func NewLoggerFunc(app *eudore.App) eudore.HandlerFunc {
	var params = []string{"action", "ram", "route", "controller"}
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
		existsAddParam(ctx, f, params)
		if requestId := ctx.GetHeader(eudore.HeaderXRequestID); len(requestId) > 0 {
			f["x-request-id"] = requestId
		}
		if parentId := ctx.GetHeader(eudore.HeaderXParentID); len(parentId) > 0 {
			f["x-parent-id"] = parentId
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

func existsAddParam(ctx eudore.Context, field eudore.Fields, names []string) {
	for _, name := range names {
		val := ctx.GetParam(name)
		if val != "" {
			field[name] = val
		}
	}
}
