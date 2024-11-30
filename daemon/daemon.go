/*
Package daemon implements application startup command and signal processing.

Multiple processes will be started and unit test coverage cannot be completed.

# Startup commands

Use [NewParseCommand] to load the full commad startup function.

	app.ParseOption(daemon.NewParseCommand())
	app.Parse()

When executing the app, use parameters to specify the startup command.

	app --command=start --pidfile=/run/run/pidfile

command:

	start   Start the program and write pid.
	daemon  Start the daemon process and write pid.
	status  Read pid to determine whether the process exists.
	reload  Read pid and send syscall.SIGUSR1 signal (10).
	restart Read pid and send syscall.SIGUSR2 signal (12).
	stop    Read pid and send syscall.SIGTERM signal (15).
	disable Skip startup command processing.

# Startup daemon

In the first line of the main function, use [AppDaemon] and
if it is not a daemon,
execute fork to start the daemon.

	func main() {
		daemon.AppDaemon()
		// start app
	}

# Signal processing

Use [NewParseSignal] to load the default [Signal] and Hot restart.

Use [Signal.Register] to register custom signal processing.

	app.ParseOption(daemon.NewParseSignal())
	app.Parse()
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

// The NewParseCommand function creates the [eudore.ConfigParseFunc] that
// executes the startup Command.
//
// If the command is not [CommandStart] or [CommandDaemon],
// it will try to disable logging by closing the [eudore.ContextKeyLogger]
// of [context.Context].
//
// not supported by windows.
func NewParseCommand() eudore.ConfigParseFunc {
	return func(ctx context.Context, conf eudore.Config) error {
		cmd := NewCommand(
			eudore.GetAny(conf.Get("command"), CommandStart),
			eudore.GetAny(conf.Get("pidfile"), eudore.DefaultDaemonPidfile),
		)
		eudore.NewLoggerWithContext(ctx).Infof(
			"daemon command is %s, PID file located at '%s', process ID: %d.",
			cmd.Command, cmd.Pidfile, os.Getpid(),
		)

		switch cmd.Command {
		case CommandDaemon:
			if !eudore.GetAny[bool](os.Getenv(eudore.EnvEudoreDaemonEnable)) {
				setValue(ctx, eudore.ContextKeyLogger, eudore.DefaultLoggerNull)
			}
		case CommandStatus, CommandStop, CommandRestart, CommandReload:
			setValue(ctx, eudore.ContextKeyLogger, eudore.DefaultLoggerNull)
		case CommandDisable:
			return nil
		}

		setValue(ctx, eudore.ContextKeyDaemonCommand, cmd)
		return cmd.Run()
	}
}

// The NewParseSignal function creates [eudore.ConfigParseFunc] for initializing
// signal management.
//
// Default registered signals: [syscall.SIGINT]
// [syscall.SIGUSR2] [syscall.SIGTERM].
//
// If [eudore.EnvEudoreDaemonParentPID] exists,
// the parent process will be shut down.
// This function must be placed after listen.
func NewParseSignal() eudore.ConfigParseFunc {
	return func(ctx context.Context, _ eudore.Config) error {
		sig := &Signal{
			Chan:  make(chan os.Signal),
			Funcs: make(map[os.Signal][]SignalFunc),
		}
		sig.Register(syscall.Signal(0x02), AppStopWithFast)
		sig.Register(syscall.Signal(0x0c), AppRestart)
		sig.Register(syscall.Signal(0x0f), AppStop)
		setValue(ctx, eudore.ContextKeyDaemonSignal, sig)
		app, ok := ctx.Value(eudore.ContextKeyApp).(context.Context)
		if ok {
			go sig.Run(app)
		}

		restartid := eudore.GetAny[int](
			os.Getenv(eudore.EnvEudoreDaemonParentPID),
		)
		if restartid != 0 {
			// Use Command to kill the parent process and write the pid.
			cmd, ok := ctx.Value(eudore.ContextKeyDaemonCommand).(*Command)
			if ok {
				err := cmd.ExecSignal(syscall.Signal(0x0f))
				if err != nil {
					return err
				}
				return cmd.writepid()
			}

			process, err := os.FindProcess(restartid)
			if err != nil {
				return fmt.Errorf("find process %d error: %w", restartid, err)
			}
			return process.Signal(syscall.Signal(0x0f))
		}

		return nil
	}
}

// The AppDaemon function directly starts the daemon process.
//
// After the daemon process is started, it may exit with an error.
func AppDaemon(envs ...string) {
	if eudore.GetAny[bool](os.Getenv(eudore.EnvEudoreDaemonEnable)) {
		return
	}

	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Env = append(os.Environ(), fmt.Sprintf("%s=%d",
		eudore.EnvEudoreDaemonEnable, 1,
	))
	cmd.Env = append(cmd.Env, envs...)
	cmd.Stdout = os.Stdout
	_ = cmd.Start()
	os.Exit(0)
}

func setValue(ctx context.Context, key, val any) {
	seter, ok := ctx.(interface{ SetValue(key any, val any) })
	if ok {
		seter.SetValue(key, val)
		return
	}

	app, ok := ctx.Value(eudore.ContextKeyApp).(*eudore.App)
	if ok {
		app.SetValue(key, val)
	}
}
