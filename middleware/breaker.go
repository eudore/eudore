package middleware

import (
	"sync"
	"time"

	"github.com/eudore/eudore"
)

// Define the breaker state.
const (
	breakerStatueClosed = iota
	breakerStatueHalfOpen
	breakerStatueOpen
)

// breakerStatues defines the breaker status string.
var breakerStatues = []string{"closed", "half-open", "open"}

type breaker struct {
	sync.RWMutex
	Index              int
	Routes             map[string]breakerEntry
	Mapping            map[int]string
	GetKeyFunc         func(eudore.Context) string
	GetBreakrEntryFunc func(int, string) breakerEntry
}

type breakerEntry interface {
	OnAccess() bool
	OnSucceed() bool
	OnFailed() bool
	SetState(state int)
}

// The NewCircuitBreakerFunc function creates middleware to implement
// handle request breaking.
//
// This middleware does not support cluster mode.
//
// options: [NewOptionKeyFunc]
// [NewOptionCircuitBreakerConfig] [NewOptionRouter].
func NewCircuitBreakerFunc(options ...Option) Middleware {
	b := &breaker{
		Routes:  make(map[string]breakerEntry),
		Mapping: make(map[int]string),
		GetKeyFunc: func(ctx eudore.Context) string {
			return ctx.GetParam(eudore.ParamRoute)
		},
		GetBreakrEntryFunc: newBreakerEntryfunc(10, 10,
			400*time.Microsecond, 10*time.Second,
		),
	}
	applyOption(b, options)

	return func(ctx eudore.Context) {
		name := b.GetKeyFunc(ctx)
		if name == "" {
			return
		}

		b.RLock()
		entry, ok := b.Routes[name]
		b.RUnlock()
		if !ok {
			b.Lock()
			b.Index++
			entry = b.GetBreakrEntryFunc(b.Index, name)
			b.Routes[name] = entry
			b.Mapping[b.Index] = name
			b.Unlock()
		}

		if entry.OnAccess() {
			// ignore panic
			ctx.Next()
			if ctx.Response().Status() < eudore.StatusInternalServerError {
				if entry.OnSucceed() {
					ctx.Infof("Breaker route %s change state to %s",
						name, breakerStatues[breakerStatueClosed],
					)
				}
			} else {
				if entry.OnFailed() {
					ctx.Infof("Breaker route %s change state to %s",
						name, breakerStatues[breakerStatueOpen],
					)
				}
			}
		} else {
			writePage(ctx, eudore.StatusServiceUnavailable,
				DefaultPageCircuitBreaker, name,
			)
			ctx.End()
		}
	}
}

func (b *breaker) data(ctx eudore.Context) {
	b.RLock()
	_ = ctx.Render(b.Routes)
	b.RUnlock()
}

func (b *breaker) get(ctx eudore.Context) {
	id := eudore.GetAnyByString(ctx.GetParam("id"), -1)
	if id < 0 || id > b.Index {
		ctx.Fatal("id is invalid")
		return
	}
	b.RLock()
	entry := b.Routes[b.Mapping[id]]
	_ = ctx.Render(entry)
	b.RUnlock()
}

func (b *breaker) putState(ctx eudore.Context) {
	id := eudore.GetAnyByString(ctx.GetParam("id"), -1)
	if id < 0 || id > b.Index {
		ctx.Fatal("id is invalid")
		return
	}
	state := eudore.GetAnyByString[int](ctx.GetParam("state"))
	if state < -1 || state > 2 {
		ctx.Fatal("state is invalid")
		return
	}
	b.RLock()
	name := b.Mapping[id]
	entry := b.Routes[name]
	b.RUnlock()
	entry.SetState(state)
	ctx.Infof("Breaker route %s set state to %s", name, breakerStatues[state])
}

// breakerEntryDefault defines the breaker data for a single entry.
type breakerEntryDefault struct {
	sync.Mutex `json:"-"`
	ID         int    `json:"id"`
	State      int    `json:"state"`
	Name       string `json:"name"`
	// config
	MaxConsecutiveSuccesses int           `json:"-"`
	MaxConsecutiveFailures  int           `json:"-"`
	HalfOpenWait            time.Duration `json:"-"`
	HalfOpenInterval        time.Duration `json:"-"`
	HalfOpenLast            time.Time     `json:"-"`
	// state
	LastTime             time.Time `json:"lastTime"`
	ConsecutiveSuccesses int       `json:"consecutiveSuccesses"`
	ConsecutiveFailures  int       `json:"consecutiveFailures"`
	TotalSuccesses       uint64    `json:"totalSuccesses"`
	TotalFailures        uint64    `json:"totalFailures"`
}

func newBreakerEntryfunc(maxSuccesses, maxFailures int, t, wait time.Duration,
) func(id int, name string) breakerEntry {
	return func(id int, name string) breakerEntry {
		return &breakerEntryDefault{
			ID:                      id,
			Name:                    name,
			MaxConsecutiveSuccesses: maxSuccesses,
			MaxConsecutiveFailures:  maxFailures,
			HalfOpenInterval:        t,
			HalfOpenWait:            wait,
		}
	}
}

func (c *breakerEntryDefault) OnAccess() bool {
	now := time.Now()
	c.Lock()
	c.LastTime = now
	allow := c.State == breakerStatueClosed ||
		(c.State == breakerStatueHalfOpen && c.OnHalfOpen(now))
	c.Unlock()
	return allow
}

func (c *breakerEntryDefault) OnHalfOpen(now time.Time) bool {
	if now.After(c.HalfOpenLast.Add(c.HalfOpenInterval)) {
		c.HalfOpenLast = now
		return true
	}
	return false
}

func (c *breakerEntryDefault) OnSucceed() bool {
	c.Lock()
	c.TotalSuccesses++
	c.ConsecutiveSuccesses++
	c.ConsecutiveFailures = 0
	change := c.State != breakerStatueClosed &&
		c.ConsecutiveSuccesses >= c.MaxConsecutiveSuccesses
	if change {
		c.ConsecutiveSuccesses = 0
		c.State = breakerStatueClosed
	}
	c.Unlock()
	return change
}

func (c *breakerEntryDefault) OnFailed() bool {
	c.Lock()
	c.TotalFailures++
	c.ConsecutiveFailures++
	c.ConsecutiveSuccesses = 0
	change := c.State != breakerStatueOpen &&
		c.ConsecutiveFailures >= c.MaxConsecutiveFailures
	if change {
		c.ConsecutiveFailures = 0
		c.State = breakerStatueOpen
		c.RetryClose()
	}
	c.Unlock()
	return change
}

func (c *breakerEntryDefault) SetState(state int) {
	c.Lock()
	c.State = state
	c.ConsecutiveSuccesses = 0
	c.ConsecutiveFailures = 0
	c.RetryClose()
	c.Unlock()
}

func (c *breakerEntryDefault) RetryClose() {
	if c.State == breakerStatueOpen {
		go func() {
			time.Sleep(c.HalfOpenWait)
			c.Lock()
			if c.State == breakerStatueOpen {
				c.State--
			}
			c.Unlock()
		}()
	}
}
