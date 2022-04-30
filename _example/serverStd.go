package main

/*
NewServerStd() 使用net/http.Server包处理http请求
NewServerStd函数可以指定结构体配置,设置http.Server属性，例如timeout
*/

import (
	"github.com/eudore/eudore"
	"time"
)

func main() {
	app := eudore.NewApp()
	app.SetValue(eudore.ContextKeyServer, eudore.NewServerStd(&eudore.ServerStdConfig{
		ReadTimeout:  eudore.TimeDuration(4 * time.Second),
		WriteTimeout: eudore.TimeDuration(12 * time.Second),
		IdleTimeout:  eudore.TimeDuration(60 * time.Second),
	}))

	app.Listen(":8088")
	app.Run()
}
