package daemon

// Use the system signal system to execute start, daemon, stop, status,
// and restart commands to operate the process.
// The process pid is stored in the pid file.

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
	// Command and Output.

	CommandStart   = "start"
	CommandDaemon  = "daemon"
	CommandStatus  = "status"
	CommandStop    = "stop"
	CommandRestart = "restart"
	CommandReload  = "reload"
	CommandDisable = "disable"
	OutputDaemon   = "Daemon process ID {{pid}} started successfully, wait time: {{time}}."
	OutputStatus   = "Process ID {{pid}} is running, PID file located at {{pidfile}}."
	OutputStop     = "Successfully stopped process ID {{pid}}, wait time: {{time}}."
	OutputRestart  = "Successfully restarted process ID {{pid}}, wait time: {{time}}."
	OutputReload   = "Process ID {{pid}} is reload, PID file located at {{pidfile}}."
	OutputError    = "Command {{command}} failed with error: {{error}}."
	OutputFailed   = "Command {{command}} failed, timed out after {{time}}."
)

// CustomWaitHook defines hook to implement custom waiting.
//
// [Command.Wait] may have inaccurate waiting time.
var CustomWaitHook = func(_, _ string, _ int) {}

// Command defines the startup command and pid file.
type Command struct {
	Command string
	Pidfile string
	Args    []string
	Envs    []string
	execpid int
}

// NewCommand function creates [Command] to manage startup commands and pid.
func NewCommand(cmd, pid string) *Command {
	return &Command{
		Command: cmd,
		Pidfile: pid,
	}
}

// Unmount method cleans up the pidfile.
func (cmd *Command) Unmount(context.Context) {
	pid, err := cmd.readpid()
	if err == nil && pid == os.Getpid() {
		_ = os.Remove(cmd.Pidfile)
	}
}

// Run method executes the Command.
func (cmd *Command) Run() error {
	var err error
	switch cmd.Command {
	case CommandStart:
		return cmd.Start()
	case CommandDaemon:
		if eudore.GetAny[bool](os.Getenv(eudore.EnvEudoreDaemonEnable)) {
			return cmd.Start()
		}
		err = cmd.Daemon()
	case CommandStatus:
		err = cmd.Status()
	case CommandStop:
		err = cmd.Stop()
	case CommandRestart:
		err = cmd.Restart()
	case CommandReload:
		err = cmd.Reload()
	default:
		cmds := []string{
			CommandStart,
			CommandDaemon,
			CommandStatus,
			CommandStop,
			CommandRestart,
			CommandReload,
		}
		err = errors.New("undefined command " + cmd.Command)
		cmd.output(fmt.Sprintf("%v. Supported commands: %s.", err, strings.Join(cmds, "/")), 0, 0)
		return err
	}

	if err != nil {
		str := strings.ReplaceAll(OutputError, "{{error}}", err.Error())
		cmd.output(str, 0, 0)
		return err
	}

	CustomWaitHook(cmd.Command, cmd.Pidfile, cmd.execpid)
	switch cmd.Command {
	case CommandStatus:
		cmd.output(OutputStatus, cmd.execpid, 0)
	case CommandReload:
		cmd.output(OutputReload, cmd.execpid, 0)
	default:
		cmd.Wait()
	}

	return context.Canceled
}

// Wait method reads the PID file to determine the execution status.
//
// [CommandStop] waits for PID file to be deleted.
//
// [CommandDaemon] waits for PID file to be created,
// but may not be listening on the port.
//
// [CommandRestart] waits for PID file content to change.
func (cmd *Command) Wait() {
	str := os.Getenv(eudore.EnvEudoreDaemonTimeout)
	t := eudore.GetAnyByString(str, 30*time.Second)
	s := time.Millisecond
	for wait := time.Duration(0); wait <= t; {
		pid, err := cmd.readpid()
		switch {
		case cmd.Command == CommandStop && err != nil:
			cmd.output(OutputStop, cmd.execpid, wait)
			return
		case cmd.Command == CommandDaemon && err == nil:
			cmd.output(OutputDaemon, pid, wait)
			return
		case cmd.Command == CommandRestart && pid != cmd.execpid:
			cmd.output(OutputRestart, pid, wait)
			return
		}

		switch {
		case wait < time.Millisecond*100:
			s *= 2
		case wait < time.Second:
			s = time.Millisecond * 100
		case wait < time.Second*3:
			s = time.Millisecond * 200
		case wait < time.Second*10:
			s = time.Millisecond * 500
		default:
			s = time.Second
		}
		time.Sleep(s)
		wait += s
	}
	if t > 0 {
		cmd.output(OutputFailed, 0, t)
	}
}

var initat = time.Now()

func (cmd *Command) output(out string, pid int, wait time.Duration) {
	if out != "" {
		wait += time.Since(initat).Truncate(time.Millisecond)
		fn := strings.ReplaceAll
		out = fn(out, "{{command}}", cmd.Command)
		out = fn(out, "{{pidfile}}", cmd.Pidfile)
		out = fn(out, "{{pid}}", strconv.Itoa(pid))
		out = fn(out, "{{time}}", wait.String())
		_, _ = fmt.Fprint(os.Stdout, out+"\r\n")
	}
}

// Start execute the startup function and write the pid to the file.
func (cmd *Command) Start() error {
	pid, err := cmd.readpid()
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	restartid := eudore.GetAny[int](os.Getenv(eudore.EnvEudoreDaemonParentPID))
	if pid != 0 && pid == restartid {
		return nil
	}

	err = cmd.Status()
	if err == nil {
		return fmt.Errorf("process exites pid %d", pid)
	}

	return cmd.writepid()
}

// Daemon Start the process in the daemon.
func (cmd *Command) Daemon() error {
	pid, err := cmd.readpid()
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	err = cmd.Status()
	if err == nil {
		return fmt.Errorf("process exites pid %d", pid)
	}

	// #nosec G204
	fork := exec.CommandContext(context.Background(), os.Args[0], os.Args[1:]...)
	fork.Args = append(fork.Args, cmd.Args...)
	fork.Env = append(os.Environ(), fmt.Sprintf("%s=%d",
		eudore.EnvEudoreDaemonEnable, 1,
	))
	fork.Env = append(fork.Env, cmd.Envs...)
	fork.Stdout = os.Stdout
	return fork.Start()
}

// Status function process sends signal 0, determine if a process exists.
func (cmd *Command) Status() error {
	return cmd.ExecSignal(syscall.Signal(0x00))
}

// Reload function process sends signal syscall.SIGUSR1.
func (cmd *Command) Reload() error {
	return cmd.ExecSignal(syscall.Signal(0x0a))
}

// Restart function process sends signal syscall.SIGUSR2.
func (cmd *Command) Restart() error {
	return cmd.ExecSignal(syscall.Signal(0x0c))
}

// Stop function process sends signal syscall.SIGTERM.
func (cmd *Command) Stop() error {
	return cmd.ExecSignal(syscall.Signal(0x0f))
}

// ExecSignal function sends the specified signal to the process
// in the pidfile.
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
	if errors.Is(err, os.ErrProcessDone) {
		_ = os.Remove(cmd.Pidfile)
		return err
	}
	cmd.execpid = pid
	return err
}

// Read the value in the pid file.
func (cmd *Command) readpid() (int, error) {
	file, err := os.Open(cmd.Pidfile)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	id, err := io.ReadAll(io.LimitReader(file, 128))
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(id)))
}

// Open and lock the pid file and write the value of pid.
func (cmd *Command) writepid() error {
	file, err := os.OpenFile(cmd.Pidfile, os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(strconv.Itoa(os.Getpid()))
	return err
}
