package main

/*
在config对象的默认解析函数eudore.NewConfigParseArgs(nil)执行命令行参数设置配置，使用eudore.Set方法设置的属性。

命令行参数使用'--{key}.{sub}={value}'格式。

如果结构体存在'flag' tag将作为该路径的缩写，tag长度需要小于5，命令行格式为'-{short}={value},短参数将会自动为长参数。

如果配置类型是ConfigEudore,那么调用Set设置的属性会按照属性一层层去选择然后设置，具体参数configEudore.go中的演示。

go run configArgs.go --help=true --k=0 --key=value

实现参考eudore.NewConfigParseArgs
*/

import (
	"github.com/eudore/eudore"
	"os"
)

var shorts = map[string][]string{
	"f": {"file"},
}

func main() {
	os.Args = append(os.Args, "--name=eudoreName", "-f=config.json")
	app := eudore.NewApp()
	// 指定NewConfigParseArgs方法的短参数映射，默认使用结构体数据tag获取。
	app.ParseOption([]eudore.ConfigParseFunc{eudore.NewConfigParseArgs(shorts)})
	app.Options(app.Parse())
	app.Debug("get name:", app.GetString("name"))
	app.Debug(app.Get(""))
	app.CancelFunc()
	app.Run()
}
