package main

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewCore()
	// 使用空证书会自动签发私人证书, Eudore也具有该方法。
	app.ListenTLS(":8088", "", "")
	app.Run()
}
