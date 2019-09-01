package main

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
)

func main() {
	app := eudore.NewCore()
	// map保存用户密码
	app.AddMiddleware(eudore.MethodAny, "", middleware.NewBasicAuthFunc("", map[string]string{
		"user": "pw",
	}))

	app.Listen(":8088")
	app.Run()
}
