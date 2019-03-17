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
		return e.Shutdown()
	})
	return nil
}

func InitConfig(app *Eudore) error {
	return app.Config.Parse()
}

func InitCommand(app *Eudore) error {
	cmd := app.Config.Get("#command").(string)
	pid := app.Config.Get("#pidfile").(string)
	return NewCommand(cmd , pid).Run()
}


func InitLogger(e *Eudore) error {
	c := e.Config.Get("#logger")
	if c != nil {
		name := GetComponetName(c)
		// load logger
		if len(name) > 0 {
			err := e.RegisterComponent(name, c)
			if err != nil {
				return err
			}
		}else {
			return fmt.Errorf("logger name is nil.")
		}
	}
	return nil
}

func InitServer(app *Eudore) error {
	// Json(e.Config)
	c := app.Config.Get("#server")
	if c != nil {
		name := GetComponetName(c)
		if len(name) > 0 {
			err := app.RegisterComponent(name, c)
			if err != nil {
				return err
			}
		}else {
			return fmt.Errorf("server name is nil.")
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

	SetComponent(app.Server, "errorhandle", app.HandleError)
	SetComponent(app.Server, "handler", app)
	go func() {
		app.stop <- app.Server.Start()
	}()
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
		// return e.RegisterComponent(ComponentServerStdName, &ServerConfigGeneral{
		// 	Addr:		":8082",
		// 	Https:		false,
		// 	Handler:	e,
		// })
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
