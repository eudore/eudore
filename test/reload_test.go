package test

import (
	"github.com/eudore/eudore"
	"testing"
)

func TestEudoreInit(t *testing.T) {
	e := eudore.NewEudore()
	e.RegisterInit("a", 02, func(*eudore.Eudore) error {
		t.Log("aaa")
		return nil
	})
	e.RegisterInit("b", 04, func(*eudore.Eudore) error {
		t.Log("bbbb")
		return nil
	})
	e.RegisterInit("c", 01, func(*eudore.Eudore) error {
		t.Log("ccc")
		return nil
	})
	t.Log(e.Init("0x02-a", "c", "aa"))
	t.Log(e.Init())
	t.Log(e.Init("a"))
}
