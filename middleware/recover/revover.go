package recover

import (
	"fmt"
	"github.com/eudore/eudore"
)

func RecoverFunc(ctx eudore.Context) {
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
