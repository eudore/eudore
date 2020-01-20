package main

/*
访问路径 /eudore/debug/vars
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/expvar"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewCore()
	httptest.NewClient(app).Stop(0)
	app.GetFunc("/eudore/debug/vars", expvar.Expvar)
	app.Listen(":8088")
	app.Run()
}
