package main

/*
ListenTLS方法一般均默认开启了h2，如果需要仅开启https，需要手动listen监听然后使用app.Serve启动服务。
*/

import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func main() {
	app := eudore.NewCore()
	httptest.NewClient(app).Stop(0)
	// 使用空证书会自动签发私人证书, Eudore也具有该方法。
	app.ListenTLS(":8088", "", "")
	app.Run()
}
