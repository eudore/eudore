package daemon

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"github.com/eudore/eudore"
)

// SignalFunc type based on the usage in Handle.
type SignalFunc func(context.Context) error

// Signal defines the structure for managing and listening for [os.Signal],
// and calling the corresponding handler functions.
type Signal struct {
	sync.Mutex
	Chan  chan os.Signal
	Funcs map[os.Signal][]SignalFunc
}

// Run method processes signals sent via [signal.Notify].
// It continuously listens for signals on the channel until the context is cancelled.
func (sig *Signal) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case s := <-sig.Chan:
			log := eudore.NewLoggerWithContext(ctx)
			log.Infof("eudore accept signal: %s", s)
			err := sig.Handle(ctx, s)
			if err != nil {
				log.Errorf("eudore handle signal %s error: %v", s, err)
			}
		}
	}
}

// Register method registers a handler function for a specified [os.Signal].
// If fn is nil, it unregisters the handlers for the signal.
func (sig *Signal) Register(s os.Signal, fn SignalFunc) {
	sig.Lock()
	defer sig.Unlock()
	if fn == nil {
		delete(sig.Funcs, s)
	} else {
		sig.Funcs[s] = append(sig.Funcs[s], fn)
	}
	// Re-notify the OS for signal listening if this is the first handler registered
	// or the only handler for this signal.
	if len(sig.Funcs[s]) <= 1 {
		sig.Notify()
	}
}

// Handle method processes the specified [os.Signal] by executing all registered functions.
func (sig *Signal) Handle(ctx context.Context, s os.Signal) error {
	sig.Lock()
	defer sig.Unlock()
	for _, fn := range sig.Funcs[s] {
		err := fn(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// Notify method re-initializes the OS signal listening to include all currently
// registered signals in the Funcs map.
func (sig *Signal) Notify() {
	sigs := make([]os.Signal, 0, 4)
	for key := range sig.Funcs {
		sigs = append(sigs, key)
	}
	// Stop previous notifications to clear the channel
	signal.Stop(sig.Chan)
	// Start listening for the new set of signals
	if sigs != nil {
		signal.Notify(sig.Chan, sigs...)
	}
}
