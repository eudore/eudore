package test

import (
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