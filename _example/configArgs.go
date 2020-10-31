package main

/*
在config对象的默认解析函数eudore.ConfigParseArgs执行命令行参数设置配置，使用eudore.Set方法设置的属性。

需要命令行参数的格式为--key=value，暂时不支持缩写参数。

如果配置类型是ConfigEudore,那么调用Set设置的属性会按照属性一层层去选择然后设置，具体参数configEudore.go中的演示。

go run configArgs.go --help=true --k=0 --key=value

实现参考eudore.ConfigParseArgs
*/

import (
	"github.com/eudore/eudore"
	"os"
)

func main() {
	app := eudore.NewApp()
	os.Args = append(os.Args, "--name=eudoreName")
	app.Options(app.Parse())

	app.CancelFunc()
	app.Run()
}
