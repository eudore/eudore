package test

import (
	"testing"
	"eudore"
	// _ "eudore/component/bone"
)

func TestSubRouter(t *testing.T) {
	e := eudore.New()
	t.Log(e.RegisterComponent("router", nil))
	t.Log(e.Router.Version())

	e.RegisterSubRoute("/api", eudore.NewRouterMust("", nil))
	r := eudore.GetSubRouter(e.Router, "/api")
	if r != nil {
		t.Log(r.Version())
	}
}