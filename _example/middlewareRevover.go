package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewCore()
	httptest.NewClient(app).Stop(0)
	app.AddMiddleware(middleware.NewRecoverFunc())
	app.Listen(":8088")
	app.Run()
}
