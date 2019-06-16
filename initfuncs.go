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
		}else {
			e.Info("eudore restart success.")
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
		app.Info("changes working directory to: " + dir)
		return os.Chdir(dir)
	}
	return nil
}

func InitCommand(app *Eudore) error {
	cmd := GetDefaultString(app.Config.Get("command"), "start")
	pid := GetDefaultString(app.Config.Get("pidfile"), "/var/run/eudore.pid")
	app.Infof("current command is %s, pidfile in %s.", cmd, pid)
	app.cmd.Reset(cmd , pid)
	return app.cmd.Run()
}


func InitLogger(app *Eudore) error {
	key := GetDefaultString(app.Config.Get("keys.logger"), "component.logger")
	c := app.Config.Get(key)
	if c != nil {
		_, err := app.RegisterComponent("", c)
		if err != nil {
			return err
		}
		ComponentSet(app.Router, "print", app.Logger.Debug)
		Set(app.Server, "print", app.Logger.Debug)
	}
	return nil
}

func InitServer(app *Eudore) error {
	key := GetDefaultString(app.Config.Get("keys.server"), "component.server")
	c := app.Config.Get(key)
	if c != nil {
		_, err := app.RegisterComponent("", c)
		if err != nil {
			return err
		}
		Set(app.Server, "print", app.Logger.Debug)
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
	ComponentSet(app.Server, "errfunc", func(err error) {
		fields := make(Fields)
		file, line := LogFormatFileLine(0)
		fields["component"] = app.Server.GetName()
		fields["file"] = file
		fields["line"] = line
		app.WithFields(fields).Errorf("server error: %v", err)
	})
	go func() {
		app.stop <- app.Server.Start()
	}()
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
		e.Session,
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
