package daemon

/*
利用系统信号进制，执行start、daemon、stop、status、restart命令来操作进程。
进程pid存储在pid文件中。
*/

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/eudore/eudore"
)

const (
	CommandStart   = "start"
	CommandDaemon  = "daemon"
	CommandStatus  = "status"
	CommandStop    = "stop"
	CommandRestart = "restart"
	CommandDisable = "disable"
)

// Command is a command parser that performs the corresponding behavior based on the current command.
//
// Command 对象是一个命令解析器，根据当前命令执行对应行为。
type Command struct {
	Command string
	Pidfile string
	Args    []string
	Envs    []string
	Print   func(string, ...any)
}

// Run 方法启动Command解析。
func (cmd *Command) Run(ctx context.Context) (err error) {
	switch cmd.Command {
	case CommandStart:
		return cmd.Start(ctx)
	case CommandDaemon:
		return cmd.Daemon(ctx)
	case CommandStatus:
		err = cmd.Status()
	case CommandStop:
		err = cmd.Stop()
	case CommandRestart:
		err = cmd.Restart()
	default:
		err = errors.New("undefined command " + cmd.Command)
		cmd.Print("undefined command %s, support command: start/status/stop/restart/daemon.", cmd.Command)
	}

	if err != nil {
		cmd.Print("%s is false, error: %w.", cmd.Command, err)
		return err
	}

	pid, _ := cmd.readpid()
	cmd.Print("%s is true, pid is %d, pidfile in %s.", cmd.Command, pid, cmd.Pidfile)
	cmd.Wait(pid)

	return fmt.Errorf("daemon is %s %w", cmd.Command, context.Canceled)
}

func (cmd *Command) Wait(p int) {
	t := eudore.GetAnyByString(os.Getenv(eudore.EnvEudoreDaemonTimeout), 60)
	if t < 0 || (cmd.Command != CommandStop && cmd.Command != CommandRestart) {
		return
	}

	for i := 0; i <= t; i++ {
		pid, err := cmd.readpid()
		switch {
		case cmd.Command == CommandStop && err != nil:
			cmd.Print("stop successfully, wait time %ds.", i)
			return
		case err != nil:
			cmd.Print("%s read pid error: %w.", cmd.Command, err)
		case cmd.Command == CommandRestart && pid != p:
			cmd.Print("restart successfully, new pid is %d, wait time %ds.", pid, i)
			return
		}
		time.Sleep(time.Second)
	}
	if t > 0 {
		cmd.Print("%s failed, wait time %ds timeout.", cmd.Command, t)
	}
}

// Start execute the startup function and write the pid to the file.
//
// Start 函数执行启动函数，并将pid写入文件。
func (cmd *Command) Start(ctx context.Context) error {
	// 测试文件是否被锁定
	pid, err := cmd.readpid()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	restartid := eudore.GetAny[int](os.Getenv(eudore.EnvEudoreDaemonRestartID))
	if pid != 0 && pid == restartid {
		return nil
	}

	err = cmd.Status()
	if err == nil {
		return fmt.Errorf("process exites pid %d", pid)
	}

	// 写入pid
	return cmd.writepid(ctx)
}

// Daemon Start the process in the background. If it is not started in the background, create a background process.
//
// Daemon 函数后台启动进程。若不是后台启动，则创建一个后台进程。
func (cmd *Command) Daemon(ctx context.Context) error {
	if eudore.GetAny[bool](os.Getenv(eudore.EnvEudoreDaemonEnable)) {
		return cmd.Start(ctx)
	}

	// 测试文件是否被锁定
	pid, err := cmd.readpid()
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	err = cmd.Status()
	if err == nil {
		return fmt.Errorf("process exites pid %d", pid)
	}

	fork := exec.Command(os.Args[0], os.Args[1:]...)
	fork.Args = append(fork.Args, cmd.Args...)
	fork.Env = append(os.Environ(), fmt.Sprintf("%s=%d", eudore.EnvEudoreDaemonEnable, 1))
	fork.Env = append(fork.Env, cmd.Envs...)
	fork.Stdout = os.Stdout
	err = fork.Start()
	if err != nil {
		return err
	}
	return fmt.Errorf("daemon start %w", context.Canceled)
}

// Status 函数调用系统命令，想进程发送信号 0。
func (cmd *Command) Status() error {
	return cmd.ExecSignal(syscall.Signal(0x00))
}

// Stop 函数调用系统命令，想进程发送信号syscall.SIGTERM。
func (cmd *Command) Stop() error {
	return cmd.ExecSignal(syscall.Signal(0x0f))
}

// Reload 函数调用系统命令，想进程发送信号syscall.SIGUSR1。
func (cmd *Command) Reload() error {
	return cmd.ExecSignal(syscall.Signal(0x0a))
}

// Restart 函数调用系统命令，想进程发送信号syscall.SIGUSR2。
func (cmd *Command) Restart() error {
	return cmd.ExecSignal(syscall.Signal(0x0c))
}

// ExecSignal 函数向pidfile内的进程发送指定命令。
func (cmd *Command) ExecSignal(sig os.Signal) error {
	pid, err := cmd.readpid()
	if err != nil {
		return err
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process %d error: %w", pid, err)
	}

	err = process.Signal(sig)
	if err != nil {
		os.Remove(cmd.Pidfile)
		return err
	}
	return nil
}

// Read the value in the pid file.
//
// 读取pid文件内的值。
func (cmd *Command) readpid() (int, error) {
	file, err := os.OpenFile(cmd.Pidfile, os.O_RDONLY, 0o644)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	id, err := io.ReadAll(file)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(id)))
}

// Open and lock the pid file and write the value of pid.
//
// 打开并锁定pid文件，写入pid的值。
func (cmd *Command) writepid(ctx context.Context) (err error) {
	file, err := os.OpenFile(cmd.Pidfile, os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = fmt.Fprintf(file, "%d", os.Getpid())
	if err != nil {
		return
	}
	go func() {
		// 关闭删除pid文件
		<-ctx.Done()
		pid, err := cmd.readpid()
		if err == nil && pid == os.Getpid() {
			os.Remove(cmd.Pidfile)
		}
	}()
	return nil
}
