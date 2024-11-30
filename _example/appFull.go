package main

import (
	"context"
	"database/sql"
	"embed"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/daemon"
	"github.com/eudore/eudore/middleware"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// App 定义一个简单app
type App struct {
	*eudore.App
	*Config
	Database *sql.DB
}

type Config struct {
	Workdir string
	Pidfile string
	Command string
	Logger  *eudore.LoggerConfig
	Server  *eudore.ServerConfig
	Client  *eudore.ClientOption
	Listen  *eudore.ServerListenConfig

	Name string `alias:"name" json:"name"`
}

//go:embed app*.go
var root embed.FS

func main() {
	app := NewApp()
	app.Parse()
	app.Run()
}

// NewApp 方法创建一个自定义app
func NewApp() *App {
	app := &App{
		App: eudore.NewApp(),
		Config: &Config{
			Name: "eudore",
			Logger: &eudore.LoggerConfig{
				Stdout:   true,
				StdColor: true,
			},
			Listen: &eudore.ServerListenConfig{
				Addr: ":8088",
			},
		},
	}
	app.SetValue(eudore.ContextKeyLogger, eudore.NewLoggerInit())
	app.SetValue(eudore.ContextKeyConfig, eudore.NewConfig(app.Config))
	app.SetValue(eudore.ContextKeyBind, eudore.NewHandlerDataFuncs(
		eudore.NewHandlerDataBinds(nil),
		eudore.NewHandlerDataValidateStruct(app),
	))
	app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))

	app.ParseOption()
	app.ParseOption(
		eudore.NewConfigParseEnvFile(),
		eudore.NewConfigParseDefault(),
		eudore.NewConfigParseJSON("config"),
		eudore.NewConfigParseEnvs("ENV_"),
		eudore.NewConfigParseArgs(),
		eudore.NewConfigParseWorkdir("workdir"),
		daemon.NewParseCommand(),
		app.NewParseLoggerFunc(),
		app.NewParseServerFunc(),
		app.NewParseClientFunc(),
		app.NewParseDatabaseFunc(),
		app.NewParseRouterFunc(),
		app.NewParseListenFunc(),
		daemon.NewParseSignal(),
	)

	return app
}

func (app *App) NewParseLoggerFunc() eudore.ConfigParseFunc {
	return func(context.Context, eudore.Config) error {
		app.SetValue(eudore.ContextKeyLogger, eudore.NewLogger(app.Config.Logger))
		app.SetValue(eudore.ContextKeyContextPool, eudore.NewContextBasePool(app))
		return nil
	}
}

func (app *App) NewParseServerFunc() eudore.ConfigParseFunc {
	return func(context.Context, eudore.Config) error {
		app.SetValue(eudore.ContextKeyServer, eudore.NewServer(app.Config.Server))
		return nil
	}
}

func (app *App) NewParseClientFunc() eudore.ConfigParseFunc {
	return func(context.Context, eudore.Config) error {
		app.SetValue(eudore.ContextKeyClient, eudore.NewClientCustom(
			eudore.NewClientHookTimeout(eudore.DefaultClientTimeout),
			eudore.NewClientHookRedirect(nil),
			eudore.NewClientHookRetry(3, nil, nil),
			eudore.NewClientHookLogger(eudore.LoggerInfo, time.Second*30),
			app.Config.Client,
		))
		return nil
	}
}

func (app *App) NewParseDatabaseFunc() eudore.ConfigParseFunc {
	return func(context.Context, eudore.Config) error {
		// db, err := sql.Open
		// if err != nil {
		// 	return nil
		// }
		// app.Database = db
		return nil
	}
}

func (app *App) NewParseRouterFunc() eudore.ConfigParseFunc {
	return func(context.Context, eudore.Config) error {
		app.AddMiddleware("global",
			middleware.NewLoggerFunc(app),
			middleware.NewHeaderDeleteFunc(nil, nil),
			middleware.NewRequestIDFunc(func(eudore.Context) string {
				return uuid.New().String()
			}),
			middleware.NewCompressionMixinsFunc(nil),
		)
		app.AddHandler("404", "", eudore.HandlerRouter404)
		app.AddHandler("405", "", eudore.HandlerRouter405)
		app.GetFunc("/metrics", promhttp.Handler())
		app.GetFunc("/heath", middleware.NewHealthCheckFunc(app))
		app.GetFunc("/static/*", eudore.NewHandlerFileSystems(root, "."))

		app.AddMiddleware(
			middleware.NewRecoveryFunc(),
			middleware.NewTimeoutFunc(app.ContextPool, time.Second*10),
		)

		return app.AddController(
			NewUsersController(app),
			NewEventsController(app),
		)
	}
}

func (app *App) NewParseListenFunc() eudore.ConfigParseFunc {
	return func(context.Context, eudore.Config) error {
		ln, err := app.Config.Listen.Listen()
		if err != nil {
			return nil
		}
		app.Infof("listen in %s %s", ln.Addr().Network(), ln.Addr().String())
		app.Serve(ln)
		return nil
	}
}

type User struct {
	ID   int
	Name string
}

type UsersController struct {
	eudore.ControllerAutoType[User]
	Config   *Config
	Database *sql.DB
}

// NewUserHandlr 方法闭包传递app对象，然后使用数据库进行操作。
func NewUsersController(app *App) eudore.Controller {
	return &UsersController{
		Config:   app.Config,
		Database: app.Database,
	}
}

func (ctl *UsersController) Get(ctx eudore.Context) {
	ctx.WriteString(ctl.Config.Name)
}

func (ctl *UsersController) Post(ctx eudore.Context, user *User) error {
	return nil
}

type EventsController struct {
	eudore.ControllerAutoRoute
}

func NewEventsController(app *App) eudore.Controller {
	return &EventsController{}
}
