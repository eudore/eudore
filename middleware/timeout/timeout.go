package timeout


import (
	"time"
	"github.com/eudore/eudore"
)

func NewTimeout(t time.Duration) func(eudore.HandlerFunc) {
	return func(ctx eudore.Context) {
		finish := make(chan struct{})

		go func() {
			ctx.Next()
			finish <- struct{}{}
		}()

		select {
		case <-time.After(t):
			ctx.WriteHeader(504)
			ctx.WriteString("timeout")
			ctx.End()
		case <-finish:
		}
	}
}