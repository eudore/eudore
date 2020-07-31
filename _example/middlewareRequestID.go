package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewApp()
	app.AddMiddleware(middleware.NewLoggerFunc(app, "route"))
	app.AddMiddleware(handlerRequestID)
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.Debug("hello 世界")
	})

	client := httptest.NewClient(app)
	client.NewRequest("GET", "/").Do()

	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}

// handlerRequestID 函数使用时间戳和随机数生成requestid，实现可替换成uuid等库。
func handlerRequestID(ctx eudore.Context) {
	randkey := make([]byte, 3)
	io.ReadFull(rand.Reader, randkey)
	requestId := fmt.Sprintf("%d-%x", time.Now().UnixNano(), randkey)

	ctx.SetLogger(ctx.Logger().WithField(eudore.HeaderXRequestID, requestId))
	ctx.Request().Header.Add(eudore.HeaderXRequestID, requestId)
}
