package middleware

import (
	"fmt"
	"github.com/eudore/eudore"
	"runtime"
	"strings"
)

// RecoverDepth 定义默认显示栈最大层数。
var RecoverDepth = 20

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

			// Ask runtime.Callers for up to 10 pcs, including runtime.Callers itself.
			pc := make([]uintptr, RecoverDepth)
			n := runtime.Callers(0, pc)
			if n == 0 {
				// No pcs available. Stop now.
				// This can happen if the first argument to runtime.Callers is large.
				ctx.WithField("error", "recover error").Fatal(err)
				return
			}

			pc = pc[:n] // pass only valid pcs to runtime.CallersFrames
			frames := runtime.CallersFrames(pc)
			stack := make([]string, 0, RecoverDepth)

			// Loop to get frames.
			// A fixed number of pcs can expand to an indefinite number of Frames.
			frame, more := frames.Next()
			for more && frame.Function != "runtime.main" {
				// remove prefix
				pos := strings.Index(frame.File, "src")
				if pos >= 0 {
					frame.File = frame.File[pos+4:]
				}
				pos = strings.LastIndex(frame.Function, "/")
				if pos >= 0 {
					frame.Function = frame.Function[pos+1:]
				}
				stack = append(stack, fmt.Sprintf("%s:%d %s", frame.File, frame.Line, frame.Function))

				frame, more = frames.Next()
			}

			ctx.WithField("error", "recover error").WithField("stack", stack).Error(err)
			ctx.WriteHeader(500)
			ctx.Render(map[string]interface{}{
				"error":        err.Error(),
				"stack":        stack,
				"status":       500,
				"x-request-id": ctx.RequestID(),
			})
		}()
		ctx.Next()
	}
}
