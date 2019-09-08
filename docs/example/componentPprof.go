package main

/*
访问路径 /eudore/debug/pprof/
*/

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
