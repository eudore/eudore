package eudore

import (
	"os"
	"fmt"
	"errors"
	"os/exec"
	"strconv"
	"syscall"
	"io/ioutil"
)

// The name of the environment variable used when the program starts in the background, which is used to indicate whether the fork is started in the background.
//
// 程序后台启动时使用的环境变量名，用于表示是否fork来后台启动。
const (
	DAEMON_ENVIRON_KEY		= "EUDORE_IS_DEAMON"
)

// Command is a command parser that performs the corresponding behavior based on the current command.
//
// Command是一个命令解析器，根据当前命令执行对应行为。
type Command struct {
	pid		int
	pidfile	string
	cmd		string
	StartHandler	func() error
}

// Returns a command to parse the object, the current command and the process pid file path,
// if the behavior is start will execute the handler.
//
// 返回一个命令解析对象，需要当前命令和进程pid文件路径，如果行为是start会执行handler。
func NewCommand(cmd , pidfile string, handler func() error) *Command {
	return &Command{	
		pid:	-1,
		cmd: 	cmd,
		pidfile:		pidfile,
		StartHandler:	handler,
	}
}

// Parse the command and execute it.
//
// 解析命令并执行。
func (c *Command) Run() (err error) {
	c.readpid()
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
		return
	}
	if err != nil {
		fmt.Printf("%s is false, %v.\n", c.cmd, err)
	}else {
		fmt.Printf("%s is true.\n", c.cmd)
	}
	return
}

// Execute the startup function and write the pid to the file.
//
// 执行启动函数，并将pid写入文件。
func (c *Command) Start() error{
	c.writepid()
	return c.StartHandler()
}

// 后台启动进程。
func (c *Command) Daemon() error {
	if os.Getenv(DAEMON_ENVIRON_KEY) == "" {
		cmd := exec.Command(os.Args[0], os.Args[1:]...)
		cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%d", DAEMON_ENVIRON_KEY, 1))
		return cmd.Start()
	}else{
		return c.Start()
	}
}

func (c *Command) Status() error {
	return execsignal(c.pid, syscall.Signal(0x00))
}

func (c *Command) Stop() error {
	return execsignal(c.pid, syscall.SIGTERM)
}

func (c *Command) Reload() error {
	return execsignal(c.pid, syscall.SIGUSR1)
}

func (c *Command) Restart() error {
	return execsignal(c.pid, syscall.SIGUSR2)
}


func execsignal(pid int, sig os.Signal) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Signal(sig)
}

func (c *Command) readpid() int {
	go func(){
			recover()
		}()
	file,err := os.Open(c.pidfile)  
	if err == nil {
		defer file.Close()  
		id, _ := ioutil.ReadAll(file)
		c.pid, _ = strconv.Atoi(string(id))
	}
	return c.pid
}

func (c *Command) writepid() {
	file, _ := os.OpenFile(c.pidfile, os.O_WRONLY | os.O_TRUNC | os.O_CREATE, 0666)
	defer file.Close()
	byteSlice := []byte(fmt.Sprintf("%d", os.Getpid()))
	file.Write(byteSlice)
	// syscall.Flock(int(file.Fd()), syscall.LOCK_EX | syscall.LOCK_NB)
}