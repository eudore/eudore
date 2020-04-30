package main

/*
在config对象的默认解析函数eudore.ConfigParseEnvs执行环境变量设置配置，使用eudore.Set方法设置的属性。

需要命令行参数的格式为ENV_KEY=value，环境变量必须是ENV_为前缀，移除ENV_后缀后的name会全部转换成小写，如果变量属性是大写需要使用结构体tag set来设置set方法使用的属性别名。

ENV_KEY=value go run configArgs.go --keys.help=true

实现参考eudore.ConfigParseEnvs
*/

import (
	"github.com/eudore/eudore"
)

func main() {
	app := eudore.NewApp()
	app.Options(app.Parse())

	app.CancelFunc()
	app.Run()
}
