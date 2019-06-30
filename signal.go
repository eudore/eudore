package eudore

import (
	"os"
	"os/signal"
	"sync"
)

var (
	graceSignalMu     sync.Mutex
	graceSignalTables []os.Signal
	graceSignalChan   chan os.Signal
	graceSignalFuncs  map[os.Signal][]SignalFunc
)

// Signal handle func.
type SignalFunc func() error

func init() {
	graceSignalChan = make(chan os.Signal)
	graceSignalFuncs = make(map[os.Signal][]SignalFunc)
	go func() {
		for {
			SignalHandle(<-graceSignalChan)
		}
	}()
}

// Trigger Signal
func SignalHandle(sig os.Signal) error {
	graceSignalMu.Lock()
	defer graceSignalMu.Unlock()
	fns, ok := graceSignalFuncs[sig]
	if !ok {
		return nil
	}
	for _, fn := range fns {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}

// Add Listen Signal
func SignalRegister(sig os.Signal, bf bool, fn SignalFunc) {
	graceSignalMu.Lock()
	defer graceSignalMu.Unlock()
	if bf {
		graceSignalFuncs[sig] = append([]SignalFunc{fn}, graceSignalFuncs[sig]...)
	} else {
		graceSignalFuncs[sig] = append(graceSignalFuncs[sig], fn)
	}

	for _, s := range graceSignalTables {
		if s == sig {
			return
		}
	}
	graceSignalTables = append(graceSignalTables, sig)
	SignalListen(graceSignalTables)
}

// Relisten Signal is sigs
func SignalListen(sigs []os.Signal) {
	signal.Stop(graceSignalChan)
	signal.Notify(
		graceSignalChan,
		sigs...,
	)
}
