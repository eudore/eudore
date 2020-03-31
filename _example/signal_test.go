package eudore_test

import (
	"errors"
	"github.com/eudore/eudore"
	"github.com/eudore/eudore/component/httptest"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestSignalRun2(t *testing.T) {
	app := eudore.NewCore()

	sig := eudore.NewSignaler()
	sig.RegisterSignal(syscall.Signal(0x01), func() error {
		t.Log("signal 1")
		return nil
	})
	sig.RegisterSignal(syscall.Signal(0x03), func() error {
		return errors.New("test signal err")
	})
	go sig.Run(app.Context)

	proc, _ := os.FindProcess(os.Getpid())
	proc.Signal(syscall.Signal(0x01))
	time.Sleep(200 * time.Microsecond)
	sig.HandleSignal(syscall.Signal(0x03))

	app.Run()
}

func TestSignalEudore2(t *testing.T) {
	app := eudore.NewEudore()
	app.RegisterSignal(syscall.Signal(0x01), func() error {
		t.Log("signal 1", os.Args)
		return nil
	})
	app.RegisterSignal(syscall.Signal(0x03), func() error {
		return errors.New("test signal err")
	})
	proc, _ := os.FindProcess(os.Getpid())
	proc.Signal(syscall.Signal(0x01))
	time.Sleep(200 * time.Microsecond)
	proc.Signal(syscall.Signal(0x01))
	time.Sleep(200 * time.Microsecond)
	proc.Signal(syscall.Signal(0x03))
	time.Sleep(200 * time.Microsecond)
	proc.Signal(syscall.Signal(0x03))
	httptest.NewClient(app).Stop(0)
	app.Run()
}
