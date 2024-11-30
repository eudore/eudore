package daemon

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/eudore/eudore"
)

var (
	listeners   = map[string]net.Listener{}
	listenersfd = map[string]uintptr{}
)

//nolint:gochecknoinits
func init() {
	addrs := os.Getenv(eudore.EnvEudoreDaemonListeners)
	for i, addr := range strings.Split(addrs, " ,") {
		if addr == "" {
			continue
		}
		listenersfd[addr] = uintptr(i + 3)
	}

	// wrap listen
	listen := eudore.DefaultServerListen
	eudore.DefaultServerListen = func(network, address string,
	) (net.Listener, error) {
		addr := fmt.Sprintf("%s://%s", network, address)
		var ln net.Listener
		var err error

		fd, ok := listenersfd[addr]
		if ok {
			ln, err = net.FileListener(os.NewFile(fd, ""))
		} else {
			ln, err = listen(network, address)
		}

		if err == nil {
			listeners[addr] = ln
		}
		return ln, err
	}
}

// The AppStopWithFast function shortens the [eudore.DefaultServerShutdownWait],
// and then uses [AppStop].
func AppStopWithFast(ctx context.Context) error {
	eudore.DefaultServerShutdownWait = time.Second
	return AppStop(ctx)
}

// The AppStop function gets [eudore.ContextKeyAppCancel] from [context.Context]
// and then closes [eudore.App].
//
// If it cannot be get, you need to register processing function for the
// [Signal].
func AppStop(ctx context.Context) error {
	cancel, ok := ctx.Value(eudore.ContextKeyAppCancel).(context.CancelFunc)
	if ok {
		cancel()
	}
	return nil
}

type filer interface {
	File() (*os.File, error)
}

// AppRestart function implements hot restart function.
//
// fork starts a new process passing listen fd and pid.
//
// After the new process is started and initialized,
// a [syscall.SIGTERM] is sent to the process,
// and finally the process is closed using signal processing.
//
// The port that app listens on needs to use the [eudore.DefaultServerListen]
// method.
// When the daemon package init(), wrap listen is used to get the listening fd.
//
// In the [NewParseSignal] function, [eudore.EnvEudoreDaemonParentPID] will be
// checked and the parent process will be closed.
func AppRestart(ctx context.Context) error {
	path := os.Args[0]
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	if filepath.Base(path) == path {
		path, err = exec.LookPath(path)
		if err != nil {
			return err
		}
	}

	addrs, files, err := getListeners()
	if err != nil {
		return err
	}
	envs := append(getEnvirons(),
		fmt.Sprintf("%s=%d", eudore.EnvEudoreDaemonEnable, 1),
		fmt.Sprintf("%s=%d", eudore.EnvEudoreDaemonParentPID, os.Getpid()),
		fmt.Sprintf("%s=%s", eudore.EnvEudoreDaemonListeners,
			strings.Join(addrs, " ,"),
		),
	)

	process, err := os.StartProcess(path, os.Args, &os.ProcAttr{
		Dir:   dir,
		Env:   envs,
		Files: append([]*os.File{os.Stdin, os.Stdout, os.Stderr}, files...),
	})
	if err == nil {
		eudore.NewLoggerWithContext(ctx).
			Infof("eudore start new process %d", process.Pid)
	}
	return err
}

func getListeners() ([]string, []*os.File, error) {
	// get addrs and socket listen fds
	addrs := make([]string, 0, len(listeners))
	files := make([]*os.File, 0, len(listeners))
	for addr, ln := range listeners {
		filer, ok := ln.(filer)
		if ok {
			fd, err := filer.File()
			if err != nil {
				return nil, nil, err
			}
			addrs = append(addrs, addr)
			files = append(files, fd)
		}
	}
	return addrs, files, nil
}

func getEnvirons() []string {
	// set graceful restart env flag
	envs := []string{}
	filters := [...]string{
		eudore.EnvEudoreDaemonListeners + "=",
		eudore.EnvEudoreDaemonParentPID + "=",
		eudore.EnvEudoreDaemonEnable + "=",
	}
	for _, value := range os.Environ() {
		switch {
		case strings.HasPrefix(value, filters[0]):
		case strings.HasPrefix(value, filters[1]):
		case strings.HasPrefix(value, filters[2]):
		default:
			envs = append(envs, value)
		}
	}
	return envs
}
