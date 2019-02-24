package test

import (
	"time"
	"testing"
	"github.com/eudore/eudore"
	// _ "eudore/component/bone"
)

func TestSubRouter(t *testing.T) {
	e := eudore.NewEudore()
	t.Log(e.RegisterComponent("router", nil))
	t.Log(e.Router.Version())

	e.SubRoute("/api", eudore.NewRouterMust("", nil))
	r := eudore.GetSubRouter(e.Router, "/api")
	if r != nil {
		t.Log(r.Version())
	}
}

func TestRouterEmpty(t *testing.T) {
	app := eudore.NewCore()
	app.Listen(":8088")
	time.AfterFunc(5 * time.Second, func() {
		app.Server.Close()
	})
	app.RegisterComponent(eudore.ComponentRouterEmptyName, eudore.HandlerFunc(func(ctx eudore.Context){
		ctx.WriteString(app.Router.Version())
		t.Log(app.Router.Version())
	}))
	// app.RegisterComponent(eudore.ComponentRouterEmptyName, nil)
	app.Run()
}