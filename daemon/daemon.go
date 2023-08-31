/*
Package daemon 实现应用进程启动命令、后台启动、信号处理、热重启的代码支持。

# 启动命令

	app --command=start/status/stop/restart/deamon/disable --pidfile=/run/run/pidfile

command:

	start	写入pid前台启动
	daemon	写入pid后台启动
	status	读取pid判断进程存在
	stop	读取pid发送syscall.SIGTERM信号(15)
	restart	读取pid发送syscall.SIGUSR2信号(12)
	disable	跳过启动命令处理

# 后台启动

	func main() {
		daemon.StartDaemon()
	}

# 信号处理

# 热重启

使用command组件或kill命令发送SIGUSR2信号。

父进程接受SIGUSR2信号后，传递当前Listen FD和ppid后台启动子进程；
子进程启动初始化后完成向父进程发送SIGTERM信号关闭父进程。
*/
package daemon

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/eudore/eudore"
)

// NewParseCommand 函数创建Command配置解析函数。
func NewParseDaemon(app *eudore.App) eudore.ConfigParseFunc {
	return func(ctx context.Context, conf eudore.Config) error {
		sig := &Signal{
			Chan:  make(chan os.Signal),
			Funcs: make(map[os.Signal][]SignalFunc),
		}
		sig.Register(syscall.Signal(0x02), AppStop)
		sig.Register(syscall.Signal(0x0f), AppStop)
		app.SetValue(eudore.ContextKeyDaemonSignal, sig)
		go sig.Run(ctx)

		cmd := &Command{
			Command: eudore.GetAny(conf.Get("command"), CommandStart),
			Pidfile: eudore.GetAny(conf.Get("pidfile"), eudore.DefaultDaemonPidfile),
			Print: func(format string, args ...any) {
				fmt.Printf(format+"\r\n", args...) //nolint:forbidigo
			},
		}
		if cmd.Command == CommandDisable {
			return nil
		}
		app.Infof("command is %s, pidfile in %s, process pid is %d.", cmd.Command, cmd.Pidfile, os.Getpid())
		if cmd.Command != CommandStart && cmd.Command != CommandDaemon {
			app.SetValue(eudore.ContextKeyLogger, eudore.DefaultLoggerNull)
		}

		app.SetValue(eudore.ContextKeyDaemonCommand, cmd)
		return cmd.Run(ctx)
	}
}

func NewParseRestart() eudore.ConfigParseFunc {
	return func(ctx context.Context, conf eudore.Config) error {
		sig, ok := ctx.Value(eudore.ContextKeyDaemonSignal).(*Signal)
		if ok {
			sig.Register(syscall.Signal(0x0c), AppRestart)
		}

		cmd, ok := ctx.Value(eudore.ContextKeyDaemonCommand).(*Command)
		if ok {
			restartid := eudore.GetAny[int](os.Getenv(eudore.EnvEudoreDaemonRestartID))
			if restartid == 0 {
				return nil
			}
			err := cmd.ExecSignal(syscall.Signal(0x0f))
			if err != nil {
				return err
			}
			return cmd.writepid(ctx)
		}
		return nil
	}
}

// StartDaemon 函数直接后台启动程序。
func StartDaemon(envs ...string) {
	if eudore.GetAny[bool](os.Getenv(eudore.EnvEudoreDaemonEnable)) {
		return
	}

	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%d", eudore.EnvEudoreDaemonEnable, 1))
	cmd.Env = append(cmd.Env, envs...)
	cmd.Stdout = os.Stdout
	_ = cmd.Start()
	os.Exit(0)
}
