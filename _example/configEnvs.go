package main

/*
在config对象的默认解析函数eudore.NewConfigParseEnvs("ENV_")执行环境变量设置配置，使用eudore.Set方法设置的属性。

环境变量将转换成小写路径，'_'下划线相当于'.'的作用

ENV_KEY=value go run configArgs.go --help=true

实现参考eudore.NewConfigParseEnvs
*/

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewApp()
	app.Parse()
	app.Debug(app.Get(""))
	app.CancelFunc()
	app.Run()
}
