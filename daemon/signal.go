package daemon

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"github.com/eudore/eudore"
)

// Signal handle func.
type SignalFunc func(context.Context) error

type Signal struct {
	sync.Mutex
	Chan  chan os.Signal
	Funcs map[os.Signal][]SignalFunc
}

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

func (sig *Signal) Register(s os.Signal, fn SignalFunc) {
	sig.Lock()
	defer sig.Unlock()
	if fn == nil {
		delete(sig.Funcs, s)
	} else {
		sig.Funcs[s] = append(sig.Funcs[s], fn)
	}
	if len(sig.Funcs[s]) <= 1 {
		sig.Notify()
	}
}

func (sig *Signal) Handle(ctx context.Context, s os.Signal) error {
	sig.Lock()
	defer sig.Unlock()
	for _, fn := range sig.Funcs[s] {
		if err := fn(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (sig *Signal) Notify() {
	sigs := make([]os.Signal, 0, 4)
	for key := range sig.Funcs {
		sigs = append(sigs, key)
	}
	signal.Stop(sig.Chan)
	if sigs != nil {
		signal.Notify(sig.Chan, sigs...)
	}
}
