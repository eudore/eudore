package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewCore()
	app.AddMiddleware(eudore.MethodAny, "", middleware.NewLoggerFunc(app.App))
	app.Listen(":8088")
	app.Run()
}
