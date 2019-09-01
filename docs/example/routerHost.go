package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/router/host"
)

func main() {
	rh := host.NewRouterHost()
	rh.RegisterHost("example", eudore.NewRouterRadix())

	app := eudore.NewEudore(rh)
	app.AddGlobalMiddleware(host.InitAddHost)
	app.AnyFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("start fasthttp server, this default page.")
	})
	app.Listen(":8088")
	app.Run()
}
