package middleware

import (
	"crypto/rand"
	"fmt"
	"io"
	"time"

	"github.com/eudore/eudore"
)

// NewRequestIDFunc 函数创建一个请求ID注入处理函数，不给定请求ID创建函数，默认使用时间戳和随机数。
func NewRequestIDFunc(fn func() string) eudore.HandlerFunc {
	if fn == nil {
		fn = func() string {
			randkey := make([]byte, 3)
			io.ReadFull(rand.Reader, randkey)
			return fmt.Sprintf("%d-%x", time.Now().UnixNano(), randkey)

		}
	}
	return func(ctx eudore.Context) {
		requestId := ctx.GetHeader(eudore.HeaderXRequestID)
		if requestId == "" {
			requestId = fn()
			ctx.Request().Header.Add(eudore.HeaderXRequestID, requestId)
		}
		ctx.SetLogger(ctx.Logger().WithField(eudore.HeaderXRequestID, requestId).WithFields(nil))
		ctx.SetHeader(eudore.HeaderXRequestID, requestId)
	}
}
