package main

/*
通过Daemon()函数后台启动程序，也可以通过命令解析启动程序。

当第一次启动时，使用os.Exec执行启动命令后台启动进程、关闭进程并附加环境变量，第二次启动时检测到环境变量即为后台启动，会忽略后台启动逻辑。然后执行正常启动。

该组件不支持win系统。
*/

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/eudore/eudore"
)

func main() {
	Daemon()

	app := eudore.NewApp()
	app.GetFunc("/*", func(ctx eudore.Context) {
		ctx.WriteString("server daemon")
	})

	go func() {
		select {
		case <-app.Done():
		case <-time.After(10 * time.Second):
			app.CancelFunc()
		}
	}()
	app.Listen(":8088")
	app.Run()
}

// Daemon 函数直接后台启动程序。
func Daemon(envs ...string) {
	if eudore.GetStringBool(os.Getenv(eudore.EnvEudoreIsDaemon)) {
		return
	}

	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%d", eudore.EnvEudoreIsDaemon, 1))
	cmd.Env = append(cmd.Env, envs...)
	cmd.Stdout = os.Stdout
	cmd.Start()
	os.Exit(0)
}
