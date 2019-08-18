package test

import (
	"fmt"
	"github.com/eudore/eudore"
	"syscall"
	"testing"
	"time"
)

func TestSignal(t *testing.T) {
	app := eudore.NewEudore()
	app.RegisterSignal(syscall.Signal(0x01), func(*eudore.Eudore) error {
		t.Log("0x01")
		return nil
	})
	app.RegisterSignal(syscall.Signal(0x00), func(*eudore.Eudore) error {
		t.Log("222")
		return nil
	})
	app.RegisterSignal(syscall.Signal(0x00), func(*eudore.Eudore) error {
		t.Log("1111")
		return nil
	})
	app.RegisterSignal(syscall.Signal(0x00), func(*eudore.Eudore) error {
		t.Log("3333")
		return fmt.Errorf("error test 333")
	})
	app.HandleSignal(syscall.Signal(0x00))
	app.HandleSignal(syscall.Signal(0x01))
	time.Sleep(1 * time.Second)
}
