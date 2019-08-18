package middleware

import (
	"github.com/eudore/eudore"
	"time"
)

// NewTimeoutFunc 未实现
//
// NewTimeoutFunc 函数创建一个处理超时中间件。
func NewTimeoutFunc(t time.Duration) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		finish := make(chan struct{})

		go func(ctx eudore.Context) {
			ctx.Next()
			finish <- struct{}{}
		}(ctx)

		select {
		case <-time.After(t):
			ctx.WriteHeader(504)
			ctx.WriteString("timeout")
			ctx.End()
		case <-finish:
		}
	}
}
