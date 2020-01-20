package main

/*
NewServerFcgi() 使用net/http/fcgi包处理fastchi
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewCore()
	httptest.NewClient(app).Stop(0)
	app.Server = eudore.NewServerFcgi()
	app.Listen(":8088")
	app.Run()
}
