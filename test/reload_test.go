package test

import (
	"testing"
	"eudore"
)

func TestReload(t *testing.T) {
	e := eudore.New()
	e.RegisterReload("a", 02, func(*eudore.Eudore) error {
			t.Log("aaa")
			return nil
		})
	e.RegisterReload("b", 04, func(*eudore.Eudore) error {
			t.Log("bbbb")
			return nil
		})
	e.RegisterReload("c", 01, func(*eudore.Eudore) error {
			t.Log("ccc")
			return nil
		})
	t.Log(e.Reload("0x02-a","c", "aa"))
	t.Log(e.Reload())
	t.Log(e.Reload("a"))
}