package main

/*
访问路径 /eudore/debug/vars
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/expvar"
)

func main() {
	app := eudore.NewCore()
	app.GetFunc("/eudore/debug/vars", expvar.Expvar)
	app.Listen(":8088")
	app.Run()
}
