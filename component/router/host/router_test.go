package host

import (
	"testing"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/router/host"
)

// TestStart 测试host路由器。
func TestStart(*testing.T) {
	router := host.NewRouterHost()
	router.SetRouter("example.com", eudore.NewRouterRadix())

	app := eudore.NewApp(router)
	app.Options(host.NewHandler(app))

	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("start fasthttp server, this default page.")
	})
	app.Listen(":8084")
	app.Run()
}
