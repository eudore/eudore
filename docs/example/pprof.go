package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/pprof"
)

func main() {
	app := eudore.NewCore()
	pprof.RoutesInject(app.Group("/eudore/debug"))

	app.Listen(":8088")
	app.Run()
}
