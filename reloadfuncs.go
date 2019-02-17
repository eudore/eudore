package eudore

import (
	"os"
	"fmt"
	"syscall"
	"github.com/eudore/eudore/config"
)

func ReloadSignal(e *Eudore) error {
	// Register signal
	// signal 9
	SignalRegister(syscall.SIGKILL, false, func() error {
		// graceOutput("received SIGKILL, graceful shutting down HTTP server.")
		return e.Stop()
	})
	// signal 10
	SignalRegister(syscall.SIGUSR1, false, func() error {
		// graceOutput("received SIGUSR1, graceful reloading HTTP server.")
		e.WithField("signal", 12).Info("eudore accept signal 12")
		err := e.Reload()
		if err != nil {
			e.Error("eudore reload error: ", err)
		}
		return err
	})
	// signal 12
	SignalRegister(syscall.SIGUSR2, false, func() error {
		// graceOutput("received SIGUSR2, graceful restarting HTTP server.")
		return e.Restart()
	})
	// signal 15
	SignalRegister(syscall.SIGTERM, false, func() error {
		// graceOutput("received SIGTERM, graceful shutting down HTTP server.")
		e.Debug("signal 15")
		return e.Shutdown()
	})
	return nil
}

func ReloadConfig(*Eudore) error {
	return nil
}

func ReloadKeys(e *Eudore) error {
	// Get #keys data.
	c := e.Config.Get("#keys")
	if c == nil {
		return nil
	}
	// Check if the type of #key is map[string]string.
	// Get all key-values data.
	keys, values, err := config.AllKeyVal(c)
	if err != nil {
		return err
	}
	// Set data.
	for i, v  := range values {
		e.Config.Set(keys[i], v)
	}
	return nil
}



func ReloadLogger(e *Eudore) error {
	c := e.Config.Get("#logger")
	if c != nil {
		name := GetComponetName(c)
		// load logger
		if len(name) > 0 {
			err := e.RegisterComponent(name, c)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func ReloadServer(e *Eudore) error {
	c := e.Config.Get("#server")
	if c != nil {
		name := GetComponetName(c)
		if len(name) > 0 {
			err := e.RegisterComponent(name, c)
			if err != nil {
				return err
			}
		}
	}
	return nil
}


func ReloadDefaultLogger(e *Eudore) error {
	if _, ok := e.Logger.(*LoggerInit); ok {
		e.Warning("eudore use default logger.")
		return e.RegisterComponent(ComponentLoggerStdName, &LoggerStdConfig{
			Std:		true,
			Level:		0,
			Format:		"json",
		})
	}
	return nil
}
func ReloadDefaultServer(e *Eudore) error {
	if e.Server == nil {
		e.Warning("eudore use default server.")
		return e.RegisterComponent(ComponentServerStdName, &ServerConfigGeneral{
			Addr:		":8082",
			Https:		false,
			Handler:	e,
		})
	}
	return nil
}


func ReloadListComponent(e *Eudore) error {
	e.Info("list all register component:", ListComponent())
	var cs []Component = []Component{
		e.Config,
		e.Server,
		e.Logger,
		e.Router,
		// e.Binder,
		// e.Renderer,
		e.View,
		e.Cache,
	}
	for _, c := range cs {
		if c != nil {
			e.Info(c.Version())		
		}
	}
	return nil
}


func ReloadStop(*Eudore) error {
	if len(os.Getenv("stop")) > 0 {
		return fmt.Errorf("stop")
	}
	return nil
}
