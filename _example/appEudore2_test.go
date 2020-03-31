package eudore_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
)

func TestAppEudoreNew2(*testing.T) {
	app := eudore.NewEudore(
		context.Background(),
		eudore.NewConfigEudore(nil),
		eudore.NewRouterFull(),
		eudore.NewLoggerStd(nil),
		eudore.NewServerStd(nil),
		eudore.Binder(eudore.BindDefault),
		eudore.Renderer(eudore.RenderDefault),
		6666,
	)
	httptest.NewClient(app).Stop(0)
	app.AnyFunc("/*path", func(ctx eudore.Context) {
		ctx.WriteString("hello eudore core")
	})

	eudore.Set(app.Server, "", eudore.ServerStdConfig{
		ReadTimeout:  12 * time.Second,
		WriteTimeout: 4 * time.Second,
	})
	app.Listen(":8088")
	app.Listen(":8089")
	app.Listen(":8088")
	app.ListenTLS(":8088", "", "")
	app.ListenTLS(":8087", "", "")

	app.Listen("localhost")
	app.ListenTLS("localhost", "", "")
	app.Run()
}

func TestAppEudoreServerLogger2(t *testing.T) {
	app := eudore.NewEudore()
	app.RegisterInit("eudore-logger", 0x009, nil)
	app.RegisterInit("test-http", 0xb26, func(app *eudore.Eudore) error {
		app.AnyFunc("/*", func(ctx eudore.Context) {
			panic(9999)
		})
		app.Listen(":8088")
		http.Get("http://127.0.0.1:8088")
		return nil
	})
	httptest.NewClient(app).Stop(0)
	app.Run()
}

func TestAppEudoreInitListener2(t *testing.T) {
	app := eudore.NewEudore()
	app.RegisterInit("init-listen", 0x126, func(app *eudore.Eudore) error {
		app.Set("keys.handler", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Log(r.URL.Path)
			app.ServeHTTP(w, r)
		}))
		app.Set("listeners", []map[string]interface{}{
			{"addr": ":8088"},
			{"addr": ":8088"},
		})
		return nil
	})
	app.RegisterInit("eudore-server", 0xf0f, eudore.InitServer)
	httptest.NewClient(app).Stop(0)
	app.Run()
}

func TestAppEudoreInitIgnore2(t *testing.T) {
	app := eudore.NewEudore()
	app.RegisterInit("eudore-logger", 0x009, nil)
	app.RegisterInit("test-http", 0xb26, func(app *eudore.Eudore) error {
		return eudore.ErrEudoreInitIgnore
	})
	httptest.NewClient(app).Stop(0)
	app.Run()
}

func TestAppEudoreLogger2(*testing.T) {
	app := eudore.NewEudore()
	app.RegisterInit("init-logger", 0x016, func(app *eudore.Eudore) error {
		app.Debug(0)
		app.Info(1)
		app.Warning(2)
		app.Error(3)
		app.Fatal(4)
		return nil
	})
	app.Run()
}
func TestAppEudoreLoggerf2(*testing.T) {
	app := eudore.NewEudore()
	app.RegisterInit("init-logger", 0x016, func(app *eudore.Eudore) error {
		app.Debugf("0")
		app.Infof("1")
		app.Warningf("2")
		app.Errorf("3")
		app.Fatalf("4")
		return nil
	})
	app.Run()
}
