package host

import (
	"testing"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/router/host"
)

// TestStart 测试host路由器。
func TestStart(*testing.T) {
	rh := host.NewRouterHost()
	rh.SetRouter("example", eudore.NewRouterRadix())

	app := eudore.NewEudore(rh)
	app.AddGlobalMiddleware(host.InitAddHost)
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("start fasthttp server, this default page.")
	})
	app.Listen(":8084")
	app.Run()
}
