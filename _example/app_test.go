package eudore_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/eudore/eudore"
	. "github.com/eudore/eudore/middleware"
	_ "golang.org/x/tools/cover"
)

func init() {
	DefaultLoggerFormatterFormatTime = "none"
	DefaultRouterLoggerKind = "~all|metadata"
}

func TestAppRun(*testing.T) {
	app := NewApp()
	app.SetValue(ContextKeyLogger, NewLogger(&LoggerConfig{
		Stdout:   true,
		StdColor: true,
		HookMeta: true,
	}))
	app.SetValue(ContextKeyRender, HandlerDataRenderJSON)
	app.SetValue(ContextKeyContextPool, NewContextBasePool(app))

	app.AddMiddleware(HandlerEmpty)
	app.AddMiddleware("global", HandlerEmpty)
	app.GetFunc("/health", NewHealthCheckFunc(app))
	app.GetFunc("/hello", func(ctx Context) {
		ctx.WriteString("hello eudore")
	})

	keys := []any{
		ContextKeyApp,
		ContextKeyAppCancel,
		ContextKeyAppValues,
		ContextKeyLogger,
		ContextKeyConfig,
		ContextKeyClient,
		ContextKeyRouter,
	}
	for _, k := range keys {
		app.Value(k)
	}
	app.SetValue("stop-hook", Unmounter(func(context.Context) {
		app.Info("stop-hook")
	}))
	app.NewRequest("GET", "/health")

	app.Parse()
	app.ParseOption(func(context.Context, Config) error {
		return errors.New("parse error")
	})
	app.Parse()
	app.SetValue(ContextKeyError, "stop app")

	app.CancelFunc()
	app.Run()
}

func TestAppListen(*testing.T) {
	app := NewApp()
	app.Listen(":8088")
	app.Listen(":8088")
	app.Listen(":8089")
	app.ListenTLS(":8088", "", "")
	app.ListenTLS(":8090", "", "")
	app.Listen("localhost")
	app.ListenTLS("localhost", "", "")

	app.CancelFunc()
	app.Run()
}
