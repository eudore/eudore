package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/show"
)

func main() {
	app := eudore.NewCore()
	show.Inject(app.Group("/eudore/debug"))
	show.RegisterObject("app", app.App)
	
	app.Listen(":8088")
	app.Run()
}