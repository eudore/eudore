package middleware

import (
	"fmt"
	"github.com/eudore/eudore"
)

// NewRecoverFunc 函数创建一个错误捕捉中间件，并返回500。
func NewRecoverFunc() eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}
				ctx.WithField("error", "recover error").Fatal(err)
			}
		}()
		ctx.Next()
	}
}
