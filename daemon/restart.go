package daemon

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/eudore/eudore"
)

var (
	listeners   = map[string]net.Listener{}
	listenersfd = map[string]uintptr{}
)

//nolint:gochecknoinits
func init() {
	for i, addr := range strings.Split(os.Getenv(eudore.EnvEudoreDaemonListeners), " ,") {
		if addr == "" {
			continue
		}
		listenersfd[addr] = uintptr(i + 3)
	}

	listen := eudore.DefaultServerListen
	eudore.DefaultServerListen = func(network, address string) (net.Listener, error) {
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

func AppStop(ctx context.Context) error {
	app, ok := ctx.Value(eudore.ContextKeyApp).(*eudore.App)
	if ok {
		app.CancelFunc()
	}
	return nil
}

type filer interface {
	File() (*os.File, error)
}

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

	// get addrs and socket listen fds
	addrs := make([]string, 0, len(listeners))
	files := make([]*os.File, 0, len(listeners))
	for addr, ln := range listeners {
		filer, ok := ln.(filer)
		if ok {
			fd, err := filer.File()
			if err != nil {
				return err
			}
			addrs = append(addrs, addr)
			files = append(files, fd)
			syscall.CloseOnExec(int(fd.Fd()))
			defer fd.Close()
		}
	}

	// set graceful restart env flag
	envs := []string{}
	for _, value := range os.Environ() {
		if !strings.HasPrefix(value, "EUDORE_DAEMON_") {
			envs = append(envs, value)
		}
	}
	envs = append(envs,
		fmt.Sprintf("%s=%d", eudore.EnvEudoreDaemonEnable, 1),
		fmt.Sprintf("%s=%d", eudore.EnvEudoreDaemonRestartID, os.Getpid()),
		fmt.Sprintf("%s=%s", eudore.EnvEudoreDaemonListeners, strings.Join(addrs, " ,")),
	)

	process, err := os.StartProcess(path, os.Args, &os.ProcAttr{
		Dir:   dir,
		Env:   envs,
		Files: append([]*os.File{os.Stdin, os.Stdout, os.Stderr}, files...),
	})
	if err == nil {
		eudore.NewLoggerWithContext(ctx).Infof("eudore start new process %d", process.Pid)
	}
	return err
}
