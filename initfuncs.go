package eudore

import (
	"os"
	"fmt"
	"syscall"
)

func InitSignal(e *Eudore) error {
	// Register signal
	// signal 9
	SignalRegister(syscall.SIGKILL, false, func() error {
		e.WithField("signal", 9).Info("eudore received SIGKILL, eudore stop HTTP server.")
		return e.Stop()
	})
	// signal 10
	SignalRegister(syscall.SIGUSR1, false, func() error {
		e.WithField("signal", 10).Info("eudore received SIGUSR1, eudore reloading HTTP server.")
		err := e.Init()
		if err != nil {
			e.Error("eudore reload error: ", err)
		}
		return err
	})
	// signal 12
	SignalRegister(syscall.SIGUSR2, false, func() error {
		e.WithField("signal", 12).Info("eudore received SIGUSR2, eudore restarting HTTP server.")
		err := e.Restart()
		if err != nil {
			e.Error("eudore reload error: ", err)
		}
		return err
	})
	// signal 15
	SignalRegister(syscall.SIGTERM, false, func() error {
		e.WithField("signal", 15).Info("eudore received SIGTERM, eudore shutting down HTTP server.")
		err := e.Shutdown()
        if err != nil {
            e.Error("eudore shutdown error: ", err)
        }
		return err
	})
	return nil
}

func InitConfig(app *Eudore) error {
	return app.Config.Parse()
}

func InitWorkdir(app *Eudore) error {
	dir := GetString(app.Config.Get("workdir"))
	if dir != "" {
		return os.Chdir(dir)
	}
	return nil
}

func InitCommand(app *Eudore) error {
	cmd := GetString(app.Config.Get("command"))
	pid := GetString(app.Config.Get("pidfile"))
	return NewCommand(cmd , pid).Run()
}


func InitLogger(app *Eudore) error {
	key := GetDefaultString(app.Config.Get("keys.logger"), "component.logger")
	c := app.Config.Get(key)
	if c != nil {
		err := app.RegisterComponent("", c)
		if err != nil {
			return err
		}
	}
	return nil
}

func InitServer(app *Eudore) error {
	key := GetDefaultString(app.Config.Get("keys.server"), "component.server")
	c := app.Config.Get(key)
	if c != nil {
		err := app.RegisterComponent("", c)
		if err != nil {
			return err
		}
	}
	return nil 
}

func InitServerStart(app *Eudore) error {
	if app.Server == nil {
		err := fmt.Errorf("Eudore can't start the service, the server is empty.")
		app.Error(err)
		return err
	}

	ComponentSet(app.Server, "config.handler", app)
	ComponentSet(app.Server, "errfunc", app.HandleError)
	go func() {
		app.stop <- app.Server.Start()
	}()
	return nil
}


func InitDefaultLogger(e *Eudore) error {
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
func InitDefaultServer(e *Eudore) error {
	if e.Server == nil {
		e.Warning("eudore use default server.")
		// return e.RegisterComponent(ComponentServerStdName, &ServerConfigGeneral{
		// 	Addr:		":8082",
		// 	Https:		false,
		// 	Handler:	e,
		// })
	}
	return nil
}


func InitListComponent(e *Eudore) error {
	e.Info("list all register component:", ComponentList())
	var cs []Component = []Component{
		e.Config,
		e.Server,
		e.Logger,
		e.Router,
		e.Cache,
		e.View,
		// e.Binder,
		// e.Renderer,
	}
	for _, c := range cs {
		if c != nil {
			e.Info(c.Version())		
		}
	}
	return nil
}


func InitStop(app *Eudore) error {
	if len(os.Getenv("stop")) > 0 {
		app.HandleError(ErrApplicationStop)
		// return ErrApplicationStop
	}
	return nil
}
