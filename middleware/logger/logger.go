package logger

import (
	"time"
	"strings"
	"github.com/eudore/eudore"
)


type Logger struct {
	GetId func() string
}
func NewLogger(fn func() string)  *Logger {
	return &Logger{
		GetId: fn,
	}
}

func (l *Logger) Handle(ctx eudore.Context) {
	now := time.Now()
	// init request id
	requestId := ctx.GetHeader(eudore.HeaderXRequestID)
	if len(requestId) == 0 {
		requestId = l.GetId()
		ctx.Request().Header().Add(eudore.HeaderXRequestID, requestId)
	}
	f := eudore.Fields{
		"x-request-id":	ctx.RequestID(),
		"method":		ctx.Method(),
		"path":				ctx.Path(),
		"remote":			ctx.RemoteAddr(),
		"proto":			ctx.Request().Proto(),
		"host":				ctx.Host(),
	}
	ctx.Next()
	f["status"] = ctx.Response().Status()
	f["time"] = time.Now().Sub(now).String()
	if parentId := ctx.GetHeader(eudore.HeaderXParentID);len(parentId) > 0 {
		f["x-parent-id"] = parentId
	}
	if action := ctx.GetParam(eudore.ParamAction);len(action) > 0 {
		f["action"] = action
	}
	if ram := ctx.GetParam(eudore.ParamRam); len(ram) > 0 {
		f["ram"] = ram
	}
	if routes := ctx.Params()[eudore.ParamRoutes]; len(routes) > 0 {
		f["routes"] = strings.Join(routes, " ")
	}
	ctx.WithFields(f).Info()
}