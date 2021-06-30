package middleware

import (
	"sync"
	"time"

	"github.com/eudore/eudore"
)

// 定义熔断器状态。
const (
	BreakerStatueClosed BreakerState = iota
	BreakerStatueHalfOpen
	BreakerStatueOpen
)

// BreakerStatues 定义熔断状态字符串
var BreakerStatues = []string{"closed", "half-open", "open"}

// BreakerState 是熔断器状态。
type BreakerState int8

// Breaker 定义熔断器。
type Breaker struct {
	sync.RWMutex            `json:"-"`
	Index                   int                      `json:"index"`
	Mapping                 map[int]string           `json:"mapping"`
	Routes                  map[string]*breakRoute   `json:"routes"`
	MaxConsecutiveSuccesses uint32                   `json:"maxconsecutivesuccesses"`
	MaxConsecutiveFailures  uint32                   `json:"maxconsecutivefailures"`
	OpenWait                time.Duration            `json:"openwait"`
	NewHalfOpen             func(string) func() bool `json:"-"`
}

// breakRoute 定义单词路由的熔断数据。
type breakRoute struct {
	sync.Mutex           `json:"-"`
	breaker              *Breaker
	ID                   int          `json:"id"`
	Name                 string       `json:"name"`
	BreakerState         BreakerState `json:"state"`
	OnHalfOpen           func() bool  `json:"-"`
	LastTime             time.Time    `json:"lasttime"`
	TotalSuccesses       uint64       `json:"totalsuccesses"`
	TotalFailures        uint64       `json:"totalfailures"`
	ConsecutiveSuccesses uint32       `json:"consecutivesuccesses"`
	ConsecutiveFailures  uint32       `json:"consecutivefailures"`
}

// NewBreakerFunc 函数创建一个路由熔断器处理函数。
func NewBreakerFunc(router eudore.Router) eudore.HandlerFunc {
	return NewBreaker().NewBreakerFunc(router)
}

// NewBreaker 函数创建一个熔断器
//
// 注意：breaker在集群模式下只能操作一个server。
func NewBreaker() *Breaker {
	return &Breaker{
		Mapping:                 make(map[int]string),
		Routes:                  make(map[string]*breakRoute),
		MaxConsecutiveSuccesses: 10,
		MaxConsecutiveFailures:  10,
		OpenWait:                10 * time.Second,
		NewHalfOpen:             NewHalfOpenTicker(400 * time.Millisecond),
	}
}

// NewHalfOpenTicker 函数创建一个方法再半开状态下，周期允许通过一个请求。
func NewHalfOpenTicker(t time.Duration) func(string) func() bool {
	return func(string) func() bool {
		last := time.Now()
		return func() bool {
			now := time.Now()
			if now.Before(last.Add(t)) {
				last = now
				return true
			}
			return false
		}
	}
}

// NewBreakerFunc 方法定义熔断器处理eudore请求上下文函数。
func (b *Breaker) NewBreakerFunc(router eudore.Router) eudore.HandlerFunc {
	if router != nil {
		router.GetFunc("/breaker/ui", HandlerAdmin)
		router.GetFunc("/breaker/data", b.data)
		router.GetFunc("/breaker/:id", b.getRoute)
		router.PutFunc("/breaker/:id/state/:state", b.putRouteState)
	}
	return func(ctx eudore.Context) {
		name := ctx.GetParam("route")
		b.RLock()
		route, ok := b.Routes[name]
		b.RUnlock()
		if !ok {
			b.Lock()
			route = &breakRoute{
				breaker:    b,
				ID:         b.Index,
				Name:       name,
				LastTime:   time.Now(),
				OnHalfOpen: b.NewHalfOpen(name),
			}
			b.Mapping[b.Index] = name
			b.Routes[name] = route
			b.Index++
			b.Unlock()
		}

		route.Handle(ctx)
	}
}

func (b *Breaker) data(ctx eudore.Context) {
	b.RLock()
	ctx.Render(b.Routes)
	b.RUnlock()
}

func (b *Breaker) getRoute(ctx eudore.Context) {
	id := eudore.GetStringInt(ctx.GetParam("id"), -1)
	if id < 0 || id >= b.Index {
		ctx.Fatal("id is invalid")
		return
	}
	b.RLock()
	route := b.Routes[b.Mapping[id]]
	b.RUnlock()
	ctx.Render(route)
}

func (b *Breaker) putRouteState(ctx eudore.Context) {
	id := eudore.GetStringInt(ctx.GetParam("id"), -1)
	state := eudore.GetStringInt(ctx.GetParam("state"), -1)
	if id < 0 || id >= b.Index {
		ctx.Fatal("id is invalid")
		return
	}
	if state < -1 || state > 2 {
		ctx.Fatal("state is invalid")
		return
	}
	b.RLock()
	route := b.Routes[b.Mapping[id]]
	b.RUnlock()
	ctx.Infof("Breaker admin set route %s change state from %s to %s", route.Name, route.BreakerState, BreakerState(state))
	route.BreakerState = BreakerState(state)
	route.ConsecutiveSuccesses = 0
	route.ConsecutiveFailures = 0
	route.RetryClose()
}

// Handle 方法实现路由条目处理熔断。
func (c *breakRoute) Handle(ctx eudore.Context) {
	c.Lock()
	isdeny := c.BreakerState == BreakerStatueOpen || (c.BreakerState == BreakerStatueHalfOpen && c.OnHalfOpen())
	c.Unlock()
	if isdeny {
		ctx.WriteHeader(503)
		ctx.Fatal("Breaker deny request: " + c.Name)
		ctx.End()
		return
	}
	ctx.Next()
	c.Lock()
	if ctx.Response().Status() < 500 {
		c.TotalSuccesses++
		c.ConsecutiveSuccesses++
		c.ConsecutiveFailures = 0
		if c.BreakerState != BreakerStatueClosed && c.ConsecutiveSuccesses > c.breaker.MaxConsecutiveSuccesses {
			ctx.Infof("Breaker route %s change state from %s to %s", c.Name, c.BreakerState, c.BreakerState-1)
			c.ConsecutiveSuccesses = 0
			c.BreakerState--
			c.LastTime = time.Now()
		}
	} else {
		c.TotalFailures++
		c.ConsecutiveFailures++
		c.ConsecutiveSuccesses = 0
		if c.BreakerState != BreakerStatueOpen && c.ConsecutiveFailures > c.breaker.MaxConsecutiveFailures {
			ctx.Infof("Breaker route %s change state from %s to %s", c.Name, c.BreakerState, c.BreakerState+1)
			c.ConsecutiveFailures = 0
			c.BreakerState++
			c.LastTime = time.Now()
			c.RetryClose()
		}
	}
	c.Unlock()
}

func (c *breakRoute) RetryClose() {
	if c.BreakerState == BreakerStatueOpen {
		go func() {
			time.Sleep(c.breaker.OpenWait)
			c.Lock()
			if c.BreakerState == BreakerStatueOpen {
				// app.Infof("Breaker route %s change state from %s to %s", c.Name, BreakerStatueOpen, BreakerStatueHalfOpen)
				c.BreakerState--
			}
			c.Unlock()
		}()
	}
}

// String 方法实现string接口
func (state BreakerState) String() string {
	return BreakerStatues[state]
}

// MarshalText 方法实现encoding.TextMarshaler接口。
func (state BreakerState) MarshalText() (text []byte, err error) {
	text = []byte(state.String())
	return
}
