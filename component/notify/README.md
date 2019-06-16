# notify 代码自动重启

实现思路：启动父进程，创建运行进程，同时使用fsnotify监听指定目录文件是否有写入，如果出发写入则执行编译命令，执行成功就关闭旧进程并启动新进程。

在启动新进程中加入一个环境变量以识别父子进程，使子进程不会重复监听文件更新自动重启。

示例：

```golang
package main


import (
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/notify"
)

func main() {
	app := eudore.NewCore()

	app.Config.Set("component.notify.buildcmd", "go build -o server coreNotify.go")
	app.Config.Set("component.notify.startcmd", "./server")
	app.Config.Set("component.notify.watchdir", ".")
	notify.NewNotify(app.App).Run()

	app.Listen(":8088")
	app.Run()
}
```

然后`go build -o server coreNotify.go`编译，`./server`启动程序，在当前目录创建一个`.go`后缀的文件并写入内容，就会触发编译重启。

**coreNotify.go是文件名称，默认实现监控.go后缀文件的写入事件**

# 事项

app需要设置Logger给notify库使用，输出信息。

需要在app.Config中设置需要的三项配置，编译目录、启动命令、监听目录三项，配置类型需要是字符串或者字符串数组类型。

如果配置存在`component.notify.disable`项，将会禁用本功能。

监听目标中`.go`后缀的文件的写入事件。

多个目录就要使用数组或者空格分隔的字符串。

fsnotify库监听的目标不是递归的，如果目录路径最后一个字符以斜杠`/`结尾，会递归监听这个目录，会忽略其中点开头的隐藏目录。

具体实现见实现源码。