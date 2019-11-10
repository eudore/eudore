package middleware

import (
	"fmt"
	"github.com/eudore/eudore"
	"runtime"
	"strings"
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

			// Ask runtime.Callers for up to 10 pcs, including runtime.Callers itself.
			pc := make([]uintptr, 10)
			n := runtime.Callers(0, pc)
			if n == 0 {
				// No pcs available. Stop now.
				// This can happen if the first argument to runtime.Callers is large.
				ctx.WithField("error", "recover error").Fatal(err)
				return
			}

			pc = pc[:n] // pass only valid pcs to runtime.CallersFrames
			frames := runtime.CallersFrames(pc)
			stack := make([]string, 0, 10)

			// Loop to get frames.
			// A fixed number of pcs can expand to an indefinite number of Frames.
			frame, more := frames.Next()
			for more {
				if strings.HasPrefix(frame.Function, "runtime.") {
					stack = stack[0:0]

				} else {
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
				}

				frame, more = frames.Next()
			}

			ctx.WithField("error", "recover error").WithField("stack", stack).Fatal(err)
		}()
		ctx.Next()
	}
}
