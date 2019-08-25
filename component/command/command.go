package command

/*
利用系统信号进制，执行start、daemon、stop、status、restart命令来操作进程。
进程pid存储在pid文件中。
*/

import (
	"context"
	"errors"
	"fmt"
	"github.com/eudore/eudore"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// The name of the environment variable used when the program starts in the background, which is used to indicate whether the fork is started in the background.
//
// 程序后台启动时使用的环境变量名，用于表示是否fork来后台启动。
const (
	DAEMON_ENVIRON_KEY = "EUDORE_IS_DEAMON"
	DAEMON_NOTPID      = "EUDORE_NOTPID"
)

// Command is a command parser that performs the corresponding behavior based on the current command.
//
// Command 对象是一个命令解析器，根据当前命令执行对应行为。
type Command struct {
	ctx     context.Context
	pidfile string
	cmd     string
	file    *os.File
}

// InitCommand 函数初始化定义程序启动命令。
func InitCommand(app *eudore.Eudore) error {
	cmd := eudore.GetDefaultString(app.Config.Get("command"), "start")
	pid := eudore.GetDefaultString(app.Config.Get("pidfile"), "/var/run/eudore.pid")
	app.Logger.Infof("current command is %s, pidfile in %s.", cmd, pid)
	return NewCommand(app, cmd, pid).Run()
}

// NewCommand returns a command to parse the object, the current command and the process pid file path,
// if the behavior is start will execute the handler.
//
// NewCommand 函数返回一个命令解析对象，需要当前命令和进程pid文件路径，如果行为是start会执行handler。
func NewCommand(ctx context.Context, cmd, pidfile string) *Command {
	if len(pidfile) == 0 {
		pidfile = "/var/run/eudore.pid"
	}
	return &Command{
		ctx:     ctx,
		cmd:     cmd,
		pidfile: pidfile,
	}
}

// Run parse the command and execute it.
//
// Run 函数解析命令并执行。
func (c *Command) Run() (err error) {
	if c.pidfile == "" {
		fmt.Println("pidfile is empty string.")
		return errors.New("pidfile is empty string.")
	}
	switch c.cmd {
	case "start":
		err = c.Start()
	case "daemon":
		err = c.Daemon()
	case "status":
		err = c.Status()
	case "stop":
		err = c.Stop()
	case "restart":
		err = c.Restart()
	default:
		err = errors.New("undefined command " + c.cmd)
		fmt.Println("undefined command ", c.cmd)
	}
	// 输出提升信息
	if err != nil {
		fmt.Printf("%s is false, %v.\n", c.cmd, err)
	} else {
		fmt.Printf("%s is true.\n", c.cmd)
	}
	if c.cmd == "daemon" {
		if os.Getenv(DAEMON_ENVIRON_KEY) == "" {
			os.Exit(0)
		}
		return
	}
	// 非启动命令结束程序
	if c.cmd != "start" {
		os.Exit(0)
	}
	return
}

// Start execute the startup function and write the pid to the file.
//
// Start 函数执行启动函数，并将pid写入文件。
func (c *Command) Start() error {
	// 测试文件是否被锁定
	_, err := c.readpid()
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	// 写入pid
	return c.writepid()
}

// Daemon Start the process in the background. If it is not started in the background, create a background process.
//
// Daemon 函数后台启动进程。若不是后台启动，则创建一个后台进程。
func (c *Command) Daemon() error {
	if os.Getenv(DAEMON_ENVIRON_KEY) != "" {
		return c.Start()
	}

	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%d", DAEMON_ENVIRON_KEY, 1))
	return cmd.Start()
}

// Status 函数调用系统命令，想进程发送信号 0。
func (c *Command) Status() error {
	return c.ExecSignal(syscall.Signal(0x00))
}

// Stop 函数调用系统命令，想进程发送信号syscall.SIGTERM。
func (c *Command) Stop() error {
	return c.ExecSignal(syscall.SIGTERM)
}

// Reload 函数调用系统命令，想进程发送信号syscall.SIGUSR1。
func (c *Command) Reload() error {
	return c.ExecSignal(syscall.SIGUSR1)
}

// Restart 函数调用系统命令，想进程发送信号syscall.SIGUSR2。
func (c *Command) Restart() error {
	return c.ExecSignal(syscall.SIGUSR2)
}

// ExecSignal 函数向pidfile内的进程发送指定命令
func (c *Command) ExecSignal(sig os.Signal) error {
	pid, err := c.readpid()
	if err != nil {
		return fmt.Errorf("read pidfile %s error: %v", c.pidfile, err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process %d error: %v", pid, err)
	}

	return process.Signal(sig)
}

// Read the value in the pid file.
//
// 读取pid文件内的值。
func (c *Command) readpid() (int, error) {
	file, err := os.OpenFile(c.pidfile, os.O_RDONLY, 0666)
	if err != nil {
		return 0, err
	}
	id, err := ioutil.ReadAll(file)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(id)))
}

// Open and lock the pid file and write the value of pid.
//
// 打开并锁定pid文件，写入pid的值。
func (c *Command) writepid() (err error) {
	if os.Getenv(DAEMON_NOTPID) != "" {
		return nil
	}
	c.file, err = os.OpenFile(c.pidfile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return
	}
	err = syscall.Flock(int(c.file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		c.file.Close()
		return
	}
	byteSlice := []byte(fmt.Sprintf("%d", os.Getpid()))
	_, err = c.file.Write(byteSlice)
	if err != nil {
		c.file.Close()
		return
	}
	c.Release()
	return nil
}

// Release Close the delete pid file and release the exclusive lock
//
// Release 函数关闭删除pid文件，并解除独占锁
func (c *Command) Release() {
	go func() {
		<-c.ctx.Done()
		syscall.Flock(int(c.file.Fd()), syscall.LOCK_UN)
		c.file.Close()
		os.Remove(c.pidfile)
	}()
}

// Daemon 函数直接后台启动程序。
func Daemon() {
	if os.Getenv(DAEMON_ENVIRON_KEY) != "" {
		return
	}

	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%d", DAEMON_ENVIRON_KEY, 1))
	cmd.Start()
	os.Exit(0)
}
