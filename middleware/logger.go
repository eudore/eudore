package middleware

import (
	"github.com/eudore/eudore"
	"time"
)

// NewLoggerFunc 函数创建一个请求日志记录中间件。
func NewLoggerFunc() eudore.HandlerFunc {
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
		if requestId := ctx.GetHeader(eudore.HeaderXRequestID); len(requestId) > 0 {
			f["x-request-id"] = requestId
		}
		if parentId := ctx.GetHeader(eudore.HeaderXParentID); len(parentId) > 0 {
			f["x-parent-id"] = parentId
		}
		if action := ctx.GetParam(eudore.ParamAction); len(action) > 0 {
			f["action"] = action
		}
		if ram := ctx.GetParam(eudore.ParamRam); len(ram) > 0 {
			f["ram"] = ram
		}
		if route := ctx.GetParam(eudore.ParamRoute); len(route) > 0 {
			f["route"] = route
		}
		// if routes := ctx.Params()[eudore.ParamRoutes]; len(routes) > 0 {
		// 	f["routes"] = strings.Join(routes, " ")
		// }
		if 300 < status && status < 400 && status != 304 {
			f["location"] = ctx.Response().Header().Get(eudore.HeaderLocation)
		}
		if status < 400 {
			ctx.WithFields(f).Info()
		} else {
			ctx.WithFields(f).Error()
		}
	}
}
