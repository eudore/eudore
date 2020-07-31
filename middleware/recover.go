package middleware

import (
	"fmt"
	"github.com/eudore/eudore"
)

// NewRecoverFunc 函数创建一个错误捕捉中间件，并返回500。
func NewRecoverFunc() eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		defer func() {
			r := recover()
			if r == nil {
				return
			}

			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("%v", r)
			}
			stack := eudore.GetPanicStack(5)
			ctx.WithField("error", "recover error").WithField("stack", stack).Error(err)

			if ctx.Response().Size() == 0 {
				ctx.WriteHeader(500)
			}
			ctx.Render(map[string]interface{}{
				"error":        err.Error(),
				"stack":        stack,
				"status":       ctx.Response().Status(),
				"x-request-id": ctx.RequestID(),
			})
		}()
		ctx.Next()
	}
}
