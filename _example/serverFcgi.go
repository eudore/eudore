package main

/*
NewServerFcgi() 使用net/http/fcgi包处理fastchi
*/

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewApp(eudore.NewServerFcgi())
	app.Listen(":8088")
	// app.CancelFunc()
	app.Run()
}
