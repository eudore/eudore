package main

import (
	"context"
	"fmt"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/middleware"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	// "github.com/glebarez/sqlite"
	// "gorm.io/plugin/opentelemetry/tracing"
	gormlogger "gorm.io/gorm/logger"
	gormutils "gorm.io/gorm/utils"
)

type App struct {
	*eudore.App
	Config   *AppConfig
	Database *gorm.DB
}

type AppConfig struct {
	Database *DatabaseConfig
}

type DatabaseConfig struct {
	Dialector     func(string) gorm.Dialector `json:"-" alias:"-"`
	LoggerLevel   eudore.LoggerLevel          `json:"loggerlevel" alias:"loggerlevel"`
	SlowThreshold time.Duration               `json:"slowthreshold" alias:"slowthreshold"`
	MaxIdle       int                         `json:"maxidle" alias:"maxidle"`
	MaxOpen       int                         `json:"maxopen" alias:"maxopen"`
	MaxLifetime   time.Duration               `json:"maxlifetime" alias:"maxlifetime"`
	Type          string                      `json:"type" alias:"type"`
	Host          string                      `json:"host" alias:"host"`
	Port          string                      `json:"port" alias:"port"`
	User          string                      `json:"user" alias:"user"`
	Password      string                      `json:"password" alias:"password"`
	Name          string                      `json:"name" alias:"name"`
	Options       string                      `json:"options" alias:"options"`
	Success       string                      `json:"success" alias:"success"`
}

func main() {
	app := NewApp()
	app.Parse()
	app.Run()
}

func NewApp() *App {
	app := &App{
		App: eudore.NewApp(),
		Config: &AppConfig{
			Database: &DatabaseConfig{},
		},
	}
	app.SetValue(eudore.ContextKeyConfig, eudore.NewConfig(app.Config))
	app.ParseOption(
		app.NewParseDatabaseFunc(),
		app.NewParseRouterFunc(),
	)
	return app
}

func (app *App) NewParseDatabaseFunc() eudore.ConfigParseFunc {
	return func(context.Context, eudore.Config) error {
		config := app.Config.Database
		ormconfig := &gorm.Config{
			Logger: NewGromLogger(app.Logger, config.LoggerLevel, config.SlowThreshold),
		}
		config.Type = eudore.GetAnyDefault(config.Type, "sqlite")

		var dsn string
		switch config.Type {
		case "sqlite":
			config.Host = eudore.GetAnyDefault(config.Host, "sqlite.db")
			config.Success = fmt.Sprintf("init database to sqlite %s", config.Host)
			config.Dialector = sqlite.Open
			dsn = config.Host
		case "postgres":
			config.Host = eudore.GetAnyDefault(config.Host, "127.0.0.1")
			config.Port = eudore.GetAnyDefault(config.Port, "5432")
			config.User = eudore.GetAnyDefault(config.User, "postgres")
			config.Name = eudore.GetAnyDefault(config.Name, "postgres")
			config.Password = eudore.GetAnyDefault(config.Password, "postgres")
			config.Options = eudore.GetAnyDefault(config.Options, "sslmode=disable")
			config.Success = fmt.Sprintf("init database to postgres %s:%s/%s",
				config.Host, config.Port, config.Name,
			)
			// config.Dialector = postgres.Open
			dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s %s",
				config.Host, config.Port, config.User,
				config.Password, config.Name, config.Options,
			)
		case "mysql":
			config.Host = eudore.GetAnyDefault(config.Host, "127.0.0.1")
			config.Port = eudore.GetAnyDefault(config.Port, "3306")
			config.User = eudore.GetAnyDefault(config.User, "mysql")
			config.Name = eudore.GetAnyDefault(config.Name, "mysql")
			config.Password = eudore.GetAnyDefault(config.Password, "mysql")
			config.Options = eudore.GetAnyDefault(config.Options,
				"charset=utf8mb4&parseTime=True&loc=Local",
			)
			config.Success = fmt.Sprintf("init database to mysql %s:%s/%s",
				config.Host, config.Port, config.Name,
			)
			// config.Dialector = mysql.Open
			dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
				config.User, config.Password, config.Host,
				config.Port, config.Name, config.Options,
			)
		default:
			return fmt.Errorf("eudore init database error: undefine database driver: '%s'", config.Type)
		}
		db, err := gorm.Open(config.Dialector(dsn), ormconfig)
		if err != nil {
			return fmt.Errorf("eudore init database error: %s", err.Error())
		}

		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		sqlDB.SetMaxIdleConns(3)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(24 * time.Hour)
		app.Database = db
		app.Info(config.Success)
		return nil
	}
}

func (app *App) NewParseRouterFunc() eudore.ConfigParseFunc {
	return func(context.Context, eudore.Config) error {
		app.AddMiddleware(
			middleware.NewLoggerFunc(app),
			middleware.NewRequestIDFunc(nil),
			middleware.NewRecoveryFunc(),
		)
		err := app.AddController(NewFilesController(app))
		if err != nil {
			return err
		}
		return app.Listen(":8089")
	}
}

type File struct {
	ID        uint `gorm:"primarykey"`
	Path      string
	Size      int
	Hash      string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

type FilesController struct {
	eudore.ControllerAutoRoute
	Database *gorm.DB
}

func NewFilesController(app *App) eudore.Controller {
	err := app.Database.AutoMigrate(new(File))
	if err != nil {
		return eudore.NewControllerError(&FilesController{}, err)
	}
	return &FilesController{
		Database: app.Database,
	}
}

func (ctl *FilesController) Get(ctx eudore.Context) (any, error) {
	tx := ctl.Database.WithContext(ctx.Context())
	result := tx.Create(&File{
		Path: "app.go",
		Hash: "e10adc3949ba59abbe56e057f20f883e",
	})
	if result.Error != nil {
		return nil, result.Error
	}

	var list []File
	result = tx.Find(&list)
	if result.Error != nil {
		return nil, result.Error
	}
	return list, nil
}

type gormLogger struct {
	Logger        eudore.Logger
	LogLevel      eudore.LoggerLevel
	SlowThreshold time.Duration
}

// NewGromLogger 函数适配eudore.Logger实现gorm/logger接口。
func NewGromLogger(logger eudore.Logger, level eudore.LoggerLevel,
	slow time.Duration,
) gormlogger.Interface {
	return &gormLogger{
		Logger:        logger,
		LogLevel:      level,
		SlowThreshold: slow,
	}
}

func (l gormLogger) getLogger(ctx context.Context) eudore.Logger {
	log, ok := ctx.Value(eudore.ContextKeyLogger).(eudore.Logger)
	if ok {
		return log
	}
	return l.Logger
}

var levelMapping = map[gormlogger.LogLevel]eudore.LoggerLevel{
	gormlogger.Silent: eudore.LoggerFatal,
	gormlogger.Error:  eudore.LoggerError,
	gormlogger.Warn:   eudore.LoggerWarning,
	gormlogger.Info:   eudore.LoggerInfo,
}

func (l *gormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	el, ok := levelMapping[level]
	if !ok {
		el = eudore.LoggerDebug
	}
	newlogger := *l
	newlogger.LogLevel = el
	return &newlogger
}

func (l gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= eudore.LoggerInfo {
		l.getLogger(ctx).Infof(msg, data...)
	}
}

func (l gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= eudore.LoggerWarning {
		l.getLogger(ctx).Warningf(msg, data...)
	}
}

func (l gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= eudore.LoggerError {
		l.getLogger(ctx).Errorf(msg, data...)
	}
}

func (l gormLogger) Trace(ctx context.Context, begin time.Time,
	fc func() (string, int64), err error,
) {
	sql, rows := fc()
	if l.LogLevel < eudore.LoggerFatal {
		elapsed := time.Since(begin)
		log := l.getLogger(ctx).WithFields(
			[]string{"sqltime", "sql", "file"},
			[]interface{}{
				fmt.Sprintf("%.3fms", float64(elapsed.Nanoseconds())/1e6),
				sql,
				gormutils.FileWithLineNum(),
			})
		if rows != -1 {
			log.WithField("rows", rows)
		}
		switch {
		case err != nil && l.LogLevel >= eudore.LoggerError:
			log.Error(err.Error())
		case elapsed > l.SlowThreshold &&
			l.SlowThreshold != 0 &&
			l.LogLevel >= eudore.LoggerWarning:
			log.Warningf("SLOW SQL >= %v", l.SlowThreshold)
		case l.LogLevel <= eudore.LoggerInfo:
			log.Info()
		}
	}
}
