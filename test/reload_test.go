package test

import (
	"github.com/eudore/eudore"
	"testing"
)

func TestEudoreInit(t *testing.T) {
	app := eudore.NewEudore()
	app.RegisterInit("eudore-config", 0, nil)
	app.RegisterInit("eudore-workdir", 0, nil)
	app.RegisterInit("eudore-signal", 0, nil)
	app.RegisterInit("eudore-server-start", 0, nil)

	app.RegisterInit("a", 02, func(*eudore.Eudore) error {
		t.Log("aaa")
		return nil
	})
	app.RegisterInit("b", 04, func(*eudore.Eudore) error {
		t.Log("bbbb")
		return nil
	})
	app.RegisterInit("c", 01, func(*eudore.Eudore) error {
		t.Log("ccc")
		return nil
	})
	t.Log(app.Init("0x02-a", "c", "aa"))
	t.Log(app.Init())
	t.Log(app.Init("a"))
}
